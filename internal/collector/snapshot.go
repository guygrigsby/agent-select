package collector

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/guygrigsby/agent-routing/routing"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// InputRef is one compile input triple: the contributing observation, its
// source and the hex config hash of the authorizing source policy.
type InputRef struct {
	ObservationID string
	SourceID      string
	PolicyVersion string
}

// CompileSnapshotParams mirrors the compile_snapshot port signature. Both
// instants are caller-supplied UTC, millisecond-aligned.
type CompileSnapshotParams struct {
	Inputs                  []InputRef
	Distribution            string
	ProjectionPolicyVersion string
	CreatedAt               time.Time
	KnowledgeCutoff         time.Time
}

// CompileSnapshotResult is the compile response: artifact hash (32 bytes),
// artifact locator and coverage digest.
type CompileSnapshotResult struct {
	ArtifactHash    []byte
	ArtifactLocator string
	CoverageDigest  []byte
}

type governingPolicy struct {
	ConfigHash     []byte
	Research       string
	Extraction     string
	Redistribution string
	DecidedAt      time.Time
}

// msAligned reports whether t carries no sub-millisecond precision.
func msAligned(t time.Time) bool { return t.Nanosecond()%1e6 == 0 }

// CompileSnapshot implements compile_snapshot: deterministic, replayable
// snapshot compilation from rights-admissible inputs.
func (p *Ports) CompileSnapshot(ctx context.Context, params CompileSnapshotParams) (CompileSnapshotResult, error) {
	if params.Distribution != "local" && params.Distribution != "distributable" {
		return CompileSnapshotResult{}, errf(CodeInvalidRequest, "distribution must be local or distributable")
	}
	if params.ProjectionPolicyVersion == "" {
		return CompileSnapshotResult{}, errf(CodeInvalidRequest, "projection policy version is required")
	}
	if params.KnowledgeCutoff.IsZero() || !msAligned(params.CreatedAt) || !msAligned(params.KnowledgeCutoff) {
		return CompileSnapshotResult{}, errf(CodeInvalidRequest, "created_at and knowledge_cutoff must be millisecond-aligned UTC instants")
	}
	if params.KnowledgeCutoff.After(params.CreatedAt) {
		return CompileSnapshotResult{}, errf(CodeInvalidRequest, "knowledge_cutoff cannot exceed created_at")
	}
	seenObs := make(map[string]bool, len(params.Inputs))
	for _, in := range params.Inputs {
		if in.ObservationID == "" || in.SourceID == "" || in.PolicyVersion == "" {
			return CompileSnapshotResult{}, errf(CodeInvalidRequest, "input triples require observation, source and policy version")
		}
		if seenObs[in.ObservationID] {
			return CompileSnapshotResult{}, errf(CodeInvalidRequest, "duplicate input observation %s", in.ObservationID)
		}
		seenObs[in.ObservationID] = true
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return CompileSnapshotResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)

	// The applied projection policy must exist.
	var policyContent []byte
	if err := tx.QueryRow(ctx,
		`SELECT content_bytes FROM projection_policies WHERE version = $1`,
		params.ProjectionPolicyVersion).Scan(&policyContent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CompileSnapshotResult{}, errf(CodeNotFound, "projection policy %s is not registered", params.ProjectionPolicyVersion)
		}
		return CompileSnapshotResult{}, storageErr(err)
	}
	var policy projectionPolicyDoc
	if err := strictDecode(policyContent, &policy); err != nil {
		return CompileSnapshotResult{}, err
	}

	// Rights gate per distinct source, evaluated at created_at.
	sources := make([]string, 0, len(params.Inputs))
	perSource := map[string][]InputRef{}
	for _, in := range params.Inputs {
		if _, ok := perSource[in.SourceID]; !ok {
			sources = append(sources, in.SourceID)
		}
		perSource[in.SourceID] = append(perSource[in.SourceID], in)
	}
	sort.Strings(sources)
	governing := map[string]*governingPolicy{}
	for _, src := range sources {
		gp, err := currentGoverningPolicy(ctx, tx, src, params.CreatedAt)
		if err != nil {
			return CompileSnapshotResult{}, err
		}
		if gp == nil {
			continue // missing, expired or same-time conflict: excluded
		}
		if gp.Research != "allowed" || gp.Extraction != "allowed" {
			continue
		}
		if params.Distribution == "distributable" && gp.Redistribution != "allowed" {
			continue
		}
		governing[src] = gp
	}

	// Contributing inputs: source admitted and policy version pinned to the
	// governing config hash.
	contributing := make([]InputRef, 0, len(params.Inputs))
	for _, in := range params.Inputs {
		gp, ok := governing[in.SourceID]
		if !ok {
			continue
		}
		if in.PolicyVersion != hexHash(gp.ConfigHash) {
			continue
		}
		contributing = append(contributing, in)
	}

	// Every input, admitted or not, must reference an existing observation
	// of its named source: a malformed reference is invalid_request, never a
	// silent exclusion.
	allObs := make([]string, len(params.Inputs))
	for i, in := range params.Inputs {
		allObs[i] = in.ObservationID
	}
	obsSource, err := observationSources(ctx, tx, allObs)
	if err != nil {
		return CompileSnapshotResult{}, err
	}
	for _, in := range params.Inputs {
		src, ok := obsSource[in.ObservationID]
		if !ok || src != in.SourceID {
			return CompileSnapshotResult{}, errf(CodeInvalidRequest, "input observation %s is not an observation of source %s", in.ObservationID, in.SourceID)
		}
	}
	obsIDs := make([]string, len(contributing))
	for i, in := range contributing {
		obsIDs[i] = in.ObservationID
	}

	// Ordered input manifest hash over the caller-supplied order.
	manifest := make([]map[string]string, len(contributing))
	for i, in := range contributing {
		manifest[i] = map[string]string{
			"observation_id": in.ObservationID,
			"source_id":      in.SourceID,
			"policy_version": in.PolicyVersion,
		}
	}
	manifestHash := hashDoc(manifest)

	// Idempotent replay by the complete parameter set.
	var existingHash []byte
	var existingCoverage []byte
	if err := tx.QueryRow(ctx, `
		SELECT artifact_hash, coverage_digest FROM snapshots
		WHERE input_manifest_hash = $1 AND distribution = $2
		  AND projection_policy_version = $3 AND created_at = $4 AND knowledge_cutoff = $5`,
		manifestHash, params.Distribution, params.ProjectionPolicyVersion,
		params.CreatedAt, params.KnowledgeCutoff).Scan(&existingHash, &existingCoverage); err == nil {
		return CompileSnapshotResult{
			ArtifactHash:    existingHash,
			ArtifactLocator: artifactLocator(existingHash),
			CoverageDigest:  existingCoverage,
		}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CompileSnapshotResult{}, storageErr(err)
	}

	offers, err := projectOffers(ctx, tx, contributing, params.KnowledgeCutoff, policy)
	if err != nil {
		return CompileSnapshotResult{}, err
	}
	coverageDigest, err := coverageDigestFor(ctx, tx, obsIDs)
	if err != nil {
		return CompileSnapshotResult{}, err
	}

	// Artifact inputs follow the routing sort (observation/source/policy).
	routingInputs := make([]routing.InputData, len(contributing))
	for i, in := range contributing {
		routingInputs[i] = routing.InputData{
			ObservationID: in.ObservationID,
			SourceID:      in.SourceID,
			PolicyVersion: in.PolicyVersion,
		}
	}
	sort.Slice(routingInputs, func(i, j int) bool {
		a, b := routingInputs[i], routingInputs[j]
		if a.ObservationID != b.ObservationID {
			return a.ObservationID < b.ObservationID
		}
		if a.SourceID != b.SourceID {
			return a.SourceID < b.SourceID
		}
		return a.PolicyVersion < b.PolicyVersion
	})
	data := routing.SnapshotData{
		Schema:          1,
		CreatedAt:       params.CreatedAt,
		KnowledgeCutoff: params.KnowledgeCutoff,
		Policy:          "educated-v1",
		Distribution:    params.Distribution,
		Inputs:          routingInputs,
		Offers:          offers,
	}
	snap, err := routing.NewSnapshot(data)
	if err != nil {
		return CompileSnapshotResult{}, errf(CodeInvalidEvidence, "snapshot projection rejected by routing constructors: %v", err)
	}
	artifactBytes, err := routing.EncodeSnapshot(snap)
	if err != nil {
		return CompileSnapshotResult{}, errf(CodeInvalidEvidence, "artifact encoding failed: %v", err)
	}
	artifactHash := canonicalHash(artifactBytes)

	// The insert runs inside a savepoint: on a 23505 the savepoint rollback
	// keeps the transaction alive so the parameter-tuple re-select can
	// actually run (Postgres aborts the whole transaction on a bare error,\t// which would make the re-select below dead code).
	sp, err := tx.Begin(ctx)
	if err != nil {
		return CompileSnapshotResult{}, storageErr(err)
	}
	if _, err := sp.Exec(ctx, `
		INSERT INTO snapshots (artifact_hash,schema_version,policy,distribution,
		  projection_policy_version,created_at,knowledge_cutoff,artifact_bytes,
		  input_manifest_hash,coverage_digest)
		VALUES ($1,1,'educated-v1',$2,$3,$4,$5,$6,$7,$8)`,
		artifactHash, params.Distribution, params.ProjectionPolicyVersion,
		params.CreatedAt, params.KnowledgeCutoff, artifactBytes,
		manifestHash, coverageDigest); err != nil {
		_ = sp.Rollback(ctx)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Artifact hash collision: identical bytes already staged. A
			// concurrent identical compile replays the stored row; a different
			// parameter set producing identical bytes is a permanent condition.
			var existHash, existCoverage []byte
			if err2 := tx.QueryRow(ctx, `
				SELECT artifact_hash, coverage_digest FROM snapshots
				WHERE input_manifest_hash = $1 AND distribution = $2
				  AND projection_policy_version = $3 AND created_at = $4
				  AND knowledge_cutoff = $5`,
				manifestHash, params.Distribution, params.ProjectionPolicyVersion,
				params.CreatedAt, params.KnowledgeCutoff).Scan(&existHash, &existCoverage); err2 == nil {
				return CompileSnapshotResult{
					ArtifactHash:    existHash,
					ArtifactLocator: artifactLocator(existHash),
					CoverageDigest:  existCoverage,
				}, nil
			}
			return CompileSnapshotResult{}, errf(CodeConflict,
				"a different parameter set produced identical artifact bytes (hash %s)", hexHash(artifactHash))
		}
		return CompileSnapshotResult{}, storageErr(err)
	}
	if err := sp.Commit(ctx); err != nil { // releases the savepoint
		return CompileSnapshotResult{}, storageErr(err)
	}
	for ordinal, in := range routingInputs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO snapshot_inputs (artifact_hash,ordinal,observation_id)
			VALUES ($1,$2,$3)`, artifactHash, ordinal, in.ObservationID); err != nil {
			return CompileSnapshotResult{}, storageErr(err)
		}
	}
	// Rights audits cover contributing sources only: a source whose inputs
	// were all excluded (pinned-version mismatch) gets no row and no inputs.
	auditSources := map[string]bool{}
	for _, in := range contributing {
		auditSources[in.SourceID] = true
	}
	for src := range auditSources {
		gp := governing[src]
		if _, err := tx.Exec(ctx, `
			INSERT INTO snapshot_rights_audits (artifact_hash,source_id,config_hash,
			  research,extraction,redistribution,decision,decided_at)
			VALUES ($1,$2,$3,$4,$5,$6,'allowed',$7)`,
			artifactHash, src, gp.ConfigHash, gp.Research, gp.Extraction,
			gp.Redistribution, gp.DecidedAt); err != nil {
			return CompileSnapshotResult{}, storageErr(err)
		}
	}
	if err := emitEvent(ctx, tx, hexHash(artifactHash), "SnapshotCompiled", 1,
		params.CreatedAt, map[string]any{
			"artifact_hash":       hexHash(artifactHash),
			"schema_version":      1,
			"policy":              "educated-v1",
			"distribution":        params.Distribution,
			"input_manifest_hash": hexHash(manifestHash),
			"coverage_digest":     hexHash(coverageDigest),
		}); err != nil {
		return CompileSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CompileSnapshotResult{}, storageErr(err)
	}
	return CompileSnapshotResult{
		ArtifactHash:    artifactHash,
		ArtifactLocator: artifactLocator(artifactHash),
		CoverageDigest:  coverageDigest,
	}, nil
}

// artifactLocator is the pinned staging and publication locator grammar.
func artifactLocator(hash []byte) string { return "snapshot://" + hexHash(hash) }

func hashDoc(v any) []byte {
	b, err := canonicalize(v)
	if err != nil {
		panic("canonicalize of in-memory values cannot fail: " + err.Error())
	}
	return canonicalHash(b)
}

// currentGoverningPolicy resolves the governing source policy at t: the
// greatest effective_at row of the current non-expired set, or nil when the
// set is empty or same-time conflicting.
func currentGoverningPolicy(ctx context.Context, tx pgx.Tx, sourceID string, t time.Time) (*governingPolicy, error) {
	rows, err := tx.Query(ctx, `
		SELECT config_hash, research, extraction, redistribution, effective_at
		FROM source_policies
		WHERE source_id = $1 AND effective_at <= $2 AND expires_at > $2
		ORDER BY effective_at DESC`, sourceID, t)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var out []*governingPolicy
	for rows.Next() {
		gp := &governingPolicy{}
		if err := rows.Scan(&gp.ConfigHash, &gp.Research, &gp.Extraction,
			&gp.Redistribution, &gp.DecidedAt); err != nil {
			return nil, storageErr(err)
		}
		out = append(out, gp)
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}
	if len(out) == 0 {
		return nil, nil
	}
	if len(out) > 1 && !out[0].DecidedAt.After(out[1].DecidedAt) {
		return nil, nil // same-time conflict
	}
	return out[0], nil
}

func observationSources(ctx context.Context, tx pgx.Tx, ids []string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := tx.Query(ctx, `
		SELECT id, source_id FROM observations WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, src string
		if err := rows.Scan(&id, &src); err != nil {
			return nil, storageErr(err)
		}
		out[id] = src
	}
	return out, rows.Err()
}

// coverageDigestFor implements the pinned coverage digest recipe.
func coverageDigestFor(ctx context.Context, tx pgx.Tx, obsIDs []string) ([]byte, error) {
	if len(obsIDs) == 0 {
		return hashDoc([]any{}), nil
	}
	rows, err := tx.Query(ctx, `
		SELECT subject_upstream_id, channel, outcome, observation_id
		FROM coverage_findings WHERE observation_id = ANY($1)`, obsIDs)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	type cov struct{ subject, channel, outcome, obs string }
	var found []cov
	for rows.Next() {
		var c cov
		if err := rows.Scan(&c.subject, &c.channel, &c.outcome, &c.obs); err != nil {
			return nil, storageErr(err)
		}
		found = append(found, c)
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}
	sort.Slice(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.subject != b.subject {
			return a.subject < b.subject
		}
		if a.channel != b.channel {
			return a.channel < b.channel
		}
		return a.obs < b.obs
	})
	doc := make([]map[string]string, len(found))
	for i, c := range found {
		doc[i] = map[string]string{
			"subject_upstream_id": c.subject,
			"channel":             c.channel,
			"outcome":             c.outcome,
		}
	}
	return hashDoc(doc), nil
}

var routingCapabilityVocab = map[string]bool{
	"text_input": true, "text_output": true, "image_input": true,
	"image_output": true, "tool_use": true, "structured_output": true,
	"image_edit": true, "image_reference": true, "image_mask": true,
	"image_alpha": true,
}

var routingPriceComponents = map[string]bool{
	"input_tokens": true, "output_tokens": true, "input_images": true,
	"output_images": true, "references": true, "searches": true,
}

// currentRevisionFor resolves the single current binding for an offer at the
// knowledge cutoff; zero or multiple current bindings yield "".
func currentRevisionFor(ctx context.Context, tx pgx.Tx, offerID string, cutoff time.Time) (string, error) {
	rows, err := tx.Query(ctx, `
		SELECT mr.revision_name
		FROM identity_bindings b
		JOIN model_revisions mr ON mr.id = b.revision_id
		WHERE b.offer_id = $1 AND b.effective_at <= $2
		  AND NOT EXISTS (
		    SELECT 1 FROM identity_binding_ends e
		    WHERE e.binding_id = b.id AND e.effective_at <= $2)
		  AND NOT EXISTS (
		    SELECT 1 FROM binding_supersessions s
		    WHERE s.old_binding_id = b.id AND s.effective_at <= $2)`, offerID, cutoff)
	if err != nil {
		return "", storageErr(err)
	}
	defer rows.Close()
	var rev string
	n := 0
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return "", storageErr(err)
		}
		rev, n = r, n+1
	}
	if err := rows.Err(); err != nil {
		return "", storageErr(err)
	}
	if n != 1 {
		return "", nil // zero or ambiguous: identity fails closed
	}
	return rev, nil
}

type offerProjectionRow struct {
	ID         string
	Provider   string
	UpstreamID string
	ListingObs string
}

// projectOffers implements the pinned projection grammar for the offer set,
// capabilities, limits, prices, revision association and signals.
func projectOffers(ctx context.Context, tx pgx.Tx, contributing []InputRef,
	cutoff time.Time, policy projectionPolicyDoc) ([]routing.OfferData, error) {
	obsIDs := make([]string, len(contributing))
	for i, in := range contributing {
		obsIDs[i] = in.ObservationID
	}
	if len(obsIDs) == 0 {
		return nil, nil
	}

	// Offer set: catalog observation is contributing, latest listing listed.
	rows, err := tx.Query(ctx, `
		SELECT o.id, o.provider, o.upstream_id FROM offers o
		WHERE o.catalog_observation_id = ANY($1)
		  AND (
		    SELECT f.action FROM offer_listing_facts f
		    WHERE f.offer_id = o.id
		    ORDER BY f.observed_at DESC, f.action DESC
		    LIMIT 1) = 'listed'`, obsIDs)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var offerRows []offerProjectionRow
	for rows.Next() {
		var r offerProjectionRow
		if err := rows.Scan(&r.ID, &r.Provider, &r.UpstreamID); err != nil {
			return nil, storageErr(err)
		}
		offerRows = append(offerRows, r)
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}

	offers := make([]routing.OfferData, 0, len(offerRows))
	for _, r := range offerRows {
		od, err := projectOneOffer(ctx, tx, r, obsIDs, cutoff, policy)
		if err != nil {
			return nil, err
		}
		offers = append(offers, od)
	}
	sort.Slice(offers, func(i, j int) bool { return offers[i].ID < offers[j].ID })
	return offers, nil
}

func projectOneOffer(ctx context.Context, tx pgx.Tx, r offerProjectionRow,
	obsIDs []string, cutoff time.Time, policy projectionPolicyDoc) (routing.OfferData, error) {
	od := routing.OfferData{
		ID:         r.Provider + "/" + r.UpstreamID,
		Provider:   r.Provider,
		UpstreamID: r.UpstreamID,
	}
	revision, err := currentRevisionFor(ctx, tx, r.ID, cutoff)
	if err != nil {
		return od, err
	}
	od.Revision = revision

	// Capabilities.
	capRows, err := tx.Query(ctx, `
		SELECT DISTINCT ca.capability FROM capability_assertions ca
		JOIN assertions a ON a.id = ca.assertion_id
		WHERE a.observation_id = ANY($1) AND a.subject_upstream_id = $2
		  AND ca.support = 'supported'`, obsIDs, r.UpstreamID)
	if err != nil {
		return od, storageErr(err)
	}
	defer capRows.Close()
	for capRows.Next() {
		var cap string
		if err := capRows.Scan(&cap); err != nil {
			return od, storageErr(err)
		}
		if routingCapabilityVocab[cap] {
			od.Capabilities = append(od.Capabilities, cap)
		}
	}
	if err := capRows.Err(); err != nil {
		return od, storageErr(err)
	}
	sort.Strings(od.Capabilities)

	limits, err := projectLimits(ctx, tx, obsIDs, r.UpstreamID)
	if err != nil {
		return od, err
	}
	od.Limits = limits

	prices, err := projectPrices(ctx, tx, r.ID)
	if err != nil {
		return od, err
	}
	od.Prices = prices

	if revision != "" {
		signals, err := projectSignals(ctx, tx, r.ID, revision, cutoff, policy)
		if err != nil {
			return od, err
		}
		od.Signals = signals
	}
	return od, nil
}

// limitResource maps a limit_assertions row onto a routing resource under
// the pinned grammar; ok=false means not projected.
func limitResource(name, scope, role string) (string, bool) {
	switch name {
	case "context_tokens":
		return "context_tokens", true
	case "input_tokens":
		return "input_tokens", true
	case "output_tokens":
		return "output_tokens", true
	case "image_count":
		if scope == "input" && role == "input_image" {
			return "input_images", true
		}
		if scope == "output" && role == "output_image" {
			return "output_images", true
		}
		return "", false
	case "reference_image_count":
		return "references", true
	case "image_width", "image_height", "image_bytes":
		return name, true
	default:
		return "", false // mask_count, image_pixels: no routing resource
	}
}

func projectLimits(ctx context.Context, tx pgx.Tx, obsIDs []string, upstreamID string) ([]routing.LimitData, error) {
	rows, err := tx.Query(ctx, `
		SELECT la.limit_name, la.scope, la.role, la.bound, la.value
		FROM limit_assertions la
		JOIN assertions a ON a.id = la.assertion_id
		WHERE a.observation_id = ANY($1) AND a.subject_upstream_id = $2`, obsIDs, upstreamID)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	type bound struct {
		min, max       int64
		hasMin, hasMax bool
		exacts         []int64
		conflict       bool
	}
	resources := map[string]*bound{}
	for rows.Next() {
		var name, scope, role, b string
		var value int64
		if err := rows.Scan(&name, &scope, &role, &b, &value); err != nil {
			return nil, storageErr(err)
		}
		res, ok := limitResource(name, scope, role)
		if !ok {
			continue
		}
		bd := resources[res]
		if bd == nil {
			bd = &bound{}
			resources[res] = bd
		}
		switch b {
		case "maximum":
			if !bd.hasMax || value < bd.max {
				bd.max, bd.hasMax = value, true
			}
		case "minimum":
			if !bd.hasMin || value > bd.min {
				bd.min, bd.hasMin = value, true
			}
		case "exact":
			for _, e := range bd.exacts {
				if e != value {
					bd.conflict = true
				}
			}
			bd.exacts = append(bd.exacts, value)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}
	var out []routing.LimitData
	for res, bd := range resources {
		if bd.conflict {
			continue // conflicting exacts: unknown
		}
		// An exact bound contributes its value to both ends of the interval.
		upper, hasUpper := bd.max, bd.hasMax
		for _, e := range bd.exacts {
			if !hasUpper || e < upper {
				upper, hasUpper = e, true
			}
		}
		if !hasUpper {
			continue
		}
		lower, hasLower := bd.min, bd.hasMin
		for _, e := range bd.exacts {
			if !hasLower || e > lower {
				lower, hasLower = e, true
			}
		}
		if hasLower && lower > upper {
			continue // empty intersection: unknown, never a fabricated one
		}
		if upper <= 0 {
			continue // zero cannot be expressed; the DTO zero means absent
		}
		out = append(out, routing.LimitData{Resource: res, Maximum: upper})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Resource < out[j].Resource })
	return out, nil
}

func projectPrices(ctx context.Context, tx pgx.Tx, offerID string) ([]routing.PriceData, error) {
	rows, err := tx.Query(ctx, `
		SELECT pa.category, pa.amount::text, pa.quote_knowledge
		FROM price_assertions pa
		WHERE pa.offer_id = $1 AND pa.currency = 'USD'
		  AND pa.unit = 'million_tokens'
		  AND pa.quote_knowledge IN ('exact','proven_upper_bound')`, offerID)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	type quote struct {
		best         string
		exactLexemes map[string]bool
	}
	qs := map[string]*quote{}
	for rows.Next() {
		var category, amount, kind string
		if err := rows.Scan(&category, &amount, &kind); err != nil {
			return nil, storageErr(err)
		}
		if !routingPriceComponents[category] {
			continue
		}
		lex := trimDecimalZeros(amount)
		q := qs[category]
		if q == nil {
			q = &quote{exactLexemes: map[string]bool{}}
			qs[category] = q
		}
		if kind == "exact" {
			q.exactLexemes[lex] = true
		}
		if q.best == "" || decimalLess(q.best, lex) {
			q.best = lex
		}
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}
	components := make([]string, 0, len(qs))
	for c := range qs {
		components = append(components, c)
	}
	sort.Strings(components)
	out := make([]routing.PriceData, 0, len(components))
	for _, c := range components {
		q := qs[c]
		if len(q.exactLexemes) > 1 {
			continue // exact quotes disagree: unknown
		}
		out = append(out, routing.PriceData{Component: c, USDPerMillion: q.best})
	}
	return out, nil
}

// trimDecimalZeros removes trailing fractional zeros from an exact decimal
// lexeme ("3.50" -> "3.5"), leaving integral lexemes untouched ("100" stays
// "100"; "0" stays "0").
func trimDecimalZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	t := strings.TrimRight(s, "0")
	t = strings.TrimRight(t, ".")
	if t == "" || t == "-" {
		return s
	}
	return t
}

// decimalLess compares nonnegative decimal lexemes without float math.
func decimalLess(a, b string) bool {
	ai, af := splitDecimal(a)
	bi, bf := splitDecimal(b)
	ai, bi = strings.TrimLeft(ai, "0"), strings.TrimLeft(bi, "0")
	if len(ai) != len(bi) {
		return len(ai) < len(bi)
	}
	if ai != bi {
		return ai < bi
	}
	n := len(af)
	if len(bf) > n {
		n = len(bf)
	}
	for i := range n {
		var da, db byte = '0', '0'
		if i < len(af) {
			da = af[i]
		}
		if i < len(bf) {
			db = bf[i]
		}
		if da != db {
			return da < db
		}
	}
	return false
}

func splitDecimal(s string) (string, string) {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func projectSignals(ctx context.Context, tx pgx.Tx, offerID, revision string,
	cutoff time.Time, policy projectionPolicyDoc) ([]routing.SignalData, error) {
	rows, err := tx.Query(ctx, `
		SELECT m.assertion_id, m.benchmark_label, m.raw_value::text, m.unit, a.evidence_class
		FROM benchmark_measurements m
		JOIN assertions a ON a.id = m.assertion_id
		JOIN measurement_revision_facts f ON f.assertion_id = m.assertion_id
		JOIN identity_bindings b ON b.revision_id = f.revision_id
		WHERE b.offer_id = $1 AND b.effective_at <= $2
		  AND NOT EXISTS (
		    SELECT 1 FROM identity_binding_ends e
		    WHERE e.binding_id = b.id AND e.effective_at <= $2)
		  AND NOT EXISTS (
		    SELECT 1 FROM binding_supersessions s
		    WHERE s.old_binding_id = b.id AND s.effective_at <= $2)
		GROUP BY m.assertion_id, m.benchmark_label, m.raw_value, m.unit, a.evidence_class`, offerID, cutoff)
	if err != nil {
		return nil, storageErr(err)
	}
	defer rows.Close()
	var out []routing.SignalData
	for rows.Next() {
		var assertionID, label, raw, unit, evidenceClass string
		if err := rows.Scan(&assertionID, &label, &raw, &unit, &evidenceClass); err != nil {
			return nil, storageErr(err)
		}
		task, ok := policy.TaskMap[label]
		if !ok {
			continue
		}
		weight, ok := policy.Weights[task]
		if !ok {
			continue
		}
		value, ok := signalValue(raw, unit)
		if !ok {
			continue
		}
		basis := "published"
		if b, ok := policy.BasisRules[evidenceClass]; ok {
			basis = b
		}
		out = append(out, routing.SignalData{
			ID:         assertionID,
			Task:       task,
			Value:      value,
			Weight:     int(weight),
			Basis:      basis,
			EvidenceID: assertionID,
			Revision:   revision,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, storageErr(err)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// signalValue converts a raw measurement to the 0..10000 signal range;
// percentage values scale to basis points. Half rounds away from zero.
func signalValue(raw, unit string) (int, bool) {
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) || f < 0 {
		return 0, false
	}
	if unit == "%" {
		f *= 100
	}
	if f > 1e15 {
		return 0, false // conversion would overflow the int range
	}
	v := int(f + 0.5)
	if v < 0 || v > 10000 {
		return 0, false // rounded outside the 0..10000 signal range: skipped
	}
	return v, true
}

// ValidateSnapshotParams mirrors validate_snapshot.
type ValidateSnapshotParams struct {
	ArtifactHash     []byte
	ValidatorVersion string
}

// ValidateSnapshotResult carries the outcome and safe findings.
type ValidateSnapshotResult struct {
	Outcome  string
	Findings []ValidationFinding
}

// ValidationFinding is one member of the closed findings array pinned by the
// snapshot_validations contract: `code`, `offer_id` (omitted for
// artifact-global findings) and `detail`.
type ValidationFinding struct {
	Code    string `json:"code"`
	OfferID string `json:"offer_id,omitempty"`
	Detail  string `json:"detail"`
}

// ValidateSnapshot implements validate_snapshot: artifact bytes re-validated
// against the routing format and the pinned rights audit.
func (p *Ports) ValidateSnapshot(ctx context.Context, params ValidateSnapshotParams) (ValidateSnapshotResult, error) {
	if len(params.ArtifactHash) != 32 {
		return ValidateSnapshotResult{}, errf(CodeInvalidRequest, "artifact hash must be 32 bytes")
	}
	if params.ValidatorVersion == "" {
		return ValidateSnapshotResult{}, errf(CodeInvalidRequest, "validator version is required")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)

	// Idempotent by (artifact_hash, validator_version).
	var storedOutcome string
	var storedFindings []byte
	err = tx.QueryRow(ctx, `
		SELECT outcome, findings_bytes FROM snapshot_validations
		WHERE artifact_hash = $1 AND validator_version = $2`,
		params.ArtifactHash, params.ValidatorVersion).Scan(&storedOutcome, &storedFindings)
	if err == nil {
		var findings []ValidationFinding
		if err := strictDecode(storedFindings, &findings); err == nil {
			return ValidateSnapshotResult{Outcome: storedOutcome, Findings: findings}, nil
		}
		return ValidateSnapshotResult{Outcome: storedOutcome}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ValidateSnapshotResult{}, storageErr(err)
	}

	var artifactBytes []byte
	err = tx.QueryRow(ctx, `
		SELECT artifact_bytes FROM snapshots WHERE artifact_hash = $1`,
		params.ArtifactHash).Scan(&artifactBytes)
	if errors.Is(err, pgx.ErrNoRows) {
		return ValidateSnapshotResult{}, errf(CodeNotFound, "no staged snapshot for hash %s", hexHash(params.ArtifactHash))
	}
	if err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}

	findings := []ValidationFinding{}
	accepted := true
	if !bytes.Equal(canonicalHash(artifactBytes), params.ArtifactHash) {
		accepted = false
		findings = append(findings, ValidationFinding{
			Code: "artifact_hash_mismatch", Detail: "artifact hash does not pin the staged bytes"})
	}
	if _, err := routing.DecodeSnapshot(artifactBytes); err != nil {
		accepted = false
		findings = append(findings, ValidationFinding{
			Code: "decode_contract_failed", Detail: "artifact bytes fail the routing decode contract"})
	}
	// Pinned rights audit must cover every contributing input's source.
	var auditCount, inputCount int
	if err := tx.QueryRow(ctx, `
		SELECT
		  (SELECT count(*) FROM snapshot_rights_audits WHERE artifact_hash = $1),
		  (SELECT count(DISTINCT o.source_id) FROM snapshot_inputs si
		   JOIN observations o ON o.id = si.observation_id
		   WHERE si.artifact_hash = $1)`, params.ArtifactHash).Scan(&auditCount, &inputCount); err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}
	if auditCount != inputCount {
		accepted = false
		findings = append(findings, ValidationFinding{
			Code:   "rights_audit_mismatch",
			Detail: fmt.Sprintf("rights audit covers %d of %d contributing sources", auditCount, inputCount)})
	}
	outcome := "rejected"
	if accepted {
		outcome = "accepted"
	}
	findingsBytes, err := canonicalize(findings)
	if err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO snapshot_validations (artifact_hash,validator_version,outcome,findings_bytes,validated_at)
		VALUES ($1,$2,$3,$4,now())`,
		params.ArtifactHash, params.ValidatorVersion, outcome, findingsBytes); err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}
	if err := emitEvent(ctx, tx, hexHash(params.ArtifactHash), "SnapshotValidated", 1,
		time.Now().UTC(), map[string]any{
			"artifact_hash":     hexHash(params.ArtifactHash),
			"validator_version": params.ValidatorVersion,
			"outcome":           outcome,
			"findings_hash":     hexHash(findingsBytes),
		}); err != nil {
		return ValidateSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ValidateSnapshotResult{}, storageErr(err)
	}
	return ValidateSnapshotResult{Outcome: outcome, Findings: findings}, nil
}

// PublishSnapshotResult carries the publication ID and artifact locator.
type PublishSnapshotResult struct {
	PublicationID   string
	ArtifactLocator string
}

// PublishSnapshot implements publish_snapshot: validated artifact, channel
// class from the registry, in-transaction rights re-evaluation, atomic
// decision persistence.
func (p *Ports) PublishSnapshot(ctx context.Context, artifactHash []byte, channel string) (PublishSnapshotResult, error) {
	if len(artifactHash) != 32 {
		return PublishSnapshotResult{}, errf(CodeInvalidRequest, "artifact hash must be 32 bytes")
	}
	if channel == "" {
		return PublishSnapshotResult{}, errf(CodeInvalidRequest, "channel is required")
	}
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}
	defer tx.Rollback(ctx)

	// Idempotent by artifact hash and channel.
	var existingID, existingLoc string
	err = tx.QueryRow(ctx, `
		SELECT id, artifact_locator FROM snapshot_publications
		WHERE artifact_hash = $1 AND channel = $2`, artifactHash, channel).Scan(&existingID, &existingLoc)
	if err == nil {
		return PublishSnapshotResult{PublicationID: existingID, ArtifactLocator: existingLoc}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PublishSnapshotResult{}, storageErr(err)
	}

	var distribution string
	var createdAt, knowledgeCutoff time.Time
	var coverage []byte
	err = tx.QueryRow(ctx, `
		SELECT distribution, created_at, knowledge_cutoff, coverage_digest
		FROM snapshots WHERE artifact_hash = $1`, artifactHash).Scan(&distribution, &createdAt, &knowledgeCutoff, &coverage)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublishSnapshotResult{}, errf(CodeNotFound, "no staged snapshot for hash %s", hexHash(artifactHash))
	}
	if err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}

	// Requires an accepted validation for this hash.
	var validated bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM snapshot_validations
		  WHERE artifact_hash = $1 AND outcome = 'accepted')`, artifactHash).Scan(&validated); err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}
	if !validated {
		return PublishSnapshotResult{}, errf(CodeConflict, "artifact %s has no accepted validation", hexHash(artifactHash))
	}

	// Channel class is read from the registry, never caller-supplied.
	var class string
	var retired bool
	if err := tx.QueryRow(ctx, `
		SELECT c.channel_class,
		       EXISTS (SELECT 1 FROM publication_channel_retirements r WHERE r.channel = c.channel)
		FROM publication_channels c WHERE c.channel = $1`, channel).Scan(&class, &retired); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PublishSnapshotResult{}, errf(CodeNotFound, "channel %s is not registered", channel)
		}
		return PublishSnapshotResult{}, storageErr(err)
	}
	if retired {
		return PublishSnapshotResult{}, errf(CodeConflict, "channel %s is retired", channel)
	}
	if distribution == "local" && class != "local" {
		return PublishSnapshotResult{}, errf(CodeForbidden, "local artifact cannot publish to a %s channel", class)
	}

	// Re-evaluate current policies inside the transaction: a revocation
	// committed before insertion fails the gate. The pinned audit rows carry
	// the contributing sources.
	auditRows, err := tx.Query(ctx, `
		SELECT source_id FROM snapshot_rights_audits WHERE artifact_hash = $1`,
		artifactHash)
	if err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}
	var sources []string
	for auditRows.Next() {
		var src string
		if err := auditRows.Scan(&src); err != nil {
			auditRows.Close()
			return PublishSnapshotResult{}, storageErr(err)
		}
		sources = append(sources, src)
	}
	auditRows.Close()
	if err := auditRows.Err(); err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}
	now := time.Now().UTC()
	for _, src := range sources {
		gp, err := currentGoverningPolicy(ctx, tx, src, now)
		if err != nil {
			return PublishSnapshotResult{}, err
		}
		if gp == nil || gp.Research != "allowed" || gp.Extraction != "allowed" ||
			(distribution == "distributable" && gp.Redistribution != "allowed") {
			return PublishSnapshotResult{}, errf(CodeForbidden,
				"rights revoked for source %s before publication", src)
		}
	}

	pubID := newID("pub")
	loc := artifactLocator(artifactHash)
	if _, err := tx.Exec(ctx, `
		INSERT INTO snapshot_publications (id,artifact_hash,channel,channel_class,
		  rights_decision,artifact_locator,published_at)
		VALUES ($1,$2,$3,$4,'allowed',$5,now())`,
		pubID, artifactHash, channel, class, loc); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return PublishSnapshotResult{}, errf(CodeConflict, "publication already exists for this artifact and channel")
		}
		return PublishSnapshotResult{}, storageErr(err)
	}
	if err := emitEvent(ctx, tx, hexHash(artifactHash), "SnapshotPublished", 1,
		now, map[string]any{
			"publication_id":   pubID,
			"artifact_hash":    hexHash(artifactHash),
			"schema_version":   1,
			"policy":           "educated-v1",
			"distribution":     distribution,
			"channel":          channel,
			"channel_class":    class,
			"coverage_digest":  hexHash(coverage),
			"artifact_locator": loc,
			"knowledge_cutoff": knowledgeCutoff.Unix(),
			"published_at":     now.Unix(),
		}); err != nil {
		return PublishSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PublishSnapshotResult{}, storageErr(err)
	}
	return PublishSnapshotResult{PublicationID: pubID, ArtifactLocator: loc}, nil
}

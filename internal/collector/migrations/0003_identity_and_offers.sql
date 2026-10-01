-- 0003: model revisions, offers, listings, coverage findings, identity policy
-- and binding lifecycle. Authority: model-identity.md#relational-contract,
-- contracts.md#data-layer.

CREATE TABLE model_revisions (
    id TEXT PRIMARY KEY,
    publisher TEXT NOT NULL,
    revision_name TEXT NOT NULL,
    identity_observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    identity_locator TEXT NOT NULL,
    UNIQUE (identity_observation_id, identity_locator)
);

CREATE TABLE offers (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    upstream_id TEXT NOT NULL,
    catalog_observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    advertised_kind TEXT NOT NULL,
    UNIQUE (provider, upstream_id)
);

CREATE TABLE offer_listing_facts (
    offer_id TEXT NOT NULL REFERENCES offers(id) ON DELETE RESTRICT,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    action TEXT NOT NULL CHECK (action IN ('listed','withdrawn')),
    observed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (offer_id, observation_id, action)
);
CREATE INDEX offer_listing_current ON offer_listing_facts (offer_id, observed_at DESC);

CREATE TABLE coverage_findings (
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    subject_upstream_id TEXT NOT NULL,
    channel TEXT NOT NULL CHECK (channel IN
        ('vercel_catalog','modelsdev','provider_api','provider_site','eval_registry','host_local')),
    outcome TEXT NOT NULL CHECK (outcome IN
        ('researched','empty','blocked','identity_ambiguous','unresearched','invalid')),
    detail TEXT NOT NULL,
    PRIMARY KEY (observation_id, subject_upstream_id, channel)
);
CREATE INDEX coverage_by_subject ON coverage_findings (subject_upstream_id, channel);

CREATE TABLE identity_authority_policies (
    version TEXT PRIMARY KEY,
    content_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(content_hash) = 32),
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE identity_authority_rules (
    policy_version TEXT NOT NULL REFERENCES identity_authority_policies(version) ON DELETE RESTRICT,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    publisher TEXT NOT NULL,
    claim_type TEXT NOT NULL CHECK (claim_type IN
        ('revision_binding','binding_end','binding_supersession')),
    PRIMARY KEY (policy_version, source_id, publisher, claim_type)
);

CREATE TABLE identity_freshness_policies (
    version TEXT PRIMARY KEY,
    content_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(content_hash) = 32),
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE identity_freshness_rules (
    policy_version TEXT NOT NULL REFERENCES identity_freshness_policies(version) ON DELETE RESTRICT,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    claim_type TEXT NOT NULL CHECK (claim_type = 'revision_binding'),
    maximum_age_seconds BIGINT NOT NULL CHECK (maximum_age_seconds > 0),
    PRIMARY KEY (policy_version, source_id, claim_type)
);

CREATE TABLE identity_bindings (
    id TEXT PRIMARY KEY,
    offer_id TEXT NOT NULL REFERENCES offers(id) ON DELETE RESTRICT,
    revision_id TEXT NOT NULL REFERENCES model_revisions(id) ON DELETE RESTRICT,
    evidence_assertion_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    authority_policy_version TEXT NOT NULL REFERENCES identity_authority_policies(version) ON DELETE RESTRICT,
    freshness_policy_version TEXT NOT NULL REFERENCES identity_freshness_policies(version) ON DELETE RESTRICT,
    effective_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    UNIQUE (offer_id, revision_id, evidence_assertion_id, authority_policy_version,
            freshness_policy_version, effective_at)
);
CREATE INDEX identity_binding_candidates ON identity_bindings (offer_id, effective_at, recorded_at);
CREATE INDEX identity_binding_provenance ON identity_bindings (revision_id);

CREATE TABLE identity_binding_confirmations (
    binding_id TEXT NOT NULL REFERENCES identity_bindings(id) ON DELETE CASCADE,
    observation_id TEXT NOT NULL REFERENCES observations(id) ON DELETE RESTRICT,
    confirmed_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (binding_id, observation_id)
);
CREATE INDEX identity_confirmation_freshness
    ON identity_binding_confirmations (binding_id, confirmed_at DESC, recorded_at);

CREATE TABLE identity_binding_ends (
    id TEXT PRIMARY KEY,
    binding_id TEXT NOT NULL REFERENCES identity_bindings(id) ON DELETE RESTRICT,
    kind TEXT NOT NULL CHECK (kind IN ('close','retract')),
    evidence_assertion_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    authority_policy_version TEXT NOT NULL REFERENCES identity_authority_policies(version) ON DELETE RESTRICT,
    freshness_policy_version TEXT NOT NULL REFERENCES identity_freshness_policies(version) ON DELETE RESTRICT,
    effective_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    UNIQUE (binding_id, kind, evidence_assertion_id, effective_at)
);
CREATE INDEX identity_end_anti_join ON identity_binding_ends (binding_id, effective_at, recorded_at);

CREATE TABLE binding_supersessions (
    id TEXT PRIMARY KEY,
    old_binding_id TEXT NOT NULL REFERENCES identity_bindings(id) ON DELETE RESTRICT,
    replacement_binding_id TEXT NOT NULL REFERENCES identity_bindings(id) ON DELETE RESTRICT,
    evidence_assertion_id TEXT NOT NULL REFERENCES assertions(id) ON DELETE RESTRICT,
    authority_policy_version TEXT NOT NULL REFERENCES identity_authority_policies(version) ON DELETE RESTRICT,
    freshness_policy_version TEXT NOT NULL REFERENCES identity_freshness_policies(version) ON DELETE RESTRICT,
    effective_at TIMESTAMPTZ NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    CHECK (old_binding_id <> replacement_binding_id),
    UNIQUE (old_binding_id, replacement_binding_id, evidence_assertion_id, effective_at)
);
CREATE INDEX binding_supersession_anti_join
    ON binding_supersessions (old_binding_id, effective_at, recorded_at);
CREATE INDEX binding_supersession_chain ON binding_supersessions (replacement_binding_id);

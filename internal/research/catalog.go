package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Catalog inputs (§7.1).
const (
	vercelCapture      = "data/raw/vercel-models.json"
	modelsDevCapture   = "data/raw/models-dev.json"
	baselineCatalogURL = "https://ai-gateway.vercel.sh/v1/models"
	modelsDevURL       = "https://models.dev/api.json"
)

// jsonNumber aliases encoding/json's Number for downstream users.
type jsonNumber = json.Number

// oobj is a JSON object preserving document key order (byte-identity
// requires iteration order to follow the input bytes, never map order).
type oobj struct {
	keys []string
	vals map[string]interface{}
}

func (o *oobj) get(k string) (interface{}, bool) {
	if o == nil {
		return nil, false
	}
	v, ok := o.vals[k]
	return v, ok
}

func (o *oobj) str(k string) string {
	v, ok := o.get(k)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func (o *oobj) keysInOrder() []string { return o.keys }

// oarr is a JSON array.
type oarr struct {
	items []interface{}
}

// decodeOrdered reads one JSON value preserving document order everywhere.
func decodeOrdered(dec *json.Decoder) (interface{}, error) {
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return decodeOrderedFrom(dec, t)
}

func decodeOrderedFrom(dec *json.Decoder, t json.Token) (interface{}, error) {
	switch tok := t.(type) {
	case json.Delim:
		switch tok {
		case '{':
			o := &oobj{vals: map[string]interface{}{}}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, _ := kt.(string)
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				if _, dup := o.vals[key]; !dup {
					o.keys = append(o.keys, key)
				}
				o.vals[key] = v
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return nil, err
			}
			return o, nil
		case '[':
			a := &oarr{}
			for dec.More() {
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				a.items = append(a.items, v)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return nil, err
			}
			return a, nil
		}
		return nil, fmt.Errorf("unexpected delimiter %v", tok)
	case json.Number:
		return tok, nil
	default:
		return t, nil // string, bool, nil
	}
}

func decodeCaptureOrdered(root, path string) (interface{}, error) {
	full := filepath.Join(root, filepath.FromSlash(path))
	raw, err := os.ReadFile(full)
	if err != nil {
		return nil, iof("read %s: %v", path, err)
	}
	if err := sizeBound(len(raw), path); err != nil {
		return nil, err
	}
	if depthCheck(raw) > 64 {
		return nil, invalidf("%s: JSON nesting depth exceeds 64", path)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return decodeOrdered(dec)
}

// leafInfo describes one numeric leaf: verbatim lexeme + /-rooted path.
type leafInfo struct {
	fieldPath string
	lexeme    string
	isString  bool
}

// walkLeaves walks document order, yielding numeric leaves: JSON numbers
// and strings whose content parses as an exact decimal. Booleans never.
func walkLeaves(v interface{}, path string, out *[]leafInfo) {
	switch x := v.(type) {
	case *oobj:
		for _, k := range x.keysInOrder() {
			walkLeaves(x.vals[k], path+"/"+k, out)
		}
	case *oarr:
		for i, it := range x.items {
			walkLeaves(it, path+"/"+fmt.Sprintf("%d", i), out)
		}
	case json.Number:
		s := x.String()
		if ValidDecimal(s) {
			*out = append(*out, leafInfo{fieldPath: path, lexeme: s})
		}
	case string:
		if ValidDecimal(x) {
			*out = append(*out, leafInfo{fieldPath: path, lexeme: x, isString: true})
		}
	}
}

func priceLeaves(v interface{}) []leafInfo {
	var out []leafInfo
	walkLeaves(v, "", &out)
	return out
}

// CatalogModel is one catalog_models row plus its child rows.
type CatalogModel struct {
	ID          string
	Name        string
	OwnedBy     string
	Type        string
	Description string
	Entry       *oobj // raw entry for downstream walks
}

// Catalog is the flattened projection of both captures.
type Catalog struct {
	Models        []*CatalogModel
	Offers        []*Offer
	PriceFacts    []PriceFact
	ExactJoins    []ExactJoin
	Disagreements []PricingDisagreement
}

// Offer is one provider_offers row (plus canonical label when present).
type Offer struct {
	Provider       string
	UID            string
	Name           string
	CanonicalLabel string
	Canonical      bool
}

// PriceFact is one price_facts row.
type PriceFact struct {
	SourcePath string
	Provider   string
	UpstreamID string
	FieldPath  string
	RawDecimal string
	Unit       string
}

// ExactJoin is one exact_catalog_joins row.
type ExactJoin struct {
	ModelID    string
	Provider   string
	UpstreamID string
}

// PricingDisagreement is one pricing_disagreements row.
type PricingDisagreement struct {
	ModelID             string
	Component           string
	VercelPerMillion    string
	ModelsDevPerMillion string
}

// vercelUnit maps a final pricing path segment to its unit.
func vercelUnit(leaf string) string {
	switch leaf {
	case "input", "output", "input_cache_read", "input_cache_write",
		"audio_input_token_cost", "audio_output_token_cost":
		return "USD_per_token"
	case "image":
		return "USD_per_image"
	default:
		return "source_specific_unresolved"
	}
}

// modelsDevUnit maps a final cost path segment to its unit.
func modelsDevUnit(leaf string) string {
	switch leaf {
	case "input", "output", "cache_read", "cache_write":
		return "USD_per_million_tokens"
	default:
		return "source_specific_unresolved"
	}
}

// LoadCatalog projects both captures per §7.1. inv gates the read paths.
func LoadCatalog(root string, inv *Inventory) (*Catalog, error) {
	for _, p := range []string{vercelCapture, modelsDevCapture} {
		if err := inv.CheckReadPath(p); err != nil {
			return nil, err
		}
	}
	vTop, err := decodeCaptureOrdered(root, vercelCapture)
	if err != nil {
		return nil, err
	}
	vTopObj, ok := vTop.(*oobj)
	if !ok {
		return nil, invalidf("%s: top level is not an object", vercelCapture)
	}
	vDataAny, ok := vTopObj.get("data")
	if !ok {
		return nil, invalidf("%s: missing data array", vercelCapture)
	}
	vData, ok := vDataAny.(*oarr)
	if !ok {
		return nil, invalidf("%s: data is not an array", vercelCapture)
	}
	mTop, err := decodeCaptureOrdered(root, modelsDevCapture)
	if err != nil {
		return nil, err
	}
	mProviders, ok := mTop.(*oobj)
	if !ok {
		return nil, invalidf("%s: top level is not an object", modelsDevCapture)
	}

	c := &Catalog{}
	byID := map[string]*CatalogModel{}

	for _, e := range vData.items {
		x, ok := e.(*oobj)
		if !ok {
			return nil, invalidf("%s: data entry is not an object", vercelCapture)
		}
		cm := &CatalogModel{
			ID:          x.str("id"),
			Name:        x.str("name"),
			OwnedBy:     x.str("owned_by"),
			Type:        x.str("type"),
			Description: x.str("description"),
			Entry:       x,
		}
		if cm.ID == "" {
			return nil, invalidf("%s: catalog entry without id", vercelCapture)
		}
		if _, dup := byID[cm.ID]; dup {
			return nil, invalidf("%s: duplicate catalog id %q", vercelCapture, cm.ID)
		}
		modsAny, ok := x.get("modalities")
		if !ok {
			return nil, invalidf("%s: model %s: modalities object is required by the frozen corpus", vercelCapture, cm.ID)
		}
		mods, ok := modsAny.(*oobj)
		if !ok {
			return nil, invalidf("%s: model %s: modalities is not an object", vercelCapture, cm.ID)
		}
		// children collected during build in document order
		_ = mods
		if prAny, ok := x.get("pricing"); ok {
			if pr, ok := prAny.(*oobj); ok {
				for _, l := range priceLeaves(pr) {
					c.PriceFacts = append(c.PriceFacts, PriceFact{
						SourcePath: vercelCapture, Provider: "vercel", UpstreamID: cm.ID,
						FieldPath: l.fieldPath, RawDecimal: l.lexeme, Unit: vercelUnit(lastSegment(l.fieldPath)),
					})
				}
			}
		}
		byID[cm.ID] = cm
		c.Models = append(c.Models, cm)
	}

	// models.dev providers in document order
	for _, prov := range mProviders.keysInOrder() {
		pAny, _ := mProviders.get(prov)
		p, ok := pAny.(*oobj)
		if !ok {
			continue
		}
		modelsAny, _ := p.get("models")
		models, ok := modelsAny.(*oobj)
		if !ok {
			continue
		}
		for _, uid := range models.keysInOrder() {
			xAny, _ := models.get(uid)
			x, ok := xAny.(*oobj)
			if !ok {
				continue
			}
			name := uid
			if s := x.str("name"); s != "" {
				name = s
			}
			offer := &Offer{Provider: prov, UID: uid, Name: name}
			if s := x.str("canonical_model_id"); s != "" {
				offer.CanonicalLabel = s
				offer.Canonical = true
			}
			c.Offers = append(c.Offers, offer)
			if costAny, ok := x.get("cost"); ok {
				if cost, ok := costAny.(*oobj); ok {
					for _, l := range priceLeaves(cost) {
						c.PriceFacts = append(c.PriceFacts, PriceFact{
							SourcePath: modelsDevCapture, Provider: prov, UpstreamID: uid,
							FieldPath: l.fieldPath, RawDecimal: l.lexeme, Unit: modelsDevUnit(lastSegment(l.fieldPath)),
						})
					}
				}
			}
		}
	}

	// exact joins + disagreements (python iterated the vercel data array in
	// document order; join insert order follows it for byte-identity)
	vercelProviderAny, _ := mProviders.get("vercel")
	vercelProvider, ok := vercelProviderAny.(*oobj)
	if !ok {
		return nil, invalidf("%s: missing vercel provider", modelsDevCapture)
	}
	vercelModelsAny, _ := vercelProvider.get("models")
	vercelModels, ok := vercelModelsAny.(*oobj)
	if !ok {
		return nil, invalidf("%s: vercel provider has no models map", modelsDevCapture)
	}
	for _, e := range vData.items {
		x := e.(*oobj)
		id := x.str("id")
		yAny, ok := vercelModels.get(id)
		if !ok {
			continue
		}
		y, ok := yAny.(*oobj)
		if !ok {
			continue
		}
		c.ExactJoins = append(c.ExactJoins, ExactJoin{ModelID: id, Provider: "vercel", UpstreamID: id})
		var mCost *oobj
		if costAny, ok := y.get("cost"); ok {
			mCost, _ = costAny.(*oobj)
		}
		if mCost == nil {
			continue
		}
		var pr *oobj
		if prAny, ok := x.get("pricing"); ok {
			pr, _ = prAny.(*oobj)
		}
		for _, pair := range []struct{ vercel, dev string }{
			{"input", "input"}, {"output", "output"},
			{"input_cache_read", "cache_read"}, {"input_cache_write", "cache_write"},
		} {
			vl, vOK := numericComponent(pr, pair.vercel)
			ml, mOK := numericComponent(mCost, pair.dev)
			if !vOK || !mOK {
				continue
			}
			product := mulDecimalPreserveScale(vl, "1000000")
			if !decimalEqual(product, ml) {
				c.Disagreements = append(c.Disagreements, PricingDisagreement{
					ModelID: id, Component: pair.vercel,
					VercelPerMillion:    product,
					ModelsDevPerMillion: ml,
				})
			}
		}
	}
	return c, nil
}

// numericComponent returns the exact-decimal lexeme for a direct pricing
// component, treating non-numeric leaves as absent.
func numericComponent(obj *oobj, key string) (string, bool) {
	if obj == nil {
		return "", false
	}
	v, ok := obj.get(key)
	if !ok {
		return "", false
	}
	switch x := v.(type) {
	case json.Number:
		s := x.String()
		return s, ValidDecimal(s)
	case string:
		return x, ValidDecimal(x)
	default:
		return "", false
	}
}

// mulDecimalPreserveScale multiplies an exact decimal lexeme by 10^6 and
// renders the product preserving the lexeme's fraction-digit count (§7.1:
// 0.000000025 × 1000000 renders 0.025000000).
func mulDecimalPreserveScale(lex, by string) string {
	neg := strings.HasPrefix(lex, "-")
	body := strings.TrimPrefix(strings.TrimPrefix(lex, "-"), "+")
	intPart, fracPart := body, ""
	if i := strings.IndexByte(body, '.'); i >= 0 {
		intPart, fracPart = body[:i], body[i+1:]
	}
	if strings.ContainsAny(body, "eE") {
		// exponent form: exact rational product, rendered with the
		// normalized fraction-digit count
		r, ok := new(big.Rat).SetString(lex)
		if !ok {
			return lex
		}
		r.Mul(r, new(big.Rat).SetInt64(1000000))
		scale := len(fracPart) - expShift(lex)
		if scale < 0 {
			scale = 0
		}
		return renderWithScale(r, scale, neg)
	}
	digits := intPart + fracPart
	trimmed := strings.TrimLeft(digits, "0")
	n := new(big.Int)
	if trimmed == "" {
		n.SetInt64(0)
	} else if _, ok := n.SetString(trimmed, 10); !ok {
		return lex
	}
	n.Mul(n, big.NewInt(1000000))
	s := n.String()
	scale := len(fracPart)
	for len(s) <= scale {
		s = "0" + s
	}
	out := s[:len(s)-scale] + "." + s[len(s)-scale:]
	if strings.HasPrefix(out, ".") {
		out = "0" + out
	}
	if neg && strings.TrimLeft(out, "0.") != "" {
		out = "-" + out
	}
	return out
}

// expShift returns the decimal exponent of a lexeme like 2.5e3.
func expShift(lex string) int {
	i := strings.IndexAny(lex, "eE")
	if i < 0 {
		return 0
	}
	exp := new(big.Int)
	if _, ok := exp.SetString(lex[i+1:], 10); !ok {
		return 0
	}
	if !exp.IsInt64() {
		return 0
	}
	return int(exp.Int64())
}

// renderWithScale renders a rational with a fixed number of fraction digits.
func renderWithScale(r *big.Rat, scale int, neg bool) string {
	if scale < 0 {
		scale = 0
	}
	shift := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	num := new(big.Int).Mul(r.Num(), shift)
	q := new(big.Int)
	rem := new(big.Int)
	q.QuoRem(num, r.Denom(), rem)
	if rem.Sign() != 0 {
		// not exactly representable at this scale; fall back to exact string
		return r.RatString()
	}
	s := q.String()
	for len(s) <= scale {
		s = "0" + s
	}
	out := s[:len(s)-scale]
	if scale > 0 {
		out += "." + s[len(s)-scale:]
	}
	if strings.HasPrefix(out, ".") {
		out = "0" + out
	}
	if neg && q.Sign() != 0 {
		out = "-" + out
	}
	return out
}

// decimalEqual compares two exact decimal lexemes.
func decimalEqual(a, b string) bool {
	ra, ok := new(big.Rat).SetString(a)
	if !ok {
		return false
	}
	rb, ok := new(big.Rat).SetString(b)
	if !ok {
		return false
	}
	return ra.Cmp(rb) == 0
}

func lastSegment(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// SortedStrings returns a sorted copy (deterministic batch order).
func SortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

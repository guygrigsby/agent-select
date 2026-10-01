package routing

// Contract-to-code drift guard for docs/specs/educated-routing.md. It parses
// the documented complete DTO field lists and closed enum token lists out of
// the actual document text and compares them, in both directions, against the
// reflected Go struct fields (names, qualified types, snake_case JSON tags)
// and the closed token maps in tokens.go. It never duplicates the documented
// lists in this file; the document itself is the expectation source.

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

const contractDocPath = "../docs/specs/educated-routing.md"

// contractOmitemptyAllowances maps each omitempty exception to the exact
// documented sentence permitting it; the guard fails if the document loses
// a sentence, so no hardcoded exception can ever go unguarded.
var contractOmitemptyAllowances = []struct {
	label, field, sentence string
}{
	{"OfferData", "ConfirmedAt", "Zero `ConfirmedAt` omits `confirmed_at`"},
	{"Decision", "Selected", "nil `Selected` omits `selected`"},
}

// contractDTOs is the single fact source for the documented complete data
// shapes: one row ties each doc label to its Go type and defining file. Error
// is deliberately absent: it is not a wire-defined DTO yet.
var contractDTOs = []struct {
	label string
	file  string
	typ   reflect.Type
}{
	{"SnapshotData", "dto.go", reflect.TypeOf(SnapshotData{})},
	{"InputData", "dto.go", reflect.TypeOf(InputData{})},
	{"OfferData", "dto.go", reflect.TypeOf(OfferData{})},
	{"LimitData", "dto.go", reflect.TypeOf(LimitData{})},
	{"PriceData", "dto.go", reflect.TypeOf(PriceData{})},
	{"SignalData", "dto.go", reflect.TypeOf(SignalData{})},
	{"RuntimeOfferData", "request.go", reflect.TypeOf(RuntimeOfferData{})},
	{"RequestData", "request.go", reflect.TypeOf(RequestData{})},
	{"Estimate", "decision.go", reflect.TypeOf(Estimate{})},
	{"CostEstimate", "decision.go", reflect.TypeOf(CostEstimate{})},
	{"Selection", "decision.go", reflect.TypeOf(Selection{})},
	{"Rejection", "decision.go", reflect.TypeOf(Rejection{})},
	{"Decision", "decision.go", reflect.TypeOf(Decision{})},
}

var contractLabelRe = regexp.MustCompile("`([A-Z][A-Za-z0-9]*)`: ")
var contractFieldRe = regexp.MustCompile("`([A-Z][A-Za-z0-9]*) ([^`]+)`")
var contractTypeRe = regexp.MustCompile(
	`^(int|int64|string|bool|time\.Time|\[\][A-Za-z][A-Za-z0-9]*|\*[A-Z][A-Za-z0-9]*|[A-Z][A-Za-z0-9]*)$`)
var contractTokenRe = regexp.MustCompile("`([a-z][a-z0-9_]*)`")

// contractEnumSpecs map each documented closed vocabulary in the real
// document (prefix ... terminator) to the token map in tokens.go.
var contractEnumSpecs = []struct {
	prefix     string
	terminator string
	name       string
	target     map[string]bool
}{
	{"Tasks: ", ".", "Tasks", taskTokens},
	{"Modes: ", ".", "Modes", modeTokens},
	{"Capabilities: ", ".", "Capabilities", capabilityTokens},
	{"Resources: ", ".", "LimitData resources", resourceTokens},
	{"Components: ", ".", "PriceData components", priceComponentTokens},
	{"Basis: ", ".", "SignalData basis", basisTokens},
	{"ImageOperation: ", ".", "ImageOperation", imageOperationTokens},
	{"Distribution is ", ";", "Distribution", distributionTokens},
}

type contractDocField struct {
	Name string
	Type string
}

func readContractDoc(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(contractDocPath)
	if err != nil {
		t.Fatalf("read %s: %v", contractDocPath, err)
	}
	return string(b)
}

// parseContractDTOFields extracts each documented `TypeName`: field list.
// A label's range ends at the next label or the end of its paragraph,
// whichever comes first, so trailing prose (including the wire-undefined
// error `Code`/`Detail` fields) is never mistaken for DTO fields.
func parseContractDTOFields(doc string) (map[string][]contractDocField, []string) {
	fields := make(map[string][]contractDocField, len(contractDTOs))
	var problems []string
	seen := map[string]bool{}
	for _, m := range contractLabelRe.FindAllStringSubmatchIndex(doc, -1) {
		name := doc[m[2]:m[3]]
		if seen[name] {
			problems = append(problems, fmt.Sprintf(
				"docs/specs/educated-routing.md defines label %q more than once; remedy: keep exactly one complete data shape per DTO", name))
			continue
		}
		seen[name] = true
		start := m[1]
		end := len(doc)
		if p := strings.Index(doc[start:], "\n\n"); p >= 0 {
			end = start + p
		}
		for i := range contractDTOs {
			if contractDTOs[i].label == name {
				continue
			}
			if n := strings.Index(doc[start:], "`"+contractDTOs[i].label+"`: "); n >= 0 && start+n < end {
				end = start + n
			}
		}
		for _, fm := range contractFieldRe.FindAllStringSubmatch(doc[start:end], -1) {
			fname, ftype := fm[1], fm[2]
			if !contractTypeRe.MatchString(ftype) {
				problems = append(problems, fmt.Sprintf(
					"docs/specs/educated-routing.md field %q of %q has unparsable Go type %q; remedy: write the field as `Name routingType` in the complete data shape",
					fname, name, ftype))
				continue
			}
			fields[name] = append(fields[name], contractDocField{Name: fname, Type: ftype})
		}
	}
	for i := range contractDTOs {
		if !seen[contractDTOs[i].label] {
			problems = append(problems, fmt.Sprintf(
				"docs/specs/educated-routing.md lost the %q complete data shape; remedy: restore the `TypeName`: field list paragraph",
				contractDTOs[i].label))
		}
	}
	if seen["Error"] {
		problems = append(problems, "docs/specs/educated-routing.md defines Error as a `TypeName`: DTO; Error is not wire-defined yet and must stay out of the complete data shapes")
	}
	return fields, problems
}

// normalizeContractType strips the routing. qualifier; time.Time is
// preserved exactly as documented.
func normalizeContractType(t reflect.Type) string {
	return strings.ReplaceAll(t.String(), "routing.", "")
}

func snakeCase(s string) string {
	var b strings.Builder
	r := []rune(s)
	lowerRun := func(j int) int {
		n := 0
		for j+n < len(r) && unicode.IsLower(r[j+n]) {
			n++
		}
		return n
	}
	for i := range r {
		if unicode.IsUpper(r[i]) {
			if i > 0 && (unicode.IsLower(r[i-1]) ||
				(unicode.IsUpper(r[i-1]) && lowerRun(i+1) > 1)) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r[i]))
		} else {
			b.WriteRune(r[i])
		}
	}
	return b.String()
}

func omitemptyPermitted(label, field string) bool {
	for _, a := range contractOmitemptyAllowances {
		if a.label == label && a.field == field {
			return true
		}
	}
	return false
}

func compareContractDTOs(doc string) []string {
	var drift []string
	for _, a := range contractOmitemptyAllowances {
		if !strings.Contains(doc, a.sentence) {
			drift = append(drift, fmt.Sprintf(
				"docs/specs/educated-routing.md lost the omitempty allowance %q; remedy: restore the sentence or remove the %s.%s exception from the drift guard",
				a.sentence, a.label, a.field))
		}
	}
	docFields, problems := parseContractDTOFields(doc)
	drift = append(drift, problems...)
	for _, e := range contractDTOs {
		df := docFields[e.label]
		codeFile := "routing/" + e.file
		byName := map[string]contractDocField{}
		for _, f := range df {
			if _, dup := byName[f.Name]; dup {
				drift = append(drift, fmt.Sprintf(
					"docs/specs/educated-routing.md lists field %q twice in the %s complete data shape; remedy: keep one entry per field", f.Name, e.label))
				continue
			}
			byName[f.Name] = f
		}
		st := e.typ
		sfByName := map[string]reflect.StructField{}
		for i := range st.NumField() {
			f := st.Field(i)
			if _, dup := sfByName[f.Name]; dup {
				drift = append(drift, fmt.Sprintf(
					"routing/%s declares field %q twice in %s; remedy: keep one field per name", e.file, f.Name, e.label))
				continue
			}
			sfByName[f.Name] = f
		}
		for _, f := range df {
			sf, ok := sfByName[f.Name]
			if !ok {
				drift = append(drift, fmt.Sprintf(
					"docs/specs/educated-routing.md documents %s field %s %s but %s has no such field; remedy: add the exported field with its snake_case JSON tag or fix the document",
					e.label, f.Name, f.Type, codeFile))
				continue
			}
			if got := normalizeContractType(sf.Type); got != f.Type {
				drift = append(drift, fmt.Sprintf(
					"docs/specs/educated-routing.md documents %s.%s as %q but %s declares %q; remedy: align the Go type with the document (routing. qualifier may be stripped, time.Time preserved)",
					e.label, f.Name, f.Type, codeFile, got))
			}
		}
		for name, sf := range sfByName {
			if _, ok := byName[name]; !ok {
				drift = append(drift, fmt.Sprintf(
					"%s declares %s field %s %s that docs/specs/educated-routing.md does not document; remedy: add `Name Type` to the %s field list or remove the field from the struct",
					codeFile, e.label, name, normalizeContractType(sf.Type), e.label))
				continue
			}
			tag := sf.Tag.Get("json")
			jsonName := tag
			if i := strings.Index(tag, ","); i >= 0 {
				jsonName, _ = tag[:i], tag[i+1:]
			}
			want := snakeCase(name)
			if jsonName != want {
				drift = append(drift, fmt.Sprintf(
					"%s JSON tag for %s.%s is %q but docs/specs/educated-routing.md requires the snake-case form %q; remedy: set the json tag to %q",
					codeFile, e.label, name, jsonName, want, want))
			}
			if strings.Contains(tag, ",omitempty") && !omitemptyPermitted(e.label, name) {
				drift = append(drift, fmt.Sprintf(
					"%s %s.%s carries omitempty but docs/specs/educated-routing.md does not permit it there; remedy: drop the option so the field is always present",
					codeFile, e.label, name))
			}
		}
	}
	return drift
}

func compareContractEnums(doc string) []string {
	var drift []string
	for i := range contractEnumSpecs {
		spec := &contractEnumSpecs[i]
		count := strings.Count(doc, spec.prefix)
		if count != 1 {
			drift = append(drift, fmt.Sprintf(
				"docs/specs/educated-routing.md must state the %q closed vocabulary exactly once (found %d); remedy: restore the single `%s ...` enumeration paragraph",
				spec.prefix, count, spec.prefix))
			continue
		}
		from := strings.Index(doc, spec.prefix) + len(spec.prefix)
		to := strings.Index(doc[from:], spec.terminator)
		if to < 0 {
			drift = append(drift, fmt.Sprintf(
				"docs/specs/educated-routing.md %q list is unterminated before %q; remedy: close the token enumeration",
				spec.prefix, spec.terminator))
			continue
		}
		docTokens := map[string]int{}
		for _, m := range contractTokenRe.FindAllStringSubmatch(doc[from:from+to], -1) {
			docTokens[m[1]]++
		}
		for tok, n := range docTokens {
			if n > 1 {
				drift = append(drift, fmt.Sprintf(
					"docs/specs/educated-routing.md lists token %q %d times for %s but the document requires no duplicate tokens; remedy: keep exactly one occurrence",
					tok, n, spec.name))
			}
			if !spec.target[tok] {
				drift = append(drift, fmt.Sprintf(
					"docs/specs/educated-routing.md lists token %q for %s but routing/tokens.go does not accept it; remedy: add the token to the closed vocabulary map or remove it from the document",
					tok, spec.name))
			}
		}
		for tok, accepts := range spec.target {
			if !accepts {
				drift = append(drift, fmt.Sprintf(
					"routing/tokens.go maps %q for %s to false; a closed vocabulary accepts only true-valued tokens and this unknown false key is dead; remedy: remove the entry or set it to true and document the token",
					tok, spec.name))
				continue
			}
			if docTokens[tok] == 0 {
				drift = append(drift, fmt.Sprintf(
					"routing/tokens.go accepts %q for %s but docs/specs/educated-routing.md does not list it; remedy: remove the token from the map or document it in the enumeration",
					tok, spec.name))
			}
		}
	}
	return drift
}

func TestContractDTODrift(t *testing.T) {
	drift := compareContractDTOs(readContractDoc(t))
	for _, d := range drift {
		t.Errorf("DTO contract drift: %s", d)
	}
}

func TestContractEnumDrift(t *testing.T) {
	drift := compareContractEnums(readContractDoc(t))
	for _, d := range drift {
		t.Errorf("enum token contract drift: %s", d)
	}
}

// TestContractDriftSeededMutation proves the comparators actually fail on
// drift: a seeded document edit removing a DTO field and adding an unknown
// enum token must be reported, while the real document passes.
func TestContractDriftSeededMutation(t *testing.T) {
	doc := readContractDoc(t)
	mutated := strings.Replace(doc, "`UpperMicros int64`, ", "", 1)
	if mutated == doc {
		t.Fatal("seed fixture: failed to remove UpperMicros from the mutated document")
	}
	mutated = strings.Replace(mutated, "`image_editing`.", "`image_editing`, `mystery_task`.", 1)
	if !strings.Contains(mutated, "`mystery_task`") {
		t.Fatal("seed fixture: failed to add mystery_task to the mutated document")
	}
	tasksSeed := "`general`, `code`, `review`"
	if strings.Count(doc, tasksSeed) != 1 || strings.Count(mutated, tasksSeed) != 1 {
		t.Fatal("seed fixture: Tasks enumeration seed text is not unique")
	}
	mutated = strings.Replace(mutated, tasksSeed, "`general`, `code`, `code`, `review`", 1)
	if strings.Count(mutated, "`code`") != strings.Count(doc, "`code`")+1 {
		t.Fatal("seed fixture: failed to duplicate `code` in the Tasks enumeration")
	}
	dtoDrift := compareContractDTOs(mutated)
	sawField := false
	for _, d := range dtoDrift {
		if strings.Contains(d, "UpperMicros") {
			sawField = true
		}
	}
	if !sawField {
		t.Errorf("seeded field removal was not detected; got drift %v", dtoDrift)
	}
	enumDrift := compareContractEnums(mutated)
	sawToken, sawDup := false, false
	for _, d := range enumDrift {
		if strings.Contains(d, "mystery_task") {
			sawToken = true
		}
		if strings.Contains(d, "duplicate") {
			sawDup = true
		}
	}
	if !sawToken {
		t.Errorf("seeded enum token addition was not detected; got drift %v", enumDrift)
	}
	if !sawDup {
		t.Errorf("seeded duplicate enum token was not detected; got drift %v", enumDrift)
	}
	// Each hardcoded omitempty exception depends on its documented
	// allowance sentence; removing either must be flagged so removal of a
	// sentence cannot leave its exception unguarded.
	for _, clause := range []string{
		"Zero `ConfirmedAt` omits `confirmed_at`;",
		"nil `Selected` omits `selected`;",
	} {
		noAllowance := strings.Replace(doc, clause, "", 1)
		if noAllowance == doc {
			t.Fatalf("seed fixture: failed to remove allowance clause %q", clause)
		}
		allowanceDrift := compareContractDTOs(noAllowance)
		sawAllowance := false
		for _, d := range allowanceDrift {
			if strings.Contains(d, "omitempty") {
				sawAllowance = true
			}
		}
		if !sawAllowance {
			t.Errorf("seeded omitempty allowance removal %q was not detected; got drift %v", clause, allowanceDrift)
		}
	}
	if d := compareContractDTOs(doc); len(d) != 0 {
		t.Errorf("real document unexpectedly drifted after seeding: %v", d)
	}
	if d := compareContractEnums(doc); len(d) != 0 {
		t.Errorf("real document enum tokens unexpectedly drifted after seeding: %v", d)
	}
	if d := compareContractDTOs(strings.Replace(doc, "Zero `ConfirmedAt` omits `confirmed_at`;", "", 1)); len(d) == 0 {
		t.Error("allowance removal is expected to drift on the real document path")
	}
}

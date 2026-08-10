package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// componentSchema is the real schema, read the way the leg reads it.
//
// The real one rather than a copy, because a copy is a second statement of what
// a record must carry and nothing would refuse the two drifting apart.
func componentSchema(t *testing.T) *schema {
	t.Helper()
	const rel = "data/schema/component-record.schema.json"
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("reading the schema: %v", err)
	}
	s, err := readSchema(rel, src)
	if err != nil {
		t.Fatalf("reading the schema: %v", err)
	}
	return s
}

func fixtureSchema(t *testing.T, name string) *schema {
	t.Helper()
	rel := filepath.Join("testdata", "data", name)
	src, err := os.ReadFile(rel)
	if err != nil {
		t.Fatalf("reading the fixture schema %s: %v", name, err)
	}
	s, err := readSchema(name, src)
	if err != nil {
		t.Fatalf("reading the fixture schema %s: %v", name, err)
	}
	return s
}

func fixtureRecord(t *testing.T, name string) []byte {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "data", name))
	if err != nil {
		t.Fatalf("reading the fixture %s: %v", name, err)
	}
	return src
}

func refusals(t *testing.T, s *schema, name string) []problem {
	t.Helper()
	found, err := s.Validate(fixtureRecord(t, name))
	if err != nil {
		t.Fatalf("validating %s: %v", name, err)
	}
	return found
}

// TestTheNearMissIsAccepted is the half of the proof that decides whether
// anybody can live with the rule. Every fixture below is this file with one edit
// in it, so a reader that refused this one would be refusing the legal shape
// rather than the illegal ones.
func TestTheNearMissIsAccepted(t *testing.T) {
	for _, p := range refusals(t, componentSchema(t), "the-near-miss.json") {
		t.Errorf("the near miss was refused: %s", p)
	}
}

// TestOneEditIsEnoughToBeRefused is the other half. Each fixture is the near
// miss with a single thing changed, and each names a different way a record can
// be wrong that a person reading a diff would not see.
//
// The counts are exact rather than a lower bound. A reader that reported one
// fault twice would pass an assertion that it found at least one, and a
// duplicated finding is how a check gets read past and then switched off.
func TestOneEditIsEnoughToBeRefused(t *testing.T) {
	s := componentSchema(t)
	for _, c := range []struct {
		fixture string
		count   int
		at      string
		says    string
		catches string
	}{
		{
			fixture: "missing-required-field.json",
			count:   1, at: "provenance", says: `is missing the required field "report_number"`,
			catches: "a record whose provenance does not reach the certificate floor",
		},
		{
			fixture: "unknown-field.json",
			count:   1, at: "", says: `carries "notes"`,
			catches: "a field the schema does not declare, which is usually a field name typed wrongly",
		},
		{
			fixture: "band-missing.json",
			count:   1, at: "airborne_lab/values", says: `is missing the required field "2000"`,
			catches: "a spectrum one band short, which is the failure this whole project is arranged against",
		},
		{
			fixture: "field-the-kind-forbids.json",
			count:   1, at: "", says: `carries "base_construction"`,
			catches: "a field that belongs to another kind of record entirely",
		},
		{
			fixture: "both-loss-factor-forms.json",
			count:   1, at: "", says: "and exactly one is allowed",
			catches: "a loss factor and a statement that there is none, which leaves a reader two answers",
		},
		{
			fixture: "spectrum-without-its-specimen.json",
			count:   2, at: "", says: `requires "specimen_area"`,
			catches: "a laboratory spectrum that cannot be corrected to a building",
		},
	} {
		t.Run(c.fixture, func(t *testing.T) {
			found := refusals(t, s, c.fixture)
			if len(found) != c.count {
				t.Fatalf("%d problem(s), want %d, in the fixture that catches %s: %v", len(found), c.count, c.catches, found)
			}
			hit := false
			for _, p := range found {
				if p.Instance == c.at && strings.Contains(p.Message, c.says) {
					hit = true
				}
			}
			if !hit {
				t.Errorf("no problem at %q saying %q; what was found: %v", c.at, c.says, found)
			}
		})
	}
}

// provenBy is where each keyword the reader implements is shown to refuse
// something, as the instance pointer its refusal appears at in
// keywords-illegal.json.
//
// This map is the count rather than the presence. A keyword added to the reader
// without an entry here turns the test below red, because a keyword that is read
// and then refuses nothing is indistinguishable from one that is enforced, and
// that is the exact shape this reader exists against.
var provenBy = map[string]string{
	"$ref":                  "ref_keyword",
	"allOf":                 "all_of_keyword",
	"anyOf":                 "any_of_keyword",
	"oneOf":                 "one_of_keyword",
	"if":                    "if_then_keyword",
	"then":                  "if_then_keyword",
	"else":                  "else_keyword",
	"properties":            "type_keyword",
	"items":                 "items_keyword/1",
	"additionalProperties":  "additional_properties_keyword/a",
	"propertyNames":         "property_names_keyword",
	"type":                  "type_keyword",
	"enum":                  "enum_keyword",
	"const":                 "const_keyword",
	"required":              "",
	"dependentRequired":     "",
	"minLength":             "min_length_keyword",
	"pattern":               "pattern_keyword",
	"minimum":               "minimum_keyword",
	"maximum":               "maximum_keyword",
	"exclusiveMinimum":      "exclusive_minimum_keyword",
	"minItems":              "min_items_keyword",
	"minProperties":         "min_properties_keyword",
	"unevaluatedProperties": provenSeparately,
}

// provenSeparately marks a keyword whose refusal cannot appear in the fixture
// above. `unevaluatedProperties` reports only where nothing else is wrong, which
// is stated on the reader, so its proof is a fixture that is legal apart from
// the one undeclared field.
const provenSeparately = "on its own fixture"

func TestEveryKeywordTheReaderImplementsIsShownToRefuseSomething(t *testing.T) {
	for keyword, kind := range keywords {
		if kind == "annotation" {
			continue
		}
		if _, ok := provenBy[keyword]; !ok {
			t.Errorf("the reader implements %q and nothing shows it refusing anything", keyword)
		}
	}
	for keyword := range provenBy {
		if _, ok := keywords[keyword]; !ok {
			t.Errorf("%q is proven here and the reader does not implement it", keyword)
		}
	}
}

func TestEveryKeywordBitesOnTheFixtureWrittenForIt(t *testing.T) {
	s := fixtureSchema(t, "keywords.schema.json")

	if found := refusals(t, s, "keywords-legal.json"); len(found) > 0 {
		t.Errorf("the legal fixture was refused: %v", found)
	}

	found := refusals(t, s, "keywords-illegal.json")
	at := map[string]int{}
	for _, p := range found {
		at[p.Instance]++
	}
	want := map[string]int{}
	for keyword, where := range provenBy {
		if where == provenSeparately {
			continue
		}
		// `required` and `dependentRequired` both refuse the document itself,
		// and each is a separate refusal at the same place.
		if where == "" {
			want[""]++
			continue
		}
		want[where] = 1
		_ = keyword
	}
	for where, count := range want {
		if at[where] != count {
			t.Errorf("%d refusal(s) at %q, want %d: %v", at[where], where, count, found)
		}
	}
	for where := range at {
		if _, expected := want[where]; !expected {
			t.Errorf("a refusal at %q that no keyword claims: %v", where, found)
		}
	}

	undeclared := refusals(t, s, "keywords-undeclared.json")
	if len(undeclared) != 1 || undeclared[0].Instance != "" || !strings.Contains(undeclared[0].Message, "undeclared_keyword") {
		t.Errorf("the undeclared field was not refused as one thing: %v", undeclared)
	}
}

// TestAKeywordTheReaderDoesNotImplementIsRefused is the property that bounds the
// reader honestly. It implements what this repository's schemas use and no more,
// and the safety of that bound is entirely this refusal: a reader that skipped
// an unknown keyword would pass every record that keyword would have refused,
// green and in silence.
func TestAKeywordTheReaderDoesNotImplementIsRefused(t *testing.T) {
	src := fixtureRecord(t, "unimplemented-keyword.schema.json")
	_, err := readSchema("unimplemented-keyword.schema.json", src)
	if err == nil {
		t.Fatal("a schema asking for patternProperties was read without complaint")
	}
	if !strings.Contains(err.Error(), "patternProperties") {
		t.Errorf("the refusal does not name the keyword: %v", err)
	}
}

// TestTheComponentSchemaIsFullyImplemented is the rule reaching this tree. It is
// the same refusal as the test above, run against the schema the database is
// actually judged by, so a keyword added to that schema is a red run rather than
// a field nothing checks.
func TestTheComponentSchemaIsFullyImplemented(t *testing.T) {
	s := componentSchema(t)
	if defects := s.defects(); len(defects) > 0 {
		names := make([]string, 0, len(defects))
		for _, d := range defects {
			names = append(names, d.String())
		}
		sort.Strings(names)
		t.Errorf("the schema asks for %d thing(s) the reader does not implement:\n%s", len(defects), strings.Join(names, "\n"))
	}
}

package main

import (
	"io"
	"testing"
)

// TestEveryPathIsPlacedSomewhere is the classification, including the near
// misses. Each illegal row is one directory or one suffix away from a legal one,
// because that is the mistake somebody actually makes: a record saved beside the
// schema, a record saved outside data/ entirely, a fixture that drifted out of a
// testdata directory.
func TestEveryPathIsPlacedSomewhere(t *testing.T) {
	for _, c := range []struct {
		path string
		want fileClass
		why  string
	}{
		{"data/schema/component-record.schema.json", schemaDocument, "the schema itself"},
		{"data/floor/ift-17-002083-pr01-x01-x02.json", dataRecord, "a record where records go"},
		{"data/wall/anything-at-all.json", dataRecord, "a record under a directory nobody has made yet"},
		{"data/README.md", dataProse, "documentation beside the data"},

		{"data/schema/notes.json", unclaimed, "a record saved into the schema directory"},
		{"data/floor/a-record.yaml", unclaimed, "a record in a format no schema claims"},
		{"a-record-at-the-root.json", unclaimed, "a record outside data/, which a directory rule would not see at all"},
		{"docs/a-record-in-the-documents.json", unclaimed, "the same mistake one directory along"},

		{"cmd/gate/testdata/data/the-near-miss.json", outsideTheData, "a fixture, which is where a file shaped like a record is deliberately not one"},
		{"store/testdata/wall-r-core.spectrum", outsideTheData, "a fixture in another package"},
		{".github/opengrep/near-miss/bounded-number.schema.json", outsideTheData, "configuration for a tool that runs this repository"},
		{"docs/decisions/data-format.md", outsideTheData, "a document"},
		{"go.mod", outsideTheData, "the module file"},
	} {
		if got, _ := classify(c.path); got != c.want {
			t.Errorf("%s was placed as %d, want %d: it is %s", c.path, got, c.want, c.why)
		}
	}
}

// TestEveryRecordInTheTreeIsClaimed states the mapping's own count. The leg
// refuses a tree in which nothing was validated, and this says the same thing
// where a reader of the suite will see it: the rule is not passing because there
// is nothing to judge.
func TestEveryRecordInTheTreeIsClaimed(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}
	files, err := trackedFiles(root)
	if err != nil {
		t.Fatalf("listing the tracked files: %v", err)
	}

	records, schemas := 0, 0
	for _, rel := range files {
		switch class, _ := classify(rel); class {
		case dataRecord:
			records++
		case schemaDocument:
			schemas++
		case unclaimed:
			t.Errorf("%s is inside the data system and no schema claims it", rel)
		}
	}
	if records == 0 {
		t.Error("no record in the tree is claimed by a schema, so a green run of this leg says nothing")
	}
	if schemas == 0 {
		t.Error("no schema in the tree, so there is nothing for a record to be judged against")
	}
}

// TestTheRuleReachesThisRepository is how the rule arrives here without anything
// being added to a list. The leg is run over the repository itself.
func TestTheRuleReachesThisRepository(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}
	if err := checkData(root, io.Discard); err != nil {
		t.Errorf("this repository holds a data file that does not match its schema:\n%v", err)
	}
}

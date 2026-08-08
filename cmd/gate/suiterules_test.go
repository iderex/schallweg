package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTheRulesBiteOnAFixtureThatBreaksThem is the proof that the reader refuses
// what it says it refuses. A rule with no fixture behind it is a rule that has
// never been observed to fire, and the way it usually fails is by matching
// nothing at all, which looks exactly like a clean tree.
//
// The expected counts are exact rather than a lower bound. A reader that started
// reporting a rule twice for one line would still pass a "found at least one"
// assertion, and a duplicate finding is how a reader gets switched off.
func TestTheRulesBiteOnAFixtureThatBreaksThem(t *testing.T) {
	want := map[string]int{
		"float-equality":    3,
		"clock":             1,
		"random-import":     1,
		"network-import":    1,
		"working-directory": 1,
		"escaping-path":     3,
	}

	got := countByRule(t, "breaks-every-rule.go.txt")

	for rule, n := range want {
		if got[rule] != n {
			t.Errorf("rule %q fired %d times, want %d", rule, got[rule], n)
		}
	}
	for rule, n := range got {
		if _, expected := want[rule]; !expected {
			t.Errorf("rule %q fired %d times and was not expected at all", rule, n)
		}
	}
}

// TestTheRulesPassTheNearMiss is the other half, and it is the half that decides
// whether anybody can live with the reader. The fixture does every one of the
// same things legally, each one a character or an import away from the fixture
// above.
func TestTheRulesPassTheNearMiss(t *testing.T) {
	got := countByRule(t, "the-near-miss.go.txt")
	if len(got) != 0 {
		t.Errorf("the near miss produced findings and must produce none: %v", got)
	}
}

// TestOrdinarySuiteObeysItsOwnRules is the rule applied to this repository.
//
// It needs a git checkout, because it asks git for the repository root rather
// than walking upwards from a working directory the toolchain is free to change.
// That dependency is real and is named here: the ordinary suite's four
// conditions are no display, no elevation, no network and no hardware, and an
// installed git is none of those, but it is still something a machine has to
// have.
func TestOrdinarySuiteObeysItsOwnRules(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("cannot find the repository root: %v", err)
	}

	found, scanned, err := checkTree(root)
	if err != nil {
		t.Fatalf("cannot read the tree: %v", err)
	}

	// Fail closed. A walker that reached no test file is not evidence of a
	// clean tree, and it is the failure that looks exactly like success.
	if scanned == 0 {
		t.Fatal("the reader scanned no test files at all; a check that reads nothing is not a check")
	}
	t.Logf("scanned %d test file(s)", scanned)

	for _, v := range found {
		t.Errorf("%s", v)
	}
}

// TestTheReaderSkipsWhatItSaysItSkips holds the two exemptions, because an
// exemption that quietly widens is how a reader stops covering the tree while
// still reporting a number.
func TestTheReaderSkipsWhatItSaysItSkips(t *testing.T) {
	dir := t.TempDir()

	write := func(rel string) {
		t.Helper()
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		src := "package p\n\nimport \"testing\"\n\nfunc TestX(t *testing.T) {\n\tif x() == 1.5 {\n\t\tt.Error(\"e\")\n\t}\n}\n"
		if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("harness/requires-network/dial_test.go")
	write("kernel/testdata/sample_test.go")
	write("kernel/real_test.go")

	found, scanned, err := checkTree(dir)
	if err != nil {
		t.Fatalf("cannot read the tree: %v", err)
	}
	if scanned != 1 {
		t.Errorf("scanned %d file(s), want 1: the harness directory and any testdata directory are skipped", scanned)
	}
	if len(found) != 1 {
		t.Fatalf("found %d finding(s), want 1: %v", len(found), found)
	}
	if found[0].Path != "kernel/real_test.go" {
		t.Errorf("the finding is in %q, want it in kernel/real_test.go", found[0].Path)
	}
}

// countByRule runs the reader over a fixture and returns how often each rule
// fired.
func countByRule(t *testing.T, name string) map[string]int {
	t.Helper()

	path := filepath.Join("testdata", "suiterules", name)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read the fixture: %v", err)
	}

	found, err := checkTestSource(name, src)
	if err != nil {
		t.Fatalf("cannot read the fixture as source: %v", err)
	}

	counts := map[string]int{}
	for _, v := range found {
		counts[v.Rule]++
	}
	return counts
}

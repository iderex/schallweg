package main

// What the harness audit refuses, and the one thing it reports without refusing.
//
// Every case is a tree written down rather than this repository's own, because
// no harness exists here today. A rule judged only against the tree it ships in
// is a rule that passes for the reason there is nothing to judge, and it would
// go on passing after somebody added the first harness wrongly.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// source reads one of the written-down Go files beside this test. They carry the
// .go.txt suffix so the build does not compile a file whose whole point is a
// build constraint.
func source(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "harness", name))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	return string(b)
}

// TestAConstraintIsReadWhereTheCompilerReadsOne is the near miss, and it is one
// blank line and one position away from the file above it.
//
// A build constraint after the package clause is not a build constraint. It
// reads in a diff exactly like one, the file builds into the ordinary suite, and
// the reviewer who approved it saw the words they were looking for.
func TestAConstraintIsReadWhereTheCompilerReadsOne(t *testing.T) {
	behind := constraintTags(source(t, "behind-the-constraint.go.txt"))
	if len(behind) != 1 || behind[0] != "requires_third_party_tool" {
		t.Fatalf("the constrained file reports %v, want the one harness tag", behind)
	}
	near := constraintTags(source(t, "the-near-miss.go.txt"))
	if len(near) != 0 {
		t.Fatalf("a line after the package clause was read as a constraint: %v", near)
	}
}

// TestEveryTagInAnExpressionIsRead covers the shape where a harness is also
// pinned to a platform. Reading only the first tag would let a harness hide
// behind an ordinary one.
func TestEveryTagInAnExpressionIsRead(t *testing.T) {
	tags := constraintTags(source(t, "two-tags-in-one-expression.go.txt"))
	if len(tags) != 2 || tags[0] != "linux" || tags[1] != "requires_instrument" {
		t.Fatalf("the expression reports %v, want both of its tags", tags)
	}
}

// TestAWellFormedHarnessIsAccepted is the case that has to pass, so that the
// refusals below are refusals of something rather than of everything.
func TestAWellFormedHarnessIsAccepted(t *testing.T) {
	found, refusals := judgeHarnesses(
		map[string][]string{"harness/requires-network/probe_test.go": {"requires_network"}},
		[]string{"requires-network"},
		map[string]bool{"requires-network": true},
	)
	if len(refusals) != 0 {
		t.Fatalf("a well-formed harness was refused: %v", refusals)
	}
	if len(found) != 1 || found[0].name != "requires-network" || len(found[0].files) != 1 {
		t.Fatalf("the harness was read as %v", found)
	}
}

// TestAHarnessWhoseNameSaysNothingIsRefused is the condition this issue leads
// with. A harness called integration tells a contributor nothing about the four
// conditions, so they try it, it fails, and the failure reads as a defect in the
// code rather than a missing prerequisite.
func TestAHarnessWhoseNameSaysNothingIsRefused(t *testing.T) {
	_, refusals := judgeHarnesses(
		map[string][]string{"harness/integration/run_test.go": {"integration"}},
		[]string{"integration"},
		map[string]bool{"integration": true},
	)
	if !refused(refusals, "does not say what it requires") {
		t.Fatalf("a harness called integration was accepted: %v", refusals)
	}
}

// TestAHarnessWithNoStatementIsRefused holds the second half of the naming rule.
// The name says which of the four it needs; the statement beside it says what
// exactly, and how a contributor would know they have it.
func TestAHarnessWithNoStatementIsRefused(t *testing.T) {
	_, refusals := judgeHarnesses(
		map[string][]string{"harness/requires-display/window_test.go": {"requires_display"}},
		[]string{"requires-display"},
		map[string]bool{},
	)
	if !refused(refusals, "carries no README.md") {
		t.Fatalf("a harness with nothing written beside it was accepted: %v", refusals)
	}
}

// TestAHarnessFileNotBehindItsConstraintIsRefused is the one that matters most,
// because the file compiles and the suite goes green. A test needing a display,
// sitting in the harness tree and built by the ordinary run, is the ordinary
// suite quietly acquiring the dependency the rule exists against.
func TestAHarnessFileNotBehindItsConstraintIsRefused(t *testing.T) {
	for name, tags := range map[string][]string{
		"behind nothing at all":   {},
		"behind another tag only": {"linux"},
	} {
		_, refusals := judgeHarnesses(
			map[string][]string{"harness/requires-display/window_test.go": tags},
			[]string{"requires-display"},
			map[string]bool{"requires-display": true},
		)
		if !refused(refusals, "the ordinary run builds it") {
			t.Errorf("a harness file %s was accepted: %v", name, refusals)
		}
	}
}

// TestAHarnessConstraintOutsideTheHarnessTreeIsRefused is the other direction,
// and it is the one a reader is least likely to think of.
//
// Such a file is built by nothing and named by nothing. The ordinary run skips
// it silently, the report below does not know it exists, and a run that skipped
// it reads exactly like a run that had nothing to skip.
func TestAHarnessConstraintOutsideTheHarnessTreeIsRefused(t *testing.T) {
	_, refusals := judgeHarnesses(
		map[string][]string{"store/reader_slow_test.go": {"requires_network"}},
		nil,
		map[string]bool{},
	)
	if !refused(refusals, "skipped by every run and named by none") {
		t.Fatalf("a harness constraint outside the harness tree was accepted: %v", refusals)
	}
}

// TestTheReportNamesEveryHarnessAndCallsNoneOfThemAPass is the reporting half.
// A skipped harness is never counted as a pass and the counts are never summed,
// because one total is the shape that lets an unrun harness disappear into a
// green figure.
func TestTheReportNamesEveryHarnessAndCallsNoneOfThemAPass(t *testing.T) {
	var out strings.Builder
	reportHarnesses(&out, []string{"example.com/a"}, []harness{
		{name: "requires-display", files: []string{"harness/requires-display/w_test.go"}, needs: "a display or a substitute for one"},
		{name: "requires-network", files: []string{"harness/requires-network/p_test.go"}, needs: "outbound network access"},
	})
	got := out.String()
	for _, want := range []string{
		"enumerated from the build",
		"none of them was run here",
		"requires-display",
		"a display or a substitute for one",
		"requires-network",
		"outbound network access",
		"not summed",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the report does not say %q; it says:\n%s", want, got)
		}
	}
}

// TestTheReportSaysSoWhenThereIsNothingToSkip covers the tree as it stands. A
// silent report and a report of nothing are different sentences, and only the
// second one can be read as an answer.
func TestTheReportSaysSoWhenThereIsNothingToSkip(t *testing.T) {
	var out strings.Builder
	reportHarnesses(&out, []string{"example.com/a"}, nil)
	if !strings.Contains(out.String(), "so this run skipped nothing") {
		t.Errorf("the report of an empty harness tree is:\n%s", out.String())
	}
}

// TestEveryDeclaredHarnessNameSaysWhatItRequires holds the vocabulary itself to
// the convention it exists to impose.
func TestEveryDeclaredHarnessNameSaysWhatItRequires(t *testing.T) {
	for _, name := range declaredHarnesses() {
		if !strings.HasPrefix(name, "requires-") {
			t.Errorf("%q is a declared harness name that does not say what it requires", name)
		}
		if constraintTag(name) == name && strings.Contains(name, "-") {
			t.Errorf("%q and its build tag are the same string, and a build tag carries no hyphen", name)
		}
	}
}

// refused reports whether any refusal says this.
func refused(refusals []string, says string) bool {
	for _, r := range refusals {
		if strings.Contains(r, says) {
			return true
		}
	}
	return false
}

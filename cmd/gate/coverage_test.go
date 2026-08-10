package main

// What the coverage bar refuses, judged against written-down reports rather than
// against this repository's own coverage on the day the test runs.
//
// A row that measures the real tree proves the state of the tree and not the
// guard: it goes green the moment somebody writes a test and says nothing about
// whether the bar would have caught the gap. Every case here is a report that
// could arrive, including the three shapes where the check cannot measure at
// all, which are the ones a check quietly passes if nobody wrote them down.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// report reads one of the written-down reports beside this test.
func report(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "coverage", name))
	if err != nil {
		t.Fatalf("reading the report: %v", err)
	}
	return string(b)
}

// judge parses a report and judges it against the deciding surfaces and the
// bar the tree ships, so a change to either is a change these cases see.
func judge(t *testing.T, name string) error {
	t.Helper()
	parsed, err := parseProfile(report(t, name))
	if err != nil {
		return err
	}
	return judgeCoverage(parsed, decidingSurfaces(), coverageBar)
}

// TestASurfaceAboveTheBarPasses is the case that has to pass, and it is here so
// that every refusal below is a refusal of something rather than of everything.
func TestASurfaceAboveTheBarPasses(t *testing.T) {
	if err := judge(t, "above-the-bar.out"); err != nil {
		t.Fatalf("a surface at 95 percent was refused against a bar of %v: %v", coverageBar, err)
	}
}

// TestASurfaceExactlyAtTheBarPasses fixes the boundary, which is the one place a
// bar is read two ways. Standing at the bar is standing at it.
func TestASurfaceExactlyAtTheBarPasses(t *testing.T) {
	if err := judge(t, "exactly-at-the-bar.out"); err != nil {
		t.Fatalf("a surface standing exactly at the bar was refused: %v", err)
	}
}

// TestOneStatementBelowTheBarIsRefused is the deliberate drop, and it is one
// statement rather than a collapse. A bar that only catches a large fall is a
// bar nothing ever reaches.
func TestOneStatementBelowTheBarIsRefused(t *testing.T) {
	err := judge(t, "one-statement-below-the-bar.out")
	if err == nil {
		t.Fatal("a surface below the bar was accepted")
	}
	for _, want := range []string{"acoustic", "92 of 100", "93.0%"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal is %q, and it does not say %q", err, want)
		}
	}
	// The message says what to do rather than only how far off it is.
	if !strings.Contains(err.Error(), "1 more statement(s)") {
		t.Errorf("the refusal is %q, and it does not say how many statements have to be reached", err)
	}
}

// TestASurfaceTheReportDoesNotMeasureIsRefused is the fail-closed case that
// matters most, because it is the one that looks like a pass.
//
// A report that measured something else is a report, and a check reading it for
// a package it does not mention finds nothing wrong. Nothing wrong and nothing
// measured are the same bytes, and the safe reading of the pair is a refusal.
func TestASurfaceTheReportDoesNotMeasureIsRefused(t *testing.T) {
	err := judge(t, "surface-not-measured.out")
	if err == nil {
		t.Fatal("a report measuring no statement of a listed surface was accepted")
	}
	if !strings.Contains(err.Error(), "measures no statement of it") {
		t.Errorf("the refusal is %q, and it does not say the surface was not measured", err)
	}
	// It also says why the surface is on the list, so the reader can decide
	// whether the list or the run is what is wrong.
	if !strings.Contains(err.Error(), "it is on the list because") {
		t.Errorf("the refusal is %q, and it does not say why the surface is on the list", err)
	}
}

// TestAReportThatCannotBeReadIsRefused covers the three shapes of an unusable
// report. Each one is a run that measured nothing, and a check that treats any
// of them as clean is decoration from the day the report stops being written.
func TestAReportThatCannotBeReadIsRefused(t *testing.T) {
	for _, c := range []struct{ file, says string }{
		{"empty.out", "empty"},
		{"mode-line-and-nothing-measured.out", "no measured statement"},
		{"no-mode-line.out", "mode line"},
		{"statement-count-is-not-a-number.out", "statement count"},
	} {
		t.Run(c.file, func(t *testing.T) {
			err := judge(t, c.file)
			if err == nil {
				t.Fatalf("%s was accepted", c.file)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Errorf("the refusal is %q, and it does not say %q", err, c.says)
			}
		})
	}
}

// TestTheWholeTreeFigureIsReportedAndNotGated holds the second half of the
// design. The figure over everything is what a reader wants beside the bar, and
// gating on it would let a gap in the arithmetic hide behind a well-covered
// command.
func TestTheWholeTreeFigureIsReportedAndNotGated(t *testing.T) {
	parsed, err := parseProfile(report(t, "surface-not-measured.out"))
	if err != nil {
		t.Fatalf("parsing the report: %v", err)
	}
	total := wholeTree(parsed)
	if total.statements != 40 || total.covered != 40 {
		t.Fatalf("the whole-tree figure is %d of %d, want 40 of 40", total.covered, total.statements)
	}
	// That report is entirely covered and it is still refused, because the
	// surface the bar names is not in it.
	if judgeCoverage(parsed, decidingSurfaces(), coverageBar) == nil {
		t.Error("a fully covered whole tree was allowed to answer for a surface nobody measured")
	}
}

// TestEveryDecidingSurfaceCarriesItsReason is the rule the list is kept under.
// A package added here without one is a package nobody can argue with later.
func TestEveryDecidingSurfaceCarriesItsReason(t *testing.T) {
	surfaces := decidingSurfaces()
	if len(surfaces) == 0 {
		t.Fatal("the bar applies to no surface, so it refuses nothing")
	}
	for _, s := range surfaces {
		if !strings.HasPrefix(s.pkg, modulePath) || len(s.pkg) <= len(modulePath) {
			t.Errorf("%q is not a package inside this module", s.pkg)
		}
		if len(strings.Fields(s.reason)) < 5 {
			t.Errorf("%s carries the reason %q, which is not one", s.pkg, s.reason)
		}
	}
}

// TestTheBarIsAboveTheGateThisOneIsMeasuredAgainst holds the argument in
// docs/gate-parity.md against the constant, so lowering the bar to the
// reference's number is a change that fails rather than one nobody notices.
func TestTheBarIsAboveTheGateThisOneIsMeasuredAgainst(t *testing.T) {
	const reference = 92.0
	if coverageBar <= reference {
		t.Errorf("the bar is %v and the reference gate's is %v; docs/gate-parity.md argues upward from it", coverageBar, reference)
	}
}

// TestNamingALegRunsThatLegAndSaysWhatItLeft is the other half of this change:
// a job on the server asks for one leg, and the run still discloses the rest.
func TestNamingALegRunsThatLegAndSaysWhatItLeft(t *testing.T) {
	asked, left, err := selectLegs(legs(), []string{"coverage"})
	if err != nil {
		t.Fatalf("asking for the coverage leg: %v", err)
	}
	if len(asked) != 1 || asked[0].name != "coverage" {
		t.Fatalf("the invocation ran %d leg(s), want the coverage leg alone", len(asked))
	}
	if len(left) != len(legs())-1 {
		t.Fatalf("it left %d leg(s) and there are %d others", len(left), len(legs())-1)
	}
}

// TestALegNameNothingAnswersToIsRefused is why the selection is not a filter. A
// job asking for a leg that has been renamed would otherwise pass having run
// nothing at all.
func TestALegNameNothingAnswersToIsRefused(t *testing.T) {
	_, _, err := selectLegs(legs(), []string{"coverge"})
	if err == nil {
		t.Fatal("a leg name nothing answers to was accepted")
	}
	if !strings.Contains(err.Error(), "coverage") {
		t.Errorf("the refusal is %q, and it does not list the legs there are", err)
	}
}

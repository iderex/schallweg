package main

// The coverage bar, and the surfaces it is pinned to.
//
// A bar over the whole tree is a number that moves when somebody adds a
// well-tested command, and it says nothing about the place where a gap costs
// something. What costs something here is the arithmetic whose output an
// engineer may put in front of a building acceptance. A defect there is not a
// crash and not a vulnerability. It is a number that looks right, that nothing
// will ever alert on, and that is read by somebody who has no way to tell it
// from a correct one.
//
// So the bar is pinned to those surfaces and the whole-tree figure is reported
// beside it without being gated. Both halves matter: a gate on the whole tree
// would let a gap in the arithmetic hide behind a well-covered command, and
// reporting nothing for the rest would leave the tree's own figure unread.

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// coverageBar is the statement coverage every surface below has to reach, as a
// percentage.
//
// What it protects is the arithmetic that decides a number somebody relies on.
// The number is chosen here rather than adopted: the gate this repository is
// measured against pins 92.0 on the modules that decide its security outcomes,
// and docs/gate-parity.md records that the argument here points upward, because
// a wrong number is not caught by anything downstream the way a failed login is.
//
// Upward, and not as far up as it will go. The surface is 282 statements today,
// so one percentage point is under three of them, and a bar set at the figure
// the tree happens to have reds on the next error branch somebody writes rather
// than on a gap. At 93.0 the surface may carry four uncovered statements, which
// is about one unexercised branch, and the fifth is what this refuses. The
// commands behind both numbers are in the pull request that set them.
const coverageBar = 93.0

// A decidingSurface is one package the bar applies to, and why it is on the
// list.
type decidingSurface struct {
	pkg    string
	reason string
}

// decidingSurfaces is the list, and it is the packages that exist rather than
// the packages that are planned.
//
// What puts a package here is that it computes a value the program's output is
// derived from. The rating procedures, the in situ correction, the path
// evaluation and the summation are all that kind of thing and all of them are
// still to be written; each joins this list in the change that adds it, and a
// package named here that the report does not measure is a failure rather than
// a line the check skips, so nothing can be added ahead of its arithmetic.
//
// Two packages that are deliberately not here. acoustic/approx decides whether
// a test passes rather than what a result is, and store reads and refuses
// numbers without computing any: its failures arrive as refusals somebody sees,
// which is the opposite of the failure this bar exists for. Both are inside the
// whole-tree figure the check reports.
func decidingSurfaces() []decidingSurface {
	return []decidingSurface{
		{
			pkg:    modulePath + "/acoustic",
			reason: "the band sets, the spectrum container and the decibel arithmetic every later quantity is built out of",
		},
	}
}

// A surfaceCoverage is what a report says about one package.
type surfaceCoverage struct {
	statements int
	covered    int
}

// percent is the statement coverage of a package, and it is only asked of a
// package that has statements.
func (c surfaceCoverage) percent() float64 {
	return 100 * float64(c.covered) / float64(c.statements)
}

// parseProfile reads a Go coverage profile into one entry per package.
//
// It refuses rather than reporting zero. An empty report, a report whose mode
// line is missing and a line it cannot read are all cases where the check
// cannot measure, and a check that cannot measure has to say so: passing there
// is how a gate turns into decoration on the day the report stops being written.
func parseProfile(text string) (map[string]surfaceCoverage, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, fmt.Errorf("the coverage report is empty")
	}
	if !strings.HasPrefix(lines[0], "mode:") {
		return nil, fmt.Errorf("the coverage report begins %q, and a report begins with its mode line", lines[0])
	}

	out := map[string]surfaceCoverage{}
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		pkg, block, err := profileLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d of the coverage report is %q: %w", i+2, line, err)
		}
		entry := out[pkg]
		entry.statements += block.statements
		entry.covered += block.covered
		out[pkg] = entry
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("the coverage report has a mode line and no measured statement")
	}
	return out, nil
}

// profileLine reads one block line: a file position, the number of statements in
// the block, and how often the block ran.
func profileLine(line string) (string, surfaceCoverage, error) {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return "", surfaceCoverage{}, fmt.Errorf("a block line has three fields and this has %d", len(parts))
	}
	file, _, found := strings.Cut(parts[0], ":")
	if !found {
		return "", surfaceCoverage{}, fmt.Errorf("the first field carries no file position")
	}
	slash := strings.LastIndex(file, "/")
	if slash < 0 {
		return "", surfaceCoverage{}, fmt.Errorf("%q names no package", file)
	}
	statements, err := strconv.Atoi(parts[1])
	if err != nil || statements < 0 {
		return "", surfaceCoverage{}, fmt.Errorf("the statement count is %q", parts[1])
	}
	count, err := strconv.Atoi(parts[2])
	if err != nil || count < 0 {
		return "", surfaceCoverage{}, fmt.Errorf("the execution count is %q", parts[2])
	}
	block := surfaceCoverage{statements: statements}
	if count > 0 {
		block.covered = statements
	}
	return file[:slash], block, nil
}

// judgeCoverage compares every deciding surface against the bar.
//
// A surface the report does not measure fails here, and that is the same
// refusal as one below the bar rather than a lesser one. A package with no
// statements produces no line in a report, so "this package has nothing to
// measure" and "this package was not run" arrive identically, and the safe
// reading of the pair is the one that refuses.
func judgeCoverage(report map[string]surfaceCoverage, surfaces []decidingSurface, bar float64) error {
	var refused []string
	for _, s := range surfaces {
		entry, measured := report[s.pkg]
		if !measured || entry.statements == 0 {
			refused = append(refused, fmt.Sprintf(
				"%s is on the bar's list and the report measures no statement of it; it is on the list because %s",
				s.pkg, s.reason))
			continue
		}
		if entry.percent() < bar {
			short := shortfall(entry, bar)
			refused = append(refused, fmt.Sprintf(
				"%s covers %d of %d statements, %.1f%%, and the bar is %.1f%%; %d more statement(s) have to be reached",
				s.pkg, entry.covered, entry.statements, entry.percent(), bar, short))
		}
	}
	if len(refused) == 0 {
		return nil
	}
	sort.Strings(refused)
	return fmt.Errorf("%s", strings.Join(refused, "\n"))
}

// shortfall is how many more statements a surface has to reach to stand at the
// bar, so the message says what to do rather than only how far off it is.
func shortfall(c surfaceCoverage, bar float64) int {
	need := bar / 100 * float64(c.statements)
	for n := c.covered; n <= c.statements; n++ {
		if float64(n) >= need {
			return n - c.covered
		}
	}
	return c.statements - c.covered
}

// wholeTree is the figure over everything the report measured. It is printed and
// never gated: it moves when a command grows and says nothing about the
// arithmetic, which is what the bar above is for.
func wholeTree(report map[string]surfaceCoverage) surfaceCoverage {
	var total surfaceCoverage
	for _, entry := range report {
		total.statements += entry.statements
		total.covered += entry.covered
	}
	return total
}

// checkCoverage is the leg: it produces a report, refuses one it cannot read,
// judges the deciding surfaces against the bar, and prints the whole-tree figure
// beside the verdict.
func checkCoverage(root string, out io.Writer) error {
	file, err := os.CreateTemp("", "schallweg-gate-coverage-*.out")
	if err != nil {
		return err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		return err
	}
	defer os.Remove(path)

	if _, err := output(root, "go", "test", "./...", "-count=1", "-covermode=count", "-coverprofile="+path); err != nil {
		return err
	}
	text, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("the coverage report at %s could not be read: %w", path, err)
	}
	report, err := parseProfile(string(text))
	if err != nil {
		return err
	}

	total := wholeTree(report)
	fmt.Fprintf(out, "           whole tree, reported and not gated: %d of %d statements, %.1f%%\n",
		total.covered, total.statements, total.percent())
	for _, s := range decidingSurfaces() {
		if entry, measured := report[s.pkg]; measured && entry.statements > 0 {
			fmt.Fprintf(out, "           %s: %d of %d statements, %.1f%%, bar %.1f%%\n",
				s.pkg, entry.covered, entry.statements, entry.percent(), coverageBar)
		}
	}
	return judgeCoverage(report, decidingSurfaces(), coverageBar)
}

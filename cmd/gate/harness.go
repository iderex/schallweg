package main

// The harness audit: what the ordinary suite is, what it is not, and what a run
// of it left alone.
//
// docs/decisions/testability.md says the ordinary suite runs with no display, no
// elevation, no network and no hardware beyond a general purpose computer, and
// that a test which cannot meet all four goes into a harness named for the thing
// it needs, behind a build constraint with the same name, under harness/. Until
// this leg existed, every part of that was held by two documents and by nobody
// checking.
//
// The reporting half is the reason this is a leg rather than a comment. A run
// that covered less than everything must not read like one that covered
// everything and found nothing, so the harnesses are read out of the tree and
// named on every run, together with what each of them would need. A list written
// into a workflow answers the same question and goes stale the first time a
// harness is added, which is the one moment the answer changes.

import (
	"fmt"
	"go/build/constraint"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// harnessRoot is where a harness lives. A harness anywhere else is a test the
// ordinary run builds, which is the thing the constraint exists to prevent.
const harnessRoot = "harness"

// harnessStatement is the file a harness carries beside it, saying exactly what
// has to be present for it to run and how a contributor would know they have it.
const harnessStatement = "README.md"

// declaredHarnesses are the five names docs/decisions/testability.md fixes, and
// the convention behind them is that a name states the requirement rather than
// the intent.
//
// A sixth name is not a typo to be corrected here. It is a category of
// prerequisite nobody has decided this project has, and the decision record is
// where that is settled, so an undeclared name is refused rather than added.
func declaredHarnesses() []string {
	return []string{
		"requires-display",
		"requires-instrument",
		"requires-third-party-tool",
		"requires-network",
		"requires-elevation",
	}
}

// A harness is one of them as the tree actually holds it.
type harness struct {
	name  string
	files []string
	needs string
}

// constraintTag is the build tag a harness of this name lives behind. A build
// constraint is a Go identifier and a harness name is written with hyphens, so
// the two differ by that substitution and by nothing else.
func constraintTag(name string) string { return strings.ReplaceAll(name, "-", "_") }

// constraintTags reads the tags in a Go file's build constraint.
//
// It reads only the lines before the package clause, which is where the
// toolchain reads one, so a line further down that looks like a constraint is
// reported as absent here exactly as the compiler treats it. That is the shape
// of the mistake worth catching: a constraint written in the right words in the
// wrong place builds into the ordinary suite and reads, in a diff, as a
// constraint.
func constraintTags(src string) []string {
	var tags []string
	for _, line := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") {
			break
		}
		expr, err := constraint.Parse(trimmed)
		if err != nil {
			continue
		}
		expr.Eval(func(tag string) bool {
			tags = append(tags, tag)
			return false
		})
	}
	sort.Strings(tags)
	return tags
}

// judgeHarnesses is the whole rule, over a tree written down as three facts: the
// build constraint tags of every Go file, which directories under harness/ exist,
// and which of those carry their statement.
//
// Taking the facts rather than the disk is what lets every refusal below be
// proven against a tree that never existed. A rule judged only against this
// repository is a rule that goes green the day somebody deletes the harness
// directory.
func judgeHarnesses(tags map[string][]string, dirs []string, hasStatement map[string]bool) ([]harness, []string) {
	declared := map[string]bool{}
	for _, n := range declaredHarnesses() {
		declared[n] = true
	}

	var refusals []string
	found := make([]harness, 0, len(dirs))
	for _, name := range dirs {
		h := harness{name: name}
		if !declared[name] {
			refusals = append(refusals, fmt.Sprintf(
				"%s/%s is a harness whose name does not say what it requires; the declared names are %s, and a sixth is a decision in docs/decisions/testability.md rather than a directory",
				harnessRoot, name, strings.Join(declaredHarnesses(), ", ")))
		}
		if !hasStatement[name] {
			refusals = append(refusals, fmt.Sprintf(
				"%s/%s carries no %s saying what has to be present for it to run and how a contributor would know they have it",
				harnessRoot, name, harnessStatement))
		}
		want := constraintTag(name)
		prefix := harnessRoot + "/" + name + "/"
		for path, fileTags := range tags {
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			h.files = append(h.files, path)
			if !contains(fileTags, want) {
				refusals = append(refusals, fmt.Sprintf(
					"%s is in a harness and is not behind the %q build constraint, so the ordinary run builds it",
					path, want))
			}
		}
		sort.Strings(h.files)
		found = append(found, h)
	}

	// The other direction, and it is the one a reader is least likely to think
	// of. A file outside harness/ behind a harness constraint is a harness the
	// harness tree does not know about: the ordinary run does not build it, the
	// report above does not name it, and nothing says it was skipped.
	for path, fileTags := range tags {
		if strings.HasPrefix(path, harnessRoot+"/") {
			continue
		}
		for _, tag := range fileTags {
			if strings.HasPrefix(tag, "requires") {
				refusals = append(refusals, fmt.Sprintf(
					"%s is behind the %q build constraint and is not under %s/, so it is skipped by every run and named by none",
					path, tag, harnessRoot))
			}
		}
	}

	sort.Strings(refusals)
	sort.Slice(found, func(i, j int) bool { return found[i].name < found[j].name })
	return found, refusals
}

// contains reports whether a tag is among a file's tags.
func contains(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

// scanForHarnesses reads the three facts out of a checkout.
func scanForHarnesses(root string) (map[string][]string, []string, map[string]bool, error) {
	tags := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel != "." && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".")) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		// Every Go file is recorded, including one with no constraint at all.
		// That is the shape the harness rule below is really for: a file in the
		// harness tree behind nothing is built by the ordinary run, and a scan
		// that only recorded constrained files would never see it.
		tags[rel] = constraintTags(string(src))
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	var dirs []string
	statement := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(root, harnessRoot))
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirs = append(dirs, e.Name())
		if _, err := os.Stat(filepath.Join(root, harnessRoot, e.Name(), harnessStatement)); err == nil {
			statement[e.Name()] = true
		}
	}
	sort.Strings(dirs)
	return tags, dirs, statement, nil
}

// testTargets is every package the build says carries tests.
//
// It is asked of the toolchain rather than written down, because a list in a
// workflow answers this question correctly on the day it is written and wrongly
// on the day a package is added, and nothing marks the difference.
func testTargets(root string) ([]string, error) {
	out, err := output(root, "go", "list", "-f", "{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}", "./...")
	if err != nil {
		return nil, err
	}
	var targets []string
	for _, line := range strings.Fields(out) {
		targets = append(targets, line)
	}
	sort.Strings(targets)
	if len(targets) == 0 {
		return nil, fmt.Errorf("the build reports no package carrying a test, which is not a tree this check can read as clean")
	}
	return targets, nil
}

// needsOf reads the first sentence of a harness's statement, so the run says
// what running it would need instead of only that it exists.
func needsOf(root, name string) string {
	b, err := os.ReadFile(filepath.Join(root, harnessRoot, name, harnessStatement))
	if err != nil {
		return "nothing is written beside it"
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return "nothing is written beside it"
}

// checkHarnesses is the leg. It enumerates the test targets from the build
// rather than from a list, names every harness the tree holds together with what
// running it would need, and refuses the shapes above.
func checkHarnesses(root string, out io.Writer) error {
	targets, err := testTargets(root)
	if err != nil {
		return err
	}
	tags, dirs, statement, err := scanForHarnesses(root)
	if err != nil {
		return err
	}
	found, refusals := judgeHarnesses(tags, dirs, statement)
	for i := range found {
		found[i].needs = needsOf(root, found[i].name)
	}
	reportHarnesses(out, targets, found)
	if len(refusals) > 0 {
		return fmt.Errorf("%s", strings.Join(refusals, "\n"))
	}
	return nil
}

// reportHarnesses writes the enumeration and the disclosure.
func reportHarnesses(out io.Writer, targets []string, found []harness) {
	fmt.Fprintf(out, "           %d package(s) carry tests, enumerated from the build:\n", len(targets))
	for _, t := range targets {
		fmt.Fprintf(out, "             %s\n", t)
	}
	if len(found) == 0 {
		fmt.Fprintf(out, "           No harness exists under %s/, so this run skipped nothing.\n", harnessRoot)
		return
	}
	fmt.Fprintf(out, "           %d harness(es) exist and none of them was run here:\n", len(found))
	for _, h := range found {
		fmt.Fprintf(out, "             %s, %d file(s); running it needs: %s\n", h.name, len(h.files), h.needs)
	}
	fmt.Fprintf(out, "           A harness that was not run is not a pass, and these counts are not summed.\n")
}

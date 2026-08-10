package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixtureDocument reads a Markdown fixture.
//
// The fixtures carry a `.md.txt` suffix rather than `.md`, and that is load
// bearing rather than tidiness. The reader under test walks every tracked
// document ending in `.md`, so a fixture named that way would be read by the
// real run, and the fixture that exists to break the rule would break it for the
// repository.
func fixtureDocument(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "references", name))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	return b
}

// dangling resolves a fixture's references against the repository as it stands
// and returns the ones that point at nothing.
//
// The fixture is resolved as though it lived at docPath, so that a relative
// target means what it would mean for a real document in that directory.
func dangling(t *testing.T, docPath, fixture string) []reference {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}
	known, err := trackedPaths(root)
	if err != nil {
		t.Fatalf("listing the tracked files: %v", err)
	}

	var out []reference
	for _, ref := range referencesIn(docPath, fixtureDocument(t, fixture)) {
		resolved := false
		for _, target := range ref.Targets {
			if known[target] {
				resolved = true
				break
			}
		}
		if !resolved {
			out = append(out, ref)
		}
	}
	return out
}

// TestTheReaderFindsEveryDanglingReference is the proof that the rule bites. The
// counts are exact rather than a lower bound, because a reader that reported one
// reference twice would still pass an assertion that it found at least four, and
// a duplicate finding is how a check gets switched off.
func TestTheReaderFindsEveryDanglingReference(t *testing.T) {
	found := dangling(t, "docs/breaks-the-rule.md", "breaks-the-rule.md.txt")

	byForm := map[string]int{}
	for _, ref := range found {
		byForm[ref.Form]++
	}

	if len(found) != 4 {
		t.Errorf("the reader found %d dangling reference(s), want 4: %v", len(found), found)
	}
	if byForm["link"] != 2 {
		t.Errorf("the reader found %d dangling link(s), want 2", byForm["link"])
	}
	if byForm["path"] != 2 {
		t.Errorf("the reader found %d dangling backticked path(s), want 2", byForm["path"])
	}
}

// TestTheReaderPassesTheNearMiss is the half that decides whether anybody can
// live with the rule. Every reference in that fixture is one character, one
// directory or one pair of backticks away from one in the fixture above, and all
// of them are legal.
func TestTheReaderPassesTheNearMiss(t *testing.T) {
	if found := dangling(t, "docs/the-near-miss.md", "the-near-miss.md.txt"); len(found) > 0 {
		for _, ref := range found {
			t.Errorf("the near miss was refused at line %d: the %s %q was read as pointing at %s",
				ref.Line, ref.Form, ref.Written, strings.Join(ref.Targets, " or "))
		}
	}
}

// TestTheNearMissIsNotEmpty is the count rather than the presence. The test
// above passes trivially against a reader that finds no reference at all, and a
// reader that stopped reading Markdown entirely would look exactly like a tree
// whose every reference resolves.
func TestTheNearMissIsNotEmpty(t *testing.T) {
	refs := referencesIn("docs/the-near-miss.md", fixtureDocument(t, "the-near-miss.md.txt"))

	byForm := map[string]int{}
	for _, ref := range refs {
		byForm[ref.Form]++
	}
	if byForm["link"] == 0 {
		t.Error("the near miss yielded no link, so the passing test above proves nothing about links")
	}
	if byForm["path"] == 0 {
		t.Error("the near miss yielded no backticked path, so the passing test above proves nothing about them")
	}
}

// TestAFencedBlockIsNotRead states the exemption as a test rather than leaving it
// to the near miss, because it is the exemption most likely to be removed by
// somebody widening the reader. Every command in these documents is quoted with
// its output, and that output names paths that were true when it ran.
func TestAFencedBlockIsNotRead(t *testing.T) {
	src := []byte(strings.Join([]string{
		"Before.",
		"```",
		"$ cat docs/never-existed.md",
		"[a link](docs/never-existed.md)",
		"```",
		"After: [a real one](decisions/layering.md).",
	}, "\n"))

	refs := referencesIn("docs/example.md", src)
	if len(refs) != 1 {
		t.Fatalf("the reader found %d reference(s), want the 1 outside the fence: %v", len(refs), refs)
	}
	if refs[0].Line != 6 {
		t.Errorf("the reference was found at line %d, want line 6", refs[0].Line)
	}
}

// TestThisRepositoryHasNoDanglingReference is the rule reaching this tree. The
// reader is run over the repository itself, which is how the rule arrives here
// without anything having to be added to a list.
func TestThisRepositoryHasNoDanglingReference(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}
	if err := checkReferences(root, io.Discard); err != nil {
		t.Errorf("this repository has a document pointing at something it does not hold:\n%v", err)
	}
}

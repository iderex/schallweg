package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The artefact sets these tests build are written into a temporary directory
// rather than committed under testdata, and the reason is what the cases are
// about. What each case turns on is one name in a set of ten, and a committed
// directory of ten near-identical placeholder files per case hides exactly the
// name a reader is trying to find. The names are written literally here, where
// the case is, and the contents are fixed strings rather than anything generated.

// setFor writes a complete artefact set for a version and returns the directory.
func setFor(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range expectedNames(version) {
		writeArtefact(t, dir, name)
	}
	return dir
}

// writeArtefact writes one file whose bytes are its own name, so that no two
// artefacts in a set hash to the same value and a digest attached to the wrong
// name is visible.
func writeArtefact(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("contents of "+name), 0o644); err != nil {
		t.Fatalf("cannot write the artefact %s: %v", name, err)
	}
}

func TestParseTagAcceptsAReleaseTag(t *testing.T) {
	got, err := parseTag("v1.2.3")
	if err != nil {
		t.Fatalf("parseTag(v1.2.3) refused a release tag: %v", err)
	}
	if got != "1.2.3" {
		t.Errorf("parseTag(v1.2.3) = %q, want 1.2.3", got)
	}
}

func TestParseTagRefusesTheNearMisses(t *testing.T) {
	// Every entry here is a tag somebody meant to be the real one. The two-part
	// tag is the shape a person types when they are thinking of a series rather
	// than a release, the leading zero is what a date-shaped habit produces, and
	// the suffix cases are the ones a tool appends.
	cases := map[string]string{
		"no leading v":     "1.2.3",
		"two parts":        "v1.2",
		"four parts":       "v1.2.3.4",
		"leading zero":     "v1.02.3",
		"empty part":       "v1..3",
		"a pre-release":    "v1.2.3-rc1",
		"build metadata":   "v1.2.3+build",
		"a word":           "vlatest",
		"nothing after v":  "v",
		"an empty tag":     "",
		"a branch name":    "main",
		"space in the tag": "v1.2.3 ",
	}
	for name, tag := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := parseTag(tag); err == nil {
				t.Fatalf("parseTag(%q) returned %q and no error, and this tag is not a release of this project", tag, got)
			}
		})
	}
}

func TestVersionFromLineReadsWhatTheProgramPrints(t *testing.T) {
	got, err := versionFromLine("schallweg 0.0.0\n")
	if err != nil {
		t.Fatalf("versionFromLine refused the line this program prints today: %v", err)
	}
	if got != "0.0.0" {
		t.Errorf("versionFromLine = %q, want 0.0.0", got)
	}
}

func TestVersionFromLineRefusesALineThatIsNotThisProgram(t *testing.T) {
	cases := map[string]string{
		"another program": "schallweg-cli 1.0.0",
		"a usage line":    "usage: schallweg [flags]",
		"nothing at all":  "",
		"only the name":   "schallweg",
		"an error":        "schallweg: cannot write to standard output",
		"a bad version":   "schallweg 1.0",
	}
	for name, line := range cases {
		t.Run(name, func(t *testing.T) {
			if got, err := versionFromLine(line); err == nil {
				t.Fatalf("versionFromLine(%q) returned %q and no error", line, got)
			}
		})
	}
}

// TestATagThatDoesNotNameWhatWasBuiltIsRefused is the near-miss this command
// exists for. Both directions are here because they come from opposite
// mistakes: tagging before the constant was raised, and raising the constant
// then tagging the commit before it.
func TestATagThatDoesNotNameWhatWasBuiltIsRefused(t *testing.T) {
	if _, err := checkTagNamesWhatWasBuilt("v0.1.0", "schallweg 0.0.0"); err == nil {
		t.Fatal("a tag ahead of the program's own version was accepted")
	}
	if _, err := checkTagNamesWhatWasBuilt("v0.0.0", "schallweg 0.1.0"); err == nil {
		t.Fatal("a tag behind the program's own version was accepted")
	}
	version, err := checkTagNamesWhatWasBuilt("v0.1.0", "schallweg 0.1.0")
	if err != nil {
		t.Fatalf("a tag naming the version that was built was refused: %v", err)
	}
	if version != "0.1.0" {
		t.Errorf("checkTagNamesWhatWasBuilt = %q, want 0.1.0", version)
	}
}

func TestACompleteSetAssembles(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)

	entries, err := assemble(dir, version)
	if err != nil {
		t.Fatalf("a complete artefact set was refused: %v", err)
	}
	if len(entries) != len(expectedNames(version)) {
		t.Fatalf("assemble returned %d entries for a set of %d artefacts", len(entries), len(expectedNames(version)))
	}

	// Every platform appears, and each digest belongs to the file it is written
	// beside. Reading the file here rather than trusting the command is the
	// point: it catches a digest attached to the wrong name, which is the way a
	// checksum file lies while looking correct.
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.Name] = true
		raw, err := os.ReadFile(filepath.Join(dir, e.Name))
		if err != nil {
			t.Fatalf("cannot read %s back: %v", e.Name, err)
		}
		sum := sha256.Sum256(raw)
		if want := hex.EncodeToString(sum[:]); e.Digest != want {
			t.Errorf("%s carries digest %s, and its bytes hash to %s", e.Name, e.Digest, want)
		}
	}
	for _, p := range platforms() {
		if !seen[p.binaryName(version)] {
			t.Errorf("the set has no binary for %s/%s", p.OS, p.Arch)
		}
		if !seen[p.documentName(version)] {
			t.Errorf("the set has no bill of materials for %s/%s", p.OS, p.Arch)
		}
	}
}

func TestASetMissingOnePlatformIsRefused(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)

	// One platform's binary, removed. This is what a build matrix that lost a
	// leg produces, and every other artefact in the set is correct.
	gone := platforms()[0].binaryName(version)
	if err := os.Remove(filepath.Join(dir, gone)); err != nil {
		t.Fatalf("cannot remove %s: %v", gone, err)
	}

	err := assembleErr(dir, version)
	if err == nil {
		t.Fatal("a set missing a platform's binary was accepted")
	}
	// The refusal has to be the one that counted the set, not the one that
	// tried to read a file and could not. Both stop the release; only the first
	// says a platform is missing, and only the first fires before anything in
	// the set has been hashed.
	if !strings.Contains(err.Error(), "is missing") {
		t.Errorf("the set was refused for some other reason than being incomplete: %v", err)
	}
	if !strings.Contains(err.Error(), gone) {
		t.Errorf("the refusal does not name the missing artefact %s: %v", gone, err)
	}
}

func TestASetMissingABillOfMaterialsIsRefused(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)

	gone := platforms()[1].documentName(version)
	if err := os.Remove(filepath.Join(dir, gone)); err != nil {
		t.Fatalf("cannot remove %s: %v", gone, err)
	}

	err := assembleErr(dir, version)
	if err == nil {
		t.Fatal("a set missing a bill of materials was accepted")
	}
	if !strings.Contains(err.Error(), "is missing") {
		t.Errorf("the set was refused for some other reason than being incomplete: %v", err)
	}
	if !strings.Contains(err.Error(), gone) {
		t.Errorf("the refusal does not name the missing document %s: %v", gone, err)
	}
}

// TestASetCarryingAFileItDoesNotNameIsRefused is the direction that is easy to
// leave out. A release that ships whatever it finds gives a checksum, and later
// a signature, to a file nobody planned for.
func TestASetCarryingAFileItDoesNotNameIsRefused(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)
	const stray = "build-notes.txt"
	writeArtefact(t, dir, stray)

	err := assembleErr(dir, version)
	if err == nil {
		t.Fatal("a set carrying an unnamed file was accepted")
	}
	if !strings.Contains(err.Error(), stray) {
		t.Errorf("the refusal does not name the stray file: %v", err)
	}
}

// TestASetBuiltForAnotherVersionIsRefused covers the case where every artefact
// is present and correct and the version in every name is the previous one,
// which is what a cached build directory produces.
func TestASetBuiltForAnotherVersionIsRefused(t *testing.T) {
	dir := setFor(t, "0.1.0")
	if err := assembleErr(dir, "0.2.0"); err == nil {
		t.Fatal("a set built for another version was accepted")
	}
}

// TestASecondRunDoesNotOverwriteTheFirstSetOfChecksums is the rule above applied
// to the one unnamed file this command itself can leave behind. It is refused by
// that rule and not by one of its own: a separate check for it existed, was
// deleted, and the suite stayed green.
func TestASecondRunDoesNotOverwriteTheFirstSetOfChecksums(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)
	if err := os.WriteFile(filepath.Join(dir, checksumFile), []byte("written by an earlier run\n"), 0o644); err != nil {
		t.Fatalf("cannot write %s: %v", checksumFile, err)
	}

	err := assembleErr(dir, version)
	if err == nil {
		t.Fatalf("a directory already holding %s was accepted, so a second run would overwrite it", checksumFile)
	}
	if !strings.Contains(err.Error(), checksumFile) {
		t.Errorf("the refusal does not name %s: %v", checksumFile, err)
	}
}

// assembleErr is assemble with the entries dropped, for the cases that are only
// about the refusal.
func assembleErr(dir, version string) error {
	_, err := assemble(dir, version)
	return err
}

func TestChecksumFileIsReadableByTheOrdinaryTools(t *testing.T) {
	entries := []entry{
		{Name: "b", Digest: strings.Repeat("a", 64)},
		{Name: "a", Digest: strings.Repeat("b", 64)},
	}
	got := checksumLines(entries)
	want := strings.Repeat("a", 64) + "  b\n" + strings.Repeat("b", 64) + "  a\n"
	if got != want {
		t.Errorf("checksumLines wrote\n%q\nand sha256sum -c reads\n%q", got, want)
	}
}

func TestPlatformsAreDistinctAndNamed(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range platforms() {
		if p.OS == "" || p.Arch == "" {
			t.Fatalf("a platform is missing part of its name: %+v", p)
		}
		key := p.String()
		if seen[key] {
			t.Errorf("%s appears twice in the platform list, so the release would build it twice and name it once", key)
		}
		seen[key] = true
	}
	if len(seen) == 0 {
		t.Fatal("the platform list is empty, so a release would carry no program at all")
	}
}

// TestRunPrintsThePlatformsTheWorkflowLoopsOver holds the interface between this
// command and the workflow. The workflow builds one binary per line of this
// output, so a change to the shape of a line is a change to the release route.
func TestRunPrintsThePlatformsTheWorkflowLoopsOver(t *testing.T) {
	var out bytes.Buffer
	if err := run(&out, true, "", "", ""); err != nil {
		t.Fatalf("-platforms failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != len(platforms()) {
		t.Fatalf("-platforms printed %d line(s) for %d platform(s)", len(lines), len(platforms()))
	}
	for i, line := range lines {
		want := platforms()[i].String()
		if line != want {
			t.Errorf("line %d is %q, want %q", i+1, line, want)
		}
	}
}

func TestRunAssemblesAndSaysItPublishedNothing(t *testing.T) {
	const version = "0.1.0"
	dir := setFor(t, version)

	var out bytes.Buffer
	if err := run(&out, false, "v"+version, "schallweg "+version, dir); err != nil {
		t.Fatalf("a complete run failed: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, checksumFile))
	if err != nil {
		t.Fatalf("the run wrote no %s: %v", checksumFile, err)
	}

	var named []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("the checksum line %q is not a digest and a name", line)
		}
		named = append(named, fields[1])
	}
	sort.Strings(named)
	want := expectedNames(version)
	if strings.Join(named, " ") != strings.Join(want, " ") {
		t.Errorf("%s names\n%v\nand the set is\n%v", checksumFile, named, want)
	}

	if !strings.Contains(out.String(), "Nothing was published") {
		t.Errorf("the run did not say that it published nothing:\n%s", out.String())
	}
}

func TestRunRefusesWithoutTheArgumentsItDecidesFrom(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]struct{ tag, line, dir string }{
		"no tag":          {"", "schallweg 0.1.0", dir},
		"no version line": {"v0.1.0", "", dir},
		"no directory":    {"v0.1.0", "schallweg 0.1.0", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			if err := run(&out, false, c.tag, c.line, c.dir); err == nil {
				t.Fatal("the run went ahead without an argument it decides from")
			}
		})
	}
}

// Command release assembles the artefact set for a tagged release and refuses a
// set that does not match the version being released. It publishes nothing.
//
// The route it belongs to is .github/workflows/release.yml, and the division
// between the two is deliberate. The workflow does what only a workflow can do:
// check out the tag, install a toolchain, run the gate, run the compiler once
// per platform, and hand the result to an upload. Everything that decides
// whether the result is a release is here, in a package the ordinary suite runs,
// because a rule written in a workflow step is a rule nothing can prove bites
// until somebody pushes a tag and watches it fail.
//
// What it decides, in order: that the tag names a version, that the program in
// the set says it is that version, that the set holds exactly the artefacts the
// platform list calls for and nothing else, and what each of them hashes to.
//
// The platform list lives here rather than in the workflow for the same reason.
// A list in the workflow and a list in this command would be two lists, and the
// day they disagree the release is short one binary and every check passes. The
// workflow asks for the list with -platforms and loops over the answer.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// programName is the name of the program this route releases. It is the name of
// its directory under cmd/ and the first word of the line the program prints.
const programName = "schallweg"

// checksumFile is the name of the file the assembly writes into the artefact
// directory. It is the only file this command creates.
const checksumFile = "checksums.txt"

// platform is one operating system and architecture pair the release is built
// for.
type platform struct {
	OS   string
	Arch string
}

// platforms is what a release contains, and the reason each entry is in it.
//
// windows/amd64 is first because it is what the audience this project is for
// actually sits at: an engineering office computing sound insulation on a
// desktop running Windows. linux/amd64 is what a reviewer, a package maintainer
// and every automated consumer will reach for. The two darwin entries are one
// audience split by a hardware transition that is not finished, and shipping
// only the newer one turns an older machine into a support question. linux/arm64
// is the cheapest entry here, because it is the same compiler invocation as the
// others, and leaving it out is what forces somebody on a small server to build
// from source.
//
// No entry is here for a platform this project cannot honestly say it built.
// Cross-compilation produces the binary; it does not run it, and nothing in this
// route claims a binary was executed on the machine it is for. That distinction
// belongs in the release notes rather than in a longer list here.
func platforms() []platform {
	return []platform{
		{OS: "windows", Arch: "amd64"},
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
	}
}

// String is the GOOS/GOARCH form, which is how the workflow names a target and
// how this command prints the list it loops over.
func (p platform) String() string { return p.OS + "/" + p.Arch }

// binaryName is what the compiler writes for one platform of one version.
//
// The version is in the file name because an artefact that has been downloaded
// and put in a directory beside another one has lost every other statement of
// which release it came from.
func (p platform) binaryName(version string) string {
	name := fmt.Sprintf("%s_%s_%s_%s", programName, version, p.OS, p.Arch)
	if p.OS == "windows" {
		name += ".exe"
	}
	return name
}

// documentName is the bill of materials that travels with one binary.
//
// One per binary rather than one per release, because the document answers what
// is inside a particular artefact, and a user who was handed one file cannot use
// a document that describes a set.
func (p platform) documentName(version string) string {
	return p.binaryName(version) + ".spdx.json"
}

// expectedNames is the complete artefact set for a version, sorted.
func expectedNames(version string) []string {
	var names []string
	for _, p := range platforms() {
		names = append(names, p.binaryName(version), p.documentName(version))
	}
	sort.Strings(names)
	return names
}

// parseTag returns the version a release tag names.
//
// The accepted shape is v followed by three non-negative decimal numbers. It is
// narrow on purpose: a tag is the one input to this route that a person types by
// hand at the moment they are least likely to be checking, and every shape this
// refuses is a shape somebody meant to be the real one.
func parseTag(tag string) (string, error) {
	if !strings.HasPrefix(tag, "v") {
		return "", fmt.Errorf("the tag %q does not begin with v, and a release tag of this project is v followed by its version", tag)
	}
	version := strings.TrimPrefix(tag, "v")
	if err := checkVersion(version); err != nil {
		return "", fmt.Errorf("the tag %q does not name a version: %w", tag, err)
	}
	return version, nil
}

// checkVersion reports why a string is not a version of this project.
func checkVersion(version string) error {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("%q has %d part(s) separated by a dot, and a version of this project has three", version, len(parts))
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("%q has an empty part", version)
		}
		if len(part) > 1 && part[0] == '0' {
			return fmt.Errorf("%q has the part %q with a leading zero, which two readers order differently", version, part)
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return fmt.Errorf("%q has the part %q, which is not a decimal number", version, part)
			}
		}
		if _, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("%q has the part %q, which is not a number this route can read: %w", version, part, err)
		}
	}
	return nil
}

// versionFromLine returns the version in the line the built program printed.
//
// The line is the program's own statement of what it is, taken from the artefact
// rather than from the source it was built from, because those are the same only
// when nothing went wrong and this is the check that says so.
func versionFromLine(line string) (string, error) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) != 2 {
		return "", fmt.Errorf("the program printed %q, and this route reads a line of exactly two words", strings.TrimSpace(line))
	}
	if fields[0] != programName {
		return "", fmt.Errorf("the program printed %q as its name, and this route releases %q", fields[0], programName)
	}
	if err := checkVersion(fields[1]); err != nil {
		return "", fmt.Errorf("the program printed a version this route cannot read: %w", err)
	}
	return fields[1], nil
}

// checkTagNamesWhatWasBuilt refuses a tag that does not name the version the
// built program says it is.
//
// This is the refusal the whole command exists for. The version is a constant in
// the source and the tag is typed by a person, so the two disagree exactly when
// somebody tagged before changing the constant or changed the constant and
// tagged the commit before it. Both produce a complete, signable, publishable
// artefact set that is wrong about what it is, and nothing downstream of here
// can tell.
func checkTagNamesWhatWasBuilt(tag, versionLine string) (string, error) {
	tagged, err := parseTag(tag)
	if err != nil {
		return "", err
	}
	built, err := versionFromLine(versionLine)
	if err != nil {
		return "", err
	}
	if tagged != built {
		return "", fmt.Errorf("the tag %q names version %s and the program in this set says it is version %s", tag, tagged, built)
	}
	return tagged, nil
}

// entry is one artefact and what it hashes to.
type entry struct {
	Name   string
	Digest string
}

// assemble reads the artefact directory, refuses a set that is not the one the
// version calls for, and returns each artefact with its digest.
//
// Both directions are refused. A missing artefact is a release that is short a
// platform, which a user discovers by not finding their own. An unexpected file
// is the more interesting one: it is a build that wrote something nobody
// planned for, and a release that ships whatever it finds is how a stray file
// from a working tree gets a checksum and a signature.
//
// The second direction is also what stops a second run overwriting the checksums
// of a first. There is no separate check for that, and there was one until it
// was deleted and the suite stayed green: the checksum file is never a name the
// version calls for, so the rule that refuses an unnamed file already refuses it,
// and a guard whose removal changes no verdict is one nobody can prove bites.
func assemble(dir, version string) ([]entry, error) {
	found, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read the artefact directory: %w", err)
	}

	present := map[string]bool{}
	for _, f := range found {
		if f.IsDir() {
			return nil, fmt.Errorf("the artefact directory holds the directory %q, and a release set is files only", f.Name())
		}
		present[f.Name()] = true
	}

	var missing []string
	expected := expectedNames(version)
	for _, name := range expected {
		if !present[name] {
			missing = append(missing, name)
		}
		delete(present, name)
	}
	var unexpected []string
	for name := range present {
		unexpected = append(unexpected, name)
	}
	sort.Strings(unexpected)

	if len(missing) > 0 {
		return nil, fmt.Errorf("the artefact set for version %s is missing %d of %d artefact(s): %s", version, len(missing), len(expected), strings.Join(missing, ", "))
	}
	if len(unexpected) > 0 {
		return nil, fmt.Errorf("the artefact set for version %s holds %d file(s) it does not name: %s", version, len(unexpected), strings.Join(unexpected, ", "))
	}

	entries := make([]entry, 0, len(expected))
	for _, name := range expected {
		digest, err := digestOf(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry{Name: name, Digest: digest})
	}
	return entries, nil
}

// digestOf is the SHA-256 of one file, lower case hexadecimal.
func digestOf(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot read an artefact: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("cannot read the artefact %s: %w", filepath.Base(path), err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// checksumLines is the content of the checksum file.
//
// The layout is the one the ordinary command line tools read, so that a user
// verifying a download does not need anything this project wrote. Two spaces
// between the digest and the name is what sha256sum -c expects and is not a
// formatting preference.
func checksumLines(entries []entry) string {
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%s  %s\n", e.Digest, e.Name)
	}
	return b.String()
}

func main() {
	var (
		list        = flag.Bool("platforms", false, "print the platforms a release is built for, one GOOS/GOARCH per line, and exit")
		tag         = flag.String("tag", "", "the tag being released, for example v0.1.0")
		versionLine = flag.String("version-line", "", "the line the built program printed when it was run")
		dir         = flag.String("dir", "", "the directory holding the artefacts")
	)
	flag.Parse()

	if err := run(os.Stdout, *list, *tag, *versionLine, *dir); err != nil {
		fmt.Fprintf(os.Stderr, "release: %v\n", err)
		os.Exit(1)
	}
}

// run is main with its output and its arguments handed to it, so that the
// decisions above are reachable from a test without starting a process.
func run(out io.Writer, list bool, tag, versionLine, dir string) error {
	if list {
		for _, p := range platforms() {
			if _, err := fmt.Fprintf(out, "%s/%s\n", p.OS, p.Arch); err != nil {
				return err
			}
		}
		return nil
	}

	switch {
	case tag == "":
		return errors.New("no -tag was given, and a release is assembled for a named tag")
	case versionLine == "":
		return errors.New("no -version-line was given, and the set is checked against what the built program says it is")
	case dir == "":
		return errors.New("no -dir was given, and there is nothing to assemble")
	}

	version, err := checkTagNamesWhatWasBuilt(tag, versionLine)
	if err != nil {
		return err
	}
	entries, err := assemble(dir, version)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, checksumFile)
	if err := os.WriteFile(path, []byte(checksumLines(entries)), 0o644); err != nil {
		return fmt.Errorf("cannot write %s: %w", checksumFile, err)
	}

	if _, err := fmt.Fprintf(out, "Release %s, assembled from the program the set carries, which says it is version %s.\n\n", tag, version); err != nil {
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(out, "  %s  %s\n", e.Digest, e.Name); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(out, "\n%d artefact(s) and %s, in %s.\n", len(entries), checksumFile, dir); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Nothing was published. This route assembles and stops there.")
	return err
}

package main

// The rule that every file a document points at exists, and the reader that
// finds a document pointing at one that does not.
//
// Documentation rots in a particular way in a repository like this one. A
// document names another document, a decision record, a schema or a source file,
// then that file is renamed or moved, and the reference is left behind. Nothing
// notices, because a reference is prose to every tool that reads this tree and
// the document still renders. The reader here is what turns that into a red run.
//
// It is offline by construction and reaches nothing outside this repository, so
// it is not the general documentation lint issue #110 asks for. An external link
// is somebody else's uptime and belongs on a route that can fail without
// blocking a merge; this one answers only the question the tree can answer about
// itself. What it does not cover is stated in checkReferences rather than left
// for a reader to infer from a green run.

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// A reference is one pointer out of one document, with where it was written.
type reference struct {
	Line int
	// Written is the text as the document has it, which is what a person has
	// to search for to repair it.
	Written string
	// Targets are the places the reference could mean, in the order they are
	// tried. A reference resolving to any one of them exists.
	Targets []string
	// Form is how it was written, so a failure says which of the two shapes
	// was wrong and the message matches what the reader will see on the line.
	Form string
}

// A danglingReference is a reference that resolved to nothing.
type danglingReference struct {
	Path string
	reference
}

func (d danglingReference) String() string {
	return fmt.Sprintf("%s:%d: the %s %q points at %s, which is not in this repository",
		d.Path, d.Line, d.Form, d.Written, strings.Join(d.Targets, " or "))
}

// linkPattern is a Markdown inline link's target.
//
// The target is taken up to the first closing parenthesis, which is what
// Markdown itself does for a target that is not wrapped in angle brackets, and
// no reference in this tree is.
var linkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// tickPattern is a token in backticks with no whitespace in it.
//
// Most references in these documents are Markdown links, and the rest are paths
// written in backticks in the middle of a sentence. The second form is narrowed
// hard below rather than here, because a backticked token is also how this
// repository writes a command, a field name, a symbol and an error string.
var tickPattern = regexp.MustCompile("`([^`\\s]+)`")

// pathExtensions are the suffixes that make a backticked token a path rather
// than prose.
//
// A list rather than "contains a dot", because `docs/decisions` is a directory
// this tree writes constantly and `R_w` is a symbol. Widening this is widening
// what the reader claims about tokens it has not been shown to understand, and
// the near-miss fixture is where that widening has to be argued.
var pathExtensions = []string{".md", ".go", ".json", ".yml", ".yaml", ".txt", ".sh", ".spectrum"}

// referencesIn finds every reference in one Markdown document.
//
// docPath is the document's path from the repository root, and is used to
// resolve a relative target and to report a finding. src is its bytes, so a
// caller may pass a fixture that is not at that path.
//
// Fenced code blocks are skipped whole. Every command in these documents is
// quoted with its output, and that output is full of paths that were true when
// the command ran: a line number, a temporary directory, a file the command was
// shown failing to find. Reading those as references would make the reader fire
// on the evidence this repository requires, which is the fastest way to have it
// switched off.
func referencesIn(docPath string, src []byte) []reference {
	dir := path.Dir(docPath)
	if dir == "." {
		dir = ""
	}

	var found []reference
	fenced := false
	for i, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}

		for _, m := range linkPattern.FindAllStringSubmatch(line, -1) {
			target := strings.Fields(m[1])
			if len(target) == 0 {
				continue
			}
			written := target[0]
			if external(written) {
				continue
			}
			// A fragment names a place inside the target rather than a
			// different file, and a link that is only a fragment names this
			// document. Neither is a reference to a file.
			clean, _, _ := strings.Cut(written, "#")
			if clean == "" {
				continue
			}
			found = append(found, reference{
				Line:    i + 1,
				Written: written,
				Targets: []string{path.Clean(path.Join(dir, clean))},
				Form:    "link",
			})
		}

		for _, m := range tickPattern.FindAllStringSubmatch(line, -1) {
			written := m[1]
			if !looksLikeAPath(written) {
				continue
			}
			// Two readings, and either one existing is enough. A backticked
			// path in these documents is written from the repository root most
			// of the time and from the document's own directory the rest of the
			// time, and both are ordinary. Refusing one of the two would be this
			// reader imposing a convention nothing else in the tree states.
			targets := []string{path.Clean(written)}
			if dir != "" {
				targets = append(targets, path.Clean(path.Join(dir, written)))
			}
			found = append(found, reference{
				Line:    i + 1,
				Written: written,
				Targets: targets,
				Form:    "path",
			})
		}
	}
	return found
}

// external reports whether a link target leaves this repository or names a place
// inside the current document.
func external(target string) bool {
	for _, prefix := range []string{"http://", "https://", "mailto:", "#"} {
		if strings.HasPrefix(target, prefix) {
			return true
		}
	}
	return false
}

// looksLikeAPath reports whether a backticked token is a path this reader will
// answer for.
//
// It has to contain a directory separator and end in one of the extensions
// above, and it may not carry a character that means the token is a pattern, a
// placeholder or a command's output rather than a file. A colon is excluded for
// that last reason: `store/spectrum.go:481` is a line reference in quoted output
// and the file it names exists, so reading it as a path would report a file that
// is there as missing.
func looksLikeAPath(token string) bool {
	if !strings.Contains(token, "/") {
		return false
	}
	if strings.ContainsAny(token, "<>*?:|$") {
		return false
	}
	if strings.HasPrefix(token, "/") || strings.HasPrefix(token, "http") {
		return false
	}
	for _, ext := range pathExtensions {
		if strings.HasSuffix(token, ext) {
			return true
		}
	}
	return false
}

// trackedPaths is every file git has, and every directory above one.
//
// Directories are included because a document pointing at a directory is
// pointing at something that exists, and git tracks files only. It asks git
// rather than walking the tree so that a build artefact, an editor's scratch
// file and anything ignored cannot make a reference resolve for a reader who
// does not have it.
func trackedPaths(root string) (map[string]bool, error) {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot list the tracked files: %w", err)
	}
	known := map[string]bool{}
	for _, f := range strings.Fields(string(b)) {
		known[f] = true
		for d := path.Dir(f); d != "." && d != "/"; d = path.Dir(d) {
			known[d] = true
		}
	}
	return known, nil
}

// checkReferences is the leg. It reads every tracked Markdown document and
// refuses one that points at a file this repository does not have.
//
// What it does not cover, said here because a green run says nothing about any
// of it. It does not resolve an external link, which is somebody else's uptime.
// It does not check that a fragment names a heading that exists. It does not
// read a path out of a fenced code block, so a command quoted with a path that
// has since moved is invisible to it. It does not judge whether a reference
// points at the right file, only that the file is there, and a document pointing
// confidently at the wrong existing file passes.
func checkReferences(root string, out io.Writer) error {
	known, err := trackedPaths(root)
	if err != nil {
		return err
	}

	docs := make([]string, 0, len(known))
	for p := range known {
		if strings.HasSuffix(p, ".md") {
			docs = append(docs, p)
		}
	}
	sort.Strings(docs)

	var dangling []danglingReference
	checked := 0
	for _, doc := range docs {
		src, err := documentBytes(root, doc)
		if err != nil {
			return err
		}
		for _, ref := range referencesIn(doc, src) {
			checked++
			resolved := false
			for _, t := range ref.Targets {
				if known[t] {
					resolved = true
					break
				}
			}
			if !resolved {
				dangling = append(dangling, danglingReference{Path: doc, reference: ref})
			}
		}
	}

	// A count rather than a presence. This reader fails by finding nothing to
	// read, which looks exactly like a tree whose every reference resolves.
	if len(docs) == 0 || checked == 0 {
		return fmt.Errorf("%d tracked document(s) yielded %d reference(s); a run that examined nothing is not a clean tree", len(docs), checked)
	}

	if len(dangling) > 0 {
		var b strings.Builder
		for _, d := range dangling {
			fmt.Fprintf(&b, "  %s\n", d)
		}
		return fmt.Errorf("%d reference(s) out of %d point at a file this repository does not have:\n%s",
			len(dangling), checked, strings.TrimRight(b.String(), "\n"))
	}

	fmt.Fprintf(out, "           %d reference(s) in %d document(s), every one of them resolved\n", checked, len(docs))
	return nil
}

// documentBytes reads a tracked document from the working tree.
//
// The working tree rather than the index, so that a contributor who has edited a
// document and not staged it gets a verdict on what they are looking at. Git
// decides which files are documents and the disk decides what they say.
func documentBytes(root, rel string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, fmt.Errorf("cannot read the document %s: %w", rel, err)
	}
	return b, nil
}

package main

// The rule that every data file in this tree is validated against the schema
// that claims it, and that a data file no schema claims is refused.
//
// The database is data in the tree, edited by people and reviewed as text, so
// the ordinary mistake is a record with a missing field, a value in the wrong
// unit or a spectrum a band short. None of those is visible to somebody reading
// a diff, and until this leg existed none of them was visible to anything else
// either: data/schema/component-record.schema.json said what a record must
// carry and no route read it.
//
// The second refusal is the one that decides whether the first is worth having.
// A check that walks a directory and validates what it finds says nothing about
// a record somebody put somewhere else, and an unvalidated file that looks
// validated is worse than one that is plainly outside the system. So the rule
// runs over every tracked file rather than over a directory, and every path is
// placed in exactly one class: a schema, a record some schema claims, prose, or
// something outside the data system for a reason written down. A path that fits
// none of those is the refusal.
//
// The classification is the whole of the rule and it is stated where somebody
// adding a data file will meet it, in data/README.md, rather than only here.

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// A claim is one schema and the files it answers for.
//
// Written out rather than derived from what is in the tree. A mapping derived
// from the files present would claim whatever arrived, which is the fail-open
// shape this leg's second refusal exists against.
type claim struct {
	Schema string
	// What describes the claim in the words a person adding a file needs.
	What string
	// Holds reports whether this schema claims that path.
	Holds func(rel string) bool
}

func claims() []claim {
	return []claim{
		{
			Schema: "data/schema/component-record.schema.json",
			What:   "every component record, which is every JSON file under data/ that is not itself in data/schema/",
			Holds: func(rel string) bool {
				return strings.HasPrefix(rel, "data/") &&
					!strings.HasPrefix(rel, "data/schema/") &&
					strings.HasSuffix(rel, ".json")
			},
		},
	}
}

// fileClass is what this leg takes a tracked path to be.
type fileClass int

const (
	// outsideTheData is a path the data system does not reach, each for a
	// reason in reasonOutside below.
	outsideTheData fileClass = iota
	// schemaDocument is a schema, which is judged by being used rather than by
	// being validated against something else.
	schemaDocument
	// dataRecord is a file some schema claims.
	dataRecord
	// dataProse is documentation sitting beside the data.
	dataProse
	// unclaimed is the refusal: a file inside the data system, or shaped like
	// one outside it, that no schema answers for.
	unclaimed
)

// classify places one tracked path.
//
// The exemptions are narrow and each is here because something in the tree
// needs it, not because it might one day. A fixture is a committed file under a
// package's own testdata directory, and a fixture written to be refused is the
// proof this leg ships with, so a testdata directory cannot also be the
// database. Under .github/ a JSON file is configuration for a tool that runs
// this repository, including the pattern analyser's own schema fixture, and none
// of it is a laboratory value.
//
// Everything else that is JSON is inside the data system whether or not it is
// under data/. That is the difference between a file this leg refuses and a file
// it cannot see: a record dropped at the repository root is a record with no
// schema behind it, and a directory rule would pass it in silence.
func classify(rel string) (fileClass, claim) {
	if inTestdata(rel) || strings.HasPrefix(rel, ".github/") {
		return outsideTheData, claim{}
	}
	if strings.HasPrefix(rel, "data/schema/") {
		if strings.HasSuffix(rel, ".schema.json") {
			return schemaDocument, claim{}
		}
		return unclaimed, claim{}
	}
	for _, c := range claims() {
		if c.Holds(rel) {
			return dataRecord, c
		}
	}
	if strings.HasPrefix(rel, "data/") {
		if strings.HasSuffix(rel, ".md") {
			return dataProse, claim{}
		}
		return unclaimed, claim{}
	}
	if strings.HasSuffix(rel, ".json") {
		return unclaimed, claim{}
	}
	return outsideTheData, claim{}
}

// inTestdata reports whether a path lies under a testdata directory at any
// depth, which is where a fixture lives and is the one place a file shaped like
// a record is deliberately not one.
func inTestdata(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part == "testdata" {
			return true
		}
	}
	return false
}

// A badRecord is one record that did not match the schema claiming it.
type badRecord struct {
	Path   string
	Schema string
	Wrong  []problem
}

func (b badRecord) String() string {
	var s strings.Builder
	fmt.Fprintf(&s, "  %s, against %s", b.Path, b.Schema)
	for _, p := range b.Wrong {
		fmt.Fprintf(&s, "\n    %s", p)
	}
	return s.String()
}

// checkData is the leg.
//
// What it does not cover, said here because a green run says nothing about any
// of it. It judges a record against the schema and never against the certificate
// the record came from, so a value transcribed from the wrong line of the right
// report passes. It judges no relation between records, so an identity used
// twice and a base construction naming a record that is not here both pass. It
// reads what git tracks, so an untracked file in the working tree is not
// examined and neither is one that is ignored.
func checkData(root string, out io.Writer) error {
	tracked, err := trackedFiles(root)
	if err != nil {
		return err
	}

	var (
		schemas   []string
		records   []string
		strays    []string
		claimedBy = map[string]claim{}
	)
	for _, rel := range tracked {
		class, c := classify(rel)
		switch class {
		case schemaDocument:
			schemas = append(schemas, rel)
		case dataRecord:
			records = append(records, rel)
			claimedBy[rel] = c
		case unclaimed:
			strays = append(strays, rel)
		}
	}

	if len(strays) > 0 {
		var b strings.Builder
		for _, s := range strays {
			fmt.Fprintf(&b, "\n  %s", s)
		}
		return fmt.Errorf("%d file(s) sit inside the data system and no schema claims them; the rule for what a data file is, and what to do with one that is not, is in data/README.md:%s", len(strays), b.String())
	}

	// A schema nothing is judged against is a schema that has stopped being a
	// rule, and it looks exactly like one that is holding. Both directions of
	// the mapping are refused here rather than only the one that fails loudly.
	declared := map[string]bool{}
	for _, c := range claims() {
		declared[c.Schema] = true
	}
	var orphans []string
	for _, s := range schemas {
		if !declared[s] {
			orphans = append(orphans, s)
		}
	}
	if len(orphans) > 0 {
		return fmt.Errorf("%d schema(s) claim no file, so nothing is judged against them: %s", len(orphans), strings.Join(orphans, ", "))
	}
	for _, c := range claims() {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(c.Schema))); err != nil {
			return fmt.Errorf("the mapping names the schema %s, which is not in this repository", c.Schema)
		}
	}

	// A count rather than a presence. This leg fails by finding nothing to
	// validate, and a run that validated no record looks exactly like a run in
	// which every record was good.
	if len(schemas) == 0 || len(records) == 0 {
		return fmt.Errorf("%d schema(s) and %d record(s); a run that validated nothing is not a clean tree", len(schemas), len(records))
	}

	loaded := map[string]*schema{}
	var bad []badRecord
	for _, rel := range records {
		c := claimedBy[rel]
		s, ok := loaded[c.Schema]
		if !ok {
			src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(c.Schema)))
			if err != nil {
				return fmt.Errorf("cannot read the schema %s: %w", c.Schema, err)
			}
			s, err = readSchema(c.Schema, src)
			if err != nil {
				return err
			}
			loaded[c.Schema] = s
		}
		src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return fmt.Errorf("cannot read the record %s: %w", rel, err)
		}
		wrong, err := s.Validate(src)
		if err != nil {
			return fmt.Errorf("%s %w", rel, err)
		}
		if len(wrong) > 0 {
			bad = append(bad, badRecord{Path: rel, Schema: c.Schema, Wrong: wrong})
		}
	}

	if len(bad) > 0 {
		var b strings.Builder
		for _, r := range bad {
			fmt.Fprintf(&b, "\n%s", r)
		}
		return fmt.Errorf("%d record(s) out of %d do not match the schema that claims them:%s", len(bad), len(records), b.String())
	}

	fmt.Fprintf(out, "           %d record(s) against %d schema(s), every field of every one of them\n", len(records), len(schemas))
	for _, c := range claims() {
		fmt.Fprintf(out, "           %s claims %s\n", path.Base(c.Schema), c.What)
	}
	return nil
}

// trackedFiles is every file git has, in a stable order.
//
// Git rather than a walk of the disk, so a build artefact, an editor's scratch
// file or anything ignored cannot become a record, and so that a record somebody
// forgot to add is absent here for the same reason it would be absent for
// everybody else. Files only: git tracks no directory, which is what makes this
// list the right one to classify.
func trackedFiles(root string) ([]string, error) {
	listed, err := output(root, "git", "ls-files")
	if err != nil {
		return nil, fmt.Errorf("cannot list the tracked files: %w", err)
	}
	var files []string
	for _, line := range strings.Split(listed, "\n") {
		if rel := strings.TrimSpace(line); rel != "" {
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	return files, nil
}

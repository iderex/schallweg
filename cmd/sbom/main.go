// Command sbom writes a bill of materials for one of the two things this
// repository ships.
//
// An engineer installing this in a consultancy is asked what is inside it, and a
// public body asks the same question in writing. The answer has to be produced by
// the build rather than reconstructed afterwards from memory, because a
// reconstruction is a claim about an artefact made from the nearest thing to
// hand instead of from the artefact itself.
//
// The program's answer is read out of the compiled binary. Go records the module
// graph and the toolchain version in the executable, so the document describes
// what was linked rather than what the source declared, and the two can differ.
// The database's answer is read out of the record tree, one entry per file with
// its checksum, because a database is a collection of facts and the question
// about it is which facts and where each came from.
//
// The two are separate documents on purpose. They are released separately, a
// user may take only one, and a single document covering both would say that
// taking the program tells you something about the data, which it does not.
//
// Nothing here reads the clock or invents a name. The creation timestamp and the
// document namespace are arguments, so the same tree at the same commit produces
// the same bytes, and a document that differs is a tree that differs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "sbom: %v\n", err)
		os.Exit(1)
	}
}

// options is everything the command was told, after the flags are parsed and
// before anything is read from disk.
type options struct {
	subject   string
	binary    string
	root      string
	created   string
	namespace string
	out       string
}

func run(args []string, stdout *os.File) error {
	var o options

	fs := flag.NewFlagSet("sbom", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&o.subject, "subject", "", "which document to write: program or database")
	fs.StringVar(&o.binary, "binary", "", "the compiled program to read the module graph out of, for -subject program")
	fs.StringVar(&o.root, "root", "data", "the directory holding the component records, for -subject database")
	fs.StringVar(&o.created, "created", "", "the creation timestamp, RFC 3339, so the document is reproducible rather than stamped with the moment it ran")
	fs.StringVar(&o.namespace, "namespace", "", "the document namespace, a URI unique to this document")
	fs.StringVar(&o.out, "out", "", "write here instead of to standard output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := build(o)
	if err != nil {
		return err
	}

	b, err := marshal(doc)
	if err != nil {
		return err
	}

	if o.out == "" {
		_, err := stdout.Write(b)
		return err
	}
	// 0o644: the document is meant to be read by whoever downloads the build.
	return os.WriteFile(o.out, b, 0o644)
}

// build decides which document was asked for and refuses an incomplete ask.
//
// The two required arguments are checked here rather than defaulted, because a
// default for either of them is a document that looks reproducible and is not:
// a timestamp taken from the clock differs on every run, and a namespace that
// is not unique makes two different documents claim to be the same one.
func build(o options) (*document, error) {
	if o.created == "" {
		return nil, errors.New("-created is required; pass the commit's own timestamp so the document is reproducible from the tree")
	}
	if o.namespace == "" {
		return nil, errors.New("-namespace is required; pass a URI unique to this document, which the commit sha is enough to make")
	}

	switch o.subject {
	case "program":
		if o.binary == "" {
			return nil, errors.New("-binary is required for -subject program; the module graph is read out of the compiled program, not out of the source")
		}
		return programDocument(o.binary, o.created, o.namespace)
	case "database":
		return databaseDocument(o.root, o.created, o.namespace)
	case "":
		return nil, errors.New("-subject is required: program or database")
	default:
		return nil, fmt.Errorf("unknown subject %q: this command writes one of program or database", o.subject)
	}
}

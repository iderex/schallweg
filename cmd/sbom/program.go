package main

// The program's bill of materials, read out of the compiled binary.

import (
	"debug/buildinfo"
	"fmt"
	"path"
	"runtime/debug"
	"strings"
)

// programComment is the creation comment on the program document. It says what
// the document was derived from, because a reader otherwise has to guess whether
// it describes the source or the artefact, and those differ.
const programComment = "Read out of the compiled program with debug/buildinfo, so this describes what was linked rather than what the source declared. The toolchain is listed as a dependency because the standard library ships inside the binary."

// programDocument reads the module graph the toolchain recorded in a compiled
// program and writes it as an SPDX document.
//
// Reading the binary rather than go.mod is the whole point. go.mod says what the
// module declares; the binary says what was linked, including the toolchain that
// linked it, and the question an engineer is asked is about the thing they were
// given.
func programDocument(binary, created, namespace string) (*document, error) {
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return nil, fmt.Errorf("cannot read the build information out of %s: %w", binary, err)
	}

	sum, err := sha256Of(binary)
	if err != nil {
		return nil, err
	}

	name := info.Path
	if name == "" {
		// A binary with no recorded main package path is one the toolchain did
		// not stamp. Say so rather than emitting a document with an empty name.
		return nil, fmt.Errorf("%s records no main package path; it was not built by the Go toolchain in a way that carries build information", binary)
	}

	doc := newDocument(path.Base(name), created, namespace, programComment)

	const programID = "SPDXRef-Program"
	doc.Packages = append(doc.Packages, pkg{
		SPDXID:           programID,
		Name:             name,
		VersionInfo:      info.Main.Version,
		DownloadLocation: noAssertion,
		FilesAnalyzed:    false,
		Checksums:        []checksum{{Algorithm: "SHA256", ChecksumValue: sum}},
		LicenseConcluded: noAssertion,
		LicenseDeclared:  noAssertion,
		CopyrightText:    noAssertion,
		Comment:          fmt.Sprintf("Built from module %s.", info.Main.Path),
	})
	doc.Relationships = append(doc.Relationships, relationship{
		SPDXElementID:      "SPDXRef-DOCUMENT",
		RelationshipType:   "DESCRIBES",
		RelatedSPDXElement: programID,
	})

	toolchainID := "SPDXRef-Toolchain"
	doc.Packages = append(doc.Packages, pkg{
		SPDXID:           toolchainID,
		Name:             "go",
		VersionInfo:      info.GoVersion,
		DownloadLocation: "https://go.dev/dl/",
		FilesAnalyzed:    false,
		LicenseConcluded: noAssertion,
		LicenseDeclared:  noAssertion,
		CopyrightText:    noAssertion,
		Comment:          "The Go toolchain, which supplies the standard library linked into the program. With an empty module graph this and the program itself are the whole supply chain, so leaving it out would make the document read as if nothing were inside the binary.",
	})
	doc.Relationships = append(doc.Relationships, relationship{
		SPDXElementID:      programID,
		RelationshipType:   "DEPENDS_ON",
		RelatedSPDXElement: toolchainID,
	})

	for i, dep := range info.Deps {
		id := fmt.Sprintf("SPDXRef-Module-%d-%s", i+1, safeIDPart(dep.Path))
		doc.Packages = append(doc.Packages, modulePackage(id, dep))
		doc.Relationships = append(doc.Relationships, relationship{
			SPDXElementID:      programID,
			RelationshipType:   "DEPENDS_ON",
			RelatedSPDXElement: id,
		})
	}

	return doc, nil
}

// modulePackage turns one module the binary records into one package entry.
//
// A replaced module is reported as what was actually linked, with the module it
// stands in for named in the comment. Reporting the original would name
// something that is not in the binary, which is the failure the whole document
// exists against.
func modulePackage(id string, m *debug.Module) pkg {
	linked := m
	comment := ""
	if m.Replace != nil {
		linked = m.Replace
		comment = fmt.Sprintf("Linked in place of %s %s.", m.Path, m.Version)
	}

	p := pkg{
		SPDXID:           id,
		Name:             linked.Path,
		VersionInfo:      linked.Version,
		DownloadLocation: noAssertion,
		FilesAnalyzed:    false,
		LicenseConcluded: noAssertion,
		LicenseDeclared:  noAssertion,
		CopyrightText:    noAssertion,
		ExternalRefs: []externalRef{{
			ReferenceCategory: "PACKAGE-MANAGER",
			ReferenceType:     "purl",
			ReferenceLocator:  fmt.Sprintf("pkg:golang/%s@%s", linked.Path, linked.Version),
		}},
		Comment: comment,
	}

	// The module hash the toolchain recorded is what go.sum checked the download
	// against, and it belongs in the document. It goes in the comment rather
	// than in a checksum field, because it is a hash over the module's file list
	// written in base64, and the checksum field is defined as hex over one
	// file's bytes. Putting it there would produce a field that reads as a
	// file checksum and fails every reader that treats it as one. A local
	// replacement has no hash, and that absence is worth seeing.
	if linked.Sum != "" {
		p.Comment = strings.TrimSpace(p.Comment + fmt.Sprintf(" Module hash recorded by the build: %s.", linked.Sum))
	}
	return p
}

// safeIDPart turns a module path into something an SPDX element identifier is
// allowed to contain.
//
// The specification restricts an identifier to letters, digits, dot, hyphen and
// plus. A module path has slashes in it, and a path with a slash left in would
// produce a document that parses and then fails validation somewhere the person
// who downloads it cannot fix.
func safeIDPart(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '+':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}

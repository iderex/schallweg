package main

// The database's bill of materials, read out of the record tree.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// databaseComment is the creation comment on the database document.
const databaseComment = "One entry per file in the record tree, with the checksum of that file. A component record is a fact taken from a published test certificate, so the question this document answers is which records are in the artefact and what each one's bytes are, not which libraries were linked."

// databaseDocument writes a document describing every file under root.
//
// Every file rather than every record, and that is deliberate. A route that
// listed only the files it recognised as records would leave a file nobody
// listed inside an artefact somebody downloaded, which is the shape this
// document exists to refuse. What the tree holds beside the records today is
// its README, and it is listed as what it is.
func databaseDocument(root, created, namespace string) (*document, error) {
	entries, err := recordFiles(root)
	if err != nil {
		return nil, err
	}

	doc := newDocument("schallweg-component-database", created, namespace, databaseComment)

	const databaseID = "SPDXRef-Database"
	sha1s := make([]string, 0, len(entries))
	ids := make([]string, 0, len(entries))

	for i, rel := range entries {
		full := filepath.Join(root, filepath.FromSlash(rel))

		sum256, err := sha256Of(full)
		if err != nil {
			return nil, err
		}
		sum1, err := sha1Of(full)
		if err != nil {
			return nil, err
		}

		id := fmt.Sprintf("SPDXRef-File-%d-%s", i+1, safeIDPart(rel))
		ids = append(ids, id)
		sha1s = append(sha1s, sum1)

		doc.Files = append(doc.Files, file{
			SPDXID:   id,
			FileName: "./" + rel,
			Checksums: []checksum{
				{Algorithm: "SHA256", ChecksumValue: sum256},
				{Algorithm: "SHA1", ChecksumValue: sum1},
			},
			LicenseConcluded: noAssertion,
			CopyrightText:    noAssertion,
		})
		doc.Relationships = append(doc.Relationships, relationship{
			SPDXElementID:      databaseID,
			RelationshipType:   "CONTAINS",
			RelatedSPDXElement: id,
		})
	}

	// The database carries no version yet. An empty versionInfo is the honest
	// answer: the version is cut by the release route when the data changes, and
	// no release has happened, so a number written here would be one this
	// command invented.
	doc.Packages = append(doc.Packages, pkg{
		SPDXID:                  databaseID,
		Name:                    "schallweg-component-database",
		DownloadLocation:        noAssertion,
		FilesAnalyzed:           true,
		PackageVerificationCode: &verification{PackageVerificationCodeValue: verificationCode(sha1s)},
		HasFiles:                ids,
		LicenseConcluded:        noAssertion,
		LicenseDeclared:         noAssertion,
		CopyrightText:           noAssertion,
		Comment:                 fmt.Sprintf("%d file(s) under %s. The database is released separately from the program and carries its own version, which does not exist yet.", len(entries), filepath.ToSlash(root)),
	})

	// The document says what it describes before it lists what that contains, so
	// a reader meets the thing before its parts.
	doc.Relationships = append([]relationship{{
		SPDXElementID:      "SPDXRef-DOCUMENT",
		RelationshipType:   "DESCRIBES",
		RelatedSPDXElement: databaseID,
	}}, doc.Relationships...)

	return doc, nil
}

// recordFiles lists every file under root, as slash-separated paths relative to
// it, in a fixed order.
//
// Sorted rather than left in walk order, because two runs over the same tree
// have to produce the same bytes, and a directory listing is not promised to be
// stable across file systems.
func recordFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("cannot read the record tree at %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory; -root names the tree the records live in", root)
	}

	var found []string
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		found = append(found, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(found)
	return found, nil
}

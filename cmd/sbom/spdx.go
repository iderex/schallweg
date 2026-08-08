package main

// The shape of an SPDX 2.3 document, and the small amount of the specification
// this command needs to write one.
//
// The format is SPDX 2.3, serialised as JSON. Why that format rather than
// another is argued in docs/decisions/bill-of-materials.md, and the version
// string below is the one place the version is written, so a document claiming
// a version this code was not written against cannot be produced by accident.
//
// What is deliberately not here: nothing in this repository validates a produced
// document against the published SPDX JSON schema. Doing that needs a schema
// validator, which is a dependency, and the dependency policy refuses one that
// cannot be cleared. What the suite checks instead is the field list this file
// declares, which is narrower: a document that passes the suite can still fail
// the published schema.

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

// spdxVersion is the specification version every document this command writes
// declares.
const spdxVersion = "SPDX-2.3"

// dataLicense is fixed by the specification: an SPDX document's own metadata is
// CC0-1.0 regardless of the licence of what it describes. It says nothing about
// the licence of this repository, which is not decided.
const dataLicense = "CC0-1.0"

// noAssertion is the specification's word for "this document does not say".
//
// It is used for every licence field here, and that is the honest value rather
// than a placeholder: this repository has no licence file, so a document
// asserting one would be asserting something nobody has decided.
const noAssertion = "NOASSERTION"

// creator names what wrote the document. The specification requires the
// "Tool: name-version" shape.
const creator = "Tool: schallweg-sbom-1"

type document struct {
	SPDXVersion       string         `json:"spdxVersion"`
	DataLicense       string         `json:"dataLicense"`
	SPDXID            string         `json:"SPDXID"`
	Name              string         `json:"name"`
	DocumentNamespace string         `json:"documentNamespace"`
	CreationInfo      creationInfo   `json:"creationInfo"`
	Packages          []pkg          `json:"packages"`
	Files             []file         `json:"files,omitempty"`
	Relationships     []relationship `json:"relationships"`
}

type creationInfo struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
	Comment  string   `json:"comment,omitempty"`
}

type pkg struct {
	SPDXID                  string        `json:"SPDXID"`
	Name                    string        `json:"name"`
	VersionInfo             string        `json:"versionInfo,omitempty"`
	DownloadLocation        string        `json:"downloadLocation"`
	FilesAnalyzed           bool          `json:"filesAnalyzed"`
	PackageVerificationCode *verification `json:"packageVerificationCode,omitempty"`
	HasFiles                []string      `json:"hasFiles,omitempty"`
	Checksums               []checksum    `json:"checksums,omitempty"`
	LicenseConcluded        string        `json:"licenseConcluded"`
	LicenseDeclared         string        `json:"licenseDeclared"`
	CopyrightText           string        `json:"copyrightText"`
	ExternalRefs            []externalRef `json:"externalRefs,omitempty"`
	Comment                 string        `json:"comment,omitempty"`
}

type file struct {
	SPDXID           string     `json:"SPDXID"`
	FileName         string     `json:"fileName"`
	Checksums        []checksum `json:"checksums"`
	LicenseConcluded string     `json:"licenseConcluded"`
	CopyrightText    string     `json:"copyrightText"`
}

type checksum struct {
	Algorithm     string `json:"algorithm"`
	ChecksumValue string `json:"checksumValue"`
}

type verification struct {
	PackageVerificationCodeValue string `json:"packageVerificationCodeValue"`
}

type externalRef struct {
	ReferenceCategory string `json:"referenceCategory"`
	ReferenceType     string `json:"referenceType"`
	ReferenceLocator  string `json:"referenceLocator"`
}

type relationship struct {
	SPDXElementID      string `json:"spdxElementId"`
	RelationshipType   string `json:"relationshipType"`
	RelatedSPDXElement string `json:"relatedSpdxElement"`
}

// newDocument fills in everything that is the same in both documents, so the
// two subjects differ only in what they describe.
func newDocument(name, created, namespace, comment string) *document {
	return &document{
		SPDXVersion:       spdxVersion,
		DataLicense:       dataLicense,
		SPDXID:            "SPDXRef-DOCUMENT",
		Name:              name,
		DocumentNamespace: namespace,
		CreationInfo: creationInfo{
			Created:  created,
			Creators: []string{creator},
			Comment:  comment,
		},
	}
}

// marshal writes the document with a trailing newline.
//
// Indented rather than compact, because this file is read in a review and by a
// person answering a question about what is inside the program. HTML escaping is
// off: a module path is not HTML and escaping it makes the document harder to
// compare against the path the toolchain prints.
func marshal(doc *document) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("cannot write the document: %w", err)
	}
	return buf.Bytes(), nil
}

// sha256Of returns the lowercase hex SHA-256 of a file's contents.
//
// This is the checksum the document makes its integrity claim with.
func sha256Of(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// sha1Of returns the lowercase hex SHA-1 of a file's contents.
//
// SHA-1 appears here for one reason: the package verification code below is
// defined by the specification in terms of SHA-1, so a document that computed it
// with anything else would carry a field that no reader can check. It is not
// used as a security property anywhere in this command, and the integrity claim
// each file entry makes is the SHA-256 beside it.
func sha1Of(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("cannot read %s: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// verificationCode is the specification's package verification code: the SHA-1
// of every file's SHA-1, written as lowercase hex, sorted as strings and
// concatenated with nothing between them.
//
// What it is for: a reader with the files in hand can recompute it and find out
// whether the package they have is the package the document describes, without
// trusting the order the document happened to list them in.
func verificationCode(fileSHA1s []string) string {
	sorted := make([]string, len(fileSHA1s))
	copy(sorted, fileSHA1s)
	sort.Strings(sorted)

	h := sha1.New()
	for _, s := range sorted {
		io.WriteString(h, s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

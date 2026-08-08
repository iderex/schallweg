package main

// What these tests hold, and what they cannot.
//
// They hold the shape of the document: that every field the specification
// requires is present, that the two subjects describe what they say they
// describe, and that the same tree produces the same bytes twice. They do not
// hold conformance to the published SPDX JSON schema, because validating
// against it needs a schema validator this repository does not have. A document
// that passes everything here can still be refused by a reader that validates
// properly, and that residual is stated in docs/decisions/bill-of-materials.md
// rather than left to be discovered.

import (
	"debug/buildinfo"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	testCreated   = "2026-01-02T03:04:05Z"
	testNamespace = "https://example.invalid/schallweg/spdx/test"
)

// databaseFixture is a record tree with a file in the root and one below it, so
// a document built from it shows whether nesting is walked and whether the path
// written into the document is relative to the tree rather than to the machine.
const databaseFixture = "testdata/database"

func TestDatabaseDocumentListsEveryFileBelowTheRoot(t *testing.T) {
	doc, err := databaseDocument(databaseFixture, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the database document: %v", err)
	}

	var names []string
	for _, f := range doc.Files {
		names = append(names, f.FileName)
	}

	want := []string{"./a.json", "./nested/b.json"}
	if len(names) != len(want) {
		t.Fatalf("the document lists %d file(s), want %d: %v", len(names), len(want), names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("file %d is %q, want %q", i, names[i], want[i])
		}
	}
}

func TestEveryDatabaseFileCarriesBothChecksums(t *testing.T) {
	doc, err := databaseDocument(databaseFixture, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the database document: %v", err)
	}

	hex40 := regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64 := regexp.MustCompile(`^[0-9a-f]{64}$`)

	for _, f := range doc.Files {
		got := map[string]string{}
		for _, c := range f.Checksums {
			got[c.Algorithm] = c.ChecksumValue
		}
		if !hex64.MatchString(got["SHA256"]) {
			t.Errorf("%s: SHA256 is %q, want 64 lowercase hex digits", f.FileName, got["SHA256"])
		}
		if !hex40.MatchString(got["SHA1"]) {
			t.Errorf("%s: SHA1 is %q, want 40 lowercase hex digits", f.FileName, got["SHA1"])
		}
	}
}

// TestTheSameTreeProducesTheSameBytes is what makes the document evidence rather
// than decoration. A document that differs run to run cannot be compared against
// the one that shipped, so a reader could never tell a changed artefact from a
// changed clock.
func TestTheSameTreeProducesTheSameBytes(t *testing.T) {
	first, err := databaseDocument(databaseFixture, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the database document: %v", err)
	}
	second, err := databaseDocument(databaseFixture, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the database document a second time: %v", err)
	}

	a, err := marshal(first)
	if err != nil {
		t.Fatalf("writing the first document: %v", err)
	}
	b, err := marshal(second)
	if err != nil {
		t.Fatalf("writing the second document: %v", err)
	}
	if string(a) != string(b) {
		t.Error("two runs over the same tree produced different bytes")
	}
}

// TestVerificationCodeMatchesAHandComputedValue pins the algorithm against a
// value computed outside this code, so a change to the implementation that keeps
// the tests self-consistent still fails.
//
// The value is the SHA-1 of the one file digest below, written as its lowercase
// hex string and hashed as text:
//
//	printf 'da39a3ee5e6b4b0d3255bfef95601890afd80709' | sha1sum
//	10a34637ad661d98ba3344717656fcc76209c2f8  -
func TestVerificationCodeMatchesAHandComputedValue(t *testing.T) {
	got := verificationCode([]string{"da39a3ee5e6b4b0d3255bfef95601890afd80709"})
	if want := "10a34637ad661d98ba3344717656fcc76209c2f8"; got != want {
		t.Errorf("verification code is %q, want %q", got, want)
	}
}

// TestVerificationCodeIsIndependentOfTheOrderGiven is the near miss. Walking a
// tree in a different order is the ordinary way this breaks, and a code that
// moved with the walk order would make two identical trees look different.
//
//	printf '7539058ad69d47657fcf72ae31ad13050dd16a0083eced1f2e62f2f7541e1ec82884e90429f1fea5' | sha1sum
//	4c0ce716c9e971f9799f7fa258c2b7554d84275b  -
func TestVerificationCodeIsIndependentOfTheOrderGiven(t *testing.T) {
	const want = "4c0ce716c9e971f9799f7fa258c2b7554d84275b"

	ascending := []string{"7539058ad69d47657fcf72ae31ad13050dd16a00", "83eced1f2e62f2f7541e1ec82884e90429f1fea5"}
	descending := []string{"83eced1f2e62f2f7541e1ec82884e90429f1fea5", "7539058ad69d47657fcf72ae31ad13050dd16a00"}

	if got := verificationCode(ascending); got != want {
		t.Errorf("in ascending order the code is %q, want %q", got, want)
	}
	if got := verificationCode(descending); got != want {
		t.Errorf("in descending order the code is %q, want %q", got, want)
	}
}

func TestVerificationCodeMovesWhenAFileDoes(t *testing.T) {
	before := verificationCode([]string{"7539058ad69d47657fcf72ae31ad13050dd16a00"})
	after := verificationCode([]string{"83eced1f2e62f2f7541e1ec82884e90429f1fea5"})
	if before == after {
		t.Error("two different file digests produced the same verification code")
	}
}

// TestProgramDocumentReadsTheLinkedGraph builds no program of its own. The test
// binary is a program this toolchain linked, so it carries exactly the build
// information the command reads out of a shipped one.
func TestProgramDocumentReadsTheLinkedGraph(t *testing.T) {
	self := os.Args[0]
	if _, err := buildinfo.ReadFile(self); err != nil {
		t.Fatalf("the test binary carries no build information, so this test cannot say anything: %v", err)
	}

	doc, err := programDocument(self, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the program document: %v", err)
	}

	var toolchain *pkg
	for i := range doc.Packages {
		if doc.Packages[i].SPDXID == "SPDXRef-Toolchain" {
			toolchain = &doc.Packages[i]
		}
	}
	if toolchain == nil {
		t.Fatal("the document names no toolchain; with an empty module graph that leaves it saying nothing is inside the program")
	}
	if !strings.HasPrefix(toolchain.VersionInfo, "go") {
		t.Errorf("the toolchain version is %q, want the version the binary records", toolchain.VersionInfo)
	}

	if len(doc.Packages) != len(doc.Relationships) {
		t.Errorf("%d package(s) against %d relationship(s); every package is either described or depended on, so these move together",
			len(doc.Packages), len(doc.Relationships))
	}
}

// TestABinaryWithNoBuildInformationIsRefused holds the direction that matters:
// a file that is not a Go program has to fail rather than produce a document
// with nothing in it, because an empty document reads as an empty dependency
// graph.
func TestABinaryWithNoBuildInformationIsRefused(t *testing.T) {
	_, err := programDocument(filepath.Join("testdata", "not-a-program.txt"), testCreated, testNamespace)
	if err == nil {
		t.Fatal("a file that is not a Go program produced a document")
	}
}

// TestEveryRequiredFieldIsPresent is the conformance leg. Each entry names the
// field and the reader that needs it, so deleting one from the writer fails here
// with the reason rather than with a diff.
func TestEveryRequiredFieldIsPresent(t *testing.T) {
	doc, err := databaseDocument(databaseFixture, testCreated, testNamespace)
	if err != nil {
		t.Fatalf("building the database document: %v", err)
	}
	b, err := marshal(doc)
	if err != nil {
		t.Fatalf("writing the document: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("the document is not valid JSON: %v", err)
	}

	for _, field := range []string{"spdxVersion", "dataLicense", "SPDXID", "name", "documentNamespace", "creationInfo", "packages", "relationships"} {
		if _, ok := got[field]; !ok {
			t.Errorf("the document carries no %q, which every SPDX reader requires", field)
		}
	}

	if got["spdxVersion"] != spdxVersion {
		t.Errorf("the document declares %v, want %q", got["spdxVersion"], spdxVersion)
	}
	if got["dataLicense"] != dataLicense {
		t.Errorf("the document declares data licence %v, want %q, which the specification fixes", got["dataLicense"], dataLicense)
	}

	packages, ok := got["packages"].([]any)
	if !ok || len(packages) == 0 {
		t.Fatal("the document describes no package")
	}
	first, ok := packages[0].(map[string]any)
	if !ok {
		t.Fatal("the first package is not an object")
	}
	for _, field := range []string{"SPDXID", "name", "downloadLocation", "filesAnalyzed", "licenseConcluded", "licenseDeclared", "copyrightText"} {
		if _, ok := first[field]; !ok {
			t.Errorf("the package carries no %q, which the specification requires of every package", field)
		}
	}
	// A package that says it analysed its files owes the code a reader
	// recomputes to check them. The two are required together, and a document
	// carrying one without the other is the failure this leg is against.
	if first["filesAnalyzed"] == true {
		if _, ok := first["packageVerificationCode"]; !ok {
			t.Error("the package says its files were analysed and carries no packageVerificationCode")
		}
	}
}

func TestAnIncompleteAskIsRefused(t *testing.T) {
	for _, c := range []struct {
		name string
		o    options
	}{
		{"no subject", options{created: testCreated, namespace: testNamespace}},
		{"unknown subject", options{subject: "everything", created: testCreated, namespace: testNamespace}},
		{"no created", options{subject: "database", namespace: testNamespace}},
		{"no namespace", options{subject: "database", created: testCreated}},
		{"program with no binary", options{subject: "program", created: testCreated, namespace: testNamespace}},
	} {
		if _, err := build(c.o); err == nil {
			t.Errorf("%s: produced a document instead of refusing", c.name)
		}
	}
}

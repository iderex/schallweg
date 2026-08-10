package store

// Fuzzing the one input format this repository parses today.
//
// The threat model is a desk-side tool: a user opens a document a colleague
// sent them. Nothing about that document is trustworthy, and the failure to
// find is not memory corruption, because this language does not have that one.
// It is a crash, a run that does not finish, and the one that matters most, a
// parse that succeeds and produces the wrong structure.
//
// The last of those cannot be found by asking whether the parser returned an
// error, because it did not. It is found differentially: read the document,
// write what was read, read that, write again, and compare the two writings.
// A reader that loses a band, transposes two of them, or attaches a value to
// the wrong band centre produces two different writings from one input.
//
// What this cannot see is stated here rather than left to be assumed. The
// comparison is against this repository's own writer, so a defect the reader
// and the writer share is invisible to it: if both agreed that the fourth band
// of the core set is 250 Hz, every round trip below would close. That class is
// what the fixtures in spectrum_test.go are for, where the expected values are
// written out by hand and not derived from the code under test.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzReadDocument reads arbitrary bytes as a spectrum document and, where they
// read, requires the document to survive being written and read again.
//
// A refusal is not a failure. Almost every input reaching this is not a
// document, and the parser refusing it is the parser working. What is a failure
// is a panic, an input that does not finish, a document that will not write, a
// written document that will not read back, and two writings of one input that
// differ.
func FuzzReadDocument(f *testing.F) {
	for _, seed := range seedDocuments(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		doc, err := Read("fuzz", strings.NewReader(string(data)))
		if err != nil {
			// The input is not a document. That is the ordinary case and it is
			// the parser doing its job, so there is nothing to compare.
			return
		}

		// A document that read has a quantity and a spectrum this repository's
		// own writer must be able to express. A reader that admits a value the
		// writer cannot write has let something through the boundary that
		// nothing downstream can hand on.
		var first strings.Builder
		if err := Write(&first, doc.Quantity, doc.Spectrum); err != nil {
			t.Fatalf("this input read as a document and then would not write: %v\ninput:\n%q", err, data)
		}

		again, err := Read("fuzz-rewritten", strings.NewReader(first.String()))
		if err != nil {
			t.Fatalf("this document was written by this package and would not read back: %v\nwritten:\n%q", err, first.String())
		}

		var second strings.Builder
		if err := Write(&second, again.Quantity, again.Spectrum); err != nil {
			t.Fatalf("the document read back and then would not write: %v\nwritten:\n%q", err, first.String())
		}

		if again.Quantity != doc.Quantity {
			t.Fatalf("the document read as %s and read back as %s\ninput:\n%q", doc.Quantity, again.Quantity, data)
		}
		if again.Spectrum.Set() != doc.Spectrum.Set() {
			t.Fatalf("the document is on the %s and read back on the %s\ninput:\n%q", doc.Spectrum.Set(), again.Spectrum.Set(), data)
		}

		// The two spectra are compared band by band, and this is the comparison
		// that carries the round trip. Comparing the two writings instead would
		// be weaker in the direction that matters: a writer that lost precision
		// would produce a writing that reads back as the value it wrote and
		// writes again identically, so the two writings would agree while the
		// number the user handed over had been changed on the way through.
		//
		// The equality is exact and has no tolerance, which is the opposite of
		// what this repository asks of a test comparing computed values. It is
		// not a computed value. The claim is that the document that came out
		// carries the number that went in, and a tolerance here would be a test
		// that permits the loss it exists to refuse. Nothing between the two
		// reads does arithmetic on the value.
		for _, band := range doc.Spectrum.Bands() {
			before, err := doc.Spectrum.At(band)
			if err != nil {
				t.Fatalf("reading %s out of the document that parsed: %v", band, err)
			}
			after, err := again.Spectrum.At(band)
			if err != nil {
				t.Fatalf("reading %s out of the document that was written and read back: %v", band, err)
			}
			if before != after {
				t.Fatalf("the value at %s went in as %v and came back as %v\ninput:\n%q\nwritten:\n%q",
					band, before, after, data, first.String())
			}
		}

		// And the two writings, which is what catches a document that is stable
		// as a spectrum and not as bytes.
		if first.String() != second.String() {
			t.Fatalf("writing this document twice gave two different documents\ninput:\n%q\nfirst:\n%q\nsecond:\n%q",
				data, first.String(), second.String())
		}
	})
}

// seedDocuments is every committed spectrum fixture in this package, as bytes.
//
// The seeds are the fixtures the ordinary suite already uses rather than a
// second set written for the fuzzer. Half of them are documents this parser
// refuses, which is what they are for: a corpus of only well-formed documents
// teaches the fuzzer the shape of the accepting path and leaves the refusals
// unexercised, and the refusals are most of this parser.
//
// They are read from the tree rather than written out here because a seed
// written twice is a seed that drifts from the fixture it was copied from.
func seedDocuments(f *testing.F) [][]byte {
	f.Helper()

	matches, err := filepath.Glob(filepath.Join("testdata", "*.spectrum"))
	if err != nil {
		f.Fatalf("looking for the fixtures: %v", err)
	}
	byteExact, err := filepath.Glob(filepath.Join("testdata", "byte-exact", "*.spectrum"))
	if err != nil {
		f.Fatalf("looking for the byte-exact fixtures: %v", err)
	}
	matches = append(matches, byteExact...)

	// A count rather than a presence. This helper failing to find anything
	// would leave the target running on whatever the corpus directory holds,
	// which looks exactly like a target that is seeded.
	if len(matches) == 0 {
		f.Fatal("no spectrum fixture was found in this package's testdata directory, so the target would run unseeded")
	}

	seeds := make([][]byte, 0, len(matches))
	for _, path := range matches {
		b, err := os.ReadFile(path)
		if err != nil {
			f.Fatalf("reading the fixture %s: %v", path, err)
		}
		seeds = append(seeds, b)
	}
	return seeds
}

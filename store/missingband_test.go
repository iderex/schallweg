package store

// The three ways a real certificate arrives without the bands a calculation
// needs, each as the document it actually is, and what happens to it.
//
// They are here rather than beside the format tests because they are one
// property rather than three parses: no route in this tree turns an absent band
// into a value. The format tests next door judge the format. These judge the
// rule.

import (
	"errors"
	"strings"
	"testing"

	"github.com/iderex/schallweg/acoustic"
)

// TestALaboratoryReportingFromOneHundredHertzCannotAnswerForFiftyHertz is the
// first case, and it is the one that reads without any refusal at all.
//
// A core certificate is a whole and valid document. Nothing is missing from it,
// because the core set is the sixteen bands from 100 Hz to 3150 Hz and it has
// all of them. The absence appears only when a calculation wants a band the set
// does not have, and it appears as a refusal from the calculation rather than
// from the reader.
//
// The conversion to octave bands is such a calculation. The 63 Hz octave is the
// energy sum of 50 Hz, 63 Hz and 80 Hz, and a core spectrum has none of them, so
// it refuses and names the band it wanted. What it does not do is produce the
// five octaves the core set does cover, which would be an octave spectrum
// standing for less energy than its input and would look entirely ordinary.
func TestALaboratoryReportingFromOneHundredHertzCannotAnswerForFiftyHertz(t *testing.T) {
	doc := mustRead(t, "wall-r-core.spectrum")
	if doc.Spectrum.Set() != acoustic.Core {
		t.Fatalf("the fixture is on %s and this case needs the core set", doc.Spectrum.Set())
	}

	_, err := acoustic.EnergySumToOctave(doc.Spectrum)
	if !errors.Is(err, acoustic.ErrOctaveNotWhole) {
		t.Fatalf("converting a core spectrum to octaves gave %v, want ErrOctaveNotWhole", err)
	}
	if !namesBand(err.Error(), "50 Hz") {
		t.Errorf("the refusal is %q, and it does not name the band it wanted", err)
	}

	// The other calculation that meets the same absence is combining this
	// spectrum with one that does carry the low bands. It refuses rather than
	// working on the thirteen nominal frequencies the two sets share, which is
	// the overlap that is always available and always wrong.
	extended := flat(t, acoustic.Extended, 40)
	if _, err := doc.Spectrum.Combine(extended, func(a, b float64) float64 { return a - b }); !errors.Is(err, acoustic.ErrDifferentBandSets) {
		t.Fatalf("combining a core and an extended spectrum gave %v, want ErrDifferentBandSets", err)
	}
}

// TestAnOlderCertificateThatStopsAtThirtyOneFiftyIsRefusedByName is the second
// case: a document that declares the extended set, carries the low bands, and
// stops where an older laboratory stopped.
//
// It is a short document rather than a wrong one, and the refusal has to say
// which bands are not there. Nineteen of twenty-one is a sentence a transcriber
// cannot act on.
func TestAnOlderCertificateThatStopsAtThirtyOneFiftyIsRefusedByName(t *testing.T) {
	_, err := read(t, "older-certificate-stops-at-3150.spectrum")
	if !errors.Is(err, ErrMissingBands) {
		t.Fatalf("reading it returned %v, want ErrMissingBands", err)
	}
	for _, band := range []string{"4000 Hz", "5000 Hz"} {
		if !namesBand(err.Error(), band) {
			t.Errorf("the refusal is %q, and it does not name %s", err, band)
		}
	}
	if namesBand(err.Error(), "3150 Hz") {
		t.Errorf("the refusal is %q, and 3150 Hz is a band the document supplied", err)
	}
}

// TestAManufacturerSummaryOfFourValuesIsNotASpectrum is the third case, and it
// is the one where the temptation to accept is strongest, because four numbers
// at four familiar frequencies look like data.
//
// They are not a spectrum on either set, and the twelve bands nobody supplied
// are all named, because a summary is corrected by going back to the certificate
// and every one of the twelve is a reason to.
func TestAManufacturerSummaryOfFourValuesIsNotASpectrum(t *testing.T) {
	_, err := read(t, "summary-four-values.spectrum")
	if !errors.Is(err, ErrMissingBands) {
		t.Fatalf("reading it returned %v, want ErrMissingBands", err)
	}
	for _, band := range []string{"100 Hz", "125 Hz", "160 Hz", "200 Hz", "315 Hz", "400 Hz",
		"630 Hz", "800 Hz", "1250 Hz", "1600 Hz", "2500 Hz", "3150 Hz"} {
		if !namesBand(err.Error(), band) {
			t.Errorf("the refusal is %q, and it does not name %s", err, band)
		}
	}
	for _, supplied := range []string{"250 Hz", "500 Hz", "1000 Hz", "2000 Hz"} {
		if namesBand(err.Error(), supplied) {
			t.Errorf("the refusal is %q, and %s is a band the document supplied", err, supplied)
		}
	}
}

// TestABandWrittenAsMissingAndABandWithNoLineAreOneRefusal is the rule the three
// cases above rest on, asserted directly.
//
// The two fixtures are different documents. One says a band is missing and the
// other says nothing about it at all. A calculation meets the same absence in
// both, so they refuse the same way and both name the band.
func TestABandWrittenAsMissingAndABandWithNoLineAreOneRefusal(t *testing.T) {
	_, said := read(t, "missing-bands.spectrum")
	_, silent := read(t, "one-band-short.spectrum")
	if !errors.Is(said, ErrMissingBands) || !errors.Is(silent, ErrMissingBands) {
		t.Fatalf("the two refusals are %v and %v, and both should be ErrMissingBands", said, silent)
	}
	if !namesBand(said.Error(), "100 Hz") {
		t.Errorf("the refusal for a band written as missing is %q, and it does not name the band", said)
	}
	if !namesBand(silent.Error(), "3150 Hz") {
		t.Errorf("the refusal for a band with no line is %q, and it does not name the band", silent)
	}
}

// TestADocumentCarryingABandItsSetDoesNotHaveIsADifferentRefusal is the bound on
// the rule above, and it is what keeps the two defects from being described as
// one.
//
// A document declaring the core set and carrying the extended bands is not short
// of anything. Its bands disagree with its declaration, and naming the core
// bands as absent would tell the reader to go and find values that are already
// in the file under other names.
func TestADocumentCarryingABandItsSetDoesNotHaveIsADifferentRefusal(t *testing.T) {
	_, err := read(t, "declared-core-actual-extended.spectrum")
	if !errors.Is(err, acoustic.ErrBandMismatch) {
		t.Fatalf("reading it returned %v, want ErrBandMismatch", err)
	}
	if errors.Is(err, ErrMissingBands) {
		t.Errorf("the refusal is %q, and it describes a disagreement as an absence", err)
	}
}

// namesBand reports whether a refusal names exactly this band.
//
// The leading space is the whole of it and it is not decoration: "250 Hz" is a
// substring of "1250 Hz", so a bare containment check reports that a refusal
// named a band the document supplied. That is the direction that matters,
// because these tests assert what a refusal does not say as well as what it
// does, and a check that cannot fail in one direction proves nothing in it.
func namesBand(message, band string) bool {
	return strings.Contains(message, " "+band)
}

// flat builds a spectrum with one value in every band of a set.
func flat(t *testing.T, set acoustic.BandSet, value float64) acoustic.Spectrum {
	t.Helper()
	nominals := set.Nominals()
	values := make([]float64, len(nominals))
	for i := range values {
		values[i] = value
	}
	s, err := acoustic.New(set, nominals, values)
	if err != nil {
		t.Fatalf("building a flat spectrum on the %s: %v", set, err)
	}
	return s
}

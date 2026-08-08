package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iderex/schallweg/acoustic"
	"github.com/iderex/schallweg/acoustic/approx"
)

// read parses a fixture from this package's testdata directory.
func read(t *testing.T, fixture string) (Document, error) {
	t.Helper()
	path := filepath.Join("testdata", fixture)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening the fixture: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("closing the fixture: %v", err)
		}
	}()
	return Read(path, f)
}

// mustRead parses a fixture that is expected to be valid.
func mustRead(t *testing.T, fixture string) Document {
	t.Helper()
	doc, err := read(t, fixture)
	if err != nil {
		t.Fatalf("reading %s: %v", fixture, err)
	}
	return doc
}

// TestAWellFormedDocumentReads is the case every refusal below is measured
// against. Without it, a reader that refused everything would satisfy all of
// them.
func TestAWellFormedDocumentReads(t *testing.T) {
	doc := mustRead(t, "wall-r-core.spectrum")

	if doc.Quantity != SoundReductionIndex {
		t.Errorf("the quantity is %s, want %s", doc.Quantity, SoundReductionIndex)
	}
	if doc.Spectrum.Set() != acoustic.Core {
		t.Fatalf("the spectrum is on the %s, want the %s", doc.Spectrum.Set(), acoustic.Core)
	}

	// Every band, so a reader that got one right and shifted the rest is caught.
	want := []float64{33.4, 36.1, 39.8, 43.2, 46.0, 48.7, 51.3, 53.9, 55.1, 57.4, 59.0, 60.2, 61.5, 62.8, 63.1, 64.4}
	bands := doc.Spectrum.Bands()
	if len(bands) != len(want) {
		t.Fatalf("the spectrum has %d bands, want %d", len(bands), len(want))
	}
	for i, band := range bands {
		got, err := doc.Spectrum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		// A tenth of the last digit the format carries here, which is far tighter
		// than any use of the value and looser than the last bit of a float64.
		approx.Equal(t, "the value at "+band.String(), got, want[i], 0.01)
	}
}

// TestTheQuantityIsReadFromTheDocument is why the quantity is in the format. The
// two quantities are the same kind of number in the same unit over the same
// bands, so nothing about the values themselves distinguishes them.
func TestTheQuantityIsReadFromTheDocument(t *testing.T) {
	r := mustRead(t, "wall-r-core.spectrum")
	ln := mustRead(t, "impact-level.spectrum")

	if r.Quantity != SoundReductionIndex {
		t.Errorf("the first document reads as %s", r.Quantity)
	}
	if ln.Quantity != NormalisedImpactLevel {
		t.Errorf("the second document reads as %s", ln.Quantity)
	}
	if r.Quantity == ln.Quantity {
		t.Fatal("two documents differing only in their quantity line read as the same quantity")
	}
}

// TestADocumentWhoseDeclaredAndActualBandsDisagreeIsRefused is the refusal the
// redundancy in the format exists for.
//
// The fixture declares the core set and carries sixteen band lines, so its count
// is right and every line is well formed. Its lowest three bands are the extended
// set's, which is what a document assembled from a certificate that reported down
// to 50 Hz looks like.
func TestADocumentWhoseDeclaredAndActualBandsDisagreeIsRefused(t *testing.T) {
	_, err := read(t, "declared-core-actual-extended.spectrum")
	if !errors.Is(err, acoustic.ErrBandMismatch) {
		t.Fatalf("reading it returned %v, want ErrBandMismatch", err)
	}
}

// TestADocumentWithTheWrongNumberOfBandsIsRefused covers the other half: the
// bands are the set's own, and one of them is not there.
func TestADocumentWithTheWrongNumberOfBandsIsRefused(t *testing.T) {
	_, err := read(t, "one-band-short.spectrum")
	if !errors.Is(err, acoustic.ErrBandCount) {
		t.Fatalf("reading it returned %v, want ErrBandCount", err)
	}
}

// TestNumbersHaveOneSpellingAndTheRestAreRefused is the locale rule.
//
// Each fixture below is a document that a European spreadsheet, a report
// generator or a scientific export produces without anybody doing anything
// wrong, and each one is a value whose meaning would depend on who read it.
func TestNumbersHaveOneSpellingAndTheRestAreRefused(t *testing.T) {
	for _, c := range []struct {
		fixture string
		want    error
		why     string
	}{
		{"decimal-comma.spectrum", ErrNumberFormat, "33,4 is thirty-three point four in one country and three hundred and thirty-four in another"},
		{"digit-group.spectrum", ErrNumberFormat, "1,033.4 groups its digits with the separator the line above uses as a decimal point"},
		{"space-grouped.spectrum", ErrLineShape, "1 033.4 groups its digits with a space, which is a field separator here"},
		{"exponent.spectrum", ErrNumberFormat, "3.34e1 is a second spelling of a value a certificate never prints that way"},
	} {
		t.Run(c.fixture, func(t *testing.T) {
			_, err := read(t, c.fixture)
			if !errors.Is(err, c.want) {
				t.Fatalf("reading it returned %v, want %v (%s)", err, c.want, c.why)
			}
		})
	}
}

// TestNoLocaleChangesWhatADocumentMeans is the other direction of the same rule,
// and it is the one the issue asks to be shown.
//
// The same fixture is read under locale variables naming a country that writes a
// decimal comma and one that writes a decimal point, and the assertion is that
// the values do not move.
//
// What this proves is narrower than it looks and the bound is stated here rather
// than left to be assumed. The language's own number parsing does not consult a
// locale in the first place, so this test does not discover anything about the
// reader as it stands today. It is a guard against the change that would break
// it: a conversion routed through anything that does consult the environment,
// which is how this breaks in every language that has the problem at all.
func TestNoLocaleChangesWhatADocumentMeans(t *testing.T) {
	first := mustRead(t, "wall-r-core.spectrum")

	for _, locale := range []string{"de_DE.UTF-8", "en_US.UTF-8", "fr_FR.UTF-8", "C"} {
		t.Setenv("LC_ALL", locale)
		t.Setenv("LC_NUMERIC", locale)
		t.Setenv("LANG", locale)

		again := mustRead(t, "wall-r-core.spectrum")
		for _, band := range again.Spectrum.Bands() {
			got, err := again.Spectrum.At(band)
			if err != nil {
				t.Fatalf("reading %s under %s: %v", band, locale, err)
			}
			want, err := first.Spectrum.At(band)
			if err != nil {
				t.Fatalf("reading %s: %v", band, err)
			}
			// Zero would be the honest tolerance here and the harness refuses one,
			// so this is the smallest value it accepts that is still far below the
			// last place of any of these numbers.
			approx.Equal(t, "the value at "+band.String()+" under "+locale, got, want, 1e-12)
		}
	}
}

// TestAValueOutsideThePlausibleRangeIsRefused is the column swap. The fixture is
// the frequency column pasted into the value column, which is the transcription
// mistake that otherwise produces a result nobody can see is wrong.
func TestAValueOutsideThePlausibleRangeIsRefused(t *testing.T) {
	_, err := read(t, "frequency-column.spectrum")
	if !errors.Is(err, ErrImplausible) {
		t.Fatalf("reading it returned %v, want ErrImplausible", err)
	}
	// The refusal names the band it stopped on, which is the 160 Hz band rather
	// than the first: 100 and 125 are inside the range. That bound is stated in
	// docs/formats/spectrum.md and this is the check of the statement.
	if !strings.Contains(err.Error(), "160 Hz") {
		t.Errorf("the refusal is %q, and it should name the band it stopped on", err)
	}
}

// TestADocumentRecordingMissingBandsIsRefusedAndNamesThemAll holds both halves
// of what the format promises about a band nobody measured: it can be written
// down, and it does not become a spectrum.
func TestADocumentRecordingMissingBandsIsRefusedAndNamesThemAll(t *testing.T) {
	_, err := read(t, "missing-bands.spectrum")
	if !errors.Is(err, ErrMissingBands) {
		t.Fatalf("reading it returned %v, want ErrMissingBands", err)
	}
	// Both of them, not the first. A transcriber who is told about one band goes
	// and finds it, comes back, and is told about the next.
	for _, band := range []string{"100 Hz", "125 Hz"} {
		if !strings.Contains(err.Error(), band) {
			t.Errorf("the refusal is %q, and it does not name %s", err, band)
		}
	}
	if !strings.Contains(err.Error(), "sound reduction index") {
		t.Errorf("the refusal is %q, and it does not name the quantity", err)
	}
}

// TestTheHeaderIsRefusedWhenItDoesNotSayWhatItShould covers the three header
// lines and the first line.
func TestTheHeaderIsRefusedWhenItDoesNotSayWhatItShould(t *testing.T) {
	for _, c := range []struct {
		fixture string
		want    error
	}{
		{"unknown-quantity.spectrum", ErrUnknownQuantity},
		{"unit-not-the-quantitys.spectrum", ErrWrongUnit},
		{"version-two.spectrum", ErrUnsupportedVersion},
	} {
		t.Run(c.fixture, func(t *testing.T) {
			_, err := read(t, c.fixture)
			if !errors.Is(err, c.want) {
				t.Fatalf("reading it returned %v, want %v", err, c.want)
			}
		})
	}
}

// TestSomethingThatIsNotADocumentIsRefused covers the bytes that never were one.
func TestSomethingThatIsNotADocumentIsRefused(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		want error
	}{
		{"empty", "", ErrNotASpectrumDocument},
		{"another format entirely", "Frequency;Level\n100;42.1\n", ErrNotASpectrumDocument},
		{"the right shape under another name", "schallweg-record 1\nquantity R\n", ErrNotASpectrumDocument},
		{"a tab where a space belongs", "schallweg-spectrum\t1\n", ErrLineShape},
		{"a blank line", "schallweg-spectrum 1\n\nquantity R\n", ErrLineShape},
		{"a header shorter than a header", "schallweg-spectrum 1\nquantity R\n", ErrHeader},
		{"a byte outside ASCII", "schallweg-spectrum 1\nquantity \xc2\xb5\n", ErrLineShape},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Read(c.name, strings.NewReader(c.in))
			if !errors.Is(err, c.want) {
				t.Fatalf("reading it returned %v, want %v", err, c.want)
			}
		})
	}
}

// TestSpacingIsRefusedAsSpacing is a guard that changes the message rather than
// the verdict, and it is here because that is the only thing it does.
//
// A line whose spacing is wrong is refused either way: the field count comes out
// wrong. What the spacing check adds is a refusal that says the spacing is
// wrong, and the alternative is a reader telling a transcriber that this line
// has four fields, which sends them to count fields on a line whose fields are
// all correct. Deleting the check leaves the document refused and this assertion
// red, which is the whole claim being made for it.
func TestSpacingIsRefusedAsSpacing(t *testing.T) {
	good, err := os.ReadFile(filepath.Join("testdata", "wall-r-core.spectrum"))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}

	for _, c := range []struct {
		name string
		from string
		to   string
	}{
		{"a trailing space", "band 100 33.4", "band 100 33.4 "},
		{"a leading space", "band 100 33.4", " band 100 33.4"},
		{"two spaces where one belongs", "band 100 33.4", "band 100  33.4"},
	} {
		t.Run(c.name, func(t *testing.T) {
			in := strings.Replace(string(good), c.from, c.to, 1)
			if in == string(good) {
				t.Fatalf("the fixture does not contain %q, so this case changes nothing", c.from)
			}
			_, err := Read(c.name, strings.NewReader(in))
			if !errors.Is(err, ErrLineShape) {
				t.Fatalf("reading it returned %v, want ErrLineShape", err)
			}
			if !strings.Contains(err.Error(), "exactly one space") {
				t.Errorf("the refusal is %q, and it should say what is wrong with the spacing", err)
			}
		})
	}
}

// TestCarriageReturnsDoNotChangeADocument reads the byte-exact fixture, whose
// line endings are the point and which is therefore exempt from this
// repository's line ending normalisation.
//
// A document written on Windows is not a different document. Without the fixture
// this could not be tested at all: a carriage return in an ordinary tracked file
// is normalised away on the way into the repository, so the test would be reading
// bytes that had already had the thing under test removed from them.
func TestCarriageReturnsDoNotChangeADocument(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "byte-exact", "crlf.spectrum"))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	if !strings.Contains(string(raw), "\r\n") {
		t.Fatal("the byte-exact fixture carries no carriage return, so it is not testing what it was written for")
	}

	crlf, err := read(t, filepath.Join("byte-exact", "crlf.spectrum"))
	if err != nil {
		t.Fatalf("reading the carriage-return fixture: %v", err)
	}
	lf := mustRead(t, "wall-r-core.spectrum")

	if crlf.Quantity != lf.Quantity {
		t.Errorf("the two read as different quantities")
	}
	for _, band := range crlf.Spectrum.Bands() {
		got, err := crlf.Spectrum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		want, err := lf.Spectrum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		approx.Equal(t, "the value at "+band.String()+" from the carriage-return document", got, want, 1e-12)
	}
}

// TestWritingAndReadingReturnsTheSameSpectrum is the round trip.
//
// It is asserted twice and the second assertion is the one that proves it is
// exact. The band comparison uses a tolerance because the harness requires one,
// which means it cannot by itself distinguish an exact round trip from one that
// loses a digit. Writing the spectrum that came back and comparing the two
// documents as text can, and a string comparison needs no tolerance.
//
// It runs over two fixtures, and the second is the one that says anything. Every
// value in the first has one decimal place, so a writer that rounded to one
// decimal place would produce the same bytes and this test would pass while that
// writer destroyed everything below a tenth of a decibel. The values in
// full-precision.spectrum need the full width, so that writer reddens the run.
func TestWritingAndReadingReturnsTheSameSpectrum(t *testing.T) {
	for _, fixture := range []string{"wall-r-core.spectrum", "full-precision.spectrum"} {
		t.Run(fixture, func(t *testing.T) { roundTrip(t, fixture) })
	}
}

func roundTrip(t *testing.T, fixture string) {
	t.Helper()
	original := mustRead(t, fixture)

	var first strings.Builder
	if err := Write(&first, original.Quantity, original.Spectrum); err != nil {
		t.Fatalf("writing the spectrum: %v", err)
	}

	back, err := Read("round trip", strings.NewReader(first.String()))
	if err != nil {
		t.Fatalf("reading back what was written: %v", err)
	}
	if back.Quantity != original.Quantity {
		t.Errorf("the quantity came back as %s, want %s", back.Quantity, original.Quantity)
	}
	if back.Spectrum.Set() != original.Spectrum.Set() {
		t.Fatalf("the band set came back as %s, want %s", back.Spectrum.Set(), original.Spectrum.Set())
	}
	for _, band := range back.Spectrum.Bands() {
		got, err := back.Spectrum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		want, err := original.Spectrum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		approx.Equal(t, "the value at "+band.String()+" after a round trip", got, want, 1e-12)
	}

	var second strings.Builder
	if err := Write(&second, back.Quantity, back.Spectrum); err != nil {
		t.Fatalf("writing the spectrum that came back: %v", err)
	}
	if first.String() != second.String() {
		t.Errorf("the round trip is not exact:\nfirst:\n%s\nsecond:\n%s", first.String(), second.String())
	}
}

// TestWhatIsWrittenIsWhatTheFormatSays pins the writer's output against
// docs/formats/spectrum.md rather than against whatever it currently produces.
func TestWhatIsWrittenIsWhatTheFormatSays(t *testing.T) {
	doc := mustRead(t, "wall-r-core.spectrum")

	var out strings.Builder
	if err := Write(&out, doc.Quantity, doc.Spectrum); err != nil {
		t.Fatalf("writing the spectrum: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")

	if len(lines) != 4+doc.Spectrum.Len() {
		t.Fatalf("the document has %d lines, want %d", len(lines), 4+doc.Spectrum.Len())
	}
	for i, want := range []string{"schallweg-spectrum 1", "quantity R", "unit dB", "band-set core"} {
		if lines[i] != want {
			t.Errorf("line %d is %q, want %q", i+1, lines[i], want)
		}
	}
	if lines[4] != "band 100 33.4" {
		t.Errorf("the first band line is %q, want %q", lines[4], "band 100 33.4")
	}
	if strings.Contains(out.String(), "\r") {
		t.Error("the writer produced a carriage return, and it writes line feeds")
	}
	if !strings.HasSuffix(out.String(), "\n") {
		t.Error("the writer produced no final newline")
	}
}

// TestWritingRefusesWhatItCannotWriteBack holds the writer's own refusals, so
// that nothing this project writes is a document it could not read again.
func TestWritingRefusesWhatItCannotWriteBack(t *testing.T) {
	doc := mustRead(t, "wall-r-core.spectrum")

	var out strings.Builder
	var none Quantity
	if err := Write(&out, none, doc.Spectrum); !errors.Is(err, ErrUnknownQuantity) {
		t.Errorf("writing with no quantity returned %v, want ErrUnknownQuantity", err)
	}

	var empty acoustic.Spectrum
	if err := Write(&out, doc.Quantity, empty); !errors.Is(err, acoustic.ErrUnknownBandSet) {
		t.Errorf("writing a spectrum that was never constructed returned %v, want ErrUnknownBandSet", err)
	}

	tooLoud, err := doc.Spectrum.Map(func(v float64) float64 { return v + 200 })
	if err != nil {
		t.Fatalf("building the out-of-range spectrum: %v", err)
	}
	if err := Write(&out, doc.Quantity, tooLoud); !errors.Is(err, ErrImplausible) {
		t.Errorf("writing an out-of-range spectrum returned %v, want ErrImplausible", err)
	}
}

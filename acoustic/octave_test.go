package acoustic

import (
	"errors"
	"strings"
	"testing"

	"github.com/iderex/schallweg/acoustic/approx"
)

// flatExtended is an extended-set spectrum with the same value in every band. It
// is the case where every implementation of an octave conversion agrees, which
// is why it proves nothing on its own and appears here only as one half of the
// pair that does.
func flatExtended(t *testing.T, value float64) Spectrum {
	t.Helper()
	values := make([]float64, len(nominalSeries))
	for i := range values {
		values[i] = value
	}
	s, err := New(Extended, Extended.Nominals(), values)
	if err != nil {
		t.Fatalf("building the flat extended spectrum: %v", err)
	}
	return s
}

// TestTheOctaveGridIsAnchoredOnTheThirdOctaveGrid checks the one thing
// octaveCentre asserts about the series: that the groups of three fall so that
// the 1000 Hz octave is centred on the 1000 Hz third-octave band, which is the
// band the exact centre frequencies are already anchored on.
//
// It is also the test that fails if nominalSeries ever grows by a band that does
// not complete an octave, because octaveCount divides and drops the remainder.
func TestTheOctaveGridIsAnchoredOnTheThirdOctaveGrid(t *testing.T) {
	if octaveCount*3 != len(nominalSeries) {
		t.Fatalf("the third-octave series has %d bands, which is not a whole number of octaves at three bands each", len(nominalSeries))
	}
	bands := OctaveBands()
	if len(bands) != octaveCount {
		t.Fatalf("OctaveBands returned %d bands and there are %d octaves", len(bands), octaveCount)
	}
	want := []int{63, 125, 250, 500, 1000, 2000, 4000}
	if len(want) != len(bands) {
		t.Fatalf("this test names %d octave bands and OctaveBands returned %d", len(want), len(bands))
	}
	for i, b := range bands {
		if b.Nominal() != want[i] {
			t.Errorf("octave band %d is %d Hz and this test expects %d Hz", i, b.Nominal(), want[i])
		}
	}

	// The anchor itself: the 1000 Hz octave's middle third-octave band is the
	// band referenceIndex points at.
	if got := nominalSeries[octaveCentre(4)]; got != nominalSeries[referenceIndex] {
		t.Errorf("the fifth octave is centred on %d Hz and the series is anchored on %d Hz", got, nominalSeries[referenceIndex])
	}
}

// TestAnOctaveBandNamesItsThreeThirdOctaveBands checks that the constituents are
// the right three, that they arrive on the extended set, and that between them
// the octave bands claim every band of the series exactly once. That last part
// is what makes the energy sum a partition rather than a selection: no band is
// counted twice and none is left out.
func TestAnOctaveBandNamesItsThreeThirdOctaveBands(t *testing.T) {
	bands := OctaveBands()
	thirds := bands[4].Thirds()
	if len(thirds) != 3 {
		t.Fatalf("the %s has %d third-octave bands", bands[4], len(thirds))
	}
	want := []int{800, 1000, 1250}
	for i, b := range thirds {
		if b.Nominal() != want[i] {
			t.Errorf("constituent %d of the %s is %d Hz and this test expects %d Hz", i, bands[4], b.Nominal(), want[i])
		}
		if b.Set() != Extended {
			t.Errorf("constituent %d of the %s is on the %s and should be on the %s", i, bands[4], b.Set(), Extended)
		}
	}

	seen := make(map[int]int, len(nominalSeries))
	for _, ob := range bands {
		for _, b := range ob.Thirds() {
			seen[b.Nominal()]++
		}
	}
	if len(seen) != len(nominalSeries) {
		t.Errorf("the octave bands between them name %d third-octave bands and the series has %d", len(seen), len(nominalSeries))
	}
	for nominal, count := range seen {
		if count != 1 {
			t.Errorf("%d Hz is named by %d octave bands", nominal, count)
		}
	}
}

// TestTheOctaveValueIsTheEnergySumOfItsThirds is the arithmetic, worked out by
// hand rather than by the route the code takes.
//
// 30, 33 and 36 dB carry 10^3.0, 10^3.3 and 10^3.6 units of energy, which is
// 1000 + 1995.262 + 3981.072 = 6976.334, and 10 log10(6976.334) is 38.4363 dB.
// The tolerance is a thousandth of a decibel, which is far finer than any
// quantity in this field is reported to, and it is that fine because the
// disagreement this test looks for is an arithmetic mistake rather than a
// rounding one.
func TestTheOctaveValueIsTheEnergySumOfItsThirds(t *testing.T) {
	values := make([]float64, len(nominalSeries))
	for i := range values {
		// A silent value everywhere else, so a sum reaching outside its own
		// octave shows up as a number that is wrong rather than plausible.
		values[i] = -100
	}
	c := octaveCentre(4)
	values[c-1], values[c], values[c+1] = 30, 33, 36

	s, err := New(Extended, Extended.Nominals(), values)
	if err != nil {
		t.Fatalf("building the spectrum: %v", err)
	}
	octaves, err := EnergySumToOctave(s)
	if err != nil {
		t.Fatalf("converting to octaves: %v", err)
	}
	got, err := octaves.At(OctaveBands()[4])
	if err != nil {
		t.Fatalf("reading the 1000 Hz octave: %v", err)
	}
	approx.Equal(t, "the 1000 Hz octave of 30, 33 and 36 dB", got, 38.4363, 0.001)
}

// TestThreeEqualThirdsSumToTheirValuePlusTenLogThree is the second hand-worked
// case and the one a reader can check without a calculator: three equal bands
// carry three times the energy of one, so the octave is 10 log10(3) above them,
// which is 4.7712 dB.
func TestThreeEqualThirdsSumToTheirValuePlusTenLogThree(t *testing.T) {
	octaves, err := EnergySumToOctave(flatExtended(t, 40))
	if err != nil {
		t.Fatalf("converting to octaves: %v", err)
	}
	for _, b := range octaves.Bands() {
		got, err := octaves.At(b)
		if err != nil {
			t.Fatalf("reading the %s: %v", b, err)
		}
		approx.Equal(t, "the "+b.String()+" of three 40 dB thirds", got, 44.7712, 0.001)
	}
}

// TestTwoDifferentSpectraReachTheSameOctaves is what the loss in this direction
// looks like when it is measured instead of described.
//
// One spectrum is flat at 40 dB. The other differs from it inside the 1000 Hz
// octave only: 44.0 dB at 800 Hz and 33.87509 dB at 1000 Hz and 1250 Hz, which
// carries 25118.86 + 2440.57 + 2440.57 = 29999.99 units of energy against the
// flat spectrum's 10000 + 10000 + 10000. The two are 4 dB and 6.1 dB apart band
// by band and their octave spectra agree to better than a thousandth of a
// decibel.
//
// This is the stronger form of the round trip the issue asks for. A round trip
// shows that one particular reverse guess does not recover the input; two
// different inputs reaching one output show that no reverse function can, which
// is the reason there is no reverse function in octave.go to round-trip
// through.
func TestTwoDifferentSpectraReachTheSameOctaves(t *testing.T) {
	flat := flatExtended(t, 40)

	shaped := make([]float64, len(nominalSeries))
	for i := range shaped {
		shaped[i] = 40
	}
	c := octaveCentre(4)
	shaped[c-1], shaped[c], shaped[c+1] = 44.0, 33.87509, 33.87509
	other, err := New(Extended, Extended.Nominals(), shaped)
	if err != nil {
		t.Fatalf("building the shaped spectrum: %v", err)
	}

	// The two inputs disagree, and by more than any tolerance in this field
	// would forgive. Said first, so that the agreement below cannot be read as
	// two names for one spectrum.
	for _, series := range []int{c - 1, c} {
		b := Band{set: Extended, series: series}
		got, err := other.At(b)
		if err != nil {
			t.Fatalf("reading %s of the shaped spectrum: %v", b, err)
		}
		if err := approx.Check("the two inputs at "+b.String(), got, 40, 1.0); err == nil {
			t.Fatalf("the two input spectra agree at %s, so this test compares one spectrum with itself", b)
		}
	}

	flatOctaves, err := EnergySumToOctave(flat)
	if err != nil {
		t.Fatalf("converting the flat spectrum: %v", err)
	}
	otherOctaves, err := EnergySumToOctave(other)
	if err != nil {
		t.Fatalf("converting the shaped spectrum: %v", err)
	}
	for _, b := range flatOctaves.Bands() {
		want, err := flatOctaves.At(b)
		if err != nil {
			t.Fatalf("reading the %s of the flat spectrum: %v", b, err)
		}
		got, err := otherOctaves.At(b)
		if err != nil {
			t.Fatalf("reading the %s of the shaped spectrum: %v", b, err)
		}
		approx.Equal(t, "the "+b.String()+" of two different third-octave spectra", got, want, 0.001)
	}
}

// TestConvertingASetThatDoesNotCoverWholeOctavesFails is the refusal the file is
// arranged around. The core set is such a set: it has 3150 Hz without the rest
// of the 4000 Hz octave and none of the 63 Hz octave, so an octave spectrum
// built from it would stand for less energy than it was given.
func TestConvertingASetThatDoesNotCoverWholeOctavesFails(t *testing.T) {
	values := make([]float64, coreLen)
	for i := range values {
		values[i] = 40
	}
	s, err := New(Core, Core.Nominals(), values)
	if err != nil {
		t.Fatalf("building the core spectrum: %v", err)
	}
	octaves, err := EnergySumToOctave(s)
	if !errors.Is(err, ErrOctaveNotWhole) {
		t.Fatalf("converting a core spectrum returned %v and %v, and should have refused with %v", octaves, err, ErrOctaveNotWhole)
	}
	if octaves.Len() != 0 {
		t.Errorf("the refusal came back with a %d band spectrum attached", octaves.Len())
	}
	// The message has to name a band. A refusal that does not say which band is
	// missing leaves the caller to work out for themselves which set to ask for.
	if !strings.Contains(err.Error(), "Hz") {
		t.Errorf("the refusal names no frequency: %v", err)
	}
}

// TestAZeroSpectrumHasNoOctaves keeps the zero value refused here as it is
// everywhere else in this package, rather than converting to a spectrum of
// zeroes that reads as a measurement of silence.
func TestAZeroSpectrumHasNoOctaves(t *testing.T) {
	octaves, err := EnergySumToOctave(Spectrum{})
	if !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("converting a zero spectrum returned %v and should have refused with %v", err, ErrUnknownBandSet)
	}
	if _, err := octaves.At(OctaveBands()[0]); !errors.Is(err, ErrNoOctaveSpectrum) {
		t.Errorf("reading a zero octave spectrum returned %v and should have refused with %v", err, ErrNoOctaveSpectrum)
	}
}

// TestAZeroOctaveBandCannotReadASpectrum is the same refusal on the reading
// side. An octave band nobody filled in would otherwise read the lowest band,
// which is a plausible number.
func TestAZeroOctaveBandCannotReadASpectrum(t *testing.T) {
	octaves, err := EnergySumToOctave(flatExtended(t, 40))
	if err != nil {
		t.Fatalf("converting to octaves: %v", err)
	}
	if _, err := octaves.At(OctaveBand{}); !errors.Is(err, ErrNotAnOctaveBand) {
		t.Errorf("reading with a zero octave band returned %v and should have refused with %v", err, ErrNotAnOctaveBand)
	}
	if got := (OctaveBand{}).Nominal(); got != 0 {
		t.Errorf("a zero octave band names %d Hz and should name nothing", got)
	}
	if (OctaveBand{}).Thirds() != nil {
		t.Error("a zero octave band names third-octave bands")
	}
}

// TestAnOctaveSpectrumNamesEveryOctaveAndSaysWhatItIs holds the two things an
// octave result carries besides its values: the bands it is on, and what it
// prints as.
//
// Bands is the one that matters. Every test that reads an octave result walks
// it, and a zero octave spectrum names no bands, so an operation returning one
// would run those loops zero times and pass. Printing is here because nothing
// read String at all, and its zero-value branch decides which of two sentences
// a reader of an error gets.
func TestAnOctaveSpectrumNamesEveryOctaveAndSaysWhatItIs(t *testing.T) {
	octaves, err := EnergySumToOctave(flatExtended(t, 40))
	if err != nil {
		t.Fatalf("converting to octaves: %v", err)
	}

	if got := len(octaves.Bands()); got != octaves.Len() {
		t.Errorf("the octave spectrum holds %d value(s) and names %d band(s)", octaves.Len(), got)
	}
	if got := len(octaves.Bands()); got != len(OctaveBands()) {
		t.Errorf("the octave spectrum names %d band(s), want the %d the grid has", got, len(OctaveBands()))
	}

	want := "spectrum on the 7 octave bands 63 Hz to 4000 Hz"
	if got := octaves.String(); got != want {
		t.Errorf("the octave spectrum prints as %q, want %q", got, want)
	}
	if got := (OctaveSpectrum{}).String(); got != "octave spectrum on no bands" {
		t.Errorf("the zero octave spectrum prints as %q", got)
	}
	if (OctaveSpectrum{}).Bands() != nil {
		t.Error("the zero octave spectrum names octave bands")
	}
}

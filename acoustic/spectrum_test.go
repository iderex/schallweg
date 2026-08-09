package acoustic

import (
	"errors"
	"math"
	"testing"

	"github.com/iderex/schallweg/acoustic/approx"
)

// coreValues is sixteen values, one per core band, distinct so that a test
// reading the wrong band gets a different number rather than a plausible one.
func coreValues() []float64 {
	v := make([]float64, coreLen)
	for i := range v {
		v[i] = 30 + float64(i)
	}
	return v
}

// extendedValues is the same idea on twenty-one bands, offset so that no value
// here equals a value at the same nominal frequency in coreValues. That is what
// makes the cross-set tests below say something: a spectrum read on the wrong
// set would return a number that is wrong rather than one that happens to match.
func extendedValues() []float64 {
	v := make([]float64, len(nominalSeries))
	for i := range v {
		v[i] = 100 + float64(i)
	}
	return v
}

// TestTheSetsHoldTheBandsTheDecisionNames pins the two band sets against
// docs/decisions/frequency-bands.md, which fixes them as sixteen bands from
// 100 Hz to 3150 Hz and twenty-one from 50 Hz to 5000 Hz.
//
// The counts are asserted as well as the endpoints, because a set with the right
// first and last band and a gap in the middle is exactly the thing this type
// exists to make impossible.
func TestTheSetsHoldTheBandsTheDecisionNames(t *testing.T) {
	for _, c := range []struct {
		set   BandSet
		count int
		first int
		last  int
	}{
		{Core, 16, 100, 3150},
		{Extended, 21, 50, 5000},
	} {
		nominals := c.set.Nominals()
		if len(nominals) != c.count {
			t.Fatalf("%s: got %d bands, want %d", c.set, len(nominals), c.count)
		}
		if c.set.Len() != c.count {
			t.Errorf("%s: Len is %d, want %d", c.set, c.set.Len(), c.count)
		}
		if nominals[0] != c.first {
			t.Errorf("%s: first band is %d Hz, want %d Hz", c.set, nominals[0], c.first)
		}
		if nominals[len(nominals)-1] != c.last {
			t.Errorf("%s: last band is %d Hz, want %d Hz", c.set, nominals[len(nominals)-1], c.last)
		}
		for i := 1; i < len(nominals); i++ {
			if nominals[i] <= nominals[i-1] {
				t.Errorf("%s: band %d (%d Hz) does not come after band %d (%d Hz)", c.set, i, nominals[i], i-1, nominals[i-1])
			}
		}
	}
}

// TestConstructingWithTheWrongNumberOfValuesFails is the first thing the issue
// asks for. One value short and one value long are both here, because a
// container that checks only one bound is a container that grows the other way.
func TestConstructingWithTheWrongNumberOfValuesFails(t *testing.T) {
	nominals := Core.Nominals()
	values := coreValues()

	for _, c := range []struct {
		name      string
		nominals  []int
		values    []float64
		wantCount int
	}{
		{"one value short", nominals, values[:len(values)-1], len(values) - 1},
		{"one value too many", nominals, append(append([]float64{}, values...), 99), len(values) + 1},
		{"no values at all", nominals, nil, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := New(Core, c.nominals, c.values)
			if !errors.Is(err, ErrBandCount) {
				t.Fatalf("New with %d values returned %v, want ErrBandCount", c.wantCount, err)
			}
		})
	}
}

// TestConstructingWithBandsThatAreNotTheSetsFails covers the other half of the
// same route: the right number of values, offered against the wrong bands.
//
// The swapped-order case is the one worth having. Two adjacent values with their
// centres swapped is a spectrum whose count is right, whose bands are all
// present and whose arithmetic is wrong by whatever those two bands differ by,
// and no later check in this project would notice.
func TestConstructingWithBandsThatAreNotTheSetsFails(t *testing.T) {
	values := coreValues()

	swapped := append([]int{}, Core.Nominals()...)
	swapped[4], swapped[5] = swapped[5], swapped[4]

	extendedCentresCoreCount := append([]int{}, Extended.Nominals()[:coreLen]...)

	for _, c := range []struct {
		name     string
		nominals []int
	}{
		{"two adjacent bands swapped", swapped},
		{"the extended set's lowest bands, with the core set's count", extendedCentresCoreCount},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := New(Core, c.nominals, values)
			if !errors.Is(err, ErrBandMismatch) {
				t.Fatalf("New returned %v, want ErrBandMismatch", err)
			}
		})
	}
}

// TestConstructingOnNoBandSetFails holds the zero value of BandSet. A struct
// field nobody filled in must not resolve to whichever set is first.
func TestConstructingOnNoBandSetFails(t *testing.T) {
	var none BandSet
	if _, err := New(none, Core.Nominals(), coreValues()); !errors.Is(err, ErrUnknownBandSet) {
		t.Fatalf("New on the zero band set returned %v, want ErrUnknownBandSet", err)
	}
	if none.Len() != 0 {
		t.Errorf("the zero band set reports %d bands, want 0", none.Len())
	}
	if none.Bands() != nil {
		t.Errorf("the zero band set produced %d bands, want none", len(none.Bands()))
	}
}

// TestAValueThatIsNotFiniteIsRefused keeps a NaN out of the container. A NaN
// passes through every arithmetic operation without complaint and arrives at the
// far end as a result somebody reads, and the band it came from is gone by then.
func TestAValueThatIsNotFiniteIsRefused(t *testing.T) {
	for _, c := range []struct {
		name string
		bad  float64
	}{
		{"not a number", math.NaN()},
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
	} {
		t.Run(c.name, func(t *testing.T) {
			values := coreValues()
			values[7] = c.bad
			if _, err := New(Core, Core.Nominals(), values); !errors.Is(err, ErrNotFinite) {
				t.Fatalf("New with %v at one band returned %v, want ErrNotFinite", c.bad, err)
			}
		})
	}
}

// TestCombiningTwoSpectraOnDifferentBandSetsFails is the second thing the issue
// asks for.
//
// It matters more than it looks. The two sets share thirteen nominal
// frequencies, so an implementation that combined the overlap would produce a
// number for every band a reader expects to see, computed from the wrong
// positions in one of the two.
func TestCombiningTwoSpectraOnDifferentBandSetsFails(t *testing.T) {
	core, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building the core spectrum: %v", err)
	}
	extended, err := New(Extended, Extended.Nominals(), extendedValues())
	if err != nil {
		t.Fatalf("building the extended spectrum: %v", err)
	}

	add := func(a, b float64) float64 { return a + b }

	if _, err := core.Combine(extended, add); !errors.Is(err, ErrDifferentBandSets) {
		t.Errorf("core.Combine(extended) returned %v, want ErrDifferentBandSets", err)
	}
	if _, err := extended.Combine(core, add); !errors.Is(err, ErrDifferentBandSets) {
		t.Errorf("extended.Combine(core) returned %v, want ErrDifferentBandSets", err)
	}
}

// TestCombiningTwoSpectraOnOneBandSetWorks is the near miss. It is one band set
// away from the test above and has to pass, or the refusal above would be
// satisfied by a Combine that refuses everything.
func TestCombiningTwoSpectraOnOneBandSetWorks(t *testing.T) {
	a, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building the first spectrum: %v", err)
	}
	b, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building the second spectrum: %v", err)
	}

	sum, err := a.Combine(b, func(x, y float64) float64 { return x + y })
	if err != nil {
		t.Fatalf("combining two spectra on one band set: %v", err)
	}
	if sum.Set() != Core {
		t.Fatalf("the result is on the %s, want the %s", sum.Set(), Core)
	}

	for i, band := range sum.Bands() {
		got, err := sum.At(band)
		if err != nil {
			t.Fatalf("reading %s: %v", band, err)
		}
		approx.Equal(t, "the sum at "+band.String(), got, 2*(30+float64(i)), 1e-9)
	}
}

// TestReadingWithABandFromAnotherSetIsRefused is the off-by-one this design
// exists to remove, in the form it actually takes.
//
// 100 Hz is position 0 of the core set and position 3 of the extended set. A
// container indexed by an integer would answer a read at 100 Hz with whatever
// sits at position 3, which is 250 Hz, and the answer would be a number in the
// right range.
func TestReadingWithABandFromAnotherSetIsRefused(t *testing.T) {
	core, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building the core spectrum: %v", err)
	}

	var hundred Band
	for _, b := range Extended.Bands() {
		if b.Nominal() == 100 {
			hundred = b
		}
	}
	if hundred.Nominal() != 100 {
		t.Fatalf("the extended set has no 100 Hz band, which the decision says it has")
	}

	if _, err := core.At(hundred); !errors.Is(err, ErrBandNotInSet) {
		t.Fatalf("reading a core spectrum with the extended set's 100 Hz band returned %v, want ErrBandNotInSet", err)
	}

	// The near miss: the same nominal frequency, from the right set, reads.
	for _, b := range core.Bands() {
		if b.Nominal() == 100 {
			got, err := core.At(b)
			if err != nil {
				t.Fatalf("reading the core set's own 100 Hz band: %v", err)
			}
			approx.Equal(t, "the core spectrum at 100 Hz", got, 30, 1e-9)
		}
	}
}

// TestTheZeroSpectrumIsNotUsable holds the promise that there is no usable empty
// spectrum. Every route into one has to refuse rather than return a zero.
func TestTheZeroSpectrumIsNotUsable(t *testing.T) {
	var empty Spectrum

	if empty.Len() != 0 {
		t.Errorf("the zero spectrum reports %d bands, want 0", empty.Len())
	}
	if _, err := empty.At(Core.Bands()[0]); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("reading the zero spectrum returned %v, want ErrUnknownBandSet", err)
	}
	if _, err := empty.Map(func(v float64) float64 { return v }); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("mapping the zero spectrum returned %v, want ErrUnknownBandSet", err)
	}

	real, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}
	if _, err := real.Combine(empty, func(a, b float64) float64 { return a }); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("combining with the zero spectrum returned %v, want ErrUnknownBandSet", err)
	}
}

// TestTheValuesGivenToNewCannotBeChangedAfterwards is the difference between a
// container and a view of somebody else's slice. Without the copy in New, a
// caller reusing its buffer for the next spectrum would rewrite the one it just
// built, and every value in it would still be plausible.
func TestTheValuesGivenToNewCannotBeChangedAfterwards(t *testing.T) {
	values := coreValues()
	s, err := New(Core, Core.Nominals(), values)
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}

	values[0] = -999

	first := s.Bands()[0]
	got, err := s.At(first)
	if err != nil {
		t.Fatalf("reading %s: %v", first, err)
	}
	approx.Equal(t, "the value at "+first.String()+" after the caller's slice was changed", got, 30, 1e-9)
}

// TestMapKeepsTheBandSetAndRefusesAValueThatIsNotFinite covers the whole-
// spectrum route, including the case where the function itself produces the
// thing New refuses.
func TestMapKeepsTheBandSetAndRefusesAValueThatIsNotFinite(t *testing.T) {
	s, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}

	shifted, err := s.Map(func(v float64) float64 { return v + 3 })
	if err != nil {
		t.Fatalf("mapping a spectrum: %v", err)
	}
	if shifted.Set() != Core {
		t.Errorf("the result is on the %s, want the %s", shifted.Set(), Core)
	}
	first := shifted.Bands()[0]
	got, err := shifted.At(first)
	if err != nil {
		t.Fatalf("reading %s: %v", first, err)
	}
	approx.Equal(t, "the mapped value at "+first.String(), got, 33, 1e-9)

	if _, err := s.Map(func(v float64) float64 { return math.NaN() }); !errors.Is(err, ErrNotFinite) {
		t.Errorf("mapping to a NaN returned %v, want ErrNotFinite", err)
	}
}

// TestExactCentreFrequenciesFollowTheStatedRule checks the one place in this
// package where a frequency is a real number.
//
// The rule is that the band n steps from 1000 Hz has its centre at
// 1000 * 10^(n/10) hertz, so 100 Hz is exact and 2000 Hz is not, and the
// tolerances below are chosen per band from what the rule gives rather than from
// what the code returned.
func TestExactCentreFrequenciesFollowTheStatedRule(t *testing.T) {
	want := map[int]float64{
		50:   50.11872336272722,
		100:  100.0,
		1000: 1000.0,
		2000: 1995.2623149688795,
		5000: 5011.872336272722,
	}

	for _, b := range Extended.Bands() {
		expected, listed := want[b.Nominal()]
		if !listed {
			continue
		}
		// A tolerance in hertz, chosen as one part in a million of the band, which
		// is far tighter than any use this number has and far looser than the last
		// bit of a float64.
		approx.Equal(t, "the exact centre of the "+b.String()+" band", b.Exact(), expected, expected/1e6)
	}
}

// TestABandFromNowhereIsNotABand holds the zero Band. It cannot be built from
// outside this package, so what is tested is that the zero value refuses rather
// than naming the first band of the first set.
func TestABandFromNowhereIsNotABand(t *testing.T) {
	var b Band

	if b.Nominal() != 0 {
		t.Errorf("the zero band names %d Hz, want nothing", b.Nominal())
	}
	if !math.IsNaN(b.Exact()) {
		t.Errorf("the zero band has an exact centre frequency, want none")
	}
	if b.String() != "invalid band" {
		t.Errorf("the zero band prints as %q", b.String())
	}

	s, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}
	if _, err := s.At(b); !errors.Is(err, ErrBandNotInSet) {
		t.Errorf("reading a spectrum with the zero band returned %v, want ErrBandNotInSet", err)
	}
}

// TestABandOneStepOutsideItsSetIsNotABand holds the boundary the Band type
// exists for.
//
// A Band carries a position in the whole series and the set it was taken from,
// and the two sets are different runs of that series. So the interesting wrong
// band is not the zero value, which TestABandFromNowhereIsNotABand already
// covers: it is a band one position past the end of its own set, or one before
// its start, which is a real position of the series and is not a band of that
// set. Both come from the same off-by-one, and both are constructible from
// inside this package.
func TestABandOneStepOutsideItsSetIsNotABand(t *testing.T) {
	s, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}

	past := Band{set: Core, series: coreOffset + coreLen}
	before := Band{set: Core, series: coreOffset - 1}

	for _, b := range []Band{past, before} {
		if _, err := s.At(b); !errors.Is(err, ErrBandNotInSet) {
			t.Errorf("reading position %d of the core set returned %v, want ErrBandNotInSet", b.series, err)
		}
		if got := b.String(); got != "invalid band" {
			t.Errorf("position %d of the core set prints as %q, want %q", b.series, got, "invalid band")
		}
		if got := b.Nominal(); got != 0 {
			t.Errorf("position %d of the core set names %d Hz, want nothing", b.series, got)
		}
	}
}

// TestASpectrumSaysWhatItIsAndNoValues covers both branches of Spectrum.String.
//
// Nothing read either of them. The branch that matters is the zero value's: an
// error message about a spectrum nobody built is the one a reader meets when
// something has already gone wrong, and it must not print as a spectrum on a
// band set with no bands.
func TestASpectrumSaysWhatItIsAndNoValues(t *testing.T) {
	s, err := New(Core, Core.Nominals(), coreValues())
	if err != nil {
		t.Fatalf("building a spectrum: %v", err)
	}

	want := "spectrum on the core third-octave bands 100 Hz to 3150 Hz, 16 bands"
	if got := s.String(); got != want {
		t.Errorf("a core spectrum prints as %q, want %q", got, want)
	}
	if got := (Spectrum{}).String(); got != "spectrum on no band set" {
		t.Errorf("the zero spectrum prints as %q", got)
	}
}

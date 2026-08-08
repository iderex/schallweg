package acoustic

import (
	"fmt"
	"math"
)

// A BandSet is which third-octave bands a spectrum is on.
//
// There are two, and there is no third state. docs/decisions/frequency-bands.md
// says why: a spectrum is on the core set or on the extended set, which one it
// is on is a property of the value and is never inferred, and there is no such
// thing as a spectrum carrying some of the extended bands. A representation that
// permits a partly populated range guarantees that some code path eventually
// reads an absent band as present.
//
// The zero value is deliberately not a set. A struct with a BandSet field that
// nobody filled in is a bug, and it fails at the first call rather than
// resolving to whichever set happened to be first.
type BandSet uint8

const (
	// Core is the sixteen bands from 100 Hz to 3150 Hz: what every laboratory
	// certificate in this field reports, and what the first release computes on.
	Core BandSet = iota + 1
	// Extended is the twenty-one bands from 50 Hz to 5000 Hz: the core set with
	// the low frequency bands a certificate carries when low frequency behaviour
	// is the point, which is where lightweight construction actually fails.
	Extended
)

// nominalSeries is the nominal centre frequency of every band either set can
// hold, in ascending order. The extended set is exactly this series; the core
// set is the run of it from 100 Hz to 3150 Hz.
//
// These are the nominal designations from the preferred-number series, which is
// what a certificate prints and what a user recognises. They are not the exact
// centre frequencies: those are computed from a band's position by the rule in
// Band.Exact rather than written down twice.
var nominalSeries = [...]int{
	50, 63, 80,
	100, 125, 160, 200, 250, 315, 400, 500, 630, 800,
	1000,
	1250, 1600, 2000, 2500, 3150,
	4000, 5000,
}

// referenceIndex is the position of the 1000 Hz band in nominalSeries, which is
// the band the exact frequency rule is anchored on.
const referenceIndex = 13

// coreOffset is where the core set starts inside nominalSeries, and coreLen is
// how many bands it has.
const (
	coreOffset = 3
	coreLen    = 16
)

// bounds returns the half-open range of nominalSeries this set covers.
func (s BandSet) bounds() (lo, hi int, err error) {
	switch s {
	case Core:
		return coreOffset, coreOffset + coreLen, nil
	case Extended:
		return 0, len(nominalSeries), nil
	default:
		return 0, 0, fmt.Errorf("%w: %d", ErrUnknownBandSet, uint8(s))
	}
}

// Len is how many bands the set has. An unknown set has no bands, which is what
// makes every loop over it run zero times rather than a plausible number of
// times.
func (s BandSet) Len() int {
	lo, hi, err := s.bounds()
	if err != nil {
		return 0
	}
	return hi - lo
}

// Bands returns the set's bands in ascending order of frequency.
//
// The slice is freshly built on every call, so a caller cannot reorder or
// truncate the set for everybody else.
func (s BandSet) Bands() []Band {
	lo, hi, err := s.bounds()
	if err != nil {
		return nil
	}
	out := make([]Band, 0, hi-lo)
	for i := lo; i < hi; i++ {
		out = append(out, Band{set: s, series: i})
	}
	return out
}

// Nominals returns the nominal centre frequencies of the set's bands, in the
// same order as Bands. It is what a constructor's caller has to supply and what
// a reader of an error message compares against.
func (s BandSet) Nominals() []int {
	lo, hi, err := s.bounds()
	if err != nil {
		return nil
	}
	out := make([]int, 0, hi-lo)
	for i := lo; i < hi; i++ {
		out = append(out, nominalSeries[i])
	}
	return out
}

// String names the set the way an error message should say it.
func (s BandSet) String() string {
	switch s {
	case Core:
		return "core third-octave bands 100 Hz to 3150 Hz"
	case Extended:
		return "extended third-octave bands 50 Hz to 5000 Hz"
	default:
		return fmt.Sprintf("unknown band set %d", uint8(s))
	}
}

// A Band is one third-octave band of one band set.
//
// Both fields are unexported and there is no literal a caller can write, so a
// Band only ever comes from the set it belongs to. That is the whole point:
// code indexes a spectrum with a Band rather than with an integer, and an
// integer index is where the off-by-one lives, in a loop, where nobody reads it.
//
// The zero Band belongs to no set and is refused by every operation that takes
// one.
type Band struct {
	set    BandSet
	series int
}

// Set is which band set this band belongs to.
func (b Band) Set() BandSet { return b.set }

// Nominal is the band's nominal centre frequency in hertz, which is its
// identity: what a certificate prints, what a file carries and what an error
// message names.
func (b Band) Nominal() int {
	if !b.valid() {
		return 0
	}
	return nominalSeries[b.series]
}

// Exact is the band's exact centre frequency in hertz.
//
// It is computed from the band's position in the series rather than stored,
// because the series is a rule that fits in a sentence: the bands are spaced a
// third of a decade apart and anchored at 1000 Hz, so the band n steps from
// 1000 Hz has its centre at 1000 * 10^(n/10) hertz. Storing a column of rounded
// values instead would be transcription, and two roundings of one band would
// eventually appear as two bands.
//
// This is the only place in the package where a frequency is a real number
// rather than a designation. Nothing compares bands by it.
func (b Band) Exact() float64 {
	if !b.valid() {
		return math.NaN()
	}
	return 1000 * math.Pow(10, float64(b.series-referenceIndex)/10)
}

// String names the band the way an error message should say it.
func (b Band) String() string {
	if !b.valid() {
		return "invalid band"
	}
	return fmt.Sprintf("%d Hz", b.Nominal())
}

// valid reports whether this band came from a set rather than from a zero value.
func (b Band) valid() bool {
	lo, hi, err := b.set.bounds()
	if err != nil {
		return false
	}
	return b.series >= lo && b.series < hi
}

package acoustic

// Octave bands, and the one route between them and third-octave bands.
//
// Laboratory data arrives in third-octave bands. Some published data, some
// requirements and some older certificates are in octave bands, so a kernel that
// meets both has to say what happens in between. The two directions are not the
// same kind of operation, and the whole of this file is arranged so that they
// cannot be mistaken for each other.
//
// Third-octave to octave is defined. An octave band is exactly three
// third-octave bands and the octave value is their energy sum, which follows
// from what the bands are and assumes nothing. It loses the shape of the
// spectrum inside the octave, and that loss is not recoverable: two different
// third-octave spectra reach the same octave spectrum, which
// TestTwoDifferentSpectraReachTheSameOctaves shows rather than describes.
//
// Octave to third-octave is not here, and its absence is the refusal rather than
// an omission. Three numbers cannot be read out of one, so any function
// producing them would have decided what the spectrum looks like inside the
// band, and that invented spectrum would then travel through the calculation
// beside values that were measured.
// docs/decisions/frequency-bands.md refuses exactly that, in the same words it
// refuses filling in a missing band, and the strongest available refusal of a
// function is that there is none to call. A caller holding octave data and
// needing thirds is holding an assumption rather than a conversion, and it
// belongs where the assumption can be stated, which is not this package.

import (
	"errors"
	"fmt"
)

// The refusals this file makes.
var (
	// ErrOctaveNotWhole is a third-octave spectrum whose band set does not hold
	// all three third-octave bands of every octave. Summing the bands that are
	// there would produce an octave spectrum standing for less energy than the
	// input carried, and nothing in the result would say so.
	ErrOctaveNotWhole = errors.New("band set does not cover whole octaves")
	// ErrNotAnOctaveBand is an octave band that did not come from OctaveBands,
	// which in practice is a zero value nobody filled in.
	ErrNotAnOctaveBand = errors.New("not an octave band")
	// ErrNoOctaveSpectrum is an operation on an octave spectrum that was never
	// built by a conversion.
	ErrNoOctaveSpectrum = errors.New("octave spectrum was never constructed")
)

// octaveCount is how many octave bands the third-octave series holds.
//
// It is divided rather than counted, because the relation is the definition:
// every octave is three third-octave bands, so a series that is a whole number
// of octaves has a third as many of them. If nominalSeries ever grows by a band
// that does not complete an octave, this constant drops the remainder, and the
// anchor test in octave_test.go is what fails.
const octaveCount = len(nominalSeries) / 3

// octaveCentre is the position in nominalSeries of the middle third-octave band
// of octave number n, counting from zero at the lowest.
//
// The rule is the anchor rather than a table: the third-octave bands run in
// groups of three from the bottom of the series, and the middle one of each
// group names the octave. That places the 1000 Hz octave on the 1000 Hz
// third-octave band, which is what a certificate prints and what referenceIndex
// already anchors the exact centre frequencies on. A written column of seven
// numbers would be that same fact transcribed a second time, and two
// transcriptions of one series eventually appear as two series.
func octaveCentre(n int) int { return 3*n + 1 }

// An OctaveBand is one octave band.
//
// Like Band it has no literal a caller can write, so an octave band only ever
// comes from OctaveBands. The zero value belongs to no set and is refused by
// every operation that takes one.
type OctaveBand struct {
	// number is the band's position among the octave bands, counted from one so
	// that the zero value is not the lowest band.
	number int
}

// OctaveBands returns the octave bands in ascending order of frequency.
//
// There is one octave set rather than two, and that follows from the refusal
// below rather than being a separate decision: the only third-octave set that
// covers whole octaves is the extended one, so the only octave spectrum this
// package can produce is the one covering all of them.
func OctaveBands() []OctaveBand {
	out := make([]OctaveBand, 0, octaveCount)
	for n := 1; n <= octaveCount; n++ {
		out = append(out, OctaveBand{number: n})
	}
	return out
}

// valid reports whether this band came from OctaveBands rather than from a zero
// value.
func (b OctaveBand) valid() bool { return b.number >= 1 && b.number <= octaveCount }

// Nominal is the band's nominal centre frequency in hertz, which is the nominal
// centre of its middle third-octave band and is what a certificate prints.
func (b OctaveBand) Nominal() int {
	if !b.valid() {
		return 0
	}
	return nominalSeries[octaveCentre(b.number-1)]
}

// Thirds returns the three third-octave bands this octave band is the energy sum
// of, in ascending order, on the extended set.
//
// They are on the extended set because that is the only set the conversion
// accepts, so a caller reading them cannot use them against a core spectrum by
// accident: Spectrum.At refuses a band from the other set rather than reading
// the value at the same position.
func (b OctaveBand) Thirds() []Band {
	if !b.valid() {
		return nil
	}
	c := octaveCentre(b.number - 1)
	out := make([]Band, 0, 3)
	for _, p := range [3]int{c - 1, c, c + 1} {
		out = append(out, Band{set: Extended, series: p})
	}
	return out
}

// String names the band the way an error message should say it.
func (b OctaveBand) String() string {
	if !b.valid() {
		return "invalid octave band"
	}
	return fmt.Sprintf("%d Hz octave", b.Nominal())
}

// An OctaveSpectrum is one value in every octave band.
//
// It is a separate type from Spectrum rather than a Spectrum on a third band
// set. The two are not interchangeable, and a function taking one must not
// quietly accept the other, which is what a shared type with a set field would
// allow. There is no constructor: the only route to an octave spectrum is
// EnergySumToOctave, so an octave spectrum always has a third-octave spectrum
// behind it and a reader can always ask where its numbers came from.
type OctaveSpectrum struct {
	// values is never handed out, for the reason Spectrum.values is not.
	values []float64
}

// Len is how many bands the spectrum has. A zero octave spectrum has none.
func (o OctaveSpectrum) Len() int { return len(o.values) }

// Bands returns the spectrum's bands in ascending order.
func (o OctaveSpectrum) Bands() []OctaveBand {
	if len(o.values) == 0 {
		return nil
	}
	return OctaveBands()
}

// At reads the value in one octave band, and refuses a band that came from a
// zero value rather than from OctaveBands.
func (o OctaveSpectrum) At(b OctaveBand) (float64, error) {
	if len(o.values) == 0 {
		return 0, fmt.Errorf("%w", ErrNoOctaveSpectrum)
	}
	if !b.valid() {
		return 0, fmt.Errorf("%w: %s", ErrNotAnOctaveBand, b)
	}
	return o.values[b.number-1], nil
}

// String says what the spectrum holds and no values, for the reason
// Spectrum.String prints none.
func (o OctaveSpectrum) String() string {
	if len(o.values) == 0 {
		return "octave spectrum on no bands"
	}
	return fmt.Sprintf("spectrum on the %d octave bands %d Hz to %d Hz",
		octaveCount, nominalSeries[octaveCentre(0)], nominalSeries[octaveCentre(octaveCount-1)])
}

// EnergySumToOctave combines a third-octave spectrum into octave bands by
// summing the energy of each octave's three third-octave bands.
//
// The name carries what the function assumes about its input, because the type
// cannot. A Spectrum holds numbers and does not know what quantity they are, and
// this arithmetic is right for a level and wrong for a ratio: a sound reduction
// index, a level difference or an improvement does not combine by energy sum,
// because the octave-band value of a ratio depends on how energy is distributed
// across the octave and that is exactly what a third-octave spectrum of ratios
// does not say. This function cannot refuse such an input, and the decibel
// quantity types in level.go do not change that: they separate a Level from a
// Delta once a value is in one of them, and a Spectrum holds neither, so a
// spectrum of indices reaches this function looking exactly like a spectrum of
// levels. The name is still the whole of the protection, which is a weaker thing
// than a refusal and is written here rather than left to be discovered. What
// would make it one is a spectrum that carries its quantity, which no issue
// holds today.
//
// It refuses a spectrum whose band set does not carry all three third-octave
// bands of every octave. The core set is such a set, so a core spectrum cannot
// be converted at all: it has 3150 Hz without the rest of the 4000 Hz octave and
// none of the 63 Hz octave. Producing the five octaves it does cover would drop
// a band on the floor, and an octave spectrum standing for less energy than its
// input is the kind of number this project exists not to produce.
func EnergySumToOctave(s Spectrum) (OctaveSpectrum, error) {
	if s.set == 0 || len(s.values) == 0 {
		return OctaveSpectrum{}, fmt.Errorf("%w: this spectrum was never constructed", ErrUnknownBandSet)
	}
	lo, hi, err := s.set.bounds()
	if err != nil {
		return OctaveSpectrum{}, err
	}
	out := make([]float64, 0, octaveCount)
	for n := 0; n < octaveCount; n++ {
		c := octaveCentre(n)
		thirds := make([]Level, 0, 3)
		for _, p := range [3]int{c - 1, c, c + 1} {
			if p < lo || p >= hi {
				return OctaveSpectrum{}, fmt.Errorf("%w: the %d Hz octave needs %d Hz and the %s does not have it",
					ErrOctaveNotWhole, nominalSeries[c], nominalSeries[p], s.set)
			}
			l, err := NewLevel(s.values[p-lo])
			if err != nil {
				return OctaveSpectrum{}, fmt.Errorf("the %d Hz octave: %w", nominalSeries[c], err)
			}
			thirds = append(thirds, l)
		}
		total, err := EnergySum(thirds...)
		if err != nil {
			return OctaveSpectrum{}, fmt.Errorf("the %d Hz octave: %w", nominalSeries[c], err)
		}
		db, err := total.Decibels()
		if err != nil {
			return OctaveSpectrum{}, fmt.Errorf("the %d Hz octave: %w", nominalSeries[c], err)
		}
		out = append(out, db)
	}
	return OctaveSpectrum{values: out}, nil
}

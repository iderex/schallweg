package acoustic

import (
	"errors"
	"fmt"
	"math"
)

// The refusals this package makes, as values a caller can test for rather than
// as sentences a caller has to match. Every one of them is a defect this
// container exists to make impossible to express.
var (
	// ErrUnknownBandSet is a band set that is not one of the two defined ones,
	// which in practice is a zero value nobody filled in.
	ErrUnknownBandSet = errors.New("unknown band set")
	// ErrBandCount is a spectrum built with a number of values that is not the
	// number of bands in its set.
	ErrBandCount = errors.New("wrong number of bands")
	// ErrBandMismatch is a value offered for a band the set does not have, or
	// offered in the wrong order.
	ErrBandMismatch = errors.New("band is not the one expected at that position")
	// ErrNotFinite is a value that is not a finite number. It is refused at
	// construction because a NaN travels through arithmetic silently and comes
	// out of the far end as a result somebody reads.
	ErrNotFinite = errors.New("value is not a finite number")
	// ErrDifferentBandSets is two spectra combined that are not on the same set.
	ErrDifferentBandSets = errors.New("spectra are on different band sets")
	// ErrBandNotInSet is a band from one set used to read a spectrum on another.
	ErrBandNotInSet = errors.New("band does not belong to this spectrum's band set")
)

// A Spectrum is a band set together with exactly one value in every band of
// that set, in the set's order.
//
// What it refuses is the product, and docs/decisions/frequency-bands.md is the
// argument. A spectrum with a band missing cannot be constructed. A spectrum
// with its bands out of order cannot be constructed. Two spectra on different
// sets cannot be combined. A value cannot be read out without naming the band it
// belongs to. None of that is a check somebody has to remember to call; it is
// the only way the type can be used.
//
// The zero Spectrum is on no band set and every operation on it refuses. There
// is no usable empty spectrum, because an empty one is what a partially filled
// one looks like on the way to being wrong.
//
// What the values mean is not decided here. They are the numbers of whatever
// quantity the caller is holding, in decibels for every quantity this project
// computes. The decibel quantity types that stop two levels being added are
// issue #40, and this container is written so that they arrive without it
// changing shape.
type Spectrum struct {
	set BandSet
	// values is never handed out. Bands are read one at a time through At, so
	// nothing can retain a slice into a spectrum and change it afterwards.
	values []float64
}

// New builds a spectrum from band centres and values given together.
//
// This is the only route to a Spectrum, and it takes the nominal centre
// frequencies alongside the values on purpose. A constructor taking a bare list
// of numbers is a constructor whose caller has an assumption about what index
// zero means, and every reader of that call afterwards has to reconstruct the
// assumption. Here the caller states which band each value is for, and a caller
// that has them in a different order, or has a band the set does not contain,
// is refused at the boundary instead of computing something.
//
// nominalCentres are in hertz, ascending, exactly the set's own bands.
func New(set BandSet, nominalCentres []int, values []float64) (Spectrum, error) {
	want, _, err := setNominals(set)
	if err != nil {
		return Spectrum{}, err
	}
	if len(nominalCentres) != len(want) || len(values) != len(want) {
		return Spectrum{}, fmt.Errorf("%w: the %s has %d bands, and %d band centres with %d values were given",
			ErrBandCount, set, len(want), len(nominalCentres), len(values))
	}
	for i, nominal := range nominalCentres {
		if nominal != want[i] {
			return Spectrum{}, fmt.Errorf("%w: position %d of the %s is %d Hz, and %d Hz was given",
				ErrBandMismatch, i, set, want[i], nominal)
		}
	}
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return Spectrum{}, fmt.Errorf("%w: the value at %d Hz is %v", ErrNotFinite, want[i], v)
		}
	}
	held := make([]float64, len(values))
	copy(held, values)
	return Spectrum{set: set, values: held}, nil
}

// setNominals is New's view of a band set: the nominal centres it expects, and
// the refusal for a set that is not one of the two.
func setNominals(set BandSet) ([]int, int, error) {
	nominals := set.Nominals()
	if nominals == nil {
		return nil, 0, fmt.Errorf("%w: %d", ErrUnknownBandSet, uint8(set))
	}
	return nominals, len(nominals), nil
}

// Set is which band set this spectrum is on. It is a property of the value and
// is never inferred from what the values look like.
func (s Spectrum) Set() BandSet { return s.set }

// Len is how many bands the spectrum has, which is how many its set has. A zero
// spectrum has none.
func (s Spectrum) Len() int { return len(s.values) }

// Bands returns the spectrum's bands in ascending order, for a caller that wants
// to walk it. Each one carries its set, so it can only be used to read a
// spectrum on that set.
func (s Spectrum) Bands() []Band { return s.set.Bands() }

// At reads the value in one band.
//
// It takes a Band rather than an integer, so there is no index to be off by one,
// and it refuses a band from another set rather than reading the value that
// happens to sit at the same position. Those two sets share thirteen nominal
// frequencies at different positions, which is exactly the case where a
// positional read returns a plausible number.
func (s Spectrum) At(b Band) (float64, error) {
	if s.set == 0 || len(s.values) == 0 {
		return 0, fmt.Errorf("%w: this spectrum was never constructed", ErrUnknownBandSet)
	}
	if b.set != s.set {
		return 0, fmt.Errorf("%w: the band is from the %s and the spectrum is on the %s", ErrBandNotInSet, b.set, s.set)
	}
	if !b.valid() {
		return 0, fmt.Errorf("%w: %s", ErrBandNotInSet, b)
	}
	lo, _, err := s.set.bounds()
	if err != nil {
		return 0, err
	}
	return s.values[b.series-lo], nil
}

// Combine builds a new spectrum by applying f to the two spectra band by band.
//
// It refuses two spectra on different band sets rather than working on the
// overlap, which is the refusal the whole container is for: the two defined sets
// share thirteen nominal frequencies, so an overlap is always available and
// always wrong. It is deliberately the only combining operation here, and it
// knows nothing about decibels. What the arithmetic is, energy sums and level
// differences, is issue #40, and it is written as functions passed to this
// rather than as methods that would put arithmetic in the container.
func (s Spectrum) Combine(other Spectrum, f func(a, b float64) float64) (Spectrum, error) {
	if s.set == 0 || other.set == 0 {
		return Spectrum{}, fmt.Errorf("%w: one of these spectra was never constructed", ErrUnknownBandSet)
	}
	if s.set != other.set {
		return Spectrum{}, fmt.Errorf("%w: one is on the %s and the other is on the %s", ErrDifferentBandSets, s.set, other.set)
	}
	out := make([]float64, len(s.values))
	for i := range s.values {
		out[i] = f(s.values[i], other.values[i])
	}
	nominals := s.set.Nominals()
	return New(s.set, nominals, out)
}

// Map builds a new spectrum by applying f to every value, on the same band set.
//
// It exists so that applying a correction to a whole spectrum does not need a
// route out of the container and back in. The result goes through New, so a
// function returning a value that is not finite is refused here rather than in
// whatever reads the result later.
func (s Spectrum) Map(f func(v float64) float64) (Spectrum, error) {
	if s.set == 0 {
		return Spectrum{}, fmt.Errorf("%w: this spectrum was never constructed", ErrUnknownBandSet)
	}
	out := make([]float64, len(s.values))
	for i, v := range s.values {
		out[i] = f(v)
	}
	return New(s.set, s.set.Nominals(), out)
}

// String says what the spectrum is on and how many bands it holds, and no
// values. A container that prints its numbers invites a reader to compare two
// printings instead of two spectra.
func (s Spectrum) String() string {
	if s.set == 0 {
		return "spectrum on no band set"
	}
	return fmt.Sprintf("spectrum on the %s, %d bands", s.set, len(s.values))
}

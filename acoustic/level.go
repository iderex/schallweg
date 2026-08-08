package acoustic

// The decibel quantities and the energy arithmetic underneath them.
//
// Everything the kernel computes is one of five operations repeated: convert a
// level to energy, add energies, convert back, take a difference, apply a
// correction. Written once they are checkable. Written at each call site they
// are where the kernel goes wrong in a way no reviewer sees, because the wrong
// version looks exactly like the right one.
//
// docs/decisions/numeric-contract.md decided the shape and this file is its
// implementation. Two things follow from that document and both are the point
// of the file rather than decoration.
//
// A decibel value is a struct. In Go a named type over float64 still supports
// the addition operator, so a Level over float64 would let two sound reduction
// indices be added into a hundred and fifteen decibels, which is the most common
// error in this domain and is invisible in review because it looks exactly like
// arithmetic. A struct with unexported fields does not support the operator at
// all. The mistake is not caught by a check, a review or a test: it does not
// build.
//
// Energy is not a type anybody holds. It exists between energyOf and levelOf,
// both unexported, and every conversion into it in this file is paired with the
// conversion back in the same function body. A stored energy quantity is a
// decibel quantity somebody eventually formats, and the moment two
// representations of one thing can both be printed, one of them gets printed by
// accident.
//
// This file is where the logarithm lives. What that is worth is stated in the
// pull request that added it rather than asserted here, because a comment
// claiming to be the only one of its kind is a claim nothing re-runs.

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// The refusals this file makes, as values a caller can test for.
var (
	// ErrNoQuantity is a Level or a Delta that was never constructed. It is the
	// distinction the whole file is built around: a quantity nobody supplied is
	// not zero decibels, and zero decibels is a real and very quiet value.
	ErrNoQuantity = errors.New("quantity was never constructed")
	// ErrNothingToSum is an energy sum over no levels at all. A sum of nothing
	// is zero energy, which is minus infinity decibels, which is not a level.
	ErrNothingToSum = errors.New("energy sum over no levels")
	// ErrNotBelowTotal is an energy difference where the part is not strictly
	// below the total. Removing a contribution that is not there leaves either
	// no energy or negative energy, and neither is a level.
	ErrNotBelowTotal = errors.New("the part is not below the total")
	// ErrRatioNotPositive is a ratio offered for weighting that is zero or
	// negative. Areas and lengths are positive, so a ratio that is not is a
	// geometry defect arriving here rather than a quantity to convert.
	ErrRatioNotPositive = errors.New("ratio is not positive")
)

// A Level is a quantity in decibels: a sound reduction index, a level difference
// between two rooms, a normalised impact sound pressure level.
//
// The zero value is not zero decibels. It is a level nobody supplied, and every
// operation on it refuses rather than reading it as a very quiet band. That
// distinction is the one docs/decisions/frequency-bands.md calls the single most
// expensive confusion available in this domain: a missing value read as zero
// gives a number that is not merely wrong but confidently wrong.
type Level struct {
	db    float64
	known bool
}

// A Delta is a difference in decibels: an improvement from a lining, a
// laboratory-to-situ correction, a vibration reduction index.
//
// It carries the same unset state as Level and for the same reason. A zero Delta
// is a correction of nothing, which is a defensible value and a common one, so a
// struct field nobody filled in would silently mean "no correction applied"
// rather than "no correction supplied".
type Delta struct {
	db    float64
	known bool
}

// NewLevel builds a level from a value in decibels.
//
// A value that is not finite is refused here rather than downstream, because a
// NaN travels through arithmetic silently and comes out of the far end as a
// number somebody reads.
func NewLevel(db float64) (Level, error) {
	if math.IsNaN(db) || math.IsInf(db, 0) {
		return Level{}, fmt.Errorf("%w: %v dB", ErrNotFinite, db)
	}
	return Level{db: db, known: true}, nil
}

// NewDelta builds a difference from a value in decibels.
func NewDelta(db float64) (Delta, error) {
	if math.IsNaN(db) || math.IsInf(db, 0) {
		return Delta{}, fmt.Errorf("%w: %v dB", ErrNotFinite, db)
	}
	return Delta{db: db, known: true}, nil
}

// Known reports whether this level came from a constructor.
func (l Level) Known() bool { return l.known }

// Decibels is the level's value, and it refuses a level nobody supplied rather
// than returning the zero that a bare field read would give.
func (l Level) Decibels() (float64, error) {
	if !l.known {
		return 0, fmt.Errorf("%w: no level", ErrNoQuantity)
	}
	return l.db, nil
}

// String says the value the way an error message should say it.
func (l Level) String() string {
	if !l.known {
		return "no level"
	}
	return fmt.Sprintf("%g dB", l.db)
}

// Known reports whether this difference came from a constructor.
func (d Delta) Known() bool { return d.known }

// Decibels is the difference's value, and it refuses one nobody supplied.
func (d Delta) Decibels() (float64, error) {
	if !d.known {
		return 0, fmt.Errorf("%w: no difference", ErrNoQuantity)
	}
	return d.db, nil
}

// String says the value the way an error message should say it.
func (d Delta) String() string {
	if !d.known {
		return "no difference"
	}
	return fmt.Sprintf("%+g dB", d.db)
}

// Difference is one level subtracted from another, which is a ratio in energy
// and a difference in decibels, and is a Delta rather than a Level.
//
// This direction is always defined. The subtraction that is not is
// EnergyDifference below, and the two are separate functions with separate names
// so that reaching for the wrong one is a word in a diff rather than a sign.
func (l Level) Difference(other Level) (Delta, error) {
	if !l.known || !other.known {
		return Delta{}, fmt.Errorf("%w: %s minus %s", ErrNoQuantity, l, other)
	}
	return NewDelta(l.db - other.db)
}

// Plus applies a correction to a level and gives a level.
func (l Level) Plus(d Delta) (Level, error) {
	if !l.known || !d.known {
		return Level{}, fmt.Errorf("%w: %s plus %s", ErrNoQuantity, l, d)
	}
	return NewLevel(l.db + d.db)
}

// Minus removes a correction from a level and gives a level.
func (l Level) Minus(d Delta) (Level, error) {
	if !l.known || !d.known {
		return Level{}, fmt.Errorf("%w: %s minus %s", ErrNoQuantity, l, d)
	}
	return NewLevel(l.db - d.db)
}

// Plus accumulates two corrections.
func (d Delta) Plus(e Delta) (Delta, error) {
	if !d.known || !e.known {
		return Delta{}, fmt.Errorf("%w: %s plus %s", ErrNoQuantity, d, e)
	}
	return NewDelta(d.db + e.db)
}

// Negated is the correction in the other direction, which is what an improvement
// becomes when it is taken off instead of put on.
func (d Delta) Negated() (Delta, error) {
	if !d.known {
		return Delta{}, fmt.Errorf("%w: no difference to negate", ErrNoQuantity)
	}
	return NewDelta(-d.db)
}

// energyOf is the conversion into energy, and levelOf is the conversion back.
// They are unexported and they are the only two in this module, so an energy
// quantity exists only inside one function body of this file.
func energyOf(db float64) float64 { return math.Pow(10, db/10) }

func levelOf(energy float64) float64 { return 10 * math.Log10(energy) }

// EnergySum combines levels into a total, which is what combining transmission
// paths is and is never anything else.
//
// A level nobody supplied refuses the whole sum and the message names its
// position. That is the guard this function exists for. An absent contribution
// read as zero decibels is one unit of energy added to a total, which for a
// spectrum of ordinary levels disappears under the others and moves the answer
// by an amount nothing in the output explains. A level that is genuinely far
// below the others is a different thing: it is supplied, it contributes what its
// energy contributes, and the total is right. Both cases are silent in a float64
// and neither is silent here.
//
// The summation order is docs/decisions/numeric-contract.md's: ascending by the
// magnitude of the summand, ties broken by the summand's position in the
// argument list, which makes the order total. Floating point addition is not
// associative, so a sum over one set in two orders can differ in the last
// places, and a corpus that reproduces exactly is one where any movement at all
// is a signal. Ascending also loses the least: adding the small contributions
// first keeps them from disappearing under a large running total, and path
// contributions in this method routinely span several orders of magnitude.
func EnergySum(levels ...Level) (Level, error) {
	if len(levels) == 0 {
		return Level{}, fmt.Errorf("%w: an energy sum needs at least one level", ErrNothingToSum)
	}
	type summand struct {
		energy float64
		at     int
	}
	terms := make([]summand, 0, len(levels))
	for i, l := range levels {
		if !l.known {
			return Level{}, fmt.Errorf("%w: the level at position %d of %d was never supplied, and it is not zero decibels",
				ErrNoQuantity, i, len(levels))
		}
		terms = append(terms, summand{energy: energyOf(l.db), at: i})
	}
	sort.Slice(terms, func(i, j int) bool {
		if terms[i].energy == terms[j].energy {
			return terms[i].at < terms[j].at
		}
		return terms[i].energy < terms[j].energy
	})
	var total float64
	for _, t := range terms {
		total += t.energy
	}
	return NewLevel(levelOf(total))
}

// EnergyDifference removes one contribution from a total and gives what is left.
//
// This is the subtraction that is defined only in one direction. It refuses a
// part that is not strictly below the total, because the energy left is then
// zero or negative and neither of those is a level: zero energy is minus
// infinity decibels and negative energy is not a quantity at all. Returning a
// number there would be an answer to a question nobody can have asked correctly,
// and the case it arises from is a measurement and a background that were not
// what the caller thought they were.
//
// The refusal is made on the decibel values and not on the energies, and there
// was a second check on the energies here that had to come out. Energy is
// monotone in the level, so a part below the total always leaves energy above
// zero, and the energy check was therefore unreachable: deleting it left the
// suite green, which is the state a guard may not ship in. What still stands
// behind the arithmetic if the two conversions ever land on one value is
// NewLevel, because the level of zero energy is minus infinity and NewLevel
// refuses a value that is not finite.
func EnergyDifference(total, part Level) (Level, error) {
	if !total.known || !part.known {
		return Level{}, fmt.Errorf("%w: %s less %s", ErrNoQuantity, total, part)
	}
	if part.db >= total.db {
		return Level{}, fmt.Errorf("%w: the total is %s and the part is %s", ErrNotBelowTotal, total, part)
	}
	return NewLevel(levelOf(energyOf(total.db) - energyOf(part.db)))
}

// RatioDelta is a dimensionless ratio expressed in decibels, which is how an
// area or a coupling length enters the arithmetic.
//
// It is the only route from a plain number into a Delta, so a weighting written
// anywhere else in the kernel is a logarithm somewhere else in the kernel and is
// visible as one. A ratio that is zero or negative is refused rather than
// converted: an area and a length are positive, so a ratio that is not one is a
// geometry defect that has reached the arithmetic, and minus infinity decibels
// is not the answer to it.
func RatioDelta(ratio float64) (Delta, error) {
	if math.IsNaN(ratio) {
		return Delta{}, fmt.Errorf("%w: %v", ErrNotFinite, ratio)
	}
	if ratio <= 0 {
		return Delta{}, fmt.Errorf("%w: %v", ErrRatioNotPositive, ratio)
	}
	return NewDelta(levelOf(ratio))
}

// LevelsOf reads a spectrum out as levels, one per band, in the set's order.
//
// A Spectrum holds plain numbers and does not know what quantity they are, which
// docs/decisions/frequency-bands.md says of it and this function does not
// change. What it does is put the numbers into the type where the arithmetic is
// defined, at one boundary, so that the arithmetic between here and the result
// cannot be written with an operator.
func LevelsOf(s Spectrum) ([]Level, error) {
	if s.set == 0 || len(s.values) == 0 {
		return nil, fmt.Errorf("%w: this spectrum was never constructed", ErrUnknownBandSet)
	}
	out := make([]Level, 0, len(s.values))
	for _, v := range s.values {
		l, err := NewLevel(v)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// SpectrumOfLevels is the way back, and it refuses a level nobody supplied
// rather than writing a zero into a band.
func SpectrumOfLevels(set BandSet, levels []Level) (Spectrum, error) {
	nominals := set.Nominals()
	if nominals == nil {
		return Spectrum{}, fmt.Errorf("%w: %d", ErrUnknownBandSet, uint8(set))
	}
	if len(levels) != len(nominals) {
		return Spectrum{}, fmt.Errorf("%w: the %s has %d bands and %d levels were given",
			ErrBandCount, set, len(nominals), len(levels))
	}
	values := make([]float64, 0, len(levels))
	for i, l := range levels {
		db, err := l.Decibels()
		if err != nil {
			return Spectrum{}, fmt.Errorf("%w at %d Hz", err, nominals[i])
		}
		values = append(values, db)
	}
	return New(set, nominals, values)
}

// EnergySumSpectra combines whole spectra band by band, which is what summing
// the transmission paths of a situation is.
//
// It refuses two spectra on different band sets, by going through the container
// rather than by checking here, so the refusal is the one Spectrum.Combine
// already makes and there is not a second opinion about what a band set is.
func EnergySumSpectra(spectra ...Spectrum) (Spectrum, error) {
	if len(spectra) == 0 {
		return Spectrum{}, fmt.Errorf("%w: an energy sum needs at least one spectrum", ErrNothingToSum)
	}
	set := spectra[0].Set()
	bands := set.Len()
	if bands == 0 {
		return Spectrum{}, fmt.Errorf("%w: the first spectrum was never constructed", ErrUnknownBandSet)
	}
	for _, s := range spectra {
		if s.Set() != set {
			return Spectrum{}, fmt.Errorf("%w: one is on the %s and another is on the %s",
				ErrDifferentBandSets, set, s.Set())
		}
	}
	totals := make([]Level, 0, bands)
	for i := 0; i < bands; i++ {
		column := make([]Level, 0, len(spectra))
		for _, s := range spectra {
			l, err := NewLevel(s.values[i])
			if err != nil {
				return Spectrum{}, err
			}
			column = append(column, l)
		}
		total, err := EnergySum(column...)
		if err != nil {
			return Spectrum{}, fmt.Errorf("at %d Hz: %w", set.Nominals()[i], err)
		}
		totals = append(totals, total)
	}
	return SpectrumOfLevels(set, totals)
}

// Corrected applies a per-band correction to a spectrum, adding the correction's
// value in each band to the spectrum's.
//
// The correction arrives as a spectrum because a correction that varies with
// frequency is what a lining, an in situ correction and a vibration reduction
// index all are. It is added and never energy summed: a correction is a
// difference in decibels, and combining it by energy would be the confusion
// between a Level and a Delta made on whole spectra at once.
func Corrected(s, correction Spectrum) (Spectrum, error) {
	if s.set == 0 || correction.set == 0 {
		return Spectrum{}, fmt.Errorf("%w: one of these spectra was never constructed", ErrUnknownBandSet)
	}
	if s.set != correction.set {
		return Spectrum{}, fmt.Errorf("%w: the spectrum is on the %s and the correction is on the %s",
			ErrDifferentBandSets, s.set, correction.set)
	}
	out := make([]Level, 0, len(s.values))
	for i := range s.values {
		l, err := NewLevel(s.values[i])
		if err != nil {
			return Spectrum{}, err
		}
		d, err := NewDelta(correction.values[i])
		if err != nil {
			return Spectrum{}, err
		}
		applied, err := l.Plus(d)
		if err != nil {
			return Spectrum{}, err
		}
		out = append(out, applied)
	}
	return SpectrumOfLevels(s.set, out)
}

// Weighted applies one area or length ratio to every band of a spectrum.
//
// The ratio is dimensionless and is formed by the caller from two quantities in
// the same unit, because this package holds no geometry and cannot check that
// the numerator and the denominator are the same kind of thing. What it can do
// is refuse a ratio that is not positive, and it does that in RatioDelta rather
// than here so there is one refusal and not two.
func Weighted(s Spectrum, ratio float64) (Spectrum, error) {
	if s.set == 0 {
		return Spectrum{}, fmt.Errorf("%w: this spectrum was never constructed", ErrUnknownBandSet)
	}
	d, err := RatioDelta(ratio)
	if err != nil {
		return Spectrum{}, err
	}
	out := make([]Level, 0, len(s.values))
	for _, v := range s.values {
		l, err := NewLevel(v)
		if err != nil {
			return Spectrum{}, err
		}
		applied, err := l.Plus(d)
		if err != nil {
			return Spectrum{}, err
		}
		out = append(out, applied)
	}
	return SpectrumOfLevels(s.set, out)
}

package acoustic

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/iderex/schallweg/acoustic/approx"
)

// The tolerances below are in decibels, and each one is chosen for what the
// assertion is about rather than copied from the one above it.
//
// exactly is for a value the arithmetic reproduces to the last places this
// project cares about: a sum of two or three terms whose expected value is
// written out in the test from the same operations in the same order. A
// thousandth of a decibel is four orders of magnitude below anything anybody
// measures and still tight enough that a wrong rounding moment fails it.
//
// audible is for an assertion about whether two results differ at all in a way a
// reader would call a difference. It is a tenth of a decibel, which is the
// precision these quantities are reported to.
//
// reordered is for the one assertion that is about the summation order rather
// than about a value: the same terms in a different argument order reaching the
// same answer. The order rule exists to make that exact, and a tolerance of zero
// is refused by approx for good reasons, so this is the smallest number that is
// still a tolerance. It is eleven orders of magnitude below audible, so a
// summation that had genuinely reordered would fail it.
const (
	exactly   = 0.001
	audible   = 0.1
	reordered = 1e-12
)

// mustLevel and mustDelta keep a test's arithmetic in the test rather than
// behind three lines of error handling per value.
func mustLevel(t *testing.T, db float64) Level {
	t.Helper()
	l, err := NewLevel(db)
	if err != nil {
		t.Fatalf("NewLevel(%v): %v", db, err)
	}
	return l
}

func mustDelta(t *testing.T, db float64) Delta {
	t.Helper()
	d, err := NewDelta(db)
	if err != nil {
		t.Fatalf("NewDelta(%v): %v", db, err)
	}
	return d
}

func decibels(t *testing.T, l Level) float64 {
	t.Helper()
	db, err := l.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	return db
}

// TestTheZeroLevelIsNotZeroDecibels is the distinction the whole file is built
// around, asserted before anything computes with it.
func TestTheZeroLevelIsNotZeroDecibels(t *testing.T) {
	var absent Level
	if absent.Known() {
		t.Fatal("a Level nobody constructed reports itself as known")
	}
	if _, err := absent.Decibels(); !errors.Is(err, ErrNoQuantity) {
		t.Fatalf("reading a level nobody supplied gave %v, want ErrNoQuantity", err)
	}

	quiet := mustLevel(t, 0)
	if !quiet.Known() {
		t.Fatal("a level of zero decibels reports itself as unknown")
	}
	approx.Equal(t, "zero decibels", decibels(t, quiet), 0, exactly)
}

// TestNonFiniteValuesAreRefusedAtConstruction keeps a NaN out of the arithmetic
// rather than finding it in a result.
func TestNonFiniteValuesAreRefusedAtConstruction(t *testing.T) {
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := NewLevel(bad); !errors.Is(err, ErrNotFinite) {
			t.Errorf("NewLevel(%v) gave %v, want ErrNotFinite", bad, err)
		}
		if _, err := NewDelta(bad); !errors.Is(err, ErrNotFinite) {
			t.Errorf("NewDelta(%v) gave %v, want ErrNotFinite", bad, err)
		}
	}
}

// TestLevelMinusLevelIsADifference covers the direction that is always defined.
//
// By hand: 62.4 dB minus 55.9 dB is 6.5 dB, and it is a difference rather than a
// level, which the type says and this test cannot.
func TestLevelMinusLevelIsADifference(t *testing.T) {
	got, err := mustLevel(t, 62.4).Difference(mustLevel(t, 55.9))
	if err != nil {
		t.Fatalf("Difference: %v", err)
	}
	db, err := got.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "62.4 dB less 55.9 dB", db, 6.5, exactly)

	// The other direction is defined too, and it is the negative of it.
	back, err := mustLevel(t, 55.9).Difference(mustLevel(t, 62.4))
	if err != nil {
		t.Fatalf("Difference: %v", err)
	}
	backDB, err := back.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "55.9 dB less 62.4 dB", backDB, -6.5, exactly)
}

// TestCorrectionsApplyAndAccumulate covers Level plus Delta, Level minus Delta,
// Delta plus Delta and Delta negated.
//
// By hand: a laboratory index of 53.0 dB with a lining improvement of 4.5 dB and
// an in situ correction of minus 1.2 dB. The two corrections accumulate to
// 3.3 dB, and applied to the index they give 56.3 dB. Taking the same 3.3 dB
// back off gives 53.0 dB again, and negating it gives minus 3.3 dB.
func TestCorrectionsApplyAndAccumulate(t *testing.T) {
	index := mustLevel(t, 53.0)
	lining := mustDelta(t, 4.5)
	situ := mustDelta(t, -1.2)

	together, err := lining.Plus(situ)
	if err != nil {
		t.Fatalf("Delta.Plus: %v", err)
	}
	togetherDB, err := together.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "4.5 dB and -1.2 dB together", togetherDB, 3.3, exactly)

	applied, err := index.Plus(together)
	if err != nil {
		t.Fatalf("Level.Plus: %v", err)
	}
	approx.Equal(t, "53.0 dB with 3.3 dB applied", decibels(t, applied), 56.3, exactly)

	removed, err := applied.Minus(together)
	if err != nil {
		t.Fatalf("Level.Minus: %v", err)
	}
	approx.Equal(t, "56.3 dB with 3.3 dB removed", decibels(t, removed), 53.0, exactly)

	negated, err := together.Negated()
	if err != nil {
		t.Fatalf("Negated: %v", err)
	}
	negatedDB, err := negated.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "3.3 dB negated", negatedDB, -3.3, exactly)
}

// TestEveryOperationRefusesAQuantityNobodySupplied walks the operations with an
// unset operand on each side.
func TestEveryOperationRefusesAQuantityNobodySupplied(t *testing.T) {
	var noLevel Level
	var noDelta Delta
	level := mustLevel(t, 50)
	delta := mustDelta(t, 2)

	if _, err := level.Difference(noLevel); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Difference with an unsupplied level gave %v", err)
	}
	if _, err := noLevel.Difference(level); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Difference on an unsupplied level gave %v", err)
	}
	if _, err := level.Plus(noDelta); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Level.Plus with an unsupplied difference gave %v", err)
	}
	if _, err := level.Minus(noDelta); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Level.Minus with an unsupplied difference gave %v", err)
	}
	if _, err := delta.Plus(noDelta); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Delta.Plus with an unsupplied difference gave %v", err)
	}
	if _, err := noDelta.Negated(); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("Negated on an unsupplied difference gave %v", err)
	}
	if _, err := EnergyDifference(level, noLevel); !errors.Is(err, ErrNoQuantity) {
		t.Errorf("EnergyDifference with an unsupplied part gave %v", err)
	}
}

// TestEnergySumIsAnEnergySum computes the worked example from
// docs/decisions/numeric-contract.md, which is the one place in this repository
// where the rounding rule is written out with every step visible.
//
// By hand: 55.0 dB is 10^5.50 units of energy and 60.1 dB is 10^6.01. Their sum
// is 1339520.7582975917, and ten times its base ten logarithm is
// 61.269494481750932 dB. The value the result carries is that one, at full
// precision, and the rounding to 61.3 dB happens at the edge and not here.
func TestEnergySumIsAnEnergySum(t *testing.T) {
	total, err := EnergySum(mustLevel(t, 55.0), mustLevel(t, 60.1))
	if err != nil {
		t.Fatalf("EnergySum: %v", err)
	}
	approx.Equal(t, "55.0 dB and 60.1 dB summed", decibels(t, total), 61.269494481750932, exactly)

	// Arithmetic addition would give a hundred and fifteen decibels. That it
	// cannot be written is the point of the type, so what is asserted here is
	// only that the energy sum is nowhere near it.
	if decibels(t, total) > 70 {
		t.Fatalf("the energy sum of 55.0 dB and 60.1 dB is %v dB, which is not an energy sum", decibels(t, total))
	}
}

// TestEnergySumDoublesOnTwoEqualLevels is the case a reader can check without a
// calculator.
//
// By hand: two equal levels carry twice the energy of one, and ten times the
// base ten logarithm of two is 3.0102999566398120 dB. So 50.0 dB and 50.0 dB
// together are 53.0102999566398130 dB.
func TestEnergySumDoublesOnTwoEqualLevels(t *testing.T) {
	total, err := EnergySum(mustLevel(t, 50), mustLevel(t, 50))
	if err != nil {
		t.Fatalf("EnergySum: %v", err)
	}
	approx.Equal(t, "50 dB twice", decibels(t, total), 53.010299956639813, exactly)
}

// TestEnergySumOrderDoesNotChangeTheAnswer is what the summation order rule buys.
//
// The terms span nine orders of magnitude in energy, which is the case the rule
// was written for, and they are offered in two different argument orders.
func TestEnergySumOrderDoesNotChangeTheAnswer(t *testing.T) {
	values := []float64{72.4, 18.1, 55.0, 61.9, 33.7, 45.2, 12.0, 68.8}
	forwards := make([]Level, 0, len(values))
	backwards := make([]Level, 0, len(values))
	for i := range values {
		forwards = append(forwards, mustLevel(t, values[i]))
		backwards = append(backwards, mustLevel(t, values[len(values)-1-i]))
	}

	one, err := EnergySum(forwards...)
	if err != nil {
		t.Fatalf("EnergySum: %v", err)
	}
	other, err := EnergySum(backwards...)
	if err != nil {
		t.Fatalf("EnergySum: %v", err)
	}
	approx.Equal(t, "the same eight levels in the other order", decibels(t, other), decibels(t, one), reordered)
}

// TestAnAbsentLevelAndANegligibleLevelAreDifferentInATotal is the guard this
// file leads with, shown in computed totals rather than argued.
//
// Three totals over the same two contributions of 2.0 dB each, differing only in
// what the third contribution is. The magnitudes are deliberately small, because
// that is where the mistake is visible: against a sound reduction index of sixty
// decibels a spurious zero is absorbed, and a test at those magnitudes would
// pass whether or not the unset state was carried.
//
// By hand. Two contributions of 2.0 dB carry 10^0.2 units of energy each, so the
// total is 5.0102999566398125 dB.
//
// Negligible, at minus 40.0 dB, carries 10^-4 against the 1.58 of each of the
// others, and the total moves to 5.0104369651251721 dB, which is the same number
// to a tenth of a decibel. That is what negligible means and the arithmetic gets
// it right without being told.
//
// Zero decibels is one whole unit of energy, and the total becomes
// 6.2011380695797715 dB. It is more than a decibel away from both of the others
// and nothing in the number says so. That total is what an absent contribution
// silently becomes if the unset state is not carried in the type.
//
// Absent is none of those three. It has no total at all: the sum refuses and the
// message names the position.
func TestAnAbsentLevelAndANegligibleLevelAreDifferentInATotal(t *testing.T) {
	one := mustLevel(t, 2.0)
	other := mustLevel(t, 2.0)

	twoAlone, err := EnergySum(one, other)
	if err != nil {
		t.Fatalf("EnergySum: %v", err)
	}
	approx.Equal(t, "two contributions of 2.0 dB", decibels(t, twoAlone), 5.0102999566398125, exactly)

	negligible, err := EnergySum(one, other, mustLevel(t, -40.0))
	if err != nil {
		t.Fatalf("EnergySum with a negligible contribution: %v", err)
	}
	approx.Equal(t, "with a negligible third", decibels(t, negligible), 5.0104369651251721, exactly)
	approx.Equal(t, "a negligible third against no third", decibels(t, negligible), decibels(t, twoAlone), audible)

	asZero, err := EnergySum(one, other, mustLevel(t, 0))
	if err != nil {
		t.Fatalf("EnergySum with a zero decibel contribution: %v", err)
	}
	approx.Equal(t, "with a third at zero decibels", decibels(t, asZero), 6.2011380695797715, exactly)
	if math.Abs(decibels(t, asZero)-decibels(t, negligible)) <= audible {
		t.Fatalf("a zero decibel contribution and a negligible one reach %v dB and %v dB, and this test needs them to differ",
			decibels(t, asZero), decibels(t, negligible))
	}

	var absent Level
	_, err = EnergySum(one, other, absent)
	if !errors.Is(err, ErrNoQuantity) {
		t.Fatalf("EnergySum with an absent third gave %v, want ErrNoQuantity", err)
	}
	if !strings.Contains(err.Error(), "position 2") {
		t.Fatalf("the refusal says %q and does not name the position", err.Error())
	}
}

// TestReadingAnAbsentLevelAsZeroIsWorstWhereTheAnswerIsADifference is the same
// guard on the operation where the substitution is not absorbed at all.
//
// By hand: a sending level of 53.0 dB against a receiving level of 48.0 dB is a
// difference of 5.0 dB. If the receiving level were absent and read as zero
// decibels, the difference would come out as 53.0 dB, which is forty-eight
// decibels wrong and is a perfectly ordinary looking number for the quantity.
func TestReadingAnAbsentLevelAsZeroIsWorstWhereTheAnswerIsADifference(t *testing.T) {
	sending := mustLevel(t, 53.0)
	got, err := sending.Difference(mustLevel(t, 48.0))
	if err != nil {
		t.Fatalf("Difference: %v", err)
	}
	db, err := got.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "53.0 dB less 48.0 dB", db, 5.0, exactly)

	var absent Level
	if _, err := sending.Difference(absent); !errors.Is(err, ErrNoQuantity) {
		t.Fatalf("a difference against an absent level gave %v, want ErrNoQuantity", err)
	}
}

// TestEnergySumRefusesNoLevelsAtAll covers the empty sum, which is zero energy
// and therefore not a level.
func TestEnergySumRefusesNoLevelsAtAll(t *testing.T) {
	if _, err := EnergySum(); !errors.Is(err, ErrNothingToSum) {
		t.Fatalf("EnergySum() gave %v, want ErrNothingToSum", err)
	}
}

// TestEnergyDifferenceRemovesAContribution is the subtraction that is defined
// only in one direction.
//
// By hand: a total of 61.269494481750932 dB is the sum of 55.0 dB and 60.1 dB,
// from the worked example above. Removing the 60.1 dB path leaves the 55.0 dB
// one.
func TestEnergyDifferenceRemovesAContribution(t *testing.T) {
	total := mustLevel(t, 61.269494481750932)
	left, err := EnergyDifference(total, mustLevel(t, 60.1))
	if err != nil {
		t.Fatalf("EnergyDifference: %v", err)
	}
	approx.Equal(t, "the total less the 60.1 dB path", decibels(t, left), 55.0, exactly)
}

// TestEnergyDifferenceRefusesTheDirectionThatIsNotDefined covers the part equal
// to the total, which leaves zero energy, and the part above it, which leaves
// negative energy.
func TestEnergyDifferenceRefusesTheDirectionThatIsNotDefined(t *testing.T) {
	total := mustLevel(t, 55.0)
	if _, err := EnergyDifference(total, mustLevel(t, 55.0)); !errors.Is(err, ErrNotBelowTotal) {
		t.Errorf("removing the whole of a total gave %v, want ErrNotBelowTotal", err)
	}
	if _, err := EnergyDifference(total, mustLevel(t, 60.0)); !errors.Is(err, ErrNotBelowTotal) {
		t.Errorf("removing more than a total gave %v, want ErrNotBelowTotal", err)
	}
}

// TestRatioDeltaIsTheOneRouteFromARatioIntoDecibels covers the weighting term.
//
// By hand: an element of 12 square metres against a separating element of 10
// gives a ratio of 1.2, and ten times the base ten logarithm of 1.2 is
// 0.7918124604762482 dB. A ratio of one is no correction at all.
func TestRatioDeltaIsTheOneRouteFromARatioIntoDecibels(t *testing.T) {
	d, err := RatioDelta(12.0 / 10.0)
	if err != nil {
		t.Fatalf("RatioDelta: %v", err)
	}
	db, err := d.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "a ratio of 1.2 in decibels", db, 0.7918124604762482, exactly)

	one, err := RatioDelta(1)
	if err != nil {
		t.Fatalf("RatioDelta(1): %v", err)
	}
	oneDB, err := one.Decibels()
	if err != nil {
		t.Fatalf("Decibels: %v", err)
	}
	approx.Equal(t, "a ratio of one in decibels", oneDB, 0, exactly)

	for _, bad := range []float64{0, -1, -0.5} {
		if _, err := RatioDelta(bad); !errors.Is(err, ErrRatioNotPositive) {
			t.Errorf("RatioDelta(%v) gave %v, want ErrRatioNotPositive", bad, err)
		}
	}
	if _, err := RatioDelta(math.NaN()); !errors.Is(err, ErrNotFinite) {
		t.Errorf("RatioDelta(NaN) gave %v, want ErrNotFinite", err)
	}
}

// coreSpectrum builds a core spectrum with one value in every band, for the
// spectrum-level operations below.
func coreSpectrum(t *testing.T, value float64) Spectrum {
	t.Helper()
	nominals := Core.Nominals()
	values := make([]float64, len(nominals))
	for i := range values {
		values[i] = value
	}
	s, err := New(Core, nominals, values)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

// TestLevelsOfAndBackAgain is the boundary between the container and the
// arithmetic, and it has to lose nothing in either direction.
func TestLevelsOfAndBackAgain(t *testing.T) {
	s := coreSpectrum(t, 48.3)
	levels, err := LevelsOf(s)
	if err != nil {
		t.Fatalf("LevelsOf: %v", err)
	}
	if len(levels) != Core.Len() {
		t.Fatalf("LevelsOf gave %d levels for a set of %d bands", len(levels), Core.Len())
	}
	back, err := SpectrumOfLevels(Core, levels)
	if err != nil {
		t.Fatalf("SpectrumOfLevels: %v", err)
	}
	for _, b := range back.Bands() {
		got, err := back.At(b)
		if err != nil {
			t.Fatalf("At(%s): %v", b, err)
		}
		approx.Equal(t, b.String(), got, 48.3, exactly)
	}
}

// TestSpectrumOfLevelsRefusesALevelNobodySupplied keeps the unset state from
// becoming a zero on the way back into the container, and names the band.
func TestSpectrumOfLevelsRefusesALevelNobodySupplied(t *testing.T) {
	levels := make([]Level, Core.Len())
	for i := range levels {
		levels[i] = mustLevel(t, 40)
	}
	levels[4] = Level{}
	_, err := SpectrumOfLevels(Core, levels)
	if !errors.Is(err, ErrNoQuantity) {
		t.Fatalf("SpectrumOfLevels with an unsupplied band gave %v, want ErrNoQuantity", err)
	}
	if want := "250 Hz"; !strings.Contains(err.Error(), want) {
		t.Fatalf("the refusal says %q and does not name %q", err.Error(), want)
	}
}

// TestEnergySumSpectraSumsBandByBand combines three paths, each flat, so the
// expected value can be written out.
//
// By hand: three equal levels carry three times the energy of one, and ten times
// the base ten logarithm of three is 4.7712125471966244 dB. So three flat
// spectra at 50.0 dB reach 54.771212547196619 dB in every band.
func TestEnergySumSpectraSumsBandByBand(t *testing.T) {
	one := coreSpectrum(t, 50)
	total, err := EnergySumSpectra(one, one, one)
	if err != nil {
		t.Fatalf("EnergySumSpectra: %v", err)
	}
	for _, b := range total.Bands() {
		got, err := total.At(b)
		if err != nil {
			t.Fatalf("At(%s): %v", b, err)
		}
		approx.Equal(t, b.String(), got, 54.771212547196619, exactly)
	}
}

// TestEnergySumSpectraRefusesTwoBandSets is the refusal the container makes,
// reached through this function so there is one opinion about it and not two.
func TestEnergySumSpectraRefusesTwoBandSets(t *testing.T) {
	core := coreSpectrum(t, 50)
	extendedNominals := Extended.Nominals()
	values := make([]float64, len(extendedNominals))
	for i := range values {
		values[i] = 50
	}
	extended, err := New(Extended, extendedNominals, values)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := EnergySumSpectra(core, extended); !errors.Is(err, ErrDifferentBandSets) {
		t.Fatalf("summing a core and an extended spectrum gave %v, want ErrDifferentBandSets", err)
	}
	if _, err := EnergySumSpectra(); !errors.Is(err, ErrNothingToSum) {
		t.Fatalf("EnergySumSpectra() gave %v, want ErrNothingToSum", err)
	}
}

// TestCorrectedAddsAndDoesNotEnergySum is the difference between a Level and a
// Delta made on whole spectra.
//
// By hand: 50.0 dB with a correction of 3.0 dB is 53.0 dB. An energy sum of
// 50.0 dB and 3.0 dB would be 50.0043 dB, so the two are far apart and the test
// asserts the first rather than merely that a number came back.
func TestCorrectedAddsAndDoesNotEnergySum(t *testing.T) {
	s := coreSpectrum(t, 50)
	correction := coreSpectrum(t, 3)
	got, err := Corrected(s, correction)
	if err != nil {
		t.Fatalf("Corrected: %v", err)
	}
	for _, b := range got.Bands() {
		v, err := got.At(b)
		if err != nil {
			t.Fatalf("At(%s): %v", b, err)
		}
		approx.Equal(t, b.String(), v, 53.0, exactly)
	}

	extendedNominals := Extended.Nominals()
	values := make([]float64, len(extendedNominals))
	extended, err := New(Extended, extendedNominals, values)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := Corrected(s, extended); !errors.Is(err, ErrDifferentBandSets) {
		t.Fatalf("correcting a core spectrum with an extended one gave %v, want ErrDifferentBandSets", err)
	}
}

// TestWeightedAppliesOneRatioToEveryBand covers the area and length term on a
// whole spectrum.
//
// By hand: a ratio of 1.2 is 0.7918124604762482 dB, so 50.0 dB becomes
// 50.7918124604762482 dB in every band.
func TestWeightedAppliesOneRatioToEveryBand(t *testing.T) {
	s := coreSpectrum(t, 50)
	got, err := Weighted(s, 12.0/10.0)
	if err != nil {
		t.Fatalf("Weighted: %v", err)
	}
	for _, b := range got.Bands() {
		v, err := got.At(b)
		if err != nil {
			t.Fatalf("At(%s): %v", b, err)
		}
		approx.Equal(t, b.String(), v, 50.7918124604762482, exactly)
	}
	if _, err := Weighted(s, 0); !errors.Is(err, ErrRatioNotPositive) {
		t.Fatalf("Weighted with a ratio of zero gave %v, want ErrRatioNotPositive", err)
	}
}

// TestTheUnconstructedSpectrumIsRefusedEverywhere walks the spectrum-level
// operations with a spectrum nobody built.
func TestTheUnconstructedSpectrumIsRefusedEverywhere(t *testing.T) {
	var none Spectrum
	s := coreSpectrum(t, 50)
	if _, err := LevelsOf(none); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("LevelsOf on an unconstructed spectrum gave %v", err)
	}
	if _, err := EnergySumSpectra(none); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("EnergySumSpectra on an unconstructed spectrum gave %v", err)
	}
	if _, err := Corrected(none, s); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("Corrected on an unconstructed spectrum gave %v", err)
	}
	if _, err := Weighted(none, 1); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("Weighted on an unconstructed spectrum gave %v", err)
	}
	if _, err := SpectrumOfLevels(0, nil); !errors.Is(err, ErrUnknownBandSet) {
		t.Errorf("SpectrumOfLevels on an unknown band set gave %v", err)
	}
	if _, err := SpectrumOfLevels(Core, nil); !errors.Is(err, ErrBandCount) {
		t.Errorf("SpectrumOfLevels with no levels gave %v", err)
	}
}

// TestEverySpectrumOperationReturnsAWholeSpectrum requires each operation that
// returns a spectrum to return one on the argument's band set, with a value in
// every band of it.
//
// Every other test here reads its result by walking got.Bands(), and the zero
// Spectrum is on no band set and has no bands. An operation that returned an
// empty spectrum and no error would therefore run those loops zero times,
// assert nothing at all, and pass. That is not hypothetical: negating the error
// check inside Weighted, inside Corrected and inside EnergySumSpectra leaves
// every test in this package green, which is how this one came to be written.
func TestEverySpectrumOperationReturnsAWholeSpectrum(t *testing.T) {
	s := coreSpectrum(t, 50)
	other := coreSpectrum(t, 1)

	cases := []struct {
		name string
		run  func() (Spectrum, error)
	}{
		{"Weighted", func() (Spectrum, error) { return Weighted(s, 12.0/10.0) }},
		{"Corrected", func() (Spectrum, error) { return Corrected(s, other) }},
		{"EnergySumSpectra", func() (Spectrum, error) { return EnergySumSpectra(s, other) }},
	}

	for _, c := range cases {
		got, err := c.run()
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if got.Set() != s.Set() {
			t.Errorf("%s returned a spectrum on the %s, want the %s", c.name, got.Set(), s.Set())
		}
		if got.Len() != s.Len() {
			t.Errorf("%s returned %d band(s), want %d", c.name, got.Len(), s.Len())
		}
		if len(got.Bands()) != s.Len() {
			t.Errorf("%s returned a spectrum naming %d band(s), want %d", c.name, len(got.Bands()), s.Len())
		}
	}
}

// Package approx is how a test in this repository compares a computed number
// against an expected one.
//
// It exists because of a specific way numerical projects grow flaky tests. A
// test asserting that a computed value equals a literal passes on the machine
// it was written on, and then a change to the order of a summation, a different
// architecture, or a compiler that fuses two operations moves the last bit and
// the test fails for a reason that has nothing to do with the number being
// wrong. The repair, applied under pressure, is usually a tolerance somebody
// invents on the spot and does not write down.
//
// So the tolerance is a required argument. It is a decision the author of the
// test makes and states, in the units of the quantity, and a reader of the test
// can see how much agreement was being asked for without reconstructing it from
// the representation.
//
// A tolerance of zero is refused. It is a bare equality wearing a helper's name,
// and it would put this package's own name behind exactly the assertion the
// package exists to prevent. A test that genuinely wants bit-identical values is
// not comparing a computed number and does not belong here.
//
// The rule that no test compares a computed value any other way is enforced in
// cmd/gate: TestOrdinarySuiteObeysItsOwnRules refuses a floating point equality
// anywhere in a test file. What that check catches and what it does not is
// stated there.
package approx

import (
	"fmt"
	"math"
	"testing"
)

// Check reports why got and want do not agree to within tolerance, or nil when
// they do.
//
// It is separate from Equal so that this package's own behaviour is testable.
// testing.TB cannot be implemented outside the testing package, so a test that
// wants to observe a failure rather than suffer one has to reach the decision
// without the reporting around it.
func Check(what string, got, want, tolerance float64) error {
	if tolerance <= 0 {
		return fmt.Errorf("%s: tolerance %v is not positive; a tolerance of zero is a bare equality and this package refuses it", what, tolerance)
	}
	if math.IsInf(tolerance, 0) || math.IsNaN(tolerance) {
		return fmt.Errorf("%s: tolerance %v is not a finite number", what, tolerance)
	}
	if math.IsNaN(got) || math.IsNaN(want) {
		return fmt.Errorf("%s: got %v, want %v; a comparison against NaN is never satisfied and is reported rather than passed", what, got, want)
	}
	if math.IsInf(got, 0) || math.IsInf(want, 0) {
		if got == want {
			return nil
		}
		return fmt.Errorf("%s: got %v, want %v", what, got, want)
	}
	if diff := math.Abs(got - want); diff > tolerance {
		return fmt.Errorf("%s: got %v, want %v, difference %v exceeds tolerance %v", what, got, want, diff, tolerance)
	}
	return nil
}

// Equal fails tb when got and want do not agree to within tolerance.
//
// what names the quantity, so a failure says which number moved rather than
// only that one did. tolerance is in the units of the quantity.
func Equal(tb testing.TB, what string, got, want, tolerance float64) {
	tb.Helper()
	if err := Check(what, got, want, tolerance); err != nil {
		tb.Error(err)
	}
}

// Slice fails tb when got and want differ in length, or when any pair of
// elements does not agree to within tolerance.
//
// It reports every element that differs rather than stopping at the first,
// because a spectrum that is wrong in one band and a spectrum that is wrong in
// all of them are different defects and the first failure looks the same in
// both.
func Slice(tb testing.TB, what string, got, want []float64, tolerance float64) {
	tb.Helper()
	if len(got) != len(want) {
		tb.Errorf("%s: got %d values, want %d", what, len(got), len(want))
		return
	}
	for i := range got {
		if err := Check(fmt.Sprintf("%s[%d]", what, i), got[i], want[i], tolerance); err != nil {
			tb.Error(err)
		}
	}
}

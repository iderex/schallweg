package approx

import (
	"math"
	"strings"
	"testing"
)

// TestCheckRefusesATolerenceThatIsNotADecision is the leg that makes this
// package worth having. A helper that accepted a tolerance of zero would let a
// bare equality through under a name that says it is not one, which is worse
// than the bare equality because it reads as careful.
func TestCheckRefusesATolerenceThatIsNotADecision(t *testing.T) {
	for _, tol := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		err := Check("q", 1, 1, tol)
		if err == nil {
			t.Errorf("Check with tolerance %v returned nil, want a refusal", tol)
		}
	}
}

// TestCheckAgreesAndDisagreesAtTheBoundary pins the comparison at the edge,
// because a helper that is off by one at the boundary is a helper that passes a
// test it should fail exactly when the difference is the size somebody chose to
// care about.
func TestCheckAgreesAndDisagreesAtTheBoundary(t *testing.T) {
	cases := []struct {
		name      string
		got, want float64
		tol       float64
		wantErr   bool
	}{
		{"inside", 1.0, 1.05, 0.1, false},
		{"exactly at the tolerance", 1.0, 1.25, 0.25, false},
		{"outside", 1.0, 1.3, 0.1, true},
		{"identical", 61.269494481750932, 61.269494481750932, 0.001, false},
		{"sign does not matter", 1.3, 1.0, 0.1, true},

		// The case that decides whether the comparison is honest. Neither 1.1
		// nor 0.1 is exactly representable in binary, and 1.1 minus 1.0 comes
		// out as 0.10000000000000009, which is larger than the tolerance. So
		// this pair does NOT agree to 0.1 and the helper says so, rather than
		// widening the comparison until the arithmetic stops being visible.
		// Anybody who wanted these two to agree has to write a tolerance that
		// says so, which is the whole reason the argument is required.
		{"a tolerance the representation cannot reach", 1.0, 1.1, 0.1, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Check(c.name, c.got, c.want, c.tol)
			if (err != nil) != c.wantErr {
				t.Fatalf("Check(%v, %v, %v) error = %v, want error: %v", c.got, c.want, c.tol, err, c.wantErr)
			}
		})
	}
}

// TestCheckRefusesNaN states the one case where the arithmetic would otherwise
// decide it. Every comparison against NaN is false, so a difference test against
// one passes by accident under some spellings and fails under others. This
// package reports it instead.
func TestCheckRefusesNaN(t *testing.T) {
	for _, c := range []struct{ got, want float64 }{
		{math.NaN(), 1},
		{1, math.NaN()},
		{math.NaN(), math.NaN()},
	} {
		err := Check("q", c.got, c.want, 0.1)
		if err == nil {
			t.Errorf("Check(%v, %v) returned nil, want a refusal", c.got, c.want)
		} else if !strings.Contains(err.Error(), "NaN") {
			t.Errorf("Check(%v, %v) error = %q, want it to name NaN", c.got, c.want, err)
		}
	}
}

// TestCheckHandlesInfinity keeps an infinite value from being compared by a
// subtraction that produces NaN and then reads as a difference.
func TestCheckHandlesInfinity(t *testing.T) {
	if err := Check("q", math.Inf(1), math.Inf(1), 0.1); err != nil {
		t.Errorf("two identical infinities disagreed: %v", err)
	}
	if err := Check("q", math.Inf(1), math.Inf(-1), 0.1); err == nil {
		t.Error("opposite infinities agreed, want a refusal")
	}
	if err := Check("q", math.Inf(1), 1, 0.1); err == nil {
		t.Error("an infinity and a finite value agreed, want a refusal")
	}
}

// TestCheckNamesTheQuantity holds the part of the failure message a reader
// actually needs, which is which number moved rather than that one did.
func TestCheckNamesTheQuantity(t *testing.T) {
	err := Check("R_w at 500 Hz", 52, 55, 0.5)
	if err == nil {
		t.Fatal("Check returned nil, want a refusal")
	}
	if !strings.Contains(err.Error(), "R_w at 500 Hz") {
		t.Errorf("error = %q, want it to name the quantity", err)
	}
}

// TestSliceReportsEveryElementThatDiffers separates a spectrum wrong in one band
// from a spectrum wrong in all of them. Stopping at the first difference makes
// those two look identical to whoever reads the failure.
func TestSliceReportsEveryElementThatDiffers(t *testing.T) {
	got := []float64{1, 2, 3}
	want := []float64{1, 9, 9}

	rec := &recorder{}
	Slice(rec, "spectrum", got, want, 0.1)

	if rec.errors != 2 {
		t.Errorf("Slice reported %d differences, want 2", rec.errors)
	}
}

// TestSliceRefusesALengthMismatch keeps a shorter slice from comparing equal on
// the elements it happens to have.
func TestSliceRefusesALengthMismatch(t *testing.T) {
	rec := &recorder{}
	Slice(rec, "spectrum", []float64{1, 2}, []float64{1, 2, 3}, 0.1)
	if rec.errors == 0 {
		t.Error("Slice accepted slices of different lengths")
	}
}

// recorder counts failures instead of causing them. It embeds a real
// *testing.T so that it satisfies testing.TB, which carries an unexported
// method and therefore cannot be implemented from outside the testing package,
// and it overrides the two methods Slice uses.
type recorder struct {
	*testing.T
	errors int
}

func (r *recorder) Helper()               {}
func (r *recorder) Error(args ...any)     { r.errors++ }
func (r *recorder) Errorf(string, ...any) { r.errors++ }

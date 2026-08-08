// Package acoustic is the numeric floor: bands, band sets, spectra, and the
// decibel and energy quantities the rest of this module computes with.
//
// It imports nothing else from this module and it performs no input or output.
// Everything it computes on arrives as an argument. That is what makes a
// validation case reproduce on any machine, and it is what makes a case that
// does not reproduce a defect in the arithmetic rather than in the surroundings.
//
// Two decisions govern what goes here. docs/decisions/frequency-bands.md fixes
// what a spectrum is: a band set with exactly one value in every band of it, so
// that a missing band, a band read twice and two spectra combined across
// different sets are all things that cannot be written down.
// docs/decisions/numeric-contract.md fixes the arithmetic: a decibel quantity is
// a struct rather than a named float, so adding two decibel values does not
// compile, and every sum runs in one defined order.
//
// The package is empty today. Issue #39 adds the spectrum type, issue #40 the
// band arithmetic, and issue #41 the conversions between octave and third-octave
// bands.
package acoustic

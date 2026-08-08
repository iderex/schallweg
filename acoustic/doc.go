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
// What is here today is the container: BandSet, Band and Spectrum, together with
// the refusals a caller can test for. It holds values and knows nothing about
// what they mean. Beside it, in octave.go, are the octave bands and the one
// conversion that is defined between them and third-octave bands, and in
// level.go the decibel quantities and the energy arithmetic every calculation in
// this module is made of.
package acoustic

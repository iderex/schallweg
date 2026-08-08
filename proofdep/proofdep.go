// Package proofdep exists on this branch only. It imports one third-party
// module so that this module has a lock file to tamper with, because on a tree
// with no dependencies there is nothing for the checksum rule to refuse and the
// rule can only be asserted.
//
// It is not part of this project. The branch carrying it is a demonstration and
// is not for merge.
package proofdep

import "github.com/google/go-cmp/cmp"

// SameSpectrum reports whether two slices of band values are equal. The
// comparison itself is beside the point; the import is the point.
func SameSpectrum(a, b []float64) bool {
	return cmp.Equal(a, b)
}

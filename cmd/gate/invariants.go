package main

// Two invariants of the calculation, read out of the syntax rather than out of
// the text.
//
// Both are stated in prose elsewhere in this project and neither was refused by
// anything. A spectrum is built in one place, so that the band count, the band
// order and the finiteness of every value are checked in one place; and a band
// centre frequency is asked for rather than written down, so that the series
// exists once and a calculation cannot carry a second copy of it that drifts.
//
// Why the syntax and not a text search. Both rules have a text form that reads
// as if it would work and fires on prose. The comment above absentBands in
// store/spectrum.go names 100 Hz, 50 Hz and 3150 Hz in one sentence, and a
// pattern over the text refuses it:
//
//	git grep -n '3150' -- store/spectrum.go
//
// A parser sees a comment as a comment, a frequency inside a string as a string,
// and a renamed import as the package it renames. That is the whole of what this
// file buys over a pattern, and it is why these two rules live beside the
// layering reader rather than in the pattern analyser's rule set.
//
// Neither rule reads a test file. A test that proves a spectrum is refused has
// to build the refused thing, and a test that proves a band is read correctly
// has to name the band it is reading, so both rules would refuse the proof
// rather than the defect.

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/iderex/schallweg/acoustic"
)

// containerRoutes are the numeric floor's container types and the one function
// each of them may be filled in.
//
// Outside the floor this rule has no work to do, because the fields of both
// types are unexported and the compiler already refuses a filled literal of one.
// Inside the floor nothing refused a second route, and a second route is what
// the container exists against: New checks the band count, the band order and
// that every value is finite, and a composite literal written beside it checks
// none of the three and produces a value indistinguishable from a good one.
//
// The map is written out rather than derived, because a type's route is a
// decision about that type rather than a property the source carries. What that
// costs is that a container type added to the floor without an entry here is not
// covered, which is the same cost the layering table carries and is why both are
// small enough to read.
var containerRoutes = map[string]string{
	"Spectrum":       "New",
	"OctaveSpectrum": "EnergySumToOctave",
}

// bandCentres are the nominal centre frequencies a calculation may not write
// down, taken from the band set that holds every band either set has.
//
// It is derived rather than restated. A list here would be a second copy of the
// series in a file whose whole purpose is to refuse second copies of the series,
// and it would drift the first time a band was added.
func bandCentres() map[int64]bool {
	out := map[int64]bool{}
	for _, nominal := range acoustic.Extended.Nominals() {
		out[int64(nominal)] = true
	}
	return out
}

// checkInvariants reports both rules for one parsed file.
//
// l decides which of the two applies. The container rule is the floor's, because
// that is the only layer where a filled literal compiles at all. The band centre
// rule is the kernel's and the data layer's: the floor is where the series is
// defined, and the command line and the harnesses hold percentages, byte counts
// and exit codes that are numbers rather than frequencies.
func checkInvariants(path string, fset *token.FileSet, file *ast.File, l layer) []violation {
	var found []violation
	add := func(pos token.Pos, rule, detail string) {
		found = append(found, violation{Path: path, Line: fset.Position(pos).Line, Rule: rule, Detail: detail})
	}

	if l == layerAcoustic {
		findFilledContainers(file, add)
	}
	if l == layerKernel || l == layerStore {
		findWrittenCentres(file, add)
	}
	return found
}

// findFilledContainers reports every composite literal of a container type that
// sets a field somewhere other than that type's route.
//
// An empty literal is not one. `return Spectrum{}, err` is how every function in
// the floor returns a failure, it carries no values, and every operation on the
// value it produces refuses. Refusing it would refuse the error path in every
// function of the layer.
func findFilledContainers(file *ast.File, add func(token.Pos, string, string)) {
	// A literal outside every function body is reported against the empty name,
	// which matches no route, so a package level var holding a filled container
	// is refused rather than invisible.
	inspect := func(enclosing string, n ast.Node) {
		ast.Inspect(n, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok || len(lit.Elts) == 0 {
				return true
			}
			name, ok := lit.Type.(*ast.Ident)
			if !ok {
				return true
			}
			route, held := containerRoutes[name.Name]
			if !held || route == enclosing {
				return true
			}
			add(lit.Pos(), "container-built-outside-its-route",
				fmt.Sprintf("fills a %s here; %s is the one route into it, and it is what checks the band count, "+
					"the band order and that every value is finite", name.Name, route))
			return true
		})
	}

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil {
				inspect(fn.Name.Name, fn.Body)
			}
			continue
		}
		inspect("", decl)
	}
}

// findWrittenCentres reports every integer literal spelled like a nominal band
// centre.
//
// A named constant is not one, and that is the repair rather than a hole. A
// declared name is a thing a reviewer reads beside the number and a thing a
// later reader can follow, and the shape this rule exists against is the centre
// written inline in an expression: a comparison against 1000, a slice of the
// series rebuilt by hand, an index derived from a frequency somebody typed.
//
// What it does not reach, in one place rather than discovered later. A centre
// written as a floating point literal, or computed, is invisible to it. A number
// that is a band centre by coincidence, a thousand millimetres in a metre among
// them, is refused and has to be given a name; that is a cost this rule imposes
// on the two calculation layers and on no other.
func findWrittenCentres(file *ast.File, add func(token.Pos, string, string)) {
	centres := bandCentres()
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			// A const declaration is where the repair lives, so the reader does
			// not descend into one. This reaches a const inside a function body
			// as well, because the parser reports that as a declaration too.
			return x.Tok != token.CONST
		case *ast.BasicLit:
			if x.Kind != token.INT {
				return true
			}
			// Base zero, so an underscore group and a hexadecimal spelling of a
			// centre are read as the number they are rather than skipped.
			value, err := strconv.ParseInt(x.Value, 0, 64)
			if err != nil || !centres[value] {
				return true
			}
			add(x.Pos(), "band-centre-written-out",
				fmt.Sprintf("writes %s, which is a nominal band centre; ask the band set for it, "+
					"through Nominals or through the band itself, or give this number a name if it is not a frequency", x.Value))
		}
		return true
	})
}

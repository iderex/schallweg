package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The two invariants, each proved by breaking it deliberately and each given a
// near miss that comes as close as the language allows without breaking it.
//
// The way a rule like this fails is by matching nothing, which looks exactly
// like a clean tree, so every test below asserts a count rather than that
// something was found.

// countInvariants reads a fixture from this package's testdata directory as the
// given layer and counts the findings by rule.
func countInvariants(t *testing.T, fixture string, l layer) map[string]int {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "invariants", fixture))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	found, err := checkSource(fixture, src, l)
	if err != nil {
		t.Fatalf("reading %s: %v", fixture, err)
	}
	return rules(found)
}

// TestAContainerFilledOutsideItsRouteIsRefused breaks the rule that a spectrum
// is built in one place.
//
// Three spellings, and the package level one is the one worth having: a filled
// container in a var declaration is outside every function, so a reader that
// only walked function bodies would report two of these and call the file clean.
func TestAContainerFilledOutsideItsRouteIsRefused(t *testing.T) {
	got := countInvariants(t, "builds-and-writes.go.txt", layerAcoustic)
	if got["container-built-outside-its-route"] != 3 {
		t.Errorf("the container rule fired %d times, want 3", got["container-built-outside-its-route"])
	}

	// Outside the floor the compiler already refuses a filled literal, because
	// the fields of both containers are unexported, so the rule has nothing to
	// say there and says nothing.
	for _, l := range []layer{layerKernel, layerStore, layerCmd, layerHarness} {
		if outside := countInvariants(t, "builds-and-writes.go.txt", l)["container-built-outside-its-route"]; outside != 0 {
			t.Errorf("the container rule fired %d times in %s", outside, l)
		}
	}
}

// TestABandCentreWrittenIntoACalculationIsRefused breaks the rule that a centre
// frequency is asked for rather than written down.
//
// Five, and two of them are one number in two spellings. A reader comparing the
// literal text would count the decimal one and miss both the underscore group
// and the hexadecimal.
func TestABandCentreWrittenIntoACalculationIsRefused(t *testing.T) {
	for _, l := range []layer{layerKernel, layerStore} {
		if got := countInvariants(t, "builds-and-writes.go.txt", l)["band-centre-written-out"]; got != 5 {
			t.Errorf("the band centre rule fired %d times in %s, want 5", got, l)
		}
	}

	// The floor is where the series is defined, and the command line and the
	// harnesses hold percentages, byte counts and exit codes.
	for _, l := range []layer{layerAcoustic, layerCmd, layerHarness} {
		if got := countInvariants(t, "builds-and-writes.go.txt", l)["band-centre-written-out"]; got != 0 {
			t.Errorf("the band centre rule fired %d times in %s", got, l)
		}
	}
}

// TestTheInvariantsPassTheNearMiss is the half that decides whether the two
// rules are livable. The fixture holds both routes, an empty literal in every
// error path, a type whose name starts with a container's, a centre in a const
// block, a centre inside a string and three centres inside a comment.
func TestTheInvariantsPassTheNearMiss(t *testing.T) {
	for _, l := range []layer{layerAcoustic, layerKernel, layerStore, layerCmd, layerHarness} {
		if got := countInvariants(t, "the-near-miss.go.txt", l); len(got) != 0 {
			t.Errorf("the near miss produced findings in %s: %v", l, got)
		}
	}
}

// TestTheCentreSeriesIsReadFromTheBandSet is the fail-closed half of the band
// centre rule.
//
// The subject is derived from the band set rather than listed here, which is
// what keeps the two from drifting, and the cost of deriving it is that an empty
// answer would leave the rule matching nothing while every run stayed green.
func TestTheCentreSeriesIsReadFromTheBandSet(t *testing.T) {
	centres := bandCentres()
	if len(centres) != 21 {
		t.Errorf("the rule was built from %d centre(s); the extended set has 21", len(centres))
	}
	for _, want := range []int64{50, 100, 1000, 3150, 5000} {
		if !centres[want] {
			t.Errorf("%d Hz is a band of the extended set and the rule does not hold it", want)
		}
	}
	if centres[10] {
		t.Error("10 is not a band centre and the rule holds it")
	}
}

// TestEveryContainerRouteExists is the other fail-closed half. A type renamed,
// or a route renamed, leaves the rule naming something the floor no longer has,
// and a rule whose subject does not exist refuses nothing while looking exactly
// like a rule that found nothing to refuse.
func TestEveryContainerRouteExists(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("cannot find the repository root: %v", err)
	}

	types, funcs, err := declarationsIn(filepath.Join(root, "acoustic"))
	if err != nil {
		t.Fatalf("cannot read the numeric floor: %v", err)
	}
	if len(types) == 0 || len(funcs) == 0 {
		t.Fatal("the floor declared no types or no functions at all, which is not a floor this rule can be about")
	}

	for container, route := range containerRoutes {
		if !types[container] {
			t.Errorf("the rule names the container %s and the floor declares no such type", container)
		}
		if !funcs[route] {
			t.Errorf("the rule names %s as the route into %s and the floor declares no such function", route, container)
		}
	}
}

// declarationsIn reads the non-test Go files of one directory and reports the
// type names and function names it declares.
//
// It is deliberately one directory rather than a walk: the floor is one package
// and a rule about it should not be satisfied by a name that appears somewhere
// underneath it.
func declarationsIn(dir string) (types, funcs map[string]bool, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	types, funcs = map[string]bool{}, map[string]bool{}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				funcs[d.Name.Name] = true
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						types[ts.Name.Name] = true
					}
				}
			}
		}
	}
	return types, funcs, nil
}

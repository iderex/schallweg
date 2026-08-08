package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// The layering rules, each proved by breaking it deliberately once.
//
// A rule whose test has never failed is a rule nobody knows works, and the way
// an architecture rule usually fails is by matching nothing at all, which looks
// exactly like a clean tree. So every rule below has a case that breaks it, a
// case that comes as close as possible without breaking it, and an assertion on
// the real module beside them.

// mod builds an import path in this module, so the fixtures below read as paths
// rather than as string concatenation.
//
// It joins rather than concatenating because this suite's own rules refuse a
// path literal beginning with a separator, and a bare separator is one. That
// rule fired on the first version of this file, which is the shortest available
// demonstration that it bites.
func mod(rest string) string { return path.Join(modulePath, rest) }

// legalGraph is the shape the module actually has: a floor that imports nothing
// of its own, a kernel over it, a data layer over that, and a command line over
// everything. Each rule's fixture is this graph with one edge changed.
func legalGraph() []pkg {
	return []pkg{
		{ImportPath: mod("acoustic"), Imports: []string{"errors", "fmt", "math"}},
		{ImportPath: mod("kernel"), Imports: []string{"fmt", mod("acoustic")}},
		{ImportPath: mod("store"), Imports: []string{"bufio", "io", "strconv", mod("acoustic"), mod("kernel")}},
		{ImportPath: mod("cmd/schallweg"), Imports: []string{"os", mod("store")}},
		{ImportPath: mod("harness"), Imports: []string{"os", "net", mod("kernel")}},
	}
}

// rules counts the findings by rule name.
func rules(found []violation) map[string]int {
	got := map[string]int{}
	for _, v := range found {
		got[v.Rule]++
	}
	return got
}

// TestTheLegalGraphBreaksNoRule is the near miss for every graph rule at once.
//
// It is the assertion that decides whether the rules are livable: the harness
// imports the network and the file system and depends on the kernel, the command
// line reads the file system, and the data layer depends on two layers below it.
// Every one of those is legal and a checker that refused any of them would be
// refusing the architecture rather than enforcing it.
func TestTheLegalGraphBreaksNoRule(t *testing.T) {
	if found := checkGraph(legalGraph()); len(found) != 0 {
		for _, v := range found {
			t.Errorf("the legal graph produced %s", v)
		}
	}
}

// TestTheKernelMayNotReachTheFileSystemOrTheNetwork breaks the first rule, in
// both the shape somebody writes on purpose and the shape that arrives by
// accident.
//
// The second case is the one worth having. Nothing in the kernel package's own
// imports is forbidden; it picks up the file system from another kernel package,
// which is how this rule is actually broken. The convenience that does it is
// loading a reference dataset inside the function that needs it, which
// docs/decisions/layering.md names as the most likely cause.
func TestTheKernelMayNotReachTheFileSystemOrTheNetwork(t *testing.T) {
	t.Run("directly", func(t *testing.T) {
		g := legalGraph()
		g[1].Imports = append(g[1].Imports, "os")

		found := checkGraph(g)
		if got := rules(found)["kernel-reaches-io"]; got != 1 {
			t.Fatalf("the rule fired %d times, want 1: %v", got, found)
		}
	})

	t.Run("through another package of this module", func(t *testing.T) {
		g := append(legalGraph(), pkg{ImportPath: mod("kernel/reference"), Imports: []string{"os"}})
		g[1].Imports = append(g[1].Imports, mod("kernel/reference"))

		found := checkGraph(g)
		// Both packages are reported: the one that asked for it, and the one whose
		// layer stopped being pure because of it.
		if got := rules(found)["kernel-reaches-io"]; got != 2 {
			t.Fatalf("the rule fired %d times, want 2: %v", got, found)
		}
	})

	t.Run("the numeric floor is held to the same rule", func(t *testing.T) {
		g := legalGraph()
		g[0].Imports = append(g[0].Imports, "net/http")

		// Two findings rather than one, and the count is asserted exactly
		// because it is the rule's reach that is being described. The floor now
		// reaches the network, and so does the kernel above it, which imports the
		// floor: both have stopped being determined by their arguments, so both
		// are named. The data access layer and the command line import the floor
		// too and are not named, because the rule is theirs to break only in the
		// two layers that promise not to.
		if got := rules(checkGraph(g))["kernel-reaches-io"]; got != 2 {
			t.Fatalf("the rule fired %d times, want 2", got)
		}
	})
}

// TestTheChainRunsOneWay breaks the rule that a layer may not import a layer
// above it, which is the same rule the issue states as the format readers
// depending on the kernel's types and not the reverse.
func TestTheChainRunsOneWay(t *testing.T) {
	for _, c := range []struct {
		name  string
		index int
		add   string
	}{
		{"the kernel imports the data access layer", 1, mod("store")},
		{"the numeric floor imports the kernel", 0, mod("kernel")},
		{"the data access layer imports the command line", 2, mod("cmd/schallweg")},
		{"the kernel imports a harness", 1, mod("harness")},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := legalGraph()
			g[c.index].Imports = append(g[c.index].Imports, c.add)

			found := checkGraph(g)
			if got := rules(found)["layer-inversion"]; got == 0 {
				t.Fatalf("the rule did not fire: %v", found)
			}
		})
	}
}

// TestTheDataAccessLayerComputesNothing breaks the rule that reading, validating
// and serialising is not calculating.
//
// The rule is expressed as the data access layer not importing math, which is
// narrower than "contains no calculation" and is the part of it a machine can
// decide. What it catches is the real shape: a converter or a plausibility bound
// creeping into the parser, at which point the same quantity is computed in two
// layers and a result depends on which one ran.
func TestTheDataAccessLayerComputesNothing(t *testing.T) {
	g := legalGraph()
	g[2].Imports = append(g[2].Imports, "math")

	found := checkGraph(g)
	if got := rules(found)["data-layer-computes"]; got != 1 {
		t.Fatalf("the rule fired %d times, want 1: %v", got, found)
	}

	// The near miss: the numeric floor imports math and must go on being allowed
	// to, because it is the layer that computes.
	if got := rules(checkGraph(legalGraph()))["data-layer-computes"]; got != 0 {
		t.Fatalf("the rule fired on the legal graph %d times", got)
	}
}

// TestAPackageInNoLayerFails is the rule that keeps the others honest. A new
// directory that nobody placed is not covered by any rule above, and the failure
// mode is silence: the package is skipped and every check stays green.
func TestAPackageInNoLayerFails(t *testing.T) {
	g := append(legalGraph(), pkg{ImportPath: mod("plugins/report"), Imports: []string{"os", mod("cmd/schallweg")}})

	found := checkGraph(g)
	if got := rules(found)["package-not-placed"]; got != 1 {
		t.Fatalf("the rule fired %d times, want 1: %v", got, found)
	}
}

// countSourceRules reads a fixture from this package's testdata directory as the
// given layer and counts the findings by rule.
func countSourceRules(t *testing.T, fixture string, l layer) map[string]int {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "architecture", fixture))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	found, err := checkSource(fixture, src, l)
	if err != nil {
		t.Fatalf("reading %s: %v", fixture, err)
	}
	return rules(found)
}

// TestOnlyTheCommandLinePrints breaks the rule that below the command line a
// thing that went wrong is a returned error.
//
// It is read out of the syntax rather than out of the import graph because every
// package that returns a formatted error imports fmt, so the import says nothing
// about whether the package prints.
func TestOnlyTheCommandLinePrints(t *testing.T) {
	got := countSourceRules(t, "prints-and-computes.go.txt", layerKernel)
	if got["prints-below-cmd"] != 5 {
		t.Errorf("the print rule fired %d times, want 5", got["prints-below-cmd"])
	}

	// The same file in the layer that is allowed to print produces nothing.
	if atCmd := countSourceRules(t, "prints-and-computes.go.txt", layerCmd); atCmd["prints-below-cmd"] != 0 {
		t.Errorf("the print rule fired %d times in the command line", atCmd["prints-below-cmd"])
	}
}

// TestOnlyTheNumericFloorComputesALogarithm breaks the rule that the decibel
// arithmetic lives in one place.
//
// The rule reaches an exponential as well as a logarithm, because an energy
// conversion is one of each and a package holding either holds half of a decibel
// arithmetic somewhere it cannot be checked.
func TestOnlyTheNumericFloorComputesALogarithm(t *testing.T) {
	got := countSourceRules(t, "prints-and-computes.go.txt", layerKernel)
	if got["logarithm-outside-the-floor"] != 4 {
		t.Errorf("the logarithm rule fired %d times, want 4", got["logarithm-outside-the-floor"])
	}

	// The same file in the layer that owns the arithmetic produces nothing.
	if atFloor := countSourceRules(t, "prints-and-computes.go.txt", layerAcoustic); atFloor["logarithm-outside-the-floor"] != 0 {
		t.Errorf("the logarithm rule fired %d times in the numeric floor", atFloor["logarithm-outside-the-floor"])
	}
}

// TestTheSourceRulesPassTheNearMiss is the other half, and it is the half that
// decides whether anybody can live with the reader. Every line in the fixture is
// one character or one identifier away from a line in the fixture above.
func TestTheSourceRulesPassTheNearMiss(t *testing.T) {
	for _, l := range []layer{layerAcoustic, layerKernel, layerStore} {
		got := countSourceRules(t, "the-near-miss.go.txt", l)
		if len(got) != 0 {
			t.Errorf("the near miss produced findings in %s: %v", l, got)
		}
	}
}

// TestARenamedImportIsStillSeen is the one-identifier mistake somebody actually
// makes, and the one a reader matching on the literal text would miss.
func TestARenamedImportIsStillSeen(t *testing.T) {
	got := countSourceRules(t, "renamed-imports.go.txt", layerKernel)
	if got["prints-below-cmd"] != 1 {
		t.Errorf("the print rule fired %d times through a renamed import, want 1", got["prints-below-cmd"])
	}
	if got["logarithm-outside-the-floor"] != 1 {
		t.Errorf("the logarithm rule fired %d times through a renamed import, want 1", got["logarithm-outside-the-floor"])
	}
}

// TestThisModuleObeysItsOwnLayering is the rules applied to this repository, on
// the graph the toolchain reports rather than on a fixture.
//
// It needs an installed toolchain and a git checkout. Both are things a run of
// this suite already has, and neither is a display, an elevated account, the
// network or a piece of hardware, which are the four conditions the ordinary
// suite is held to.
func TestThisModuleObeysItsOwnLayering(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("cannot find the repository root: %v", err)
	}

	pkgs, err := moduleGraph(root)
	if err != nil {
		t.Fatalf("cannot read the module graph: %v", err)
	}
	// Fail closed. A graph with no packages in it is not evidence of a clean
	// module, and it is the failure that looks exactly like success.
	if len(pkgs) == 0 {
		t.Fatal("the toolchain reported no packages at all; a check that reads nothing is not a check")
	}
	t.Logf("read %d package(s) from the build", len(pkgs))

	for _, v := range checkGraph(pkgs) {
		t.Errorf("%s", v)
	}

	found, scanned, err := checkSourceTree(root)
	if err != nil {
		t.Fatalf("cannot read the tree: %v", err)
	}
	if scanned == 0 {
		t.Fatal("the reader scanned no source files at all")
	}
	t.Logf("scanned %d source file(s)", scanned)

	for _, v := range found {
		t.Errorf("%s", v)
	}
}

// TestEveryPackageOfThisModuleIsPlaced is the same rule stated the other way
// round, so a failure says which directory nobody placed rather than only that
// the module has a package too many.
func TestEveryPackageOfThisModuleIsPlaced(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("cannot find the repository root: %v", err)
	}
	pkgs, err := moduleGraph(root)
	if err != nil {
		t.Fatalf("cannot read the module graph: %v", err)
	}

	for _, p := range pkgs {
		if !strings.HasPrefix(p.ImportPath, modulePath) {
			t.Errorf("%s is not a package of this module and the toolchain listed it", p.ImportPath)
			continue
		}
		if _, placed := layerOf(p.ImportPath); !placed {
			t.Errorf("%s is in no layer; docs/layout.md is where a new one is added", p.ImportPath)
		}
	}
}

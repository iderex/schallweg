package main

// The layering rules, read out of the build and out of the source.
//
// docs/decisions/layering.md says what may depend on what, and until this file
// existed it also said that nothing enforced any of it. A document refuses
// nothing, and the dependency that breaks a layering is added by somebody who
// never read the document.
//
// Two readers, because the rules are of two kinds. Three of them are properties
// of the import graph, which the toolchain already computes, so they are checked
// against the graph rather than against a list written here. Two are properties
// of what a function calls rather than of what a package imports, because `fmt`
// is imported by every package that returns an error and says nothing about
// whether that package prints, so those are read out of the syntax.
//
// What this cannot see is stated where it is decided, in layerImports below and
// in the two source rules. The largest bound is that the graph rule reads what a
// package asks for by name, not what the linker eventually pulls in: every
// package that imports `fmt` has `os` in its transitive dependencies, so a rule
// written over the full dependency list would fire on all of them and prove
// nothing.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// modulePath is this module, and it is what separates a package of this project
// from one of the standard library in an import list.
const modulePath = "github.com/iderex/schallweg"

// A layer is one rung of docs/decisions/layering.md. The order of the constants
// is the chain, and it is the whole of what "may import" means: a package may
// import a lower layer and never a higher one.
type layer int

const (
	layerAcoustic layer = iota
	layerKernel
	layerStore
	layerCmd
	// layerHarness sits outside the chain. It may import anything and nothing
	// may import it, which is what keeps the rest of the tree honest about what
	// it needs.
	layerHarness
)

func (l layer) String() string {
	switch l {
	case layerAcoustic:
		return "the numeric floor"
	case layerKernel:
		return "the kernel"
	case layerStore:
		return "the data access layer"
	case layerCmd:
		return "the command line"
	case layerHarness:
		return "the harnesses"
	default:
		return "an unplaced package"
	}
}

// layerByDirectory places a package by the top-level directory it is in.
//
// This is the mapping docs/layout.md makes visible in the tree. A package
// somewhere else is not placed, and an unplaced package is a failure rather than
// something the rules skip.
var layerByDirectory = map[string]layer{
	"acoustic": layerAcoustic,
	"kernel":   layerKernel,
	"store":    layerStore,
	"cmd":      layerCmd,
	"harness":  layerHarness,
}

// layerImports are the standard library packages the numeric floor and the
// kernel may not reach for, by name.
//
// The list is what those two layers are forbidden to do rather than everything
// that performs input or output: reaching the file system, reaching the network,
// running another program, and reading the process's own surroundings. A layer
// with none of these is a layer whose behaviour is fully determined by its
// arguments, which is what makes a validation case that reproduces mean
// something.
//
// What it does not cover, said here rather than left to be discovered. It reads
// what a package imports by name, so a package reaching the file system through
// another package of this module is caught only because the module's own edges
// are followed too, and one reaching it through a third-party package would not
// be caught at all. The tree has no third-party packages today and the
// dependency policy is what keeps that true.
var layerImports = map[string]string{
	"os":                "the file system and the process's surroundings",
	"os/exec":           "running another program",
	"os/signal":         "the process's surroundings",
	"os/user":           "the machine's accounts",
	"io/fs":             "the file system",
	"net":               "the network",
	"net/http":          "the network",
	"net/rpc":           "the network",
	"net/smtp":          "the network",
	"net/textproto":     "the network",
	"net/http/httptest": "the network",
	"syscall":           "the operating system directly",
}

// A pkg is one package of this module as the toolchain reports it.
type pkg struct {
	// ImportPath is the package's full import path.
	ImportPath string
	// Imports is what the package's own non-test source imports, directly.
	Imports []string
}

// layerOf places an import path of this module, and reports whether it could be
// placed at all.
func layerOf(importPath string) (layer, bool) {
	rest, inModule := strings.CutPrefix(importPath, modulePath)
	if !inModule {
		return 0, false
	}
	rest = strings.TrimPrefix(rest, "/")
	dir, _, _ := strings.Cut(rest, "/")
	l, placed := layerByDirectory[dir]
	return l, placed
}

// checkGraph reports every layering rule the module's import graph breaks.
//
// It follows this module's own edges transitively, so a kernel package that
// reaches the file system through another kernel package is reported at the
// package that asked for it and at the one that carried it. Both are true and
// both are worth seeing: the first is where the import is, the second is why the
// layer is no longer pure.
func checkGraph(pkgs []pkg) []violation {
	var found []violation

	byPath := make(map[string]pkg, len(pkgs))
	for _, p := range pkgs {
		byPath[p.ImportPath] = p
	}

	for _, p := range pkgs {
		l, placed := layerOf(p.ImportPath)
		if !placed {
			found = append(found, violation{
				Path: p.ImportPath,
				Rule: "package-not-placed",
				Detail: "this package is in no directory the layering places, so no rule below reaches it; " +
					"docs/layout.md is where a new layer is added",
			})
			continue
		}

		for _, imported := range reach(p, byPath) {
			if why, forbidden := layerImports[imported]; forbidden && (l == layerAcoustic || l == layerKernel) {
				found = append(found, violation{
					Path:   p.ImportPath,
					Rule:   "kernel-reaches-io",
					Detail: "reaches " + imported + ", which is " + why + "; " + l.String() + " computes only on what arrives as an argument",
				})
			}
			other, inModule := layerOf(imported)
			if !inModule {
				continue
			}
			if l == layerHarness {
				continue
			}
			if other == layerHarness {
				found = append(found, violation{
					Path:   p.ImportPath,
					Rule:   "layer-inversion",
					Detail: "imports " + imported + ", and nothing may depend on the harnesses",
				})
				continue
			}
			if other > l {
				found = append(found, violation{
					Path:   p.ImportPath,
					Rule:   "layer-inversion",
					Detail: "is in " + l.String() + " and imports " + imported + ", which is in " + other.String() + "; the chain runs one way",
				})
			}
		}

		if l == layerStore {
			for _, imported := range p.Imports {
				if imported == "math" {
					found = append(found, violation{
						Path:   p.ImportPath,
						Rule:   "data-layer-computes",
						Detail: "imports math; the data access layer reads, validates and serialises, and every quantity it holds was computed below it",
					})
				}
			}
		}
	}

	sort.SliceStable(found, func(i, j int) bool {
		if found[i].Path != found[j].Path {
			return found[i].Path < found[j].Path
		}
		return found[i].Rule < found[j].Rule
	})
	return found
}

// reach is everything p imports, directly and through this module's own
// packages, with each import path reported once.
//
// Only this module's edges are followed. A standard library package's own
// dependencies are not, because following them makes every package that returns
// a formatted error reach the file system and the rule stops saying anything.
func reach(p pkg, byPath map[string]pkg) []string {
	seen := map[string]bool{}
	var out []string

	var walk func(imports []string)
	walk = func(imports []string) {
		for _, imported := range imports {
			if seen[imported] {
				continue
			}
			seen[imported] = true
			out = append(out, imported)
			if next, ours := byPath[imported]; ours {
				walk(next.Imports)
			}
		}
	}
	walk(p.Imports)

	sort.Strings(out)
	return out
}

// printCalls are the ways a package writes to a terminal, by the name of the
// function in the fmt package.
var printCalls = map[string]bool{
	"Print": true, "Printf": true, "Println": true,
}

// transcendentalCalls are the functions that take a logarithm or an exponential,
// by their name in the math package.
//
// Pow is here with the logarithms because that is what an energy conversion is:
// ten to the power of a level over ten, and back through a base ten logarithm.
// A package holding one of those holds half of a decibel arithmetic, and the
// point of the rule is that the two halves live in one place.
var transcendentalCalls = map[string]bool{
	"Log": true, "Log10": true, "Log2": true, "Log1p": true, "Logb": true,
	"Exp": true, "Exp2": true, "Expm1": true,
	"Pow": true, "Pow10": true,
}

// checkSource reports the rules that are properties of what a package's source
// says rather than of what it imports.
//
// Two of them are its own, and they are about what a package calls: printing
// below the command line, and decibel arithmetic outside the floor. The rest are
// in invariants.go, which is handed the same parsed file rather than parsing it
// a second time, and which is about what a calculation writes down.
//
// path is used only for reporting, so a caller may pass a fixture's name. l is
// the layer the file is in.
//
// Neither rule reads a test file, and the reasons differ. A test writes to the
// testing package rather than to a terminal, so the print rule would have
// nothing to say about one. And a test proving an arithmetic operation is
// expected to write the arithmetic out by hand, which is a logarithm in a test
// file of a layer that may not compute one, and refusing it would refuse the
// proof rather than the defect.
func checkSource(path string, src []byte, l layer) ([]violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var found []violation
	add := func(pos token.Pos, rule, detail string) {
		found = append(found, violation{Path: path, Line: fset.Position(pos).Line, Rule: rule, Detail: detail})
	}

	// The local name each import was given, so a renamed import is still seen and
	// a variable that happens to be called math is not mistaken for one.
	local := map[string]string{}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		local[p] = name
	}
	fmtName, mathName, osName := local["fmt"], local["math"], local["os"]

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			base, ok := x.X.(*ast.Ident)
			if !ok {
				return true
			}
			if l != layerCmd && l != layerHarness {
				if fmtName != "" && base.Name == fmtName && printCalls[x.Sel.Name] {
					add(x.Pos(), "prints-below-cmd", "calls "+base.Name+"."+x.Sel.Name+
						"; below the command line a thing that went wrong is a returned error and a thing worth saying is a field in the result")
				}
				if osName != "" && base.Name == osName && (x.Sel.Name == "Stdout" || x.Sel.Name == "Stderr") {
					add(x.Pos(), "prints-below-cmd", "names "+base.Name+"."+x.Sel.Name+
						"; a library with an opinion about who is reading is a library nobody can embed")
				}
			}
			if l != layerAcoustic && mathName != "" && base.Name == mathName && transcendentalCalls[x.Sel.Name] {
				add(x.Pos(), "logarithm-outside-the-floor", "calls "+base.Name+"."+x.Sel.Name+
					"; decibel arithmetic lives in one place so that it can be checked in one place")
			}
		case *ast.CallExpr:
			id, ok := x.Fun.(*ast.Ident)
			if !ok || l == layerCmd || l == layerHarness {
				return true
			}
			if id.Name == "print" || id.Name == "println" {
				add(x.Pos(), "prints-below-cmd", "calls the built-in "+id.Name+
					"; below the command line nothing writes to a terminal")
			}
		}
		return true
	})

	found = append(found, checkInvariants(path, fset, file, l)...)

	sort.SliceStable(found, func(i, j int) bool { return found[i].Line < found[j].Line })
	return found, nil
}

// moduleGraph asks the toolchain for this module's packages and what each one
// imports.
//
// It is derived from the build rather than from a list written here, which is
// the difference between a rule that stays true when a package is added and one
// that quietly stops covering it. The cost is a dependency on an installed
// toolchain, which a run of the suite already has.
//
// Test files are deliberately not included. The layering is a property of the
// package a caller gets, and a test reading its own fixture from disk is the
// ordinary way a test in any layer works.
func moduleGraph(root string) ([]pkg, error) {
	out, err := output(root, "go", "list", "-f", "{{.ImportPath}}|{{join .Imports \" \"}}", "./...")
	if err != nil {
		return nil, err
	}

	var pkgs []pkg
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path, imports, _ := strings.Cut(line, "|")
		p := pkg{ImportPath: path}
		if imports != "" {
			p.Imports = strings.Fields(imports)
		}
		pkgs = append(pkgs, p)
	}
	return pkgs, nil
}

// checkSourceTree reports every source rule broken by a non-test Go file in the
// module rooted at root, and how many files it read.
//
// A directory named testdata is skipped, because the toolchain does not build
// one and a fixture in it is source meant to be read rather than run.
func checkSourceTree(root string) ([]violation, int, error) {
	var found []violation
	scanned := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if d.Name() == "testdata" || (strings.HasPrefix(d.Name(), ".") && rel != ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir, _, _ := strings.Cut(rel, "/")
		l, placed := layerByDirectory[dir]
		if !placed {
			// A Go file outside every layer is the graph rule's finding, and
			// reporting it twice would say one thing in two voices.
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		scanned++
		v, checkErr := checkSource(rel, src, l)
		if checkErr != nil {
			return checkErr
		}
		found = append(found, v...)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	sort.SliceStable(found, func(i, j int) bool {
		if found[i].Path != found[j].Path {
			return found[i].Path < found[j].Path
		}
		return found[i].Line < found[j].Line
	})
	return found, scanned, nil
}

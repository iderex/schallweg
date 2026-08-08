package main

// The rules the ordinary test suite has to obey, and the reader that finds a
// test breaking one.
//
// A numerical project grows flaky tests in a particular way, and it is not
// timing. It is a test asserting on a floating point value with an implicit
// tolerance, a fixture generated at run time from something that changes, and a
// test that reads a file whose path depends on where it was run from. All three
// pass on the machine they were written on.
//
// These rules are read out of the source rather than observed at run time, which
// decides both what they catch and what they cannot. A test that never runs on
// anybody's machine still breaks them visibly, and a change that introduces one
// is refused at the first run rather than on the day the environment differs.
// What that costs is that nothing here sees through a helper: a test that
// reaches the network inside a function in another package is invisible to every
// rule below. The runtime half of that guard is the test workflow, which denies
// outbound network to the user the suite runs as, and the two are not
// substitutes for each other.

import (
	"fmt"
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

// violation is one rule broken at one place.
type violation struct {
	Path   string
	Line   int
	Rule   string
	Detail string
}

func (v violation) String() string {
	return fmt.Sprintf("%s:%d: %s: %s", v.Path, v.Line, v.Rule, v.Detail)
}

// forbiddenTestImports maps an import path a test may not use to the rule that
// refuses it.
//
// The network entries include the test server package. A server listening on
// loopback is still a network dependency: it takes a port, it can collide with
// something else on the machine, and it is exactly the "local mock that happens
// to listen on a port" the testability decision names.
//
// The randomness entries are here because a fixture generated at run time is a
// fixture nobody can look at. A test that fails on one seed in a thousand fails
// for somebody else, once, and cannot be reproduced from what they report.
var forbiddenTestImports = map[string]string{
	"net":               "network-import",
	"net/http":          "network-import",
	"net/http/httptest": "network-import",
	"net/rpc":           "network-import",
	"net/smtp":          "network-import",
	"net/textproto":     "network-import",
	"math/rand":         "random-import",
	"math/rand/v2":      "random-import",
	"crypto/rand":       "random-import",
}

// forbiddenTimeCalls are the calls that make a test depend on when it ran.
var forbiddenTimeCalls = map[string]bool{
	"Now":   true,
	"Since": true,
	"Until": true,
	"Local": true,
}

// checkTestSource reports every rule the given test source breaks.
//
// path is used only for reporting, so a caller may pass a fixture's name.
func checkTestSource(path string, src []byte) ([]violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}

	var found []violation
	at := func(pos token.Pos) int { return fset.Position(pos).Line }
	add := func(pos token.Pos, rule, detail string) {
		found = append(found, violation{Path: path, Line: at(pos), Rule: rule, Detail: detail})
	}

	// The local name each import was given, so a renamed import is still seen
	// and a variable that happens to be called time is not mistaken for one.
	local := map[string]string{}
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if rule, bad := forbiddenTestImports[p]; bad {
			add(imp.Pos(), rule, fmt.Sprintf("a test may not import %q", p))
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		local[p] = name
	}

	timeName := local["time"]
	osName := local["os"]

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if x.Op != token.EQL && x.Op != token.NEQ {
				return true
			}
			if lit, ok := floatLiteral(x.X); ok {
				add(x.Pos(), "float-equality", fmt.Sprintf("compares against the floating point literal %s; use acoustic/approx with an explicit tolerance", lit))
			} else if lit, ok := floatLiteral(x.Y); ok {
				add(x.Pos(), "float-equality", fmt.Sprintf("compares against the floating point literal %s; use acoustic/approx with an explicit tolerance", lit))
			}
		case *ast.SelectorExpr:
			base, ok := x.X.(*ast.Ident)
			if !ok {
				return true
			}
			if timeName != "" && base.Name == timeName && forbiddenTimeCalls[x.Sel.Name] {
				add(x.Pos(), "clock", fmt.Sprintf("uses %s.%s; a test that reads the clock is a test whose result depends on when it ran", base.Name, x.Sel.Name))
			}
			if osName != "" && base.Name == osName && x.Sel.Name == "Chdir" {
				add(x.Pos(), "working-directory", "uses os.Chdir; the toolchain already runs a test in its own package directory, and changing it makes the run order matter")
			}
		case *ast.BasicLit:
			if x.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(x.Value)
			if err != nil || !escapingPath(s) {
				return true
			}
			add(x.Pos(), "escaping-path", fmt.Sprintf("the path %q leaves the package directory or is absolute; a fixture lives in this package's testdata directory", s))
		}
		return true
	})

	sort.SliceStable(found, func(i, j int) bool { return found[i].Line < found[j].Line })
	return found, nil
}

// floatLiteral reports whether e is a floating point literal, with or without a
// leading sign, and returns how it was written.
func floatLiteral(e ast.Expr) (string, bool) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind == token.FLOAT {
			return x.Value, true
		}
	case *ast.UnaryExpr:
		if x.Op == token.SUB || x.Op == token.ADD {
			if lit, ok := floatLiteral(x.X); ok {
				return x.Op.String() + lit, true
			}
		}
	case *ast.ParenExpr:
		return floatLiteral(x.X)
	}
	return "", false
}

// escapingPath reports whether s reads as a file path that leaves the package
// directory or names a fixed place on one machine.
func escapingPath(s string) bool {
	if s == ".." || strings.HasPrefix(s, "../") || strings.HasPrefix(s, `..\`) {
		return true
	}
	if strings.HasPrefix(s, "/") {
		return true
	}
	// A Windows path: one letter, a colon, then a separator.
	if len(s) >= 3 && s[1] == ':' && (s[2] == '\\' || s[2] == '/') {
		c := s[0]
		return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	}
	return false
}

// exemptDirs are the directories the rules above do not reach, by name at the
// root of the module.
//
// harness is exempt by construction rather than by concession: it holds the
// tests that need a display, an instrument, another program, the network or an
// elevated account, each behind a build constraint named for what it needs.
// Applying the network rule there would refuse the one place a network test is
// allowed to live.
var exemptDirs = map[string]bool{"harness": true}

// checkTree reports every rule broken by any test file in the module rooted at
// root.
//
// Directories named testdata are skipped, because the toolchain does not build
// them and a fixture in one is source that is meant to be read rather than run.
func checkTree(root string) ([]violation, int, error) {
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
		if d.IsDir() {
			name := d.Name()
			if name == "testdata" || strings.HasPrefix(name, ".") && rel != "." {
				return fs.SkipDir
			}
			if exemptDirs[rel] {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		scanned++
		v, checkErr := checkTestSource(filepath.ToSlash(rel), src)
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

// Command gate runs this repository's checks locally, in order, and stops at the
// first failure.
//
// It exists so that a contributor gets the same verdict before pushing that the
// gate on the server gives afterwards, from one command rather than from a list
// of commands in a document that drifts.
//
// It also says what it examined. A leg that was not asked for prints that it was
// not asked for and what asking would cost, so a run that covered less than the
// whole set cannot be read as one that covered it and found nothing. Naming legs
// on the command line runs those and prints the rest under the same rule, which
// is how a job on the server asks for one of them without the run reading as the
// whole gate.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "\ngate: %v\n", err)
		os.Exit(1)
	}
}

// leg is one check. Its name is what the run prints, and its check reports a
// failure by returning an error whose message is what the contributor has to act
// on.
//
// A check writes to out for what it measured and did not gate on. A number a
// leg reports without refusing it is not a failure and must still be readable,
// and the alternative is a leg that keeps the figure to itself.
type leg struct {
	name  string
	what  string
	check func(root string, out io.Writer) error
}

// notAsked is a check this repository has, which this command does not run.
// Naming it here is the difference between a run that covered less than
// everything and a run that looks like it covered everything.
type notAsked struct {
	name string
	cost string
}

func legs() []leg {
	return []leg{
		{
			name: "format",
			what: "every Go file is gofmt formatted",
			check: func(root string, _ io.Writer) error {
				out, err := output(root, "gofmt", "-l", ".")
				if err != nil {
					return err
				}
				if strings.TrimSpace(out) != "" {
					return fmt.Errorf("these files are not formatted, run 'gofmt -w .':\n%s", strings.TrimSpace(out))
				}
				return nil
			},
		},
		{
			name: "vet",
			what: "go vet over every package",
			check: func(root string, _ io.Writer) error {
				_, err := output(root, "go", "vet", "./...")
				return err
			},
		},
		{
			name: "build",
			what: "every package builds, into a temporary directory outside the tree",
			check: func(root string, _ io.Writer) error {
				dir, err := os.MkdirTemp("", "schallweg-gate-build-")
				if err != nil {
					return err
				}
				defer os.RemoveAll(dir)
				_, err = output(root, "go", "build", "-o", dir+string(os.PathSeparator), "./...")
				return err
			},
		},
		{
			name: "test",
			what: "the ordinary suite, with the result cache defeated",
			check: func(root string, _ io.Writer) error {
				out, err := output(root, "go", "test", "./...", "-count=1")
				if err != nil {
					return err
				}
				if !strings.Contains(out, "ok  \t") {
					return fmt.Errorf("no package reported a passing test run:\n%s", strings.TrimSpace(out))
				}
				return nil
			},
		},
		{
			name: "docs",
			what: "no hard tab and no trailing whitespace in tracked Markdown",
			check: func(root string, _ io.Writer) error {
				for _, scan := range []struct {
					pattern string
					wrong   string
				}{
					{`\t`, "a hard tab"},
					{`[ \t]+$`, "trailing whitespace"},
				} {
					out, code, err := grep(root, scan.pattern)
					if err != nil {
						return err
					}
					switch code {
					case 0:
						return fmt.Errorf("%s in tracked Markdown:\n%s", scan.wrong, strings.TrimSpace(out))
					case 1:
						// Nothing matched, which is the passing case.
					default:
						// Fail closed: a broken scanner is not a clean tree.
						return fmt.Errorf("the Markdown scan for %s exited %d, which is neither a match nor a clean tree", scan.wrong, code)
					}
				}
				return nil
			},
		},
		{
			name:  "refs",
			what:  "every file a tracked document points at is in this repository",
			check: checkReferences,
		},
		{
			name:  "harnesses",
			what:  "the test targets the build reports, and every harness the tree holds",
			check: checkHarnesses,
		},
		{
			name:  "coverage",
			what:  "the surfaces that decide a number stand above the coverage bar",
			check: checkCoverage,
		},
	}
}

func notAskedFor() []notAsked {
	return []notAsked{
		{
			name: "the workflow security audit",
			cost: "it installs its analyser from the network on every run, so asking for it here would put a download in the middle of a local check",
		},
		{
			name: "the supply-chain self-audit and the dependency review",
			cost: "both read this repository through the hosting service's own interface, which a local run has no equivalent of",
		},
		{
			name: "the sign-off check",
			cost: "it judges the commits in a pull request against its base, which does not exist yet when this command runs",
		},
		{
			name: "the text rules check",
			cost: "it reads what git recorded for every tracked file, which this command could do, and doing it here would be a second implementation of one rule with nothing refusing the two drifting apart; .gitattributes already normalises on the way in, so what that check catches is a weakened attributes file rather than an ordinary mistake",
		},
	}
}

// selectLegs is which legs an invocation runs and which it leaves, from the
// names given on the command line.
//
// No name means every leg. An unknown name is refused rather than ignored,
// because a job asking for a leg that has been renamed would otherwise report a
// pass having run nothing, which is the failure the run's own disclosure exists
// against.
func selectLegs(all []leg, names []string) (run, left []leg, err error) {
	if len(names) == 0 {
		return all, nil, nil
	}
	wanted := map[string]bool{}
	for _, n := range names {
		known := false
		for _, l := range all {
			if l.name == n {
				known = true
				break
			}
		}
		if !known {
			var have []string
			for _, l := range all {
				have = append(have, l.name)
			}
			return nil, nil, fmt.Errorf("there is no leg called %q; the legs are %s", n, strings.Join(have, ", "))
		}
		wanted[n] = true
	}
	for _, l := range all {
		if wanted[l.name] {
			run = append(run, l)
		} else {
			left = append(left, l)
		}
	}
	return run, left, nil
}

func run(out io.Writer, names []string) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "gate: %s\n\n", root)

	asked, left, err := selectLegs(legs(), names)
	if err != nil {
		return err
	}
	for _, l := range asked {
		fmt.Fprintf(out, "  %-8s %s\n", l.name, l.what)
		if err := l.check(root, out); err != nil {
			fmt.Fprintf(out, "  %-8s FAILED\n", l.name)
			return fmt.Errorf("%s: %w", l.name, err)
		}
	}

	fmt.Fprintf(out, "\n%d leg(s) ran and passed.\n", len(asked))
	if len(left) > 0 {
		fmt.Fprintf(out, "\nLegs this invocation did not ask for:\n")
		for _, l := range left {
			fmt.Fprintf(out, "  %-8s %s\n", l.name, l.what)
		}
	}
	fmt.Fprintf(out, "\nNot asked for by this command:\n")
	for _, n := range notAskedFor() {
		fmt.Fprintf(out, "  %s\n    asking would cost: %s\n", n.name, n.cost)
	}
	fmt.Fprintf(out, "\nThose run on the server. A green run here is not a green run there.\n")
	return nil
}

// repoRoot asks git rather than walking upwards, so the answer is the same one
// every other command in this repository would get.
func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot find the repository root, is this a git checkout: %w", err)
	}
	return filepath.Clean(strings.TrimSpace(string(b))), nil
}

// output runs a command in the repository root and returns everything it wrote.
// A failure carries the command's own output, because that is what says what is
// wrong.
func output(root, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = root
	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), fmt.Errorf("%s %s:\n%s", name, strings.Join(args, " "), strings.TrimSpace(string(b)))
	}
	return string(b), nil
}

// grep runs git grep and returns its output and its exit code separately,
// because git grep reports "nothing matched" as exit code 1 and a scanner error
// as 2 or more, and those two must not be collapsed.
func grep(root, pattern string) (string, int, error) {
	cmd := exec.Command("git", "grep", "-nP", pattern, "--", "*.md")
	cmd.Dir = root
	b, err := cmd.CombinedOutput()
	if err == nil {
		return string(b), 0, nil
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(b), exit.ExitCode(), nil
	}
	return string(b), -1, fmt.Errorf("cannot run git grep: %w", err)
}

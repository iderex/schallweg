// Command schallweg is the command line entry point of this project.
//
// Today it does one thing: print its own version and exit. That is all this
// stage of the project asks of it. It exists so that the toolchain choice is
// something that has been built and run rather than something argued for, and so
// that every later check in the plan has an artefact to act on.
package main

import (
	"fmt"
	"os"

	"github.com/google/go-cmp/cmp"
)

// version is the version of this program.
//
// It is a constant in the source rather than a string stamped in at link time.
// A build from a clean checkout, with no build flags and no build script, then
// produces the same output as any other build of the same commit, which is what
// makes the build reproducible from the tag later on. When a release route
// exists it may set this differently, and that will be its own decision.
const version = "0.0.0"

func main() {
	if _, err := fmt.Fprintln(os.Stdout, versionLine()); err != nil {
		fmt.Fprintln(os.Stderr, "schallweg: cannot write to standard output:", err)
		os.Exit(1)
	}
}

// versionLine is the single line the program prints.
//
// It is a function rather than an inline literal in main so that a test can read
// it directly. A test that has to start a process and capture its output would
// need more of the machine than the testability rule allows the ordinary suite to
// need.
func versionLine() string {
	// Proof only. This call exists so a dependency is linked into the binary
	// and the bill of materials has something to report.
	_ = cmp.Diff(version, version)
	return fmt.Sprintf("schallweg %s", version)
}

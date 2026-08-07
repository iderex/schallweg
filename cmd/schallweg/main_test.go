package main

import (
	"strconv"
	"strings"
	"testing"
)

// TestVersionLineIsAProgramNameAndAVersion checks the shape of the one line this
// program prints: exactly two space-separated fields, the program name and a
// dotted numeric version.
//
// The shape is worth holding because the line is the first thing anything else
// will parse. An empty version, a version with a space in it, or a line that
// grew a third field would all still print something that looks fine to a person
// and breaks whatever reads it.
func TestVersionLineIsAProgramNameAndAVersion(t *testing.T) {
	fields := strings.Fields(versionLine())
	if len(fields) != 2 {
		t.Fatalf("versionLine() = %q, want exactly two space-separated fields, got %d", versionLine(), len(fields))
	}
	if fields[0] != "schallweg" {
		t.Errorf("first field = %q, want %q", fields[0], "schallweg")
	}

	parts := strings.Split(fields[1], ".")
	if len(parts) != 3 {
		t.Fatalf("version %q has %d dot-separated parts, want 3", fields[1], len(parts))
	}
	for i, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			t.Errorf("version %q part %d is %q, want a number: %v", fields[1], i, p, err)
		}
	}
}

// TestVersionLineNeedsNothingFromTheMachine is a statement of the testability
// rule at the one place there is code to state it about. The whole of this
// package's behaviour is reachable without a display, without elevation, without
// the network and without any hardware beyond the machine running the test: the
// call below is the entire surface.
func TestVersionLineNeedsNothingFromTheMachine(t *testing.T) {
	if versionLine() == "" {
		t.Fatal("versionLine() returned an empty string")
	}
}

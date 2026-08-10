package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iderex/schallweg/acoustic"
)

// The four worked constructions of docs/decisions/element-model.md, as fixtures.
//
// Their laboratory values are invented and each file says so in its own
// provenance, in the field a reader looks at first. They are here rather than in
// data/ because a fixture is not a record: the gate's data leg places a file
// under a testdata directory outside the data system, so nothing in this list
// can be validated as a record or read as one.
const (
	heavyWall       = "construction-heavy-wall.json"
	lightweightWall = "construction-lightweight-wall.json"
	concreteFloor   = "construction-concrete-floor.json"
	window          = "construction-window.json"
	ratingOnly      = "construction-rating-only.json"
)

func recordFixture(t *testing.T, name string) []byte {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading the fixture %s: %v", name, err)
	}
	return src
}

func readFixture(t *testing.T, name string) Construction {
	t.Helper()
	c, err := ReadConstruction(name, recordFixture(t, name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return c
}

// TestTheFourWorkedConstructions is the model against the four cases the
// decision record works through. Each one tests something the others do not: the
// heavy wall is the baseline, the lightweight wall is why the extended band set
// exists and why a certificate printing no loss factor still enters, the floor
// carries both laboratory quantities, and the window is the case that must not
// be asked for things a window does not have.
func TestTheFourWorkedConstructions(t *testing.T) {
	for _, c := range []struct {
		fixture string
		kind    Kind
		layers  int
		set     acoustic.BandSet
		impact  bool
	}{
		{heavyWall, Wall, 1, acoustic.Core, false},
		{lightweightWall, Wall, 5, acoustic.Extended, false},
		{concreteFloor, Floor, 1, acoustic.Core, true},
		{window, Window, 3, acoustic.Core, false},
	} {
		t.Run(c.fixture, func(t *testing.T) {
			got := readFixture(t, c.fixture)
			if got.Kind() != c.kind {
				t.Errorf("kind is %s, want %s", got.Kind(), c.kind)
			}
			if got.Basis() != Measured {
				t.Errorf("basis is %s, want measured", got.Basis())
			}
			if len(got.Layers()) != c.layers {
				t.Errorf("%d layer(s), want %d", len(got.Layers()), c.layers)
			}
			airborne, ok := got.AirborneLab()
			if !ok {
				t.Fatal("no airborne laboratory spectrum")
			}
			if airborne.Set() != c.set {
				t.Errorf("the airborne spectrum is on the %s, want the %s", airborne.Set(), c.set)
			}
			if airborne.Quantity() != SoundReductionIndex {
				t.Errorf("the airborne spectrum holds %s", airborne.Quantity())
			}
			if len(airborne.Values()) != c.set.Len() {
				t.Errorf("%d band(s), want the %d of the %s", len(airborne.Values()), c.set.Len(), c.set)
			}
			if _, ok := got.ImpactLab(); ok != c.impact {
				t.Errorf("carries an impact spectrum: %v, want %v", ok, c.impact)
			}
			if _, err := got.ForDetailedModel(); err != nil {
				t.Errorf("it cannot enter the detailed model: %v", err)
			}
		})
	}
}

// TestAWindowIsNotAskedForAnImpactSpectrum states the case the fixture exists
// for, because a model that required one of every kind would refuse a complete
// window as an incomplete floor.
func TestAWindowIsNotAskedForAnImpactSpectrum(t *testing.T) {
	if _, ok := readFixture(t, window).ImpactLab(); ok {
		t.Fatal("the window fixture carries an impact spectrum, so it cannot test the absence of one")
	}
	if _, err := readFixture(t, window).ForDetailedModel(); err != nil {
		t.Errorf("a window with no impact spectrum was refused: %v", err)
	}
}

// TestABoundedBandIsNotAMeasurement is the refusal at the point of reading. The
// window fixture's two top bands are stated against the facility's own limit,
// and a route that handed them over as measurements would record measurements
// nobody made.
func TestABoundedBandIsNotAMeasurement(t *testing.T) {
	airborne, ok := readFixture(t, window).AirborneLab()
	if !ok {
		t.Fatal("no airborne spectrum")
	}

	bounded := 0
	for _, v := range airborne.Values() {
		if !v.Measured() {
			bounded++
			if _, err := v.Decibels(); !errors.Is(err, ErrNotMeasured) {
				t.Errorf("the %d Hz band handed over a measurement: %v", v.Nominal(), err)
			}
		}
	}
	if bounded != 2 {
		t.Fatalf("%d bounded band(s) in the fixture, want 2; the test below proves nothing without them", bounded)
	}

	_, err := airborne.Measured()
	if !errors.Is(err, ErrNotMeasured) {
		t.Fatalf("the spectrum was handed over as measured: %v", err)
	}
	for _, want := range []string{"2500 Hz", "3150 Hz", "2 band(s)"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
}

// TestAMeasuredSpectrumReachesTheKernel is the other direction, and it is the
// count rather than the presence: the test above passes against a route that
// refuses every spectrum.
func TestAMeasuredSpectrumReachesTheKernel(t *testing.T) {
	airborne, ok := readFixture(t, heavyWall).AirborneLab()
	if !ok {
		t.Fatal("no airborne spectrum")
	}
	s, err := airborne.Measured()
	if err != nil {
		t.Fatalf("a spectrum with no bound in it was refused: %v", err)
	}
	if s.Len() != acoustic.Core.Len() {
		t.Errorf("the spectrum has %d bands, want %d", s.Len(), acoustic.Core.Len())
	}
}

// TestEachWayARecordIsRefused is the proof that each refusal bites, one fixture
// per way. Every one of them is a record that would otherwise become a
// laboratory value nobody can trace or a calculation nobody can defend.
func TestEachWayARecordIsRefused(t *testing.T) {
	for _, c := range []struct {
		fixture string
		is      error
		says    string
	}{
		{"construction-without-provenance.json", ErrRecordIncomplete, "no provenance"},
		{"construction-version-three.json", ErrRecordVersion, "version 3"},
		{"construction-with-a-supersession.json", ErrRecordNotCarried, "superseded value"},
		{"construction-that-is-a-lining.json", ErrNotAConstruction, "lining"},
		{"construction-described-with-values.json", ErrRecordIncomplete, "described construction carries no laboratory value"},
	} {
		t.Run(c.fixture, func(t *testing.T) {
			_, err := ReadConstruction(c.fixture, recordFixture(t, c.fixture))
			if err == nil {
				t.Fatal("the record was read without complaint")
			}
			if !errors.Is(err, c.is) {
				t.Errorf("the refusal is %v, want one that is %v", err, c.is)
			}
			if !strings.Contains(err.Error(), c.says) {
				t.Errorf("the refusal does not say %q: %v", c.says, err)
			}
			if !strings.Contains(err.Error(), c.fixture) {
				t.Errorf("the refusal does not name the file it is about: %v", err)
			}
		})
	}
}

// TestARatingOnlyRecordIsDistinguishableBeforeACalculationStarts is the
// condition this issue turns on. A certificate that publishes the weighted
// ratings and no spectrum is a complete record and an incomplete input, and the
// difference is visible before anything is computed rather than half way
// through.
func TestARatingOnlyRecordIsDistinguishableBeforeACalculationStarts(t *testing.T) {
	c := readFixture(t, ratingOnly)
	if _, ok := c.AirborneSingle(); !ok {
		t.Fatal("the fixture carries no weighted rating, so it cannot test this")
	}
	_, err := c.ForDetailedModel()
	if !errors.Is(err, ErrNotDetailed) {
		t.Fatalf("a record with no spectrum reached the detailed model: %v", err)
	}
	for _, want := range []string{c.Identity(), "no laboratory spectrum", "tested specimen"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
}

// TestReadingAndWritingIsTheIdentity is the serialisation condition. It compares
// the written document against the file it came from as decoded values rather
// than as bytes, because the order of an object's keys and the spelling of a
// number are not what "without loss" is about.
func TestReadingAndWritingIsTheIdentity(t *testing.T) {
	for _, name := range []string{heavyWall, lightweightWall, concreteFloor, window, ratingOnly} {
		t.Run(name, func(t *testing.T) {
			src := recordFixture(t, name)
			first, err := ReadConstruction(name, src)
			if err != nil {
				t.Fatalf("reading: %v", err)
			}

			var out bytes.Buffer
			if err := WriteConstruction(&out, first); err != nil {
				t.Fatalf("writing: %v", err)
			}

			var was, is any
			if err := json.Unmarshal(src, &was); err != nil {
				t.Fatalf("decoding the fixture: %v", err)
			}
			if err := json.Unmarshal(out.Bytes(), &is); err != nil {
				t.Fatalf("decoding what was written: %v", err)
			}
			if !reflect.DeepEqual(was, is) {
				t.Errorf("the written record is not the one that was read:\nwas %v\nis  %v", was, is)
			}

			// And again through the reader, so that what was written is a
			// record this reader accepts rather than merely one that decodes.
			second, err := ReadConstruction(name, out.Bytes())
			if err != nil {
				t.Fatalf("reading what was written: %v", err)
			}
			if !reflect.DeepEqual(first, second) {
				t.Error("the construction read back is not the one that was written")
			}
		})
	}
}

// TestTheFixturesSayTheyAreFixtures is the guard on the risk the fixtures
// create. Their values are invented, they are shaped exactly like records, and
// the failure to prevent is one of them being copied into data/ one day and
// entering the database with no source. Every one of them says so in the field a
// reader of a record looks at first.
func TestTheFixturesSayTheyAreFixtures(t *testing.T) {
	for _, name := range []string{heavyWall, lightweightWall, concreteFloor, window, ratingOnly} {
		p := readFixture(t, name).Provenance()
		if !strings.Contains(p.Laboratory, "fixture") {
			t.Errorf("%s does not say in the laboratory field that it is a fixture: %q", name, p.Laboratory)
		}
		if !strings.Contains(p.ObtainedFrom, "not a component record") {
			t.Errorf("%s does not say where it was obtained that it is not a record: %q", name, p.ObtainedFrom)
		}
	}
}

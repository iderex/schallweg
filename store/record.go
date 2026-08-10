package store

// A construction as this module holds it, and the one route from a component
// record on disk into that value.
//
// The vocabulary is docs/decisions/element-model.md and it is worth stating at
// the top because this file is named for a record and holds a construction. A
// construction is a type: what a laboratory tested or what somebody described,
// with no position and no building around it, and it is what a database record
// holds. An element is one instance of a construction in one building, it has an
// area and a position, and it exists only inside a project. The two are kept
// apart because the in situ correction needs the dimensions of the specimen that
// was tested and the dimensions of the thing that was built, and a model with one
// object has to hold both and then decide per field which one is meant. The
// project file is issue #91 and the element arrives with it.
//
// Every field carries its unit here, in SI, which is docs/decisions/numeric-contract.md.
//
// What this reader refuses, and why each refusal is at the boundary rather than
// later. A record whose schema version this reader does not know, because a major
// version exists to say the fields may no longer mean what the reader thinks. A
// lining or a covering, which are their own objects with their own improvement
// quantities and are issues #49 and #50. A construction with no provenance, which
// is a laboratory value with no source and is the thing the whole database exists
// to be better than. And a record carrying superseded values, because the route
// that writes and resolves those is issue #78 and a reader that dropped them
// would turn a correction history into a silent loss.

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/iderex/schallweg/acoustic"
)

// The refusals a caller can distinguish.
var (
	// ErrRecordVersion is a record whose schema major version this reader does
	// not know.
	ErrRecordVersion = errors.New("the record declares a schema version this reader does not know")
	// ErrNotAConstruction is a record that is a lining or a covering.
	ErrNotAConstruction = errors.New("the record is not a construction")
	// ErrRecordIncomplete is a record missing something its own shape requires.
	ErrRecordIncomplete = errors.New("the record is incomplete")
	// ErrRecordNotCarried is a record carrying something this reader does not
	// hold yet, which is refused rather than dropped.
	ErrRecordNotCarried = errors.New("the record carries something this reader does not hold")
	// ErrNotMeasured is a band that the test bounded rather than measured, asked
	// for as though it were a measurement.
	ErrNotMeasured = errors.New("the band was bounded rather than measured")
	// ErrNotDetailed is a construction that does not carry what the detailed
	// model needs.
	ErrNotDetailed = errors.New("the construction cannot enter the detailed model")
)

// recordSchemaVersion is the major version of data/schema/component-record.schema.json
// this reader was written against.
const recordSchemaVersion = 2

// Basis is whether a construction carries laboratory values or a description.
//
// Never inferred from whether the values happen to be present: an absent
// spectrum and a description are different statements, and the first release
// computes from measured constructions only.
type Basis uint8

const (
	// Measured carries laboratory values from a test, with provenance.
	Measured Basis = iota + 1
	// Described carries the makeup and no laboratory values.
	Described
)

func (b Basis) String() string {
	switch b {
	case Measured:
		return "measured"
	case Described:
		return "described"
	default:
		return "no basis"
	}
}

// Kind is what a construction is. It decides which calculations the
// construction may enter and it is not free text.
type Kind uint8

const (
	Wall Kind = iota + 1
	Floor
	Roof
	Window
	Door
)

var kindNames = map[string]Kind{"wall": Wall, "floor": Floor, "roof": Roof, "window": Window, "door": Door}

func (k Kind) String() string {
	for name, kind := range kindNames {
		if kind == k {
			return name
		}
	}
	return "no kind"
}

// Attachment is how a layer is connected to the one below. It is what
// distinguishes constructions that are identical on a materials list and tens of
// decibels apart in practice.
type Attachment uint8

const (
	Bonded Attachment = iota + 1
	MechanicallyFixed
	ResilientlyFixed
	Free
)

var attachmentNames = map[string]Attachment{
	"bonded":             Bonded,
	"mechanically fixed": MechanicallyFixed,
	"resiliently fixed":  ResilientlyFixed,
	"free":               Free,
}

func (a Attachment) String() string {
	for name, at := range attachmentNames {
		if at == a {
			return name
		}
	}
	return "no attachment"
}

// A Layer is one part of a construction's makeup.
//
// It has no acoustic values of its own and is never looked up on its own. A
// sheet is naturally described by mass per area and a solid by density, so one
// of the two is present and forcing either form would make half the records
// carry a converted number nobody can check against its source.
type Layer struct {
	// Material is a name from a controlled vocabulary rather than free text.
	Material string
	// Thickness is in metres.
	Thickness float64
	// Density is in kg/m3, and is zero where the layer is given by mass.
	Density float64
	// MassPerArea is in kg/m2, and is zero where the layer is given by density.
	MassPerArea float64
	// Attachment is how this layer is connected to the one below.
	Attachment Attachment
}

// A BandValue is what one band of a laboratory spectrum holds.
//
// Either a measurement or a bound, and they are different states rather than a
// number with a flag beside it. A test facility has a limit of its own, and a
// specimen that out-performs it is reported band by band against that limit; a
// reader that took the limit for a measurement would record a measurement nobody
// made. Decibels.
type BandValue struct {
	nominal   int
	value     float64
	bounded   bool
	atLeast   bool
	limitedBy string
}

// Nominal is the band's nominal centre frequency in hertz.
func (v BandValue) Nominal() int { return v.nominal }

// Measured reports whether this band holds a measurement.
func (v BandValue) Measured() bool { return !v.bounded }

// Decibels is the value, and it refuses a band the test bounded rather than
// measured. The refusal is here rather than several steps later with a plausible
// number.
func (v BandValue) Decibels() (float64, error) {
	if v.bounded {
		return 0, fmt.Errorf("%w: the %d Hz band is stated as %s %g dB, limited by %s",
			ErrNotMeasured, v.nominal, v.boundWord(), v.value, v.limitedBy)
	}
	return v.value, nil
}

// Bound is the bound the report printed, for a caller that has been written for
// one. It refuses a band that holds a measurement, so the two states cannot be
// read through one route.
func (v BandValue) Bound() (atLeast bool, value float64, limitedBy string, err error) {
	if !v.bounded {
		return false, 0, "", fmt.Errorf("the %d Hz band holds a measurement rather than a bound", v.nominal)
	}
	return v.atLeast, v.value, v.limitedBy, nil
}

func (v BandValue) boundWord() string {
	if v.atLeast {
		return "at least"
	}
	return "at most"
}

// A LabSpectrum is a laboratory quantity band by band, as the record holds it.
//
// It is not an acoustic.Spectrum, because that type holds a number in every band
// and a record may hold a bound in one. Measured is the route between them and it
// is the place a bound stops the calculation.
type LabSpectrum struct {
	quantity Quantity
	set      acoustic.BandSet
	values   []BandValue
}

// Quantity is what the spectrum is of.
func (s LabSpectrum) Quantity() Quantity { return s.quantity }

// Set is the band set the spectrum is on.
func (s LabSpectrum) Set() acoustic.BandSet { return s.set }

// Values are the bands in ascending order, each carrying its own state.
func (s LabSpectrum) Values() []BandValue {
	out := make([]BandValue, len(s.values))
	copy(out, s.values)
	return out
}

// Measured is this spectrum as a value the kernel can compute on.
//
// It refuses where any band is a bound rather than a measurement, and it names
// every such band rather than the first, because a certificate that bounds one
// band usually bounds several and a caller correcting the input wants the list.
func (s LabSpectrum) Measured() (acoustic.Spectrum, error) {
	nominals := make([]int, 0, len(s.values))
	values := make([]float64, 0, len(s.values))
	var bounded []string
	for _, v := range s.values {
		nominals = append(nominals, v.nominal)
		values = append(values, v.value)
		if v.bounded {
			bounded = append(bounded, fmt.Sprintf("%d Hz (%s %g dB, limited by %s)", v.nominal, v.boundWord(), v.value, v.limitedBy))
		}
	}
	if len(bounded) > 0 {
		return acoustic.Spectrum{}, fmt.Errorf("%w: %s of %s, in %d band(s): %s",
			ErrNotMeasured, s.quantity, s.set, len(bounded), joinAnd(bounded))
	}
	return acoustic.New(s.set, nominals, values)
}

// Uncertainty is what the laboratory states for a rated value, as the report
// prints it, with the document the figure comes from. It is recorded rather than
// computed: it is a fact the certificate asserts.
type Uncertainty struct {
	// Value is the half-width in decibels.
	Value float64
	// Source is the document the figure comes from, as the report names it.
	Source string
}

// AdaptationTerms is one term per band range, keyed by the range it covers,
// written as the lowest and highest nominal centre frequency in hertz with a
// hyphen between them, or "unstated" for the pair a certificate prints on its
// rating line without naming its range. Decibels.
type AdaptationTerms map[string]int

// AirborneRating is the weighted sound reduction index with its adaptation
// terms, as a certificate prints them. The terms are carried beside the rating
// and never folded into it.
type AirborneRating struct {
	// Rw is the weighted sound reduction index in decibels.
	Rw int
	// C and Ctr are the spectrum adaptation terms in decibels.
	C, Ctr AdaptationTerms
	// Uncertainty is what the laboratory states, where it states one.
	Uncertainty *Uncertainty
}

// ImpactRating is the weighted normalised impact sound pressure level with its
// adaptation term.
type ImpactRating struct {
	// Lnw is the weighted normalised impact sound pressure level in decibels.
	Lnw int
	// CI is the impact adaptation term in decibels.
	CI AdaptationTerms
	// Uncertainty is what the laboratory states, where it states one.
	Uncertainty *Uncertainty
}

// TestStandards is one standard per role, each with its part and edition as the
// report prints it. A role nobody read is absent rather than guessed.
type TestStandards struct {
	Airborne, Impact, Facility string
	AirborneRating             string
	ImpactRating               string
}

// Provenance is where a record's values came from.
//
// It is not optional on a measured construction and it can never be added
// retroactively to a record entered without it. The floor is
// docs/decisions/certificate-extraction.md and what provenance is as a data
// model beyond that floor is issue #74.
type Provenance struct {
	Laboratory          string
	ReportNumber        string
	ReportDate          string
	TestStandards       TestStandards
	Client              string
	SpecimenDesignation string
	ProductDesignation  string
	ObtainedFrom        string
	DescribedFrom       string
	EnteredBy           string
	EnteredOn           string
}

// A Construction is what a laboratory tested or what somebody described.
//
// The fields are unexported and there is one route in, ReadConstruction, so a
// value of this type is one that passed the refusals rather than one somebody
// filled in beside them. A composite literal of a type whose invariants are
// checked in a constructor is the second route that checks none of them.
type Construction struct {
	id          string
	kind        Kind
	basis       Basis
	massPerArea float64
	thickness   float64
	layers      []Layer

	airborneLab *LabSpectrum
	impactLab   *LabSpectrum

	airborneSingle *AirborneRating
	impactSingle   *ImpactRating

	specimenArea     float64
	specimenEdges    []float64
	labLossFactor    float64
	lossFactorAbsent string

	provenance Provenance
}

// Identity is the stable identity a result cites. The scheme is
// docs/decisions/identity.md.
func (c Construction) Identity() string { return c.id }

// Kind is what the construction is.
func (c Construction) Kind() Kind { return c.kind }

// Basis is whether it carries laboratory values or a description.
func (c Construction) Basis() Basis { return c.basis }

// MassPerArea is in kg/m2.
func (c Construction) MassPerArea() float64 { return c.massPerArea }

// Thickness is the total, in metres.
func (c Construction) Thickness() float64 { return c.thickness }

// Layers is the makeup, outermost first. A measured construction whose makeup
// was not published carries none, and that emptiness is a fact rather than a
// gap to be filled with a guess.
func (c Construction) Layers() []Layer {
	out := make([]Layer, len(c.layers))
	copy(out, c.layers)
	return out
}

// AirborneLab is the laboratory sound reduction index, and reports whether the
// record carries one at all.
func (c Construction) AirborneLab() (LabSpectrum, bool) {
	if c.airborneLab == nil {
		return LabSpectrum{}, false
	}
	return *c.airborneLab, true
}

// ImpactLab is the laboratory normalised impact sound pressure level, and
// reports whether the record carries one at all.
func (c Construction) ImpactLab() (LabSpectrum, bool) {
	if c.impactLab == nil {
		return LabSpectrum{}, false
	}
	return *c.impactLab, true
}

// AirborneSingle is the weighted rating, which is what the simplified model
// runs on. It is never derived from the spectrum here and the spectrum is never
// derived from it: docs/decisions/model-shape.md says each model reads its own
// input.
func (c Construction) AirborneSingle() (AirborneRating, bool) {
	if c.airborneSingle == nil {
		return AirborneRating{}, false
	}
	return *c.airborneSingle, true
}

// ImpactSingle is the weighted impact rating, for the simplified impact model.
func (c Construction) ImpactSingle() (ImpactRating, bool) {
	if c.impactSingle == nil {
		return ImpactRating{}, false
	}
	return *c.impactSingle, true
}

// Provenance is where the values came from.
func (c Construction) Provenance() Provenance { return c.provenance }

// A Detailed is a construction carrying everything the detailed model needs.
//
// It is a separate type rather than a flag, and that is the whole of what this
// answers. A calculation that takes a Detailed cannot be started on a
// construction that does not have the data, because there is no way to make one
// except through ForDetailedModel, which is where the refusal lives. A flag on
// one type puts the check at every call site and leaves the kernel to discover
// halfway through a calculation that a specimen dimension is missing.
type Detailed struct {
	c Construction
}

// Construction is the construction underneath, for a caller that has the
// detailed view and wants the ordinary fields.
func (d Detailed) Construction() Construction { return d.c }

// SpecimenArea is the area of the tested specimen, in m2. Per record rather
// than a constant, because a window's specimen is much smaller than a wall's.
func (d Detailed) SpecimenArea() float64 { return d.c.specimenArea }

// SpecimenEdges are the edge lengths of the tested specimen in metres, going
// round its boundary, so that their sum is its perimeter.
func (d Detailed) SpecimenEdges() []float64 {
	out := make([]float64, len(d.c.specimenEdges))
	copy(out, d.c.specimenEdges)
	return out
}

// LabLossFactor is the total loss factor of the specimen as tested,
// dimensionless, and reports whether the report stated one. Where it did not,
// the sentence recording that is LossFactorAbsent, and the laboratory value can
// only be used uncorrected, which is an assumption the result then carries.
func (d Detailed) LabLossFactor() (float64, bool) {
	return d.c.labLossFactor, d.c.lossFactorAbsent == ""
}

// LossFactorAbsent is the sentence recording that the report printed no loss
// factor, and is empty where one was printed.
func (d Detailed) LossFactorAbsent() string { return d.c.lossFactorAbsent }

// ForDetailedModel is the one route to a Detailed.
//
// It names everything that is missing rather than the first thing, because a
// record is corrected by somebody reading a list, and it names the construction,
// so a refusal arriving from a calculation over forty elements says which one.
func (c Construction) ForDetailedModel() (Detailed, error) {
	var missing []string
	if c.basis != Measured {
		missing = append(missing, "it is described rather than measured, and the first release computes from measured constructions only")
	}
	if c.airborneLab == nil && c.impactLab == nil {
		missing = append(missing, "it carries no laboratory spectrum, and a rating alone is the simplified model's input rather than this one's")
	}
	if c.specimenArea == 0 {
		missing = append(missing, "the area of the tested specimen")
	}
	if len(c.specimenEdges) == 0 {
		missing = append(missing, "the edge lengths of the tested specimen")
	}
	if c.labLossFactor == 0 && c.lossFactorAbsent == "" {
		missing = append(missing, "the loss factor of the specimen as tested, or the sentence recording that the report printed none")
	}
	if len(missing) > 0 {
		return Detailed{}, fmt.Errorf("%w: %s is %s, and the detailed model wants %s",
			ErrNotDetailed, c.id, c.kind, joinAnd(missing))
	}
	return Detailed{c: c}, nil
}

// ReadConstruction reads one component record and refuses one that is not a
// construction this module can hold.
//
// name is the file the bytes came from and appears in every refusal, because a
// message about a record nobody can find is a message about nothing.
func ReadConstruction(name string, src []byte) (Construction, error) {
	var raw recordDocument
	dec := json.NewDecoder(bytesReader(src))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return Construction{}, fmt.Errorf("%s is not a component record this reader knows: %w", name, err)
	}

	if raw.Schema == nil {
		return Construction{}, fmt.Errorf("%s: %w: it declares none, and the reader was written against version %d",
			name, ErrRecordVersion, recordSchemaVersion)
	}
	if *raw.Schema != recordSchemaVersion {
		return Construction{}, fmt.Errorf("%s: %w: it declares version %d and the reader was written against version %d",
			name, ErrRecordVersion, *raw.Schema, recordSchemaVersion)
	}
	if len(raw.Superseded) > 0 {
		return Construction{}, fmt.Errorf("%s: %w: %d superseded value(s); the route that writes and resolves those is issue #78, and dropping them here would turn a correction history into a silent loss",
			name, ErrRecordNotCarried, len(raw.Superseded))
	}

	kind, ok := kindNames[raw.Kind]
	if !ok {
		switch raw.Kind {
		case "lining", "covering":
			return Construction{}, fmt.Errorf("%s: %w: it is a %s, which improves a construction rather than being one, and has its own type",
				name, ErrNotAConstruction, raw.Kind)
		default:
			return Construction{}, fmt.Errorf("%s: %w: %q is not a kind this reader knows", name, ErrNotAConstruction, raw.Kind)
		}
	}

	c := Construction{
		id:               raw.ID,
		kind:             kind,
		massPerArea:      raw.MassPerArea,
		thickness:        raw.Thickness,
		specimenArea:     raw.SpecimenArea,
		specimenEdges:    append([]float64(nil), raw.SpecimenEdges...),
		labLossFactor:    raw.LabLossFactor,
		lossFactorAbsent: raw.LabLossFactorAbsent,
	}

	switch raw.Basis {
	case "measured":
		c.basis = Measured
	case "described":
		c.basis = Described
	default:
		return Construction{}, fmt.Errorf("%s: %w: the basis is %q, and it is measured or described rather than inferred from which fields are present",
			name, ErrRecordIncomplete, raw.Basis)
	}

	if c.id == "" {
		return Construction{}, fmt.Errorf("%s: %w: it has no identity, and a record a result cannot cite is a record nothing can reproduce", name, ErrRecordIncomplete)
	}

	var err error
	if c.layers, err = readLayers(name, raw.Layers); err != nil {
		return Construction{}, err
	}
	if c.airborneLab, err = readLabSpectrum(name, "airborne_lab", SoundReductionIndex, raw.AirborneLab); err != nil {
		return Construction{}, err
	}
	if c.impactLab, err = readLabSpectrum(name, "impact_lab", NormalisedImpactLevel, raw.ImpactLab); err != nil {
		return Construction{}, err
	}
	c.airborneSingle = raw.AirborneSingle.value()
	c.impactSingle = raw.ImpactSingle.value()

	if c.provenance, err = readProvenance(name, c.basis, raw.Provenance); err != nil {
		return Construction{}, err
	}
	if err := c.consistent(name); err != nil {
		return Construction{}, err
	}
	return c, nil
}

// consistent is the refusals that are about the record as a whole rather than
// about one field.
func (c Construction) consistent(name string) error {
	if c.massPerArea <= 0 {
		return fmt.Errorf("%s: %w: no mass per unit area, which the in situ correction and the junction route both need and neither can proceed without",
			name, ErrRecordIncomplete)
	}
	if c.thickness <= 0 {
		return fmt.Errorf("%s: %w: no thickness", name, ErrRecordIncomplete)
	}
	if c.basis == Described {
		if len(c.layers) == 0 {
			return fmt.Errorf("%s: %w: a described construction carries the makeup, and this one carries none", name, ErrRecordIncomplete)
		}
		if c.airborneLab != nil || c.impactLab != nil || c.airborneSingle != nil || c.impactSingle != nil {
			return fmt.Errorf("%s: %w: a described construction carries no laboratory value, and this one does; the first release computes from measured constructions only, and a described one carrying values is that restriction quietly undone",
				name, ErrRecordIncomplete)
		}
		return nil
	}
	if c.airborneLab == nil && c.impactLab == nil && c.airborneSingle == nil && c.impactSingle == nil {
		return fmt.Errorf("%s: %w: a measured construction carries at least one measured quantity and this one carries none", name, ErrRecordIncomplete)
	}
	if c.airborneLab != nil || c.impactLab != nil {
		if c.labLossFactor != 0 && c.lossFactorAbsent != "" {
			return fmt.Errorf("%s: %w: a loss factor and a sentence saying the report printed none, which leaves a reader two answers and no rule for choosing",
				name, ErrRecordIncomplete)
		}
		if c.labLossFactor == 0 && c.lossFactorAbsent == "" {
			return fmt.Errorf("%s: %w: a laboratory spectrum with neither a loss factor nor the sentence recording that the report printed none",
				name, ErrRecordIncomplete)
		}
		if c.specimenArea <= 0 || len(c.specimenEdges) == 0 {
			return fmt.Errorf("%s: %w: a laboratory spectrum without the specimen it was measured on, which cannot be corrected to a building",
				name, ErrRecordIncomplete)
		}
	}
	return nil
}

func readLayers(name string, raw []layerDocument) ([]Layer, error) {
	var out []Layer
	for i, l := range raw {
		attachment, ok := attachmentNames[l.Attachment]
		if !ok {
			return nil, fmt.Errorf("%s: %w: layer %d is attached %q, and that is not one of the four ways this model distinguishes",
				name, ErrRecordIncomplete, i+1, l.Attachment)
		}
		if l.Material == "" || l.Thickness <= 0 {
			return nil, fmt.Errorf("%s: %w: layer %d has no material or no thickness", name, ErrRecordIncomplete, i+1)
		}
		if l.Density <= 0 && l.MassPerArea <= 0 {
			return nil, fmt.Errorf("%s: %w: layer %d has neither a density nor a mass per unit area, and one of the two says how much of it there is",
				name, ErrRecordIncomplete, i+1)
		}
		out = append(out, Layer{
			Material:    l.Material,
			Thickness:   l.Thickness,
			Density:     l.Density,
			MassPerArea: l.MassPerArea,
			Attachment:  attachment,
		})
	}
	return out, nil
}

func readProvenance(name string, basis Basis, raw *provenanceDocument) (Provenance, error) {
	if raw == nil {
		return Provenance{}, fmt.Errorf("%s: %w: no provenance; a laboratory value with no source is a rumour, and this is the field that can never be added back later",
			name, ErrRecordIncomplete)
	}
	p := Provenance{
		Laboratory:          raw.Laboratory,
		ReportNumber:        raw.ReportNumber,
		ReportDate:          raw.ReportDate,
		Client:              raw.Client,
		SpecimenDesignation: raw.SpecimenDesignation,
		ProductDesignation:  raw.ProductDesignation,
		ObtainedFrom:        raw.ObtainedFrom,
		DescribedFrom:       raw.DescribedFrom,
		EnteredBy:           raw.EnteredBy,
		EnteredOn:           raw.EnteredOn,
	}
	if raw.TestStandards != nil {
		p.TestStandards = TestStandards{
			Airborne:       raw.TestStandards.Airborne,
			Impact:         raw.TestStandards.Impact,
			Facility:       raw.TestStandards.Facility,
			AirborneRating: raw.TestStandards.AirborneRating,
			ImpactRating:   raw.TestStandards.ImpactRating,
		}
	}

	// The floor a measured record has to reach, and the smaller one a described
	// record reaches because it has no certificate behind it.
	want := map[string]string{"entered_by": p.EnteredBy, "entered_on": p.EnteredOn}
	if basis == Measured {
		want["laboratory"] = p.Laboratory
		want["report_number"] = p.ReportNumber
		want["report_date"] = p.ReportDate
		want["client"] = p.Client
		want["specimen_designation"] = p.SpecimenDesignation
		want["obtained_from"] = p.ObtainedFrom
	} else {
		want["described_from"] = p.DescribedFrom
	}
	var missing []string
	for field, value := range want {
		if value == "" {
			missing = append(missing, field)
		}
	}
	if basis == Measured && raw.TestStandards == nil {
		missing = append(missing, "test_standards")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return Provenance{}, fmt.Errorf("%s: %w: the provenance is missing %s", name, ErrRecordIncomplete, joinAnd(missing))
	}
	return p, nil
}

func readLabSpectrum(name, field string, quantity Quantity, raw *spectrumDocument) (*LabSpectrum, error) {
	if raw == nil {
		return nil, nil
	}
	if raw.Quantity != quantity.Symbol() {
		return nil, fmt.Errorf("%s: %w: %s holds %q, and %s is %s; a reader taking one for the other is wrong by tens of decibels and looks ordinary while being so",
			name, ErrRecordIncomplete, field, raw.Quantity, field, quantity.Symbol())
	}
	// The same map the exchange format reads a band set through, so a record
	// and a document cannot come to disagree about what "core" means.
	set, known := bandSetByName[raw.BandSet]
	if !known {
		return nil, fmt.Errorf("%s: %w: %s is on %q, which is not a band set this project has", name, ErrRecordIncomplete, field, raw.BandSet)
	}
	nominals := set.Nominals()
	values := make([]BandValue, 0, len(nominals))
	for _, nominal := range nominals {
		entry, present := raw.Values[strconv.Itoa(nominal)]
		if !present {
			return nil, fmt.Errorf("%s: %w: %s has no entry at %d Hz, and a band read as absent is not a band read as zero",
				name, ErrRecordIncomplete, field, nominal)
		}
		v, err := readBandValue(name, field, nominal, entry)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	if len(raw.Values) != len(nominals) {
		return nil, fmt.Errorf("%s: %w: %s has %d entries and %s has %d bands",
			name, ErrRecordIncomplete, field, len(raw.Values), set, len(nominals))
	}
	return &LabSpectrum{quantity: quantity, set: set, values: values}, nil
}

func readBandValue(name, field string, nominal int, entry json.RawMessage) (BandValue, error) {
	var measured float64
	if err := json.Unmarshal(entry, &measured); err == nil {
		return BandValue{nominal: nominal, value: measured}, nil
	}
	var bound boundDocument
	dec := json.NewDecoder(bytesReader(entry))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&bound); err != nil {
		return BandValue{}, fmt.Errorf("%s: %w: %s at %d Hz is neither a measured value nor a bound: %v",
			name, ErrRecordIncomplete, field, nominal, err)
	}
	if bound.LimitedBy == "" {
		return BandValue{}, fmt.Errorf("%s: %w: %s at %d Hz is a bound with nothing saying what limited the measurement, and a consumer deciding whether the bound is usable has to read that",
			name, ErrRecordIncomplete, field, nominal)
	}
	switch bound.Bound {
	case "at_least":
		return BandValue{nominal: nominal, value: bound.Value, bounded: true, atLeast: true, limitedBy: bound.LimitedBy}, nil
	case "at_most":
		return BandValue{nominal: nominal, value: bound.Value, bounded: true, limitedBy: bound.LimitedBy}, nil
	default:
		return BandValue{}, fmt.Errorf("%s: %w: %s at %d Hz has the bound %q, and a bound is at_least or at_most",
			name, ErrRecordIncomplete, field, nominal, bound.Bound)
	}
}

func joinAnd(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		out := ""
		for i, p := range parts {
			switch {
			case i == 0:
				out = p
			case i == len(parts)-1:
				out += " and " + p
			default:
				out += ", " + p
			}
		}
		return out
	}
}

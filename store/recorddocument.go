package store

// The shape of a component record on disk, and the route back out to it.
//
// These types mirror data/schema/component-record.schema.json field for field.
// They are separate from the types in record.go on purpose: this half knows what
// the file looks like and nothing about what is required of a construction, and
// that half knows the construction and nothing about JSON. Collapsing the two
// would put the refusals inside the decoder, where a field's absence and a
// field's zero value are the same thing.
//
// Every optional field is a pointer or a slice for that reason. A record that
// omits the mass per unit area and a record that states it as zero are different
// records, and a decoder into a plain float64 cannot tell them apart.
//
// WriteConstruction is the way back. It is not a general JSON writer: it writes
// exactly the fields the construction carries, so that reading a record and
// writing it again is the identity on everything the type holds. What the type
// does not hold is refused on the way in rather than dropped here, which is what
// makes that sentence true rather than nearly true.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

type recordDocument struct {
	Schema              *int                `json:"schema"`
	ID                  string              `json:"id"`
	Kind                string              `json:"kind"`
	Basis               string              `json:"basis,omitempty"`
	MassPerArea         float64             `json:"mass_per_area"`
	Thickness           float64             `json:"thickness"`
	Layers              []layerDocument     `json:"layers,omitempty"`
	AirborneLab         *spectrumDocument   `json:"airborne_lab,omitempty"`
	ImpactLab           *spectrumDocument   `json:"impact_lab,omitempty"`
	Improvement         *spectrumDocument   `json:"improvement,omitempty"`
	AirborneSingle      *airborneDocument   `json:"airborne_single,omitempty"`
	ImpactSingle        *impactDocument     `json:"impact_single,omitempty"`
	SpecimenArea        float64             `json:"specimen_area,omitempty"`
	SpecimenEdges       []float64           `json:"specimen_edges,omitempty"`
	LabLossFactor       float64             `json:"lab_loss_factor,omitempty"`
	LabLossFactorAbsent string              `json:"lab_loss_factor_absent,omitempty"`
	BaseConstruction    string              `json:"base_construction,omitempty"`
	Superseded          []json.RawMessage   `json:"superseded,omitempty"`
	Provenance          *provenanceDocument `json:"provenance"`
}

type layerDocument struct {
	Material    string  `json:"material"`
	Thickness   float64 `json:"thickness"`
	Density     float64 `json:"density,omitempty"`
	MassPerArea float64 `json:"mass_per_area,omitempty"`
	Attachment  string  `json:"attachment"`
}

type spectrumDocument struct {
	BandSet  string                     `json:"band_set"`
	Quantity string                     `json:"quantity"`
	Unit     string                     `json:"unit"`
	Values   map[string]json.RawMessage `json:"values"`
}

type boundDocument struct {
	Bound     string  `json:"bound"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	LimitedBy string  `json:"limited_by"`
}

type uncertaintyDocument struct {
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Source string  `json:"source"`
}

type airborneDocument struct {
	Rw          int                  `json:"Rw"`
	C           map[string]int       `json:"C,omitempty"`
	Ctr         map[string]int       `json:"Ctr,omitempty"`
	Uncertainty *uncertaintyDocument `json:"uncertainty,omitempty"`
	Unit        string               `json:"unit"`
}

type impactDocument struct {
	Lnw         int                  `json:"Lnw"`
	CI          map[string]int       `json:"CI,omitempty"`
	Uncertainty *uncertaintyDocument `json:"uncertainty,omitempty"`
	Unit        string               `json:"unit"`
}

type provenanceDocument struct {
	Laboratory          string             `json:"laboratory,omitempty"`
	ReportNumber        string             `json:"report_number,omitempty"`
	ReportDate          string             `json:"report_date,omitempty"`
	TestStandards       *standardsDocument `json:"test_standards,omitempty"`
	Client              string             `json:"client,omitempty"`
	SpecimenDesignation string             `json:"specimen_designation,omitempty"`
	ProductDesignation  string             `json:"product_designation,omitempty"`
	ObtainedFrom        string             `json:"obtained_from,omitempty"`
	DescribedFrom       string             `json:"described_from,omitempty"`
	EnteredBy           string             `json:"entered_by,omitempty"`
	EnteredOn           string             `json:"entered_on,omitempty"`
}

type standardsDocument struct {
	Airborne       string `json:"airborne,omitempty"`
	Impact         string `json:"impact,omitempty"`
	Facility       string `json:"facility,omitempty"`
	AirborneRating string `json:"airborne_rating,omitempty"`
	ImpactRating   string `json:"impact_rating,omitempty"`
}

// value turns a decoded rating into the one the construction carries, and keeps
// absent absent. A record with no rating and a record with a rating of zero
// decibels are different records.
func (d *airborneDocument) value() *AirborneRating {
	if d == nil {
		return nil
	}
	r := AirborneRating{Rw: d.Rw, C: AdaptationTerms(d.C), Ctr: AdaptationTerms(d.Ctr)}
	if d.Uncertainty != nil {
		r.Uncertainty = &Uncertainty{Value: d.Uncertainty.Value, Source: d.Uncertainty.Source}
	}
	return &r
}

func (d *impactDocument) value() *ImpactRating {
	if d == nil {
		return nil
	}
	r := ImpactRating{Lnw: d.Lnw, CI: AdaptationTerms(d.CI)}
	if d.Uncertainty != nil {
		r.Uncertainty = &Uncertainty{Value: d.Uncertainty.Value, Source: d.Uncertainty.Source}
	}
	return &r
}

func bytesReader(src []byte) io.Reader { return bytes.NewReader(src) }

// WriteConstruction writes a construction as a component record.
//
// Two spaces of indentation and a final newline, which is what the records and
// the schema in this tree already use, so a record written here and one entered
// by hand are the same shape in a diff.
func WriteConstruction(w io.Writer, c Construction) error {
	doc, err := c.document()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// document is the construction as the file holds it.
func (c Construction) document() (recordDocument, error) {
	version := recordSchemaVersion
	doc := recordDocument{
		Schema:              &version,
		ID:                  c.id,
		Kind:                c.kind.String(),
		Basis:               c.basis.String(),
		MassPerArea:         c.massPerArea,
		Thickness:           c.thickness,
		SpecimenArea:        c.specimenArea,
		SpecimenEdges:       c.specimenEdges,
		LabLossFactor:       c.labLossFactor,
		LabLossFactorAbsent: c.lossFactorAbsent,
	}
	for _, l := range c.layers {
		doc.Layers = append(doc.Layers, layerDocument{
			Material:    l.Material,
			Thickness:   l.Thickness,
			Density:     l.Density,
			MassPerArea: l.MassPerArea,
			Attachment:  l.Attachment.String(),
		})
	}
	var err error
	if doc.AirborneLab, err = c.airborneLab.document(); err != nil {
		return recordDocument{}, err
	}
	if doc.ImpactLab, err = c.impactLab.document(); err != nil {
		return recordDocument{}, err
	}
	if c.airborneSingle != nil {
		doc.AirborneSingle = &airborneDocument{
			Rw:   c.airborneSingle.Rw,
			C:    c.airborneSingle.C,
			Ctr:  c.airborneSingle.Ctr,
			Unit: SoundReductionIndex.Unit(),
		}
		if u := c.airborneSingle.Uncertainty; u != nil {
			doc.AirborneSingle.Uncertainty = &uncertaintyDocument{Value: u.Value, Unit: SoundReductionIndex.Unit(), Source: u.Source}
		}
	}
	if c.impactSingle != nil {
		doc.ImpactSingle = &impactDocument{
			Lnw:  c.impactSingle.Lnw,
			CI:   c.impactSingle.CI,
			Unit: NormalisedImpactLevel.Unit(),
		}
		if u := c.impactSingle.Uncertainty; u != nil {
			doc.ImpactSingle.Uncertainty = &uncertaintyDocument{Value: u.Value, Unit: NormalisedImpactLevel.Unit(), Source: u.Source}
		}
	}
	p := c.provenance
	doc.Provenance = &provenanceDocument{
		Laboratory:          p.Laboratory,
		ReportNumber:        p.ReportNumber,
		ReportDate:          p.ReportDate,
		Client:              p.Client,
		SpecimenDesignation: p.SpecimenDesignation,
		ProductDesignation:  p.ProductDesignation,
		ObtainedFrom:        p.ObtainedFrom,
		DescribedFrom:       p.DescribedFrom,
		EnteredBy:           p.EnteredBy,
		EnteredOn:           p.EnteredOn,
	}
	if s := p.TestStandards; s != (TestStandards{}) {
		doc.Provenance.TestStandards = &standardsDocument{
			Airborne:       s.Airborne,
			Impact:         s.Impact,
			Facility:       s.Facility,
			AirborneRating: s.AirborneRating,
			ImpactRating:   s.ImpactRating,
		}
	}
	return doc, nil
}

// document is one laboratory spectrum as the file holds it, band by band, in
// whichever of the two shapes each band was read in.
func (s *LabSpectrum) document() (*spectrumDocument, error) {
	if s == nil {
		return nil, nil
	}
	setName, err := bandSetName(s.set)
	if err != nil {
		return nil, err
	}
	doc := &spectrumDocument{
		BandSet:  setName,
		Quantity: s.quantity.Symbol(),
		Unit:     s.quantity.Unit(),
		Values:   map[string]json.RawMessage{},
	}
	for _, v := range s.values {
		key := strconv.Itoa(v.nominal)
		if !v.bounded {
			raw, err := json.Marshal(v.value)
			if err != nil {
				return nil, fmt.Errorf("the %d Hz band cannot be written: %w", v.nominal, err)
			}
			doc.Values[key] = raw
			continue
		}
		bound := "at_most"
		if v.atLeast {
			bound = "at_least"
		}
		raw, err := json.Marshal(boundDocument{Bound: bound, Value: v.value, Unit: s.quantity.Unit(), LimitedBy: v.limitedBy})
		if err != nil {
			return nil, fmt.Errorf("the %d Hz band cannot be written: %w", v.nominal, err)
		}
		doc.Values[key] = raw
	}
	return doc, nil
}

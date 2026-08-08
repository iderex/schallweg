package store

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/iderex/schallweg/acoustic"
)

// The refusals reading a spectrum document makes, as values a caller can test
// for rather than sentences a caller has to match. docs/formats/spectrum.md is
// the authority for the format and states what each one is for.
var (
	// ErrNotASpectrumDocument is a first line that does not name this format.
	ErrNotASpectrumDocument = errors.New("not a spectrum document")
	// ErrUnsupportedVersion is a document at a major version this reader does
	// not know. A higher major exists to say that the fields this reader
	// recognises may no longer mean what it thinks, so it refuses rather than
	// reading the parts it recognises.
	ErrUnsupportedVersion = errors.New("unsupported format version")
	// ErrHeader is a header line that is absent, repeated, out of order or not
	// one of the three.
	ErrHeader = errors.New("malformed header")
	// ErrUnknownQuantity is a quantity symbol this reader does not define.
	ErrUnknownQuantity = errors.New("unknown quantity")
	// ErrWrongUnit is a unit that is not the one the declared quantity is in.
	ErrWrongUnit = errors.New("unit is not the quantity's unit")
	// ErrLineShape is a line whose fields, spacing or bytes are not what the
	// format allows.
	ErrLineShape = errors.New("malformed line")
	// ErrNumberFormat is a value that is not a decimal number in the one
	// spelling this format has. It is the refusal that keeps a decimal comma,
	// a digit group and an exponent out.
	ErrNumberFormat = errors.New("not a number in this format's one spelling")
	// ErrImplausible is a value outside the range any measurement of these
	// quantities occupies, which is how a frequency column pasted into the
	// value column is caught.
	ErrImplausible = errors.New("value is outside the plausible range")
	// ErrMissingBands is a document that has no value in one or more bands of
	// the set it declares. The format can express that and a spectrum cannot
	// hold it.
	//
	// Two ways in and one refusal out. A band written as missing says so, and a
	// band with no line at all says nothing, and to a calculation the two are
	// the same absence. Refusing them apart would let the second arrive as a
	// count that names no band, which is the message a transcriber cannot act
	// on.
	ErrMissingBands = errors.New("document has no value in a band of its set")
)

// The plausible range for a band value, in decibels.
//
// It is a judgement of this project and not a quantity from the standard: wide
// enough to hold anything a laboratory reports for either quantity defined here,
// and narrow enough to refuse a frequency column pasted into the value column.
// docs/formats/spectrum.md says what that catches and what it does not.
const (
	plausibleLow  = -20.0
	plausibleHigh = 150.0
)

// magic is the first field of the first line, and formatVersion is the major
// version this reader implements.
const (
	magic         = "schallweg-spectrum"
	formatVersion = 1
	missingToken  = "missing"
)

// A Quantity is what the values in a document are.
//
// It is in the document because a sound reduction index and a normalised impact
// sound pressure level are the same kind of number, in the same unit, over the
// same bands. A tool reading one as the other is wrong by tens of decibels and
// looks entirely ordinary while being so.
//
// The zero value is not a quantity and every route through it refuses.
type Quantity uint8

const (
	// SoundReductionIndex is R, the sound reduction index.
	SoundReductionIndex Quantity = iota + 1
	// NormalisedImpactLevel is Ln, the normalised impact sound pressure level.
	NormalisedImpactLevel
)

// Symbol is how the quantity is written in a document.
func (q Quantity) Symbol() string {
	switch q {
	case SoundReductionIndex:
		return "R"
	case NormalisedImpactLevel:
		return "Ln"
	default:
		return ""
	}
}

// Unit is the unit the quantity is in, as it is written in a document.
func (q Quantity) Unit() string {
	switch q {
	case SoundReductionIndex, NormalisedImpactLevel:
		return "dB"
	default:
		return ""
	}
}

// String names the quantity the way an error message should say it.
func (q Quantity) String() string {
	switch q {
	case SoundReductionIndex:
		return "the sound reduction index R"
	case NormalisedImpactLevel:
		return "the normalised impact sound pressure level Ln"
	default:
		return fmt.Sprintf("unknown quantity %d", uint8(q))
	}
}

// quantityBySymbol is the reverse of Symbol, and adding a quantity is adding an
// entry here and a row to the table in docs/formats/spectrum.md.
var quantityBySymbol = map[string]Quantity{
	"R":  SoundReductionIndex,
	"Ln": NormalisedImpactLevel,
}

// bandSetByName is the two band sets as a document names them. There is no third
// value, and a document on some other set of bands does not exist.
var bandSetByName = map[string]acoustic.BandSet{
	"core":     acoustic.Core,
	"extended": acoustic.Extended,
}

// A Document is one spectrum and what is needed to read it correctly.
//
// It carries no provenance, no date and no element identity. What a document is
// and is not is docs/formats/spectrum.md.
type Document struct {
	Quantity Quantity
	Spectrum acoustic.Spectrum
}

// Read parses one spectrum document.
//
// name is what the refusals call the document. It is a parameter rather than
// something read from a file handle because this reader takes bytes from
// anywhere, and a refusal that cannot say which file it is about is a refusal
// somebody has to go and correlate by hand.
func Read(name string, r io.Reader) (Document, error) {
	lines, err := readLines(name, r)
	if err != nil {
		return Document{}, err
	}
	if len(lines) == 0 {
		return Document{}, fmt.Errorf("%s: %w: the document is empty", name, ErrNotASpectrumDocument)
	}

	if err := readMagic(name, lines[0]); err != nil {
		return Document{}, err
	}

	quantity, set, err := readHeader(name, lines)
	if err != nil {
		return Document{}, err
	}

	return readBands(name, lines, quantity, set)
}

// readLines splits the input into lines, accepting a line feed or a carriage
// return and a line feed, and refuses any byte the format does not allow.
//
// The byte check is here rather than at each field because it is the one place
// that sees every byte: a tab, a stray carriage return inside a line, a control
// character and anything outside ASCII are all one refusal made once.
func readLines(name string, r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			return nil, fmt.Errorf("%s:%d: %w: a document has no blank line", name, n, ErrLineShape)
		}
		for i := 0; i < len(line); i++ {
			if line[i] < 0x20 || line[i] > 0x7e {
				return nil, fmt.Errorf("%s:%d: %w: byte %d is 0x%02x, and every byte of a document is printable ASCII",
					name, n, ErrLineShape, i+1, line[i])
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: reading the document: %w", name, err)
	}
	return lines, nil
}

// fields splits a line on single spaces and refuses the spacings the format does
// not allow: a leading space, a trailing space, and two spaces together all
// produce an empty field.
func fields(name string, n int, line string, want int) ([]string, error) {
	parts := strings.Split(line, " ")
	for _, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("%s:%d: %w: a field separator is exactly one space, with none at either end",
				name, n, ErrLineShape)
		}
	}
	if len(parts) != want {
		return nil, fmt.Errorf("%s:%d: %w: %d fields were given and this line has %d",
			name, n, ErrLineShape, len(parts), want)
	}
	return parts, nil
}

// readMagic checks the first line names this format and a major version this
// reader implements.
func readMagic(name, line string) error {
	parts, err := fields(name, 1, line, 2)
	if err != nil {
		return fmt.Errorf("%s:1: %w: the first line is %q, and it names the format and its major version",
			name, ErrNotASpectrumDocument, line)
	}
	if parts[0] != magic {
		return fmt.Errorf("%s:1: %w: the first field is %q and this format is %q", name, ErrNotASpectrumDocument, parts[0], magic)
	}
	major, err := wholeNumber(parts[1])
	if err != nil {
		return fmt.Errorf("%s:1: %w: the major version is %q", name, ErrNotASpectrumDocument, parts[1])
	}
	if major != formatVersion {
		return fmt.Errorf("%s:1: %w: the document is version %d and this reader implements version %d",
			name, ErrUnsupportedVersion, major, formatVersion)
	}
	return nil
}

// headerKeys are the three header lines, in the order a document carries them.
var headerKeys = [...]string{"quantity", "unit", "band-set"}

// readHeader reads the three header lines and checks the unit against the
// quantity.
func readHeader(name string, lines []string) (Quantity, acoustic.BandSet, error) {
	if len(lines) < 1+len(headerKeys) {
		return 0, 0, fmt.Errorf("%s: %w: the document has %d lines and the header alone is %d",
			name, ErrHeader, len(lines), 1+len(headerKeys))
	}

	values := map[string]string{}
	for i, key := range headerKeys {
		n := i + 2
		parts, err := fields(name, n, lines[i+1], 2)
		if err != nil {
			return 0, 0, err
		}
		if parts[0] != key {
			return 0, 0, fmt.Errorf("%s:%d: %w: this line is %q and the header's line %d is %q",
				name, n, ErrHeader, parts[0], i+1, key)
		}
		values[key] = parts[1]
	}

	quantity, known := quantityBySymbol[values["quantity"]]
	if !known {
		return 0, 0, fmt.Errorf("%s:2: %w: %q is not a quantity this reader defines",
			name, ErrUnknownQuantity, values["quantity"])
	}
	if values["unit"] != quantity.Unit() {
		return 0, 0, fmt.Errorf("%s:3: %w: the document says %q and %s is in %q",
			name, ErrWrongUnit, values["unit"], quantity, quantity.Unit())
	}
	set, known := bandSetByName[values["band-set"]]
	if !known {
		return 0, 0, fmt.Errorf("%s:4: %w: %q is not a band set, which is %q or %q",
			name, ErrHeader, values["band-set"], "core", "extended")
	}
	return quantity, set, nil
}

// readBands reads the band lines and builds the spectrum.
//
// The band centres are read out of the document rather than taken from the
// declared set, so a document whose declared set and actual bands disagree is
// refused by the spectrum's own constructor instead of being trusted on one of
// the two.
func readBands(name string, lines []string, quantity Quantity, set acoustic.BandSet) (Document, error) {
	body := lines[1+len(headerKeys):]
	nominals := make([]int, 0, len(body))
	values := make([]float64, 0, len(body))
	var recordedMissing []int

	for i, line := range body {
		n := i + 2 + len(headerKeys)
		parts, err := fields(name, n, line, 3)
		if err != nil {
			return Document{}, err
		}
		if parts[0] != "band" {
			return Document{}, fmt.Errorf("%s:%d: %w: this line is %q and a band line begins with %q",
				name, n, ErrLineShape, parts[0], "band")
		}
		nominal, err := wholeNumber(parts[1])
		if err != nil {
			return Document{}, fmt.Errorf("%s:%d: %w: the band centre is %q, which is a whole number of hertz",
				name, n, ErrLineShape, parts[1])
		}
		nominals = append(nominals, nominal)

		if parts[2] == missingToken {
			recordedMissing = append(recordedMissing, nominal)
			values = append(values, 0)
			continue
		}
		v, err := decimal(parts[2])
		if err != nil {
			return Document{}, fmt.Errorf("%s:%d: %w: the value at %d Hz is %q", name, n, err, nominal, parts[2])
		}
		if v < plausibleLow || v > plausibleHigh {
			return Document{}, fmt.Errorf("%s:%d: %w: the value at %d Hz is %s dB, and a value lies between %g dB and %g dB",
				name, n, ErrImplausible, nominal, parts[2], plausibleLow, plausibleHigh)
		}
		values = append(values, v)
	}

	if absent := absentBands(set, nominals, recordedMissing); len(absent) > 0 {
		return Document{}, fmt.Errorf("%s: %w: %s has no value at %s, and a spectrum has a value in every band of its set",
			name, ErrMissingBands, quantity, strings.Join(absent, ", "))
	}

	spectrum, err := acoustic.New(set, nominals, values)
	if err != nil {
		return Document{}, fmt.Errorf("%s: the document declares the %s: %w", name, set, err)
	}
	return Document{Quantity: quantity, Spectrum: spectrum}, nil
}

// absentBands is the bands of the declared set the document supplied no value
// for, named and in the set's order.
//
// It is where a band written as missing and a band with no line at all become
// one refusal. A laboratory that reports from 100 Hz where the calculation wants
// 50 Hz, a certificate that stops at 3150 Hz, and a manufacturer's summary of
// four values are the same defect arriving in three shapes, and a reader that
// answered the first with a named band and the others with a count would be
// telling a transcriber to go and find out which band on two of the three.
//
// The second half is computed only when every band centre in the document
// belongs to the declared set. A document carrying a band the set does not have
// is a different defect, the declared set and the actual bands disagreeing, and
// naming the set's own bands as absent there would describe it wrongly: the
// spectrum's constructor is what names that one and it says which position
// disagrees.
func absentBands(set acoustic.BandSet, supplied, recordedMissing []int) []string {
	inSet := map[int]bool{}
	for _, n := range set.Nominals() {
		inSet[n] = true
	}
	present := map[int]bool{}
	for _, n := range supplied {
		if !inSet[n] {
			return names(recordedMissing)
		}
		present[n] = true
	}
	for _, n := range recordedMissing {
		present[n] = false
	}

	out := make([]string, 0, len(set.Nominals()))
	for _, n := range set.Nominals() {
		if !present[n] {
			out = append(out, fmt.Sprintf("%d Hz", n))
		}
	}
	return out
}

// names writes band centres the way a refusal says them.
func names(nominals []int) []string {
	out := make([]string, 0, len(nominals))
	for _, n := range nominals {
		out = append(out, fmt.Sprintf("%d Hz", n))
	}
	return out
}

// wholeNumber parses a run of ASCII digits and nothing else. It refuses a sign,
// a separator and anything the language's own integer parser would accept beyond
// the digits, so a band centre and a version have one spelling each.
func wholeNumber(s string) (int, error) {
	if s == "" {
		return 0, ErrLineShape
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, ErrLineShape
		}
	}
	return strconv.Atoi(s)
}

// decimal parses this format's one spelling of a number: an optional minus sign,
// one or more digits, and optionally a full stop with one or more digits after
// it.
//
// The grammar is written out rather than handed to a general-purpose parser
// because what has to be refused is exactly what such a parser accepts: an
// exponent, a leading plus, a hexadecimal float, and the words for infinity and
// for not-a-number. Nothing here reads a locale, so a document means the same
// thing wherever it is read, which is the property docs/formats/spectrum.md
// exists to hold. The value is handed to the language's parser only once the
// spelling has been accepted, so the digits become a float64 by the same route
// as everywhere else rather than by arithmetic written here.
func decimal(s string) (float64, error) {
	rest := strings.TrimPrefix(s, "-")
	whole, fraction, hasPoint := strings.Cut(rest, ".")
	if !allDigits(whole) {
		return 0, ErrNumberFormat
	}
	if hasPoint && !allDigits(fraction) {
		return 0, ErrNumberFormat
	}
	return strconv.ParseFloat(s, 64)
}

// allDigits reports whether s is one or more ASCII digits and nothing else.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// Write writes a spectrum as a document.
//
// The output is in one form: line feeds, a final newline, and each value in the
// shortest decimal spelling that reads back as the same number. That is what
// makes writing a document read back as the same spectrum, and writing that one
// again produce the same bytes.
//
// There is no route by which this writes a missing band, because a spectrum has
// a value in every band of its set.
func Write(w io.Writer, quantity Quantity, s acoustic.Spectrum) error {
	if quantity.Symbol() == "" {
		return fmt.Errorf("%w: %s", ErrUnknownQuantity, quantity)
	}
	bands := s.Bands()
	if len(bands) == 0 {
		return fmt.Errorf("%w: this spectrum was never constructed", acoustic.ErrUnknownBandSet)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %d\n", magic, formatVersion)
	fmt.Fprintf(&b, "quantity %s\n", quantity.Symbol())
	fmt.Fprintf(&b, "unit %s\n", quantity.Unit())
	name, err := bandSetName(s.Set())
	if err != nil {
		return err
	}
	fmt.Fprintf(&b, "band-set %s\n", name)

	for _, band := range bands {
		v, err := s.At(band)
		if err != nil {
			return fmt.Errorf("reading the spectrum at %s: %w", band, err)
		}
		if v < plausibleLow || v > plausibleHigh {
			return fmt.Errorf("%w: the value at %s is %g dB, and a document holds a value between %g dB and %g dB",
				ErrImplausible, band, v, plausibleLow, plausibleHigh)
		}
		fmt.Fprintf(&b, "band %d %s\n", band.Nominal(), strconv.FormatFloat(v, 'f', -1, 64))
	}

	_, err = io.WriteString(w, b.String())
	return err
}

// bandSetName is the reverse of bandSetByName, and it refuses a set the format
// has no name for rather than writing a document nothing can read back.
func bandSetName(set acoustic.BandSet) (string, error) {
	for name, s := range bandSetByName {
		if s == set {
			return name, nil
		}
	}
	return "", fmt.Errorf("%w: %s", acoustic.ErrUnknownBandSet, set)
}

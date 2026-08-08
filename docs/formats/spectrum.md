# The spectrum exchange format

Version 1. This document is the authority for the format. The reader in `store/`
implements it and does not extend it, so a document this text calls valid and the
reader refuses is a defect in the reader.

## What it is for

A spectrum arrives from a certificate transcribed by hand, from a laboratory's
export, or from a spreadsheet somebody has kept for fifteen years. It leaves
again into a report and into whatever the reader of that report does next. This
is the one shape those bytes take on the way in and on the way out.

Everything about the format is chosen so that a document means the same thing to
every reader on every machine. That is why it is strict: a format that accepts
two spellings of a number is a format where two people disagree about a value
and both can point at the file.

## The document

Every line is one directive. Line one names the format and its major version,
the next three are the header, and the rest are the band values.

    schallweg-spectrum 1
    quantity R
    unit dB
    band-set core
    band 100 42.1
    band 125 44.3
    ...
    band 3150 61.4

Lines are separated by a line feed or by a carriage return and a line feed. Both
are accepted because a document written on Windows is not a different document,
and the reader is proved against a fixture whose bytes carry the carriage
returns.

A field separator is exactly one space. There is no leading space, no trailing
space and no tab anywhere in a document, and no blank line. Every byte of a
document is printable ASCII; the file may end with a newline or without one.

## The header

The three header lines appear once each, in the order above.

`quantity` names what the values are. Two are defined:

| Symbol | What it is | Unit |
| --- | --- | --- |
| `R` | Sound reduction index | `dB` |
| `Ln` | Normalised impact sound pressure level | `dB` |

The quantity is in the document so that a sound reduction spectrum cannot be
read as an impact level spectrum. Those two are the same kind of number, in the
same unit, over the same bands, and a tool that reads one as the other produces a
result that is wrong by fifty decibels and looks entirely ordinary.

`unit` names the unit and has to be the one the quantity is in. It is stated
rather than implied because a document that carries its unit can be checked, and
a document that implies it can only be trusted.

`band-set` is `core` or `extended`, the two band sets
[../decisions/frequency-bands.md](../decisions/frequency-bands.md) defines. There
is no third value and there is no document on some other set of bands.

Adding a quantity is adding a row to the table above and a value to the reader's
list, and it does not change the version. Removing one, changing what one means,
or changing a unit is a new major version.

## The band lines

One line per band of the declared set, in the set's order, lowest frequency
first:

    band <nominal centre frequency in hertz> <value>

The nominal centre frequency is the designation a certificate prints: `100`,
`125`, `160`. It is a whole number of hertz and never an exact centre frequency,
because two roundings of one band would otherwise arrive as two bands.

The band centres are in the document alongside the values even though `band-set`
already determines them. That is deliberate redundancy: it is what makes a
document self-describing, and it is what lets the reader refuse a file whose
declared set and actual bands disagree instead of trusting one of the two.

A value is a decimal number or the word `missing`.

## Numbers, and the one thing that breaks in Europe

A decimal number is an optional minus sign, one or more digits, and optionally a
full stop followed by one or more digits. That is the whole grammar and
everything else is refused:

| Refused | Why |
| --- | --- |
| `42,1` | A comma is the decimal separator in most of the countries this project is written for and the thousands separator in the rest. A reader that guesses reads it as forty-two point one or as four hundred and twenty-one, and both are defensible. |
| `1 234.5` and `1,234.5` | Digit grouping, for the same reason in the other direction. |
| `4.21e1` | An exponent is a second spelling of a number a certificate never prints. |
| `+42.1` | A second spelling of the same value. |
| `NaN`, `Inf`, `-Inf` | Not measurements. A spectrum cannot hold one. |
| `0x1p+5` | A spelling of a floating point number that the language accepts and no certificate uses. |
| `42.` and `.1` | A separator with nothing on one side of it, which is how a truncated line looks. |

Nothing in the reader consults a locale, an environment variable or the machine's
regional settings, so a document reads identically wherever it is read. That is
the property this section exists for, and it is the reason the grammar is written
out rather than delegated to a general-purpose number parser.

A value has to lie between -20 dB and 150 dB. That range is a judgement of this
project rather than anything from the standard: it is wide enough to hold
anything a laboratory reports for either quantity and narrow enough to refuse a
frequency column pasted into the value column, which is the transcription mistake
that otherwise produces a plausible-looking result. It catches that mistake from
the 160 Hz band upward and not below it, and the reader refuses the document on
the first band that is out of range.

## A band that was not measured

`missing` is how a document says a band has no value:

    band 160 missing

The format can express it and the reader will not turn it into a spectrum. A
document with any missing band is refused, and the refusal names the quantity and
every band that was missing rather than the first one, so a transcriber learns
the size of what they have to go and find.

A band with no line at all is refused the same way and by the same name. To a
calculation the two are one absence, so the reader treats them as one: a
document that declares the core set and carries fifteen of its bands is refused
naming the sixteenth, rather than by a count that leaves a transcriber to work
out which one it was. That holds only where every band centre in the document
belongs to the declared set. A document carrying a band the set does not have is
a different defect, the declaration and the bands disagreeing, and it is refused
as that.

Missing rather than omitted still says more, and that is why the word exists. A
band written as missing says a laboratory looked and has no value; a band left
out says nothing, and is indistinguishable from a document that was truncated on
the way here. Neither is a band whose value is zero, which in decibels is not
absence but a very quiet band.
[../decisions/frequency-bands.md](../decisions/frequency-bands.md) is where that
is argued, and it is why the reader refuses rather than filling the band in.

## Refusals

The reader names the document, the line and what it expected. Every refusal is a
value a caller can test for rather than a sentence a caller has to match, and
they are listed at the top of `store/spectrum.go`.

A first line that is not `schallweg-spectrum` followed by a major version is
refused. A major version above the one the reader knows is refused, naming both
numbers: a higher major exists precisely to say that the fields the reader
recognises may no longer mean what it thinks, which is the rule
[../decisions/data-format.md](../decisions/data-format.md) already sets for
records.

## Writing

A document this project writes is always in one form: line feeds, no trailing
newline omitted, and each value in the shortest decimal spelling that reads back
as the same number. Writing a spectrum and reading it again gives the same
spectrum, and writing that one again gives the same bytes.

The writer cannot produce a `missing` line, because a spectrum has a value in
every band of its set and there is nothing for it to write.

## What this format is not

It is not the component record format, which is JSON and is
[../decisions/data-format.md](../decisions/data-format.md). A record carries
provenance, a history of superseded values and much else; this carries one
spectrum and what is needed to read it correctly.

It carries no provenance, no measurement date, no laboratory and no element
identity. A document is a spectrum and its quantity, and anything that wants to
say where the numbers came from says it somewhere else.

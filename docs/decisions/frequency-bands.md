# Decision: the frequency band model

Status: decided. This is the most load-bearing type in the kernel and it is
decided once, before there is arithmetic to break.

Almost every defect this kernel will ever have is one of three shapes: a band
index off by one, a missing band treated as zero, or two spectra combined that
were not on the same bands. All three are defects of the container rather than of
the physics, and a container can be built so that none of them can be written
down.

## Resolution

Third-octave bands are the kernel's only internal resolution. Octave bands are a
representation at the edge, never a second internal one.

The reason is that the method works on third octaves, the certificates that feed
it publish third octaves, and a kernel with two internal resolutions has a
conversion in the middle of every operation and a question about which one is
authoritative. Converting once, at the boundary, means the question is asked in
one place and answered there.

The direction of that conversion is not symmetric and the asymmetry is the
useful part. Third-octave values combine into octave values by an energy sum,
which is defined and loses information deliberately, so it is available as an
output. Octave values do not determine the third-octave values inside them, so
the reverse is not a conversion at all: it is an invention with a plausible shape.
The kernel refuses it. Which conversions exist and what is refused is issue #41,
and this document is the reason it has something to refuse.

## What each direction loses

Answered by issue #41 and written here because the section above delegated it.

Third-octave to octave loses the shape of the spectrum inside each octave, and it
loses it completely. Two third-octave spectra that differ by six decibels band by
band inside one octave reach the same octave value when they carry the same
energy, which `TestTwoDifferentSpectraReachTheSameOctaves` in `acoustic` shows
with two such spectra rather than asserting. That is the whole of the loss and it
is deliberate: an octave value is a statement about an octave, and everything
finer than an octave is what the caller gave up by asking for one.

The refusal that came with it is stricter than the sentence above implies, and
the reason is worth having in writing. The conversion refuses a band set that
does not carry all three third-octave bands of every octave, which the core set
does not: it holds 3150 Hz without the rest of the 4000 Hz octave, and none of
the 63 Hz octave. So a core spectrum has no octave form at all. Producing the
five octaves it does cover whole would silently drop 3150 Hz, and an octave
spectrum standing for less energy than its input is exactly the number this
document's next section refuses to let a missing band produce. Extended data is
what an octave form needs, and asking for it is a question about the source
rather than a rounding somebody can absorb.

Octave to third-octave loses nothing, because it does not exist. There is no
function to call, which is the strongest refusal available and the one that
reports at compile time rather than at run time. Anything that produced three
numbers from one would have chosen a shape for the spectrum inside the band, and
that invented shape would then travel through the calculation beside values that
were measured. It is the same invention this document refuses when it refuses to
extrapolate a missing band, and it is refused for the same reason.

There is one thing the conversion cannot refuse and it is stated here rather than
left in the source. A `Spectrum` holds numbers and does not know what quantity
they are, and an energy sum is right for a level and wrong for a ratio: the
octave value of a sound reduction index depends on how energy is distributed
across the octave, which a third-octave spectrum of indices does not say. The
function is named `EnergySumToOctave` so that a caller reaching for it on the
wrong quantity has been told, and a name is not a refusal. The decibel quantity
types that would turn it into one are issue #40.

## Range, and the extended bands

There is one band set definition in the tree, and a spectrum is always on a
declared band set rather than on whatever bands its source happened to carry.

Two band sets are defined. The core set is the sixteen third-octave bands from
100 Hz to 3150 Hz, which is what every laboratory certificate in this field
reports and what every calculation in the first release runs on. The extended set
adds the bands from 50 Hz to 80 Hz below and 4000 Hz to 5000 Hz above, twenty-one
bands in all, which is what a certificate carries when the low frequency
behaviour is the point, and low frequency behaviour is where lightweight
construction actually fails.

The extended bands are optional, and optional has a precise meaning here, which
is the meaning that follows. It does not mean a spectrum may contain
some of them. It means a spectrum is on the core set or on the extended set, that
which one it is on is a property of the value and is never inferred, and that an
operation combining two spectra on different sets refuses rather than padding,
truncating or promoting either of them. There is no such thing as a spectrum with
four of the five extended bands.

The reason for two whole sets rather than one set with holes is that a hole is
the failure. A representation that permits a partly populated extended range
guarantees that some code path will one day treat an absent 50 Hz band as
present, and the arithmetic that does so will be correct in every visible respect
and wrong by whatever that band would have contributed.

Adding a band set later, for a range this project has not planned for, is adding
a definition rather than changing a type. That is deliberate: the set is data
about which bands exist and in what order, and the code is written against the
set rather than against a hard-coded count.

## What is in memory

A spectrum is a band set together with exactly one value per band of that set, in
the set's order. It is constructed only through a constructor that checks the
count, and there is no route that produces one otherwise: no zero value that is a
usable empty spectrum, no exported field to assign to, no way to grow one.

A band is its own type rather than an integer. Code indexes a spectrum with a
band, and the type of a band carries which set it belongs to, so reading a
spectrum at a band from another set does not compile into a neighbouring value.
An integer index would permit the off-by-one that this whole design exists to
remove, and it would permit it in a loop, which is where nobody reads it.

What the type makes impossible to express, stated as a list because it is the
justification for the cost:

- A spectrum with a band missing. The count is checked at construction and cannot
  change afterwards.
- A spectrum with a band present twice, or with its bands out of order.
- Reading or writing a band that is not in the set.
- Combining two spectra on different band sets. The operation refuses, by
  returning an error, rather than working on the overlap.
- An absent band arriving at arithmetic as a zero, which in decibels is not
  absence but a very quiet band, and is the single most expensive confusion
  available in this domain.

The alternative was a mapping from centre frequency to value, and it was rejected
for what it makes easy rather than for what it makes hard. A mapping is pleasant
to build from a parsed file, it accepts any subset without complaining, and it
turns every arithmetic operation into a question about which keys are present
that the author of that operation has to remember to ask. The failure mode is
uniform: the code is correct for the data the author had, and quietly different
for data with one key fewer. A fixed vector is worse to build and refuses to
express the defect at all, and the building happens once at the edge while the
arithmetic happens everywhere.

The cost of the fixed vector is paid at the parser, and it is real. Reading a
file that carries an unfamiliar or partial set of bands has to become either a
spectrum on a defined set or a refusal, and there is no third answer that defers
the decision. That work lands in the data access layer, which is where
[layering.md](layering.md) puts every question about untrusted bytes.

## An input with missing bands

Refuse, naming the element, the quantity and the band that was absent.

Not extrapolate. Extrapolating invents a measurement, and the invention would
then travel through the whole calculation and appear in a result beside values
that were measured, indistinguishable from them at the point where somebody
decides whether to believe the number. That is precisely the failure that
[model-shape.md](model-shape.md) refuses when it declines to synthesise a
spectrum from a single number, and the argument here is the same argument.

Not flag and continue. A flagged partial result is a result, and a result gets
quoted. The flag is a field somebody has to read and a report somebody has to
print, and the number in front of it is what gets copied into a letter.

There is one distinction that has to be kept sharp, because
[result-contents.md](result-contents.md) does allow the kernel to supply a value
it did not receive, and this section must not be read as contradicting it. The
kernel may default a correction it can name, mark it as defaulted, and set the
result incomplete. It may not invent a measurement. A laboratory-to-situ
correction that was not supplied has a defensible neutral value and a name to
record it under. A missing 125 Hz sound reduction index has neither: there is no
neutral value for a band level, and any number put there is a guess about a
physical object.

What the refusal owes the user is a route forward rather than a dead end. It
names what was missing, it says which band set would have been satisfied by what
was supplied, and where a narrower calculation is available on the bands that are
present it says so and leaves the choice to the user. Making the choice for them
is the flag route with better manners.

## Nominal against exact centre frequencies

Decided separately for the code and for the file formats, because they answer to
different readers.

In the code, a band's identity is its nominal designation, and the exact centre
frequency is derived from the band's position in the series by the stated rule
rather than stored. Identity is what the code compares, indexes and prints, and
the nominal designation is unambiguous, short and free of any question about
rounding. The exact frequency is a computed quantity used only where an equation
needs a frequency as a number, and computing it from the rule keeps it in one
place. This is also a tier one number under
[standard-text.md](standard-text.md): the series is a rule that can be written in
a sentence, so the tree computes it and states the rule instead of transcribing a
column of values.

In the file formats, only the nominal designation appears. It is what a
certificate prints, what a user recognises and what a spreadsheet already
contains. Exact frequencies in a file would invite two different roundings of the
same band to appear as two different bands, which is a matching failure that
looks like missing data and would be diagnosed as one.

A file carrying a frequency the nominal series does not name is not silently
matched to the nearest band. It is refused, and the refusal says what was read
and what was expected, because the nearest band is usually right and the case
where it is wrong is undetectable afterwards.

## What downstream work owes this decision

Every kernel issue that touches a spectrum is written against the container
described here rather than against a slice of numbers. The type itself is issue
#39, the conversions are `acoustic/octave.go`, and the exchange format is issue
#47.

Where a later issue names this document, a reader sees the constraint before the
work starts. Where it does not, the constraint applies anyway, because the type
is the mechanism and prose is not.

## What would reopen this decision

A calculation this project takes on that is not defined on third-octave bands. A
certificate source that publishes on a band set neither defined set covers, which
is a new set rather than a change to these. Or evidence from the validation work
that a refused input is common enough in real projects that refusing it makes the
tool unusable, which would be an argument to make the narrower calculation easier
to ask for and not an argument to fill the band in.

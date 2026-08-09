# Sources for the validation corpus

This is the survey that decides what this project can claim about its own
accuracy. It is written before any case is encoded, because a corpus assembled
from whatever turned up first is a corpus whose gaps nobody can see.

Every entry names the address that was read and the date it was read on. Where a
source was found but not read, the entry says so and asserts nothing about its
contents. Where a claim could not be sourced it is marked `UNSUPPORTED` in place.
A marked sentence is not a weaker sourced sentence, it is a different kind of
sentence, and it is left in so that the thinness of the record is visible.

All addresses below were read on 2026-08-09 unless the entry says otherwise.
Publisher and repository pages change without notice, so a later reader should
re-read rather than trust what is written here.

## What was looked for

Four kinds of material, in the order they would be worth most.

A comparison exercise in which several independent implementations computed one
specified case. That is the only kind of source that separates a defect in this
implementation from a limitation of the method, because every participant faced
the same inputs.

A published measurement of a real building alongside a prediction for it. That is
the common kind and it answers a different question: how far the method lands
from the building, with this project's own implementation as one more variable.

A worked example in the standard's own annexes or in a national application
document.

A test case published by an academic group, with inputs complete enough to be
re-run.

## The entries

### Timber buildings, three case studies

Di Nocco, Morandi, Barbaresi and Di Bella, "Use of the ISO 12354 Standard for the
Prediction of the Sound Insulation of Timber Buildings: Application to Three Case
Studies", Building Simulation Applications 2019, pages 141 to 147. Read at
`https://publications.ibpsa.org/proceedings/bsa/2019/papers/bsa2019_9788860461766_18.pdf`.

What it provides. Three buildings under ISO 12354-1 and ISO 12354-2 together: a
timber frame building in Rimini with a 180 mm cross laminated timber floor, a
cross laminated timber building in Bologna with a 180 mm floor and 120 mm walls,
and a cross laminated timber building in Graz with a 160 mm floor and 100 mm
walls. Airborne and impact sound insulation were measured in situ under ISO
16283-1 and ISO 16283-2 and are plotted against calculated values in third octave
bands from 100 Hz to 3150 Hz. The separating element area and the receiving room
volume are printed for each case: 8.7 m squared and 23.4 m cubed, 14.6 m squared
and 46.7 m cubed, 26.1 m squared and 70.6 m cubed.

Its most useful property is that each case is computed several ways rather than
once. Case A carries five variants that differ only in where the input came from,
including one where the vibration reduction indices come from the annex formulas
and one where they come from site measurement. Case B carries three and case C
carries five. That is a study of input sensitivity, which is the quantity this
project most needs to know and the one a single predicted curve cannot show.

Completeness. Not complete. The element sound reduction indices, the impact
levels and the vibration reduction indices are shown as figures rather than
tabulated, so the cases cannot be re-run by a third party without reading values
off a plot. The junction arrangements are given as drawings.

Expected result and uncertainty. It reports curves and a discussion. No numerical
tolerance, agreement criterion or uncertainty figure was found in it.

Availability. Free to download from the address above.

Redistribution. No licence statement was found in the paper. Its inputs are in
any case not in a form that could be redistributed as data.

One sentence in it bears directly on this repository's own position: the authors
say their implementation was calibrated against the examples in the appendix to
ISO 12354. Those examples are the standard's own text, which is tier three under
the standard text decision, so they are not a route this project can take.

### Cross laminated timber floors against a large measurement set

Simmons, "A systematic comparison between EN ISO 12354 calculations of CLT floors
with a large set of laboratory and field measurements", Euronoise 2021, Madeira.
Read at `https://www.sea-acustica.es/archivo/congresos/Madeira21/ID13.pdf`.

What it provides. Calculations under EN ISO 12354-1 and EN ISO 12354-2 compared
against 23 field measurements taken in 8 buildings with cross laminated timber
floors, from the AkuLite, Aku20 and AkuTimber projects. Results are given as the
average difference between calculated and measured values and the standard
deviation of that difference, in third octave bands from 50 Hz to 5000 Hz and on
the single numbers. The paper states a practical safety margin of 5 dB, and the
margin curve it plots is the average deviation increased by 1.35 times the
standard deviation, described as a 90 per cent margin.

Completeness. Not complete, and not intended to be. Inputs are described in prose
and shown as figures. The calculations were made with commercial software and
with a commercial element database, so re-running the study would need both.

Expected result and uncertainty. This is the entry that carries an agreement
figure rather than a curve, and it is the one worth reading before the tolerance
question is settled. It is a statement about a population of buildings rather
than a per-case tolerance.

Availability. Free to download from the address above.

Redistribution. No licence statement was found in the paper.

### Ventilated facades

Di Nocco, Barbaresi, Morandi and Garai, "Measurements and prediction of sound
insulation of innovative ventilated facade solutions", 23rd International Congress
on Acoustics 2019, Aachen, pages 3498 onward. Read at
`https://pub.dega-akustik.de/ICA2019/data/articles/000971.pdf`.

What it provides. Facade sound insulation measured on two twin test buildings in
San Mauro Pascoli with five facade configurations, compared against the ISO
12354-3 model, with vibration reduction indices for the facade junctions measured
under ISO 10848-1.

This entry is recorded and then set aside. ISO 12354-3 is facade insulation
against outdoor sound, which is not what this project computes, and whether it is
ever in scope is entry 8 of issue #1. It is listed because a survey that quietly
dropped it would look like a survey that did not find it.

Completeness. Not complete. The paper states that laboratory certificates for the
elements could not be retrieved and that the input sound reduction indices were
calculated with commercial software instead, which makes the case a comparison
against a second prediction rather than against measured element data.

Availability. Free to download from the address above.

### Laboratory floor dataset, 30 wooden and 8 concrete constructions

Hongisto, Alakoivu, Virtanen, Hakala, Saarinen, Laukka, Linderholt, Olsson,
Jarnero and Keranen, "Sound insulation dataset of 30 wooden and 8 concrete floors
tested in laboratory conditions", Data in Brief volume 49 (2023), article 109393,
`https://doi.org/10.1016/j.dib.2023.109393`. The data itself is at Mendeley Data,
`https://doi.org/10.17632/y83p8mpryd.1`, where the version read on 2026-08-09
states the licence as CC BY 4.0; the article points at version 2 of the same
dataset.

What it provides. Airborne and impact sound insulation of 38 full scale floor
specimens measured in an accredited laboratory under ISO 10140-2 and ISO 10140-3,
in third octave bands from 20 Hz to 5000 Hz, with single numbers under ISO 717-1
and ISO 717-2, structural drawings per construction, and dynamic stiffness of the
resilient materials.

This is not a comparison exercise and it cannot validate anything on its own. It
is element data, which is the input side, and it is the only material found in
this survey that is both complete enough to use and licensed for redistribution.
Its value to this project is therefore larger than its position in this list
suggests: it is a candidate for the seed database and for the input half of an
encoded case, and it is the one entry here where the answer to "may this be
carried in the tree" is yes with attribution rather than no or unknown.

Completeness. Complete for what it is. It carries no in situ result, no junction
and no flanking transmission, so no building level case can be built from it
alone.

Availability. Both the article and the dataset are free.

Redistribution. CC BY 4.0 on the dataset version read. Attribution obligations
would attach to any record derived from it, and the provenance work in the
database milestone is where that is carried.

### The standard's own annex examples

Named in the timber paper above as the material its implementation was calibrated
against. Not read here and not readable here: the annexes are part of a document
that is sold, and under the standard text decision no part of it comes into this
repository in any form.

Recorded because it is the source a reader will ask about first, and because the
answer is a position rather than an absence.

## Sources found and not read

Each of these appeared in the searches that produced the entries above. None was
read, so nothing is asserted about any of them beyond the address and the title
as the search result gave it.

`https://www.researchgate.net/publication/228487088` , "Development and use of
prediction models in Building Acoustics as in EN 12354". Search summaries
attached to this title describe an inter-laboratory comparison with eight
laboratories and a comparison of 40 building cases against EN 12354 calculations,
and a recommended 3 dB safety margin for heavy partitions. `UNSUPPORTED`: none of
that was read at the source and none of it is asserted. If the description is
accurate this is the most valuable entry in the whole survey and it should be
obtained and read before any case is encoded.

`https://doi.org/10.3390/buildings15203753` , "Measurement and Prediction of
Airborne Sound Insulation Performance of Different Vertical Partition Walls in
Indoor Environments: A Case Study", 2025, described in search results as open
access.

`https://www.mdpi.com/2624-599X/8/1/11` , "Review of Modelling and Prediction
Methods for Flanking Transmissions". The address returned HTTP 403 to the fetch
attempted on 2026-08-09. A review rather than a case, so its worth here is its
reference list.

`https://www.academia.edu/81991324` , "Validation of EN 12354-1 prediction models
by means of Intensity and Vibration measurement techniques in Spanish buildings
involving flanking airborne sound transmission".

`https://www.researchgate.net/publication/269380541` , "Verification of
Calculating Sound Insulation of Building Structures According to EN 12354 with
the Results of Measurements in Site".

`https://www.sciencedirect.com/science/article/abs/pii/S0360132316302098` ,
"Uncertainty of facade sound insulation by a Round Robin Test". A round robin on
measurement rather than on prediction, and on facades, so it is out of scope on
two counts; listed because its title is what a later reader searching for a round
robin will find first.

## What no usable source was found for

Each line below is a negative result over the searches that produced this
document, not a proof that nothing exists.

No comparison exercise was found in which several independent implementations
computed one specified case under parts 1 or 2. Every round robin found was a
round robin on measurement, and the one comparison of implementations that turned
up compared two editions of the standard inside one commercial program rather
than two programs against each other. This is the gap that matters most, because
without such an exercise a disagreement between this kernel and a published
result cannot be attributed.

No case was found whose inputs are tabulated completely enough to be re-run by a
third party. Every comparison read here shows its inputs as figures. That is the
practical obstacle to the whole corpus and it is a property of how this field
publishes rather than of these particular papers.

No published case was found for the impact flanking paths specifically. The
timber study reports impact results at building level, and it says in its own
conclusion that impact sound insulation is less affected by flanking than
airborne is, which is a reason to expect the impact flanking paths to be the
least evidenced part of the model rather than a reason to skip them.

No published case was found for the simplified single number models. Everything
read here computes in third octave bands.

No published case was found that exercises the derivation of the vibration
reduction index from a junction description across the junction types this project
will support. The timber study exercises it for cross laminated timber junctions
and reports that the annex formulas give very different results from the measured
values there, which is evidence about one construction type rather than a case.

## What would be needed to fill each gap

For the attribution gap. Either the multi-laboratory comparison named in the
unread list turns out to be what its description says, or this project builds the
comparison itself under issue #87 by running one input through a second
implementation. The second route needs access to a commercial program and a
reading of its licence, and it produces findings rather than verdicts.

For the tabulated input gap. Ask the authors. Three of the four read entries name
their authors with institutional addresses, and a request for the input tables
behind a published figure is an ordinary thing to send. Whether the answer may be
carried in this tree is a separate question from whether it may be read, and both
have to be settled per source before a case is encoded.

For the impact flanking gap and the junction derivation gap. No route was found
that does not need either a laboratory campaign or material this survey did not
reach. Recording the gap is what this document can do; closing it is not.

For the simplified models. A case computed both ways from one building would
serve, and none was found. This may be the place where a case invented here and
marked as a regression fixture is the honest answer, which is the shape issue #45
already sets out for the rating procedures.

## How this document is kept

An entry is added when a source is found and read, and an entry that was read on
one date is not silently updated: the date it carries is the date its claims were
true at the address it names. A source that moves to a new address gets a new
line rather than an edited one, so that a reader can see the move.

The negative results above are the part most likely to go stale and the part
least likely to be revisited, because nobody re-runs a search that found nothing.
Issue #85 encodes the first case and is where this document is next read.

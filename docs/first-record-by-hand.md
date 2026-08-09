# Entering the first record by hand

One record was entered into the schema from a real published test report, before
any import route exists, to find out what the schema gets wrong while getting it
wrong costs one record instead of a thousand.

The record is `data/floor/ift-17-002083-pr01-x01-x02.json`. This note is the
other half of the exercise and it is the more useful half.

## What was entered, and from where

ift Rosenheim GmbH, Labor Bauakustik, report 17-002083-PR01 (PB X1-F03-04-de-02)
of 2019-03-21, commissioned by Holzbau Deutschland - Institut e.V., read on
2026-08-09 at
`https://informationsdienst-holz.de/fileadmin/Publikationen/6_Arbeitshilfen/PB_Decken_17-002083-PR01_PB_X1-F03-04-de-02.pdf`.

It is 89 pages and covers 30 specimens across 60 measurement sheets: a matrix of
timber joist floors and cross laminated timber floors against variants of screed,
resilient layer, cavity insulation, ballast and suspended ceiling. One specimen
was taken, the one on measurement sheets 1 and 2, measurement numbers X01 and
X02.

Pages 1 to 4 and 24 to 31 were read. The rest was not, and nothing below is a
claim about a page that was not read.

## What the schema required that the report does not print

**A total loss factor for the specimen as tested.** The schema makes
`lab_loss_factor` a dependency of either laboratory spectrum, so a spectrum
cannot be entered without it. The two measurement sheets for this specimen print
the test area, the volumes of both test rooms, the climate, the drying time and
the static air pressure. Neither prints a loss factor, and none was found on any
page that was read.

This is not a gap in one report. The in situ correction exists because a
laboratory value has to be moved to a building's loss factor, and the value it
moves from is the one the schema is asking for. If published certificates do not
carry it, the detailed model's input cannot be assembled from published
certificates at all, and this project needs to know that before it seeds a
database rather than after.

The refusal is real rather than supposed, and it was run rather than reasoned
about. Nothing in this tree validates a record against the schema yet, which is issue
#38, so the run below is an outside validator rather than a route this repository
offers. It reads the record as committed and adds an airborne spectrum to it:

    $ python - <<'EOF'
    import json
    from jsonschema import Draft202012Validator
    schema = json.load(open("data/schema/component-record.schema.json", encoding="utf-8"))
    record = json.load(open("data/floor/ift-17-002083-pr01-x01-x02.json", encoding="utf-8"))
    bands = ["100","125","160","200","250","315","400","500",
             "630","800","1000","1250","1600","2000","2500","3150"]
    record["airborne_lab"] = {"band_set": "core", "quantity": "R", "unit": "dB",
                              "values": {b: 0.0 for b in bands}}
    errs = sorted(Draft202012Validator(schema).iter_errors(record), key=lambda e: e.message)
    print(len(errs), "error(s)")
    for e in errs:
        print(" -", e.message)
    EOF
    2 error(s)
     - 'lab_loss_factor' is a dependency of 'airborne_lab'
     - 'specimen_edge_length' is a dependency of 'airborne_lab'

The same script without the two added lines prints `0 error(s)`, which is the
record as it stands.

**A single specimen edge length.** `specimen_edge_length` is one number. The
report gives the test opening as 4.0 m by 5.0 m. Which single length the schema
wants out of two dimensions is not stated by the schema, and the four candidates
a reader could pick, either dimension, the perimeter, or something the in situ
correction defines, are not interchangeable. Nothing was entered for it.

## What the report carries that the schema cannot hold

**Band values that are lower bounds rather than values.** Every third octave
value of the sound reduction index on this specimen's airborne sheet is printed
with a "greater than or equal" sign, against a column of the facility's own limit
and a stated maximum of 83 dB for the test area. The specimen out-performs the
test facility, so what was measured is a bound on the specimen and not the
specimen.

The schema's spectrum is a plain number per band. Entering those numbers would
record measurements that were not made, and would do it invisibly, which is the
failure this whole database is arranged against. No airborne spectrum was
entered.

The impact sheet is different and the difference matters: its values carry no
such mark over the bands the core band set covers, and only the two highest bands
outside that set carry a background noise mark. So the impact spectrum for this
specimen is a spectrum, and it was still not entered, for the loss factor reason
above.

**A rating stated twice at two precisions.** The measurement sheets print
Rw (C; Ctr) = 79 (-5; -13) dB and Ln,w (CI) = 42 (0) dB. A summary table in the
same report lists the same two measurements as 79.6 and 41.3 in tenths of a
decibel. The schema has one integer field for each, and the integers from the
measurement sheets were taken.

Recorded as an observation and not as a defect in the report: the pairs are not
related by ordinary rounding in either case, 79.6 going to 79 and 41.3 going to
42, so the two presentations are the outputs of two procedures rather than one
number shown twice. Which procedure a consumer of this database should receive is
a question this exercise raises and does not answer, and it needs the rating
standard to answer it.

**Adaptation terms for four band ranges.** Beside the pair on the rating line the
sheet prints the airborne terms for three further ranges, each named by the range
it covers. The schema holds one C and one Ctr with no field saying which range
they are for. The rating line pair was taken. Issue #43 holds the same gap on the
kernel side; the record side has it too and nothing currently says so.

**A measurement uncertainty.** The report states the uncertainty of the rated
values as plus or minus 1.2 dB for the airborne rating and plus or minus 1.5 dB
for the impact rating, and names the standard the figures come from. The schema
has no field for it, so the record asserts 79 and 42 with no indication that the
laboratory itself brackets them wider than the margins a user will design to.

**The makeup.** The specimen is a joisted floor: a floating screed on a resilient
layer, over a boarded deck fixed to joists at a stated spacing, with insulation
in the cavity and a plasterboard ceiling on resilient hangers. The schema's layer
is a continuous layer with a thickness and either a density or a mass per unit
area. A joist course is not continuous, and what it does acoustically depends on
its spacing, for which the layer type has no field.

So the makeup was published and could not be reduced to the fields the model has.
`layers` is not required of a measured construction, so the record omits it. It
does not carry an empty list: the schema's own description reserves the empty list
for a makeup that was never published, and these two states are different facts
that a consumer would otherwise be unable to tell apart.

**One field for five standards.** `test_standard` is one string, and this report
names one standard for the airborne measurement, one for the impact measurement,
one for the facility and two for the rating. All five went into the field as a
sentence, which makes it readable and defeats the reason the field exists, since
nothing can now compare the edition of the rating standard across two records.

**A product designation for something nobody sells.** The field asks for the
product as the manufacturer names it, and the reason given is that a user who
cannot map a record onto something they can buy cannot use it. This report was
commissioned by an industry association and tests generic constructions. There is
no manufacturer of the floor and no product designation to enter, so the
laboratory's own designation of the specimen was used.

This is worth more attention than a field mismatch usually deserves, because the
certificates that are published openly and are free to read are disproportionately
the ones commissioned by associations and public bodies. The field as written fits
the manufacturer certificates behind a login better than it fits the documents
this project can actually reach.

## Every value that needed interpretation

Ten decisions were made that a schema check would not have caught either way.

1. The ratings were taken as the integers on the measurement sheets rather than
   the tenths in the summary table, because the schema's field is an integer and
   the sheet's line is the one the report labels as the rating.
2. The adaptation terms were taken for the range on the rating line, because the
   other three are printed with their range in their name and this one is not.
3. The thickness came from the measurement sheet, 387 mm, because the specimen
   description points at the measurement sheets for it rather than stating one.
4. The mass per unit area came from the same place and for the same reason,
   180.1 kg per square metre.
5. The specimen area was taken as the product printed on the sheet, 20.0 square
   metres, rather than recomputed from the two dimensions.
6. The product designation is the laboratory's designation of the specimen, for
   the reason above.
7. The test standard field carries a sentence naming five standards.
8. The identity is built from the report number and the two measurement numbers.
   It is traceable, which is the most that can be said for it until issue #56
   decides what makes an identity stable.
9. No spectrum was entered, although the report prints two.
10. No layers were entered, although the report describes the makeup.

Six of those ten are decisions about what a field means rather than about what
the report says, and none of them is recorded anywhere in the record itself. A
second person entering the same report would have to make all ten again and could
differ on at least four.

## The time, and what it does and does not imply

The interval between opening the document and the record file being complete was
under five minutes. That figure is recorded because the issue asks for it, and it
is not offered as an estimate of what entering a thousand records costs. It is one
entry by one route, it excludes finding the document, and it excludes the ten
decisions above, which were settled while this note was written rather than while
the values were typed. A number measured under those conditions does not survive
being multiplied.

What can be counted rather than timed, and what does scale:

- 12 of the report's 89 pages had to be read to enter one record, and 2 of them
  carried the values.
- 8 measured numbers went into the record, out of 23 filled fields, 9 of which
  are provenance.
- 10 decisions were needed that no check would catch, listed above.
- 30 records are available from this one report, and 29 of them would share every
  provenance field with the first.

Those counts point the same way and it is not the way the question assumed. The
transcription is the cheap part. The expensive parts are finding a published
report at all, and answering the ten questions once so that they stop being
answered per record by whoever is entering it. What would actually measure hand
entry is a second report entered after those ten are settled, and that has not
been done.

## What this forces

The schema changes this exercise argues for are filed as issue #161 rather than
made here, because the schema landed under its own issue and a change to it
should be argued against the same reasons that shaped it rather than as a side
effect of one record.

What #161 carries: whether a spectrum value can be marked as a bound, whether
a rating can say which band range its adaptation terms cover, whether the loss
factor dependency should refuse a spectrum or admit one that records it as
absent, what a single specimen edge length means when a report prints two
dimensions, whether the uncertainty a laboratory states belongs in the record,
and whether the product designation field fits a certificate with no product.

The makeup problem is not a schema problem. A joisted assembly is a shape the
element model does not have, and it belongs with the element model rather than
with the file format.

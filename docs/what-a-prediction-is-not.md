# What this is not for, and what a prediction is not

This program computes numbers. It does not measure a building and it does not
decide whether a building is acceptable. The distance between those three things
is where this tool will be misused, and the misuse will not be unlawful. It will
be a computed number treated as a measured one.

The audience this is written for already knows most of it. It is written down
anyway, because the person holding the result is often not the person who ran it.

## A prediction is not a measurement

A measurement is what a building did, once, with the instruments in the rooms and
the doors closed. A prediction is what a model says a building of that
description should do, given values measured on other specimens somewhere else.

The two answer different questions and they can differ by more than the margin
anybody is designing to. A building that computes as meeting a requirement can
fail its acceptance test, and a building that computes as failing can pass. If
the number matters, it gets measured.

Nothing here replaces an acceptance measurement. Where a contract, a building
control body or a client asks whether a building performs, that question is
answered by a test in the finished building under the field measurement
standards, not by this program.

## What the numbers are worth

The honest position on accuracy has two parts and they should not be run
together.

The accuracy the method claims for itself is stated in the standard's own text.
That text is not in this repository and it was not read for this document, so no
figure from it is quoted here. Somebody with a copy can read what it says. The
absence of that figure here is deliberate and is explained in
[docs/decisions/standard-text.md](docs/decisions/standard-text.md).

The accuracy other people have measured is a separate thing, it is published, and
one figure from it is worth carrying. Simmons compared calculations under EN ISO
12354 parts 1 and 2 against 23 field measurements taken in 8 buildings with cross
laminated timber floors and proposed a practical safety margin of 5 dB, built as
the average deviation increased by 1.35 times the standard deviation of the
comparison. That is one population of one construction type, reported by one
author, and it is not a general accuracy figure for this method or for this
program. It is quoted because a reader deciding how much margin to leave is
better served by one sourced number with its limits attached than by a sentence
saying accuracy varies. The entry with its address is in
[docs/validation-sources.md](docs/validation-sources.md).

What this program's own agreement with published results is has not been measured
yet. When it has, it is reported in the validation record, including where it is
weak. Until then, no claim about this implementation's accuracy is made anywhere,
and any sentence you find that seems to make one is a defect worth reporting.

## Where the inputs are already optimistic

A laboratory value is close to an upper bound on what the same element does in a
building. The specimen was built by people who build specimens, mounted in a
facility designed to suppress everything except the path being measured, and
tested once. The wall in the building is none of those things.

The model knows about this and corrects for some of it. It does not know about
the rest, and the rest is not small. A penetration nobody drew. A socket box
back to back with another one. A screed bridged to the wall at one corner during
the pour. A resilient layer compressed by a stack of plasterboard for three weeks.
A junction detail built the way the crew usually builds it rather than the way it
was specified. These are routine and each of them can cost several decibels.

None of that appears in an input file, so none of it appears in the result. A
result that assumes perfect execution is what this program produces, and it says
so here rather than in a footnote somebody reads afterwards.

## It does not say whether anything complies with anything

The method computes quantities. National building rules say what a building has
to reach, and they change on their own schedule.

This program prints the quantities it computed. Where a requirement is compared
against a result, that requirement is a number the operator supplied, and the
output says that it came from the operator. No output names a law, a national
rule set or a regulation, and no output prints a verdict of compliance.

Whether requirement profiles are ever shipped as data is entry 4 of issue #1 and
is not decided. Until it is decided, the position in this section is the whole
position.

## It does not replace the person

Choosing which elements to model, deciding whether a junction detail is what the
drawing shows, judging whether a laboratory value applies to what will actually be
built, and reading a result against what a building is for, are the work of
somebody qualified to do it. This program does none of them.

Running it without that judgement is the misuse worth warning about, and it is
more likely than any of the others, because the output looks finished.

## Where this sits

[NOTICE.md](NOTICE.md) carries the general intended-use notice and this document
is the specific one. They do not conflict: the notice is about lawful use and
this is about a number being read as more than it is.

## What is not yet true

A short version of this belongs in every report the program hands over, because
the report is what leaves the operator's control and reaches somebody who did not
run the calculation. There is no report generator in this tree yet, so that
version does not exist and this document is the only place the warning is
written. Issue #96 builds the report and owes this text a place in it.

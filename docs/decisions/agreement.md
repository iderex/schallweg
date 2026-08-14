# Decision: what agreement means, decided before the first case is run

Status: decided. It is written before any case is encoded, because a tolerance
chosen after seeing a result is not a test, and the condition that reopens it is
at the end.

## What is being compared

A validation case in this project compares a number this implementation computes
against a number somebody else published for the same inputs. It does not
compare a computed number against a measured building.

Those are different questions and only the first one is about this
implementation. A method that lands four decibels from a building is a property
of the method, and it is the same property whichever program computed it. A
program that lands four decibels from another program given identical inputs has
a defect in it. The corpus this repository builds is for the second question, and
every number below follows from that.

[validation-sources.md](../validation-sources.md) is the survey behind this. It
records that no comparison exercise was found in which several independent
implementations computed one specified case, which is the material this question
most wants and which does not appear to exist in reachable form. What was found
instead is method-against-building material, and the one entry carrying a
numerical agreement figure at all is the Euronoise 2021 comparison of calculated
cross laminated timber floors against 23 field measurements in 8 buildings. That
paper states a practical safety margin of 5 decibels and plots a margin curve of
the average deviation increased by 1.35 times its standard deviation.

That figure is not the tolerance here and using it as one would be a mistake
worth naming. It describes how far the method sits from a population of
buildings. Adopting it would let this implementation compute a value several
decibels away from the value the same inputs produce elsewhere and still report
agreement, which is the one outcome this corpus exists to catch.

## The tolerances

**Per band, half of the last digit the source printed, and never smaller than
0.05 decibels.**

A published number is a rounded number. A reproduction of it agrees when it
rounds to the same printed value, and half of the last printed digit is exactly
that condition written as a tolerance. A source printing 52.3 dB is reproduced
within 0.05 dB; a source printing 52 dB is reproduced within 0.5 dB.

The floor exists because the source can print more precision than the procedure
carries. [numeric-contract.md](numeric-contract.md) records that the band values
entering the reference curve procedure are taken to one decimal place before the
search begins, so 0.05 dB is half of the smallest step the arithmetic itself
distinguishes, and a tolerance below it would be asserting agreement on digits
the procedure does not have. That document marks its own precision figure
`CLAIM, NOT VERIFIED` against the clause it comes from, so this floor inherits
that mark: if the clause says something else, the floor moves with it and the
rule above does not.

**On a single-number rating, exact agreement.**

The ratings and the adaptation terms are integers, produced by an integer search.
Half of the last printed digit of an integer is half a decibel, and any
disagreement smaller than that is not representable, so the rule above already
says exact and this states it in the form somebody will look for. A rating that
differs by one is a procedural difference, in the band range, the stopping
condition or where the rounding happened, and it is the defect the corpus is
most likely to find. It is never within tolerance.

**The tolerance is derived from the source and is not a number in this document
to be adjusted.** That is deliberate. A stated constant is a thing somebody
widens on the day a case fails; a rule that reads the published precision has
nothing to widen.

## Symmetry

Symmetric.

The asymmetric argument is real and it belongs somewhere else. For a compliance
tool, a result that is optimistic about a building is worse than one that is
pessimistic, because the optimistic error is the one nobody discovers until the
acceptance measurement. That asymmetry applies to the gap between a prediction
and a building, which is where a safety margin like the 5 decibel figure above
lives, and this project makes no such comparison and states no such margin.

Against another implementation's published result there is no favourable
direction. A value too high and a value too low are the same defect seen from two
sides, and an asymmetric tolerance would hide half of them.

## The four outcomes

A case has one of four outcomes and they are not three.

**Pass.** Every band and every rating within tolerance.

**Fail.** A rating outside tolerance, or a band outside tolerance in a case whose
rating also moved.

**Divergent.** The ratings agree and one or more bands do not. This is the
outcome the issue behind this document asked to be given a name, and it is a real
result rather than a rounding artefact: band-level disagreement partially cancels
in a rating, so an implementation can be wrong in two bands in opposite
directions and produce the right single number. Divergent is not a pass. It is
recorded with the bands that disagreed and their signed deviations, it is
reported beside the passes, and it files a finding against the case rather than
against the run.

**Not comparable.** The source did not publish the inputs as numbers. Where a
value has to be read off a plotted curve, it has no printed precision, so the
rule above has nothing to compute a tolerance from, and a tolerance invented to
cover plot reading would be an invented number, which is what the rule above
keeps out. Such a case
is encoded, run and reported, and it never decides anything. The survey records
that the timber case studies are in this position, which is most of the
building-level material found.

A case's outcome is a property of the case and of the run, and a case moving from
pass to divergent between two runs is itself a finding.

## What a failure does

A fail reds the check that runs the corpus, and a red corpus blocks a release.

A fail is closed in one of two ways and widening a tolerance is neither. Either
this implementation is corrected, or the deviation is recorded as accepted, with
the case, the size and direction of the deviation, and the reason it is not a
defect in this implementation. The register of accepted deviations is the subject
of its own issue and every entry in it is a public statement that this program
disagrees with a published result and why.

A divergent case does not block a release and does not pass silently. It appears
in the published report with its bands.

A not comparable case blocks nothing and is reported as what it is, so that a
corpus of thirty cases of which twenty decide nothing cannot read as thirty cases
that all agreed.

## Changing any of this

A change to the tolerance rule, to the outcome definitions or to what a failure
does is its own issue and its own pull request, argued on its own. The argument
goes in that pull request's body and the result comes back into this document. It
is never a line edited in passing inside a change about something else.

A change made while a case is failing, or immediately after one has failed, is
refused. Not because the change would necessarily be wrong, but because at that
moment nobody can tell the difference between a rule that was too tight and a
rule that was in the way, and the corpus is worth nothing if it can be adjusted
to agree.

## What would reopen this

A comparison exercise in which several independent implementations computed one
specified case, with its inputs published as numbers. That is the material the
survey looked for and did not find, and it would replace the reasoning above with
a measured spread between implementations, which is the quantity every number
here is standing in for.

Until then the tolerances are derived from what the sources printed rather than
measured from anything, and that is a weaker footing than a measured one. It is
stated here rather than left for a reader to work out.

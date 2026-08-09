# The mutation score, and every mutant nothing killed

Coverage says a line ran. Mutation testing alters the line and asks whether any
test noticed. For arithmetic that is the whole question, because a test that
computes a value and asserts it is a number covers every line it touches and
proves nothing about any of them.

`.github/workflows/mutation-testing.yml` runs it weekly and by hand. This file
is what the run leaves behind that a person has to write: the argument for each
mutant nothing killed.

## What it covers, and what it does not

It covers `acoustic` and `acoustic/approx`, measured separately. Those are the
arithmetic: bands, band sets, spectra, the decibel quantities, the energy
operations and the octave conversion. `kernel` is empty today and joins this
list when it is not.

It does not cover `store`, `cmd/schallweg`, `cmd/gate`, `cmd/sbom` or `harness`.
A mutant in a file reader is a question about a parser under bytes it did not
expect, and the answer to that is fuzzing rather than mutation, which is issue
#106. A mutant in the gate is a question about a check, and every check here
already owes a fixture that proves it bites.

It reports and does not gate. A mutation score is a measurement, and a
measurement turned into a bar teaches a suite to kill mutants rather than to
check behaviour. Several of the mutants below are equivalent by construction: no
test can kill them, and a bar would make somebody write one that pretends to.

## The score

Measured by the job itself, on the runner:

    acoustic
    Killed: 116, Lived: 12, Not covered: 2
    Timed out: 0, Not viable: 0, Skipped: 0
    Test efficacy: 90.62%
    Mutator coverage: 98.46%

    acoustic/approx
    Killed: 9, Lived: 0, Not covered: 1
    Timed out: 0, Not viable: 0, Skipped: 0
    Test efficacy: 100.00%
    Mutator coverage: 90.00%

    gh run view 31304524236 --json status,conclusion --jq '"\(.status) \(.conclusion)"'
    completed success

Those are the numbers to compare a later run against. The same commands on
Windows, where the triage below was written, report one mutant fewer killed and
one timed out, for the reason under the loop entry.

Before the tests this triage produced, the same command over `acoustic` reported
`Killed: 102, Lived: 20, Not covered: 6`, efficacy 83.61%, coverage 95.31%.

The timeout coefficient is not tuning. This suite runs in under a second, so the
timeout the tool derives from it is shorter than the compile a mutant needs, and
without the coefficient the run fills with timeouts that say nothing: the same
tree reported 61 of them at the default. A timed-out mutant is inconclusive. It
is never read as a pass.

## The mutants nothing killed

Each entry says where it is, what changes, and why nothing kills it. An entry
that could not make that argument would be a test that is owed rather than a
line here.

### Equivalent: a slice capacity nobody can observe

`acoustic/band.go:91` and `acoustic/band.go:106`, arithmetic and negation, four
mutants.

Both lines are `make(..., 0, hi-lo)`: a length of zero and a capacity hint, for a
slice then filled by `append`. Capacity is not observable through anything this
package exports, and `append` grows the slice whatever the hint was. No test can
distinguish the mutants from the original, and one that tried would be asserting
an allocation rather than a band set.

### Equivalent: the summation comparator, which is about repetition and not about a value

`acoustic/level.go:234`, `:235` and `:237`, negation and boundary, five mutants.

`EnergySum` sorts its terms before adding them, and these three lines are the
comparator. The sort exists so that one input gives the same last bit on every
run and on every machine, which is a property about repeating a calculation
rather than about its answer. Mutating the comparator changes the order and
therefore only the floating point rounding, and with the count of terms this
function is given, at most the twenty-one bands of a band set, no ordering
difference approaches any tolerance a test here may state.

`:235` is the strongest of the five: it is the tie-break between two terms of
equal energy, and adding two equal values in either order gives an identical
result.

The property those mutants would have to break is already asserted from the
other side, in `TestEnergySumOrderDoesNotChangeTheAnswer`, which requires the
answer not to move when the caller shuffles the input.

### Equivalent: unreachable with the two band sets that exist

`acoustic/octave.go:218`, boundary, one mutant. The guard is `p < lo || p >= hi`
and the mutant makes it `p > hi`. Reaching the difference needs `p == hi`. On the
extended set `hi` is 21 and the largest position this loop forms is 20. On the
core set the loop errors out on the first octave, at `p == 0` against `lo == 3`,
long before any position could reach `hi`. So no spectrum this package can
construct distinguishes the two.

`acoustic/octave.go:222`, arithmetic and negation, two mutants. The line reads
`s.values[p-lo]`, and the only band set that reaches it is the extended one,
whose `lo` is zero, for the reason just above. With `lo` zero, `p-lo`, `p+lo` and
the negation of either are the same index.

### Killed by running out of memory: a loop that stops terminating

`acoustic/band.go:92` and `acoustic/octave.go:89`, increment to decrement, two
mutants, both killed on the runner.

The lines are `for i := lo; i < hi; i++` and its octave counterpart. Counting
down never reaches the bound, so neither mutant produces a wrong answer: each
appends to a slice for as long as it is allowed to run. The tool's timeout is
derived from how long the unmutated suite takes, and on a runner the memory is
gone first. The first dispatched run of this job was cancelled by exactly that,
with everything it had measured lost, which is issue #167.

What stops them is the address space ceiling the measuring step sets before it
starts. The test binary dies out of memory, its test fails, and the mutant is
killed, which is the verdict it deserved. On a Windows machine, with no such
ceiling and a slower allocation, the same two are reported as timed out instead,
and a timed-out mutant is inconclusive rather than a pass.

### Not reachable by the tool: a constant declaration carries no coverage counter

`acoustic/band.go:25`, arithmetic, one mutant. `Core BandSet = iota + 1`. A
constant declaration is not an executable statement, so it has no coverage
counter, so the tool never runs the mutant at all. Made by hand, it does not
compile:

    acoustic\band.go:25:17: cannot use iota - 1 (untyped int constant -1) as BandSet value in constant declaration (overflows)

and `go test ./acoustic/` reports the package as a build failure rather than
running it.

`acoustic/octave.go:56`, arithmetic, one mutant. `const octaveCount =
len(nominalSeries) / 3`, the same class. Made by hand, multiplying instead of
dividing, `go test ./acoustic/` fails.

### Not reachable by the tool: a test cannot watch a testing.TB fail

`acoustic/approx/approx.go:69`, negation, one mutant. It is the error check
inside `Equal`, which reports through `tb.Error`. This package's own tests reach
the decision through `Check`, which returns an error instead, precisely because
a test cannot observe a real `testing.TB` failing without failing itself. So the
line is uncovered by the tests of the package it is in.

It is not uncovered by the module. Made by hand, `go test ./...` fails across
`acoustic`, in ten tests and more.

## What this triage added

Six survivors and four uncovered mutants were not equivalent. They were one
defect wearing two faces, and the tests that kill them are named here so that
deleting one has a visible consequence.

**A result read by walking its bands is not read at all when it has none.**
`Weighted`, `Corrected` and `EnergySumSpectra` each return `Spectrum{}` on an
error, and the zero spectrum is on no band set and names no bands. Negating any
of the six error checks in those functions made them return an empty spectrum
and a nil error, and the whole suite stayed green, because every test that reads
one of those results walks `got.Bands()` and a loop over nothing asserts nothing.
`TestEverySpectrumOperationReturnsAWholeSpectrum` requires the band set and the
band count of the result before anything walks it.

The same shape in the octave direction: `OctaveSpectrum.Bands` returns nil for a
spectrum with no values, so negating that check silently emptied every octave
walk. `TestAnOctaveSpectrumNamesEveryOctaveAndSaysWhatItIs` holds it.

**Two `String` methods were read by nothing.** `Spectrum.String` and
`OctaveSpectrum.String` each have a branch for the value nobody built, and that
branch is what a reader meets in an error message when something has already
gone wrong. Both branches are now asserted.

**A band one step outside its own set.** `Band.valid` ends `b.series < hi`, and
the boundary mutant survived because every band a caller can obtain comes from
its set and stops one short. From inside the package a band at exactly `hi` is
constructible, and it is the off-by-one this whole type exists to refuse.
`TestABandOneStepOutsideItsSetIsNotABand` builds one on each side and requires
both to be refused.

## What this file does not claim

A score is not a quality. Efficacy of 90.55% over `acoustic` says that of the
mutants the tool could run and judge, nine in ten were noticed; it says nothing
about the mutants this tool does not generate, and the operator list it does
generate is short and syntactic.

Nothing here reaches the defects this project has most reason to fear. A wrong
reference curve, a band index off by one against the standard, an equation from
the wrong edition: every one of those is a correct program computing the wrong
thing, and every mutant of it dies against tests that assert the same wrong
number. That is the validation corpus's work and not this file's.

No check reads this document, and nothing refuses a survivor that is added to
the run without being argued for here.

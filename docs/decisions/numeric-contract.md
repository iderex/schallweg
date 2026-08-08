# Decision: units, decibel arithmetic and where rounding happens

Status: decided, with one constant deliberately left open and marked at the place
it is left.

Two implementations of this method disagree in the third decimal place for
uninteresting reasons and in the first for interesting ones. Validation cannot
tell those apart unless the uninteresting reasons have been removed in advance,
which is what this document does.

## The numeric type

`float64` everywhere a physical quantity is held, with no second type and no
per-quantity precision.

Sound reduction indices, level differences and impact levels occupy a range of
roughly a hundred and forty decibels, which is fourteen orders of magnitude in
energy. `float32` has about seven decimal digits and would put the rounding error
of an energy sum into the range where it competes with the differences this
method is used to detect. Sixty-four bit binary floating point has about fifteen
digits and leaves the arithmetic error four or five orders of magnitude below
anything anybody measures.

A decimal or rational type was considered and refused. The arithmetic here is
logarithmic: every energy sum goes through a power and a logarithm, and neither is
exact in any representation. A decimal type would give exact addition of numbers
that are not exactly representable in the first place, at the cost of a
dependency, slower arithmetic and the appearance of a precision the physics does
not have.

## Where decibels live, where energy lives, and what refuses to compile

A decibel value is a struct. That sentence is the mechanism, and everything else
in this section is its consequence.

In Go a named type over `float64` still supports `+`, so `Level(55) + Level(60)`
would compile and produce a hundred and fifteen decibels, which is the single most
common error in this domain and is invisible in review because it looks exactly
like arithmetic. A struct with an unexported field does not support `+` at all.
The mistake is not caught by a check, it is not caught by a review, and it is not
caught by a test: it does not build.

Three types, and what each one may do:

- `Level` is a quantity in decibels: a sound reduction index, a level difference
  between two rooms, a normalised impact sound pressure level.
- `Delta` is a difference in decibels: an improvement from a lining, a
  laboratory-to-situ correction, a vibration reduction index. Two levels
  subtracted give one of these.
- Energy is not a type a caller ever holds. It exists inside the two functions
  that convert, and every conversion into it is paired with the conversion back
  in the same function body, so no energy quantity is ever stored, passed or
  serialised.

What is available:

- `Level` minus `Level` gives `Delta`. Two quantities compared is a ratio in
  energy and a difference in decibels, so this is defined and it is the useful
  case.
- `Level` plus `Delta` gives `Level`, and minus likewise. Applying a correction
  to a quantity is what most of this method does.
- `Delta` plus `Delta` gives `Delta`. Corrections accumulate.
- `EnergySum` over a collection of `Level` gives `Level`, by the name that says
  what it is. Combining transmission paths is an energy sum and never anything
  else.

What is not available, by construction rather than by convention: `Level` plus
`Level`. There is no method for it, there is no operator for it, and a
contributor who wants it has to add a function and name it, which is a line in a
diff that says what it is.

Energy quantities exist only between the conversion and its inverse. The reason
they are never stored is that a stored energy quantity is a decibel quantity
somebody will eventually format, and the moment two representations of the same
thing can both be printed, one of them gets printed by accident.

## Units

Every quantity is in its SI unit and there is no alternate. Areas in square
metres, volumes in cubic metres, lengths in metres, frequencies in hertz,
reverberation times in seconds, mass per unit area in kilograms per square metre.
Nothing is stored in a non-SI unit and converted on the way out, because a value
whose unit depends on where it came from is a value nobody can compare.

Units are carried in the type where the type already exists for another reason,
which is the decibel quantities above, and in the identifier otherwise. A plain
`float64` parameter that is an area is named so that its unit is in its name. A
full dimensional type system for the geometric quantities was refused: it would
be a large amount of code guarding the arithmetic that is not where the mistakes
happen, and the mistakes happen in decibels.

The dimensionless quantities are named as such: the share of the total that a
path accounts for, a loss factor, a ratio of masses. A dimensionless quantity
never gets a unit in output, and where output would be more readable as a
percentage the conversion happens in the formatting and not in the value.

## Summation order

Every sum in the kernel is performed in one defined order: ascending by the
magnitude of the summand, ties broken by the summand's identity, which is the
band index or the path identifier and is therefore total.

Floating point addition is not associative, so a sum over the same set in two
orders can differ in the last places. That difference is far below anything
physically meaningful and it is still worth removing, because a validation corpus
that reproduces exactly is a corpus where any movement at all is a signal, and a
corpus that reproduces to within a wobble is one where somebody has to decide how
big a wobble is allowed every time it moves.

Ascending by magnitude rather than by band or path order, because that ordering
also happens to be the one that loses the least: adding the small contributions
first keeps them from disappearing under a large running total. Path
contributions in this method routinely span several orders of magnitude, so the
ordering is not academic.

No compensated summation. The number of terms in any sum here is small, at most a
few tens of paths or twenty-one bands, and the ordering above already covers the
case compensation is for. What would change that is a summation over hundreds of
terms, which this method does not have.

## Whether results are bit-identical across platforms

Within one architecture and one toolchain version, yes, and it is required.

Across architectures, not claimed, and the reason is specific rather than
general. Two things in the way. The language permits an implementation to fuse a
multiplication and an addition into a single operation with one rounding where
the intermediate is not explicitly rounded, and whether it does so depends on the
target. And the standard library's logarithm and power functions are not
guaranteed to be correctly rounded, so an implementation for one architecture may
differ in the last place from another.

What is done about the first: the kernel writes an explicit conversion around any
product that feeds an addition, which is what the language specification says
makes the intermediate rounding explicit and therefore unfusable. That is a
convention in the source rather than a check, and stating it here is not the same
as enforcing it.

What is done about the second: nothing, because nothing can be, short of this
project shipping its own logarithm. Instead the cross-architecture claim is
demoted from bit-identity to agreement within a stated tolerance, and it is a
measured property rather than an assumption. Measuring it is the numerical
regression gate, issue #111, which runs the validation corpus on more than one
architecture and reports the largest deviation it found.

NOT MEASURED. No result of this kernel has been compared across architectures,
because the kernel does not exist yet. The paragraph above states what will be
required and what will be measured, and it must not be read as a report of a
measurement that happened.

## Rounding: two kinds, kept apart

Rounding to presentation precision happens exactly once, at the edge, in the
layer that formats output. The kernel returns full precision and stores full
precision, and no value that has been rounded for presentation ever re-enters a
calculation.

The reason for the edge and not anywhere else is that rounding is destruction,
and destruction inside a pipeline is invisible to everything downstream. A value
rounded in the middle carries an error that the rest of the calculation then
amplifies or hides, and two implementations that round in different middles are
two decibels apart with no way to say which step did it.

There is a second kind of rounding and it is not the same thing, so it has a
different name and a different rule. A method rounding is a defined step of a
procedure: the procedure says that at this point the values are taken to a stated
precision, and the answer the procedure gives is different if they are not. That
rounding happens inside the kernel, because it is part of the arithmetic rather
than part of the display, and it is subject to three requirements. It happens at
exactly one place in the code. That place names the clause it implements. And the
value it produces is used only by the procedure that asked for it and is never
substituted for the full precision value elsewhere.

Presentation rounding is done by the formatter and by nothing else. Method
rounding is done by the procedure that defines it and by nothing else. A function
that does both is the defect this separation exists to make nameable.

## The rule that feeds the reference curve procedure

This gets its own section because it is where the rounding rule earns its keep.

The weighted single-number ratings are found by shifting a reference curve against
the measured spectrum in whole decibel steps until a stated condition on the sum
of the shortfalls is met. It is an integer search over rounded values, so the
precision of the values entering it is not a display question: it changes which
step the search stops on, and therefore the rating, by a whole decibel.

The rule, stated as three requirements rather than as one number:

- The spectrum is taken to the procedure's stated precision once, before the
  search begins, at a single named function.
- Nothing inside the search loop rounds anything. The loop compares values that
  were already at the procedure's precision when it started.
- The shortfall sum the stopping condition tests is computed from those values,
  not from the full precision spectrum, and not from a mixture.

The failure each requirement prevents is a distinct defect and all three produce
a rating that is wrong by one decibel while looking entirely reasonable. Rounding
inside the loop makes the search's own comparisons inconsistent between
iterations. Feeding full precision values to a search that was defined over
rounded ones makes the stopping condition trip a step early or a step late.
Mixing the two makes the answer depend on which branch of the comparison ran.

CLAIM, NOT VERIFIED. The precision the procedure asks for is one decimal place on
the band values, and the resulting rating is an integer. That is stated here from
common practice in how these quantities are reported rather than from the
standard's own clause, because this project holds no copy of the clause. Issue
#42, which implements the procedure, owes the check against the clause text
before it ships, and if the clause says otherwise then the number changes and the
three requirements above do not.

## Worked example, with every rounding step visible

Two transmission paths contribute 55.0 dB and 60.1 dB. What the operator is shown
is one number. What happens between those is below.

The numbers were produced by the program at the end of this section. They are an
illustration of the rounding rule and are not a result of this kernel, which does
not exist yet.

The correct route, rounding once at the edge:

    input                55.0 dB and 60.1 dB, as supplied
    to energy            316227.76601683797 and 1023292.9922807537
    summed               1339520.7582975917
    back to a level      61.269494481750932 dB
    held in the result   61.269494481750932 dB, full precision
    presented            61.3 dB

The edge is the last line and nothing above it was rounded. The value the result
carries is the full precision one, so a report at a different precision, a
machine-readable output, and a comparison against a requirement all round the
same value once rather than rounding a rounded one.

The same calculation with one rounding moved earlier, which is the mistake:

    input                55.0 dB and 60.1 dB
    rounded first        55 dB and 60 dB
    to energy, summed, back to a level
                         61.193310480660941 dB
    presented            61.2 dB

One tenth of a decibel apart, both plausible, and nothing in either output says
which route produced it. On the reference curve procedure the same displacement
lands on the wrong integer step and the answer moves by a whole decibel.

The summation order rule is visible here too: the sum of the two energies is the
same number in either order, which the program checks and reports as true. Two
terms is the case where it is obviously true; the rule exists for the case of
twelve flanking paths spanning several orders of magnitude, where it is not
obvious and is not checked by anything except the ordering being fixed.

To reproduce, put this in an empty directory beside a `go.mod` declaring any
module name, and run `go run .`:

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    a, b := 55.0, 60.1
    ea, eb := math.Pow(10, a/10), math.Pow(10, b/10)
    lvl := 10 * math.Log10(ea+eb)
    fmt.Printf("ea=%.17g eb=%.17g sum=%.17g\n", ea, eb, ea+eb)
    fmt.Printf("level=%.17g presented=%.1f\n", lvl, lvl)
    fmt.Printf("order independent: %v\n", lvl == 10*math.Log10(eb+ea))

    ra, rb := math.Round(a), math.Round(b)
    w := 10 * math.Log10(math.Pow(10, ra/10)+math.Pow(10, rb/10))
    fmt.Printf("rounded first: level=%.17g presented=%.1f\n", w, w)
}
```

## What would reopen this decision

A validation deviation that survives after the rules above are in force, which
would mean the disagreement is in the method rather than in the arithmetic and
belongs on the validation issues. Evidence that the standard library's
transcendental functions differ across the architectures this project targets by
more than the tolerance the regression gate sets, which would force the kernel to
carry its own. Or the clause check owed by issue #42 giving a different precision
than the claim above, which changes the constant and nothing else.

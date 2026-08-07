# Decision: how uncertainty is carried

Status: decided for the first release. The condition that reopens it is named at
the end and is expected to be met.

## The decision

The first release carries no numeric uncertainty on a result.

The reason is that this project cannot yet state one it could defend. A model
uncertainty taken from the literature would be a number this project did not
measure, applied to an implementation whose agreement with the published cases
has not been established. A propagation of input uncertainties would require an
uncertainty on every laboratory value, and a test report does not always carry
one, so most of the inputs would arrive with an invented figure that then
propagates into the result and comes out looking measured. Both routes produce a
number that a user would reasonably read as this tool's own estimate of its own
error, and neither would be that. A wrong uncertainty is worse than none,
because it is believed.

That is a statement about what is available today, not a position that
uncertainty does not matter.

## What the output does instead

Every result carries, as mandatory structure rather than as a note, the material
a reader needs to judge how far to trust it. The fields are specified in
[result-contents.md](result-contents.md); what they are for is here.

The path breakdown. Every transmission path's contribution appears beside the
total, so a reader sees whether the answer rests on one dominant path or on many
comparable ones. A result dominated by a single well-measured element is a
different kind of answer from one where five paths contribute within two decibels
of each other, and no single figure expresses that difference.

The origin of every input. Each value says whether it came from a test
certificate, from somebody's estimate, from a derivation the kernel performed, or
from a default the kernel supplied. The count of each is summarised in
`input_basis`.

The completeness flag. A result computed with a defaulted input is marked
`incomplete` and names what was missing and what was used instead.

The assumptions list. Every choice made on the user's behalf is named in the
result.

None of that is an uncertainty and none of it is presented as one. `input_basis`
in particular is a count of where numbers came from. It is not a confidence, it
does not order two results, and any presentation that shows it as a bar, a score
or a percentage of trust would turn a negative disclosure into a positive
assurance and is refused.

## Is a result without an uncertainty representable

Yes, and today it is the only representable kind, since no uncertainty field
exists. Nothing in the structure or in the kernel treats an uncertainty as
optional, because there is nothing to make optional.

The converse is what matters for the future. When an uncertainty field is added,
it is added as a field that is either present with a stated basis or absent with
a stated reason, never present with a placeholder. A result that carries a zero
or an unexplained default in that field would be the exact failure this decision
avoids.

## Presentation rule

A result is never presented as a bare number.

Wherever a value is shown, whether in the terminal, in a machine-readable output
or in a handover report, it is accompanied by the model that produced it, the
completeness flag, and the dominant path. The machine-readable output carries the
whole structure. The report carries the whole structure as well, in a form a
reader can follow, with the input list and the assumptions in full, because the
report is the artefact a third party reads without access to the tool.

Rounding follows the standard's own rule for the quantity being reported, and
the result never shows more precision than the method supports. Where a value is
rounded for presentation, the unrounded value stays in the machine-readable
output, so a consumer is never forced to work from a rounded figure.

No output presents a computed value as a compliance verdict against a national
rule. That is a separate question, held open as an open maintainer decision on
issue #1, entry 4, and nothing here anticipates its answer.

## What would change this decision

The validation work in this plan, where this implementation is run against the
published comparison exercises and every deviation is recorded with its case.
That produces a measured distribution of deviations between this implementation
and published results, obtained by this project, on cases anybody can look up.

Once that exists, a model uncertainty can be stated with the measurement behind
it and the corpus it was measured on, which is the only form this project would
ship. Adding the field then is an additive change under the versioning promise in
[result-contents.md](result-contents.md), and it does not disturb any consumer
already reading results.

Two things would still be needed at that point and neither is assumed here.
The corpus has to be large enough and varied enough that a distribution measured
on it says something about a case outside it, and the document adding the field
has to say where it stops. Neither is settled by this decision.

# Decision: one structure, two evaluation strategies

Status: decided for the first release. What would reopen it is at the end.

## The decision

The detailed model on frequency bands and the simplified model on single numbers
share one description of the transmission situation and one path enumeration, and
have separate arithmetic. Neither is expressed in terms of the other.

The reason is that the two models differ where the numbers are made and agree
where the building is described. Two rooms, the elements between and around them,
and the junctions where those elements meet are the same physical facts whichever
model is then run over them, and the set of transmission paths between a source
room and a receiving room is the same set. The arithmetic along a path is not the
same, and the simplified model is not the detailed model evaluated on one wide
band: it is a separate empirical fit that happens to answer a similar question.
Writing it as a degenerate case of the other would encode a physical claim that
is false, and it would be encoded in the one place nobody looks, which is the
band loop.

## What the other two shapes would have cost

One implementation, with the simplified model as the detailed model on a
degenerate single band. It is the smallest amount of code and it is the shape
that reads best in a diagram. Its cost is that every place where the empirical
fit departs from the band method becomes a special case inside a loop that is
supposed to be uniform, and the fit's own correction terms have nowhere honest to
live. The failure mode is silent: the number comes out, it is wrong by a few
decibels, and the structure of the code argues that it cannot be.

Two independent implementations. Honest, and each can be validated against the
published cases for its own model without any argument about which shared piece
caused a deviation. Its cost is that the situation model, the path enumeration
and the geometry validation get written twice, and the second copy is the one
that drifts. Path enumeration is where most of the complexity of this standard
actually is, so duplicating it duplicates the expensive part to avoid sharing the
cheap part.

## What is shared and what is not

Shared, one implementation used by both:

- The situation model: rooms, their separating element, the flanking elements,
  and the junctions between them.
- Path enumeration: deriving the direct path and the flanking path set for a
  situation.
- Geometry validation: refusing a situation whose surfaces or junctions do not
  close.
- Element and junction identity, and the provenance carried with every input
  value.
- Result assembly: collecting per-path contributions into a result and reporting
  each contribution beside the total.
- The rating step that reduces a spectrum to a single number, which the detailed
  model uses at the end and the simplified model does not use at all.

Not shared, one implementation per model:

- The arithmetic along a single path, direct or flanking.
- The junction contribution, including how the vibration reduction index enters.
- The in situ correction between laboratory and building conditions.
- The quantity a path contributes and the rule for summing contributions.
- The input requirements for an element, which are a spectrum for one model and
  a single number with its adaptation terms for the other.

The seam between the two lists is deliberately at the path. Everything that
answers "which paths are there and what are they made of" is shared. Everything
that answers "how much comes down this path" is not.

## A result says which model produced it

A result is not interpretable without knowing which of the two produced it, since
the two answer with different quantities of different precision and a user
comparing them will otherwise compare them directly.

The field that carries this is part of the result structure specified in
[result-contents.md](result-contents.md), which is where every field of a result
is defined and versioned. This document does not restate the field list; it
states the requirement that one of those fields is the model identity, and that
it is not optional and not defaulted.

## Asking for the detailed model with simplified data

The kernel refuses, and it says which element and which quantity.

It does not synthesise a spectrum from a single number. A weighted rating and its
adaptation terms do not determine a third-octave curve, and any reconstruction is
an invented input that would then be carried through the whole calculation and
reported as if it were measured. That is exactly the failure the provenance
requirement in [result-contents.md](result-contents.md) exists to make visible,
and producing it inside the kernel would defeat it at the source.

The refusal names the input that caused it, so the user can either supply the
spectrum or ask for the simplified model. Where an element has both a spectrum
and a single number, the detailed model uses the spectrum and the simplified
model uses the single number, and neither derives one from the other.

The reverse direction is different and is allowed: the simplified model may run
on data that also carries spectra, because reducing a spectrum to the rating the
simplified model wants is the defined rating procedure rather than an invention.
Where that reduction happens, the result records the quantity as derived rather
than supplied.

## What would reopen this decision

Validation against the published comparison cases showing that the shared path
enumeration cannot express the path set one of the two models needs. If that
happens, the seam is in the wrong place, and the answer is to move the seam
rather than to duplicate the tree.

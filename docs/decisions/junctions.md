# Decision: how a junction is described and where its values come from

Status: decided for the first release, with one enumeration marked as a claim and
owed to issue #53.

Flanking transmission is what makes this standard worth implementing, and the
junction is where flanking is decided. It is also the input a real project is
least likely to have measured, so the decision here is mostly about what the
kernel does when the good input is absent, and about making sure a number built
on an estimate never looks like a number built on a measurement.

## How a junction is described

A junction is the elements that meet, the way they meet, and the edge they share.

- `id`, required. Unique within the project.
- `topology`, required. The arrangement of the elements meeting at the edge,
  from an enumerated set: cross, T, corner. Not free text, because it decides
  which routes may run.
- `type`, required. What the connection itself is, from an enumerated set: rigid,
  elastic interlayer, and the connection kinds a framed construction uses. This
  is separate from topology because the same three elements in the same
  arrangement behave differently depending on whether they are cast together or
  separated by a resilient strip, and a model with one field would have to encode
  that in the name of the other.
- `elements`, required. The elements meeting at this junction, each with the role
  it plays: which are the separating elements and which are flanking on each
  side. An element appears here by its identity in the project, so the junction
  refers to the built thing rather than to a construction type.
- `coupling_length`, m, required. The length of the common edge. It is a property
  of the junction rather than of any element, because it is what the elements
  share.
- `kij`, optional. Measured vibration reduction index values, keyed by ordered
  element pair, because the quantity is directional and the value from element i
  to element j is not the value from j to i. Each entry is a spectrum or a single
  value on the band set the calculation runs on.
- `provenance`, required on every `kij` entry, by
  [certificate-extraction.md](certificate-extraction.md).

The value is keyed by ordered pair rather than held once per junction because
that is what the quantity is. A junction of four elements has twelve directed
pairs, a project usually has a value for none of them, and a model that held one
number per junction would have to average what it does not have.

## Where a value comes from, and which wins

Two sources. A measurement, which is rare and good. An empirical route computed
by the kernel from the mass ratio and the junction geometry, which is common and
is an estimate.

The empirical route is computed by the kernel rather than supplied as data. The
alternative was considered: a data-supplied route would let the expressions be
corrected without a release and would keep the arithmetic out of the tree. It was
refused because an expression supplied as data is an expression nobody validates.
The validation corpus in this project compares computed results against published
cases, and it can only do that for arithmetic the kernel actually performs. A
route that arrives as data is outside every test in this repository, and its
first wrong answer would be a user's.

Precedence, when both are available for a direction: the measurement wins, and
the estimate is not computed for that direction at all. A measured value is a
fact about a junction somebody built and tested, and an estimate that agrees with
it adds nothing while an estimate that disagrees with it would have to be
resolved by whoever wrote the code, in advance, for all cases.

Where a junction has measured values for some directions and not others, the
remaining directions are estimated and the result records that they were, naming
both sources. This is deliberately different from the rule for a spectrum with
missing bands, which is refused outright by
[frequency-bands.md](frequency-bands.md), and the difference is not an
inconsistency. A band is one part of one measurement of one quantity, and filling
it from elsewhere invents part of a measurement. A direction pair is a separate
quantity that is measured and published separately, so using a measured value for
one and an estimate for another is combining two sources rather than fabricating
one.

## Where the source is recorded

In the result, on the input entry, in the structure already fixed by
[result-contents.md](result-contents.md). This section names the places rather
than inventing new ones.

Each junction value the calculation used appears once in `inputs`, with
`quantity` naming the vibration reduction index, `junction` naming the junction,
and `origin` saying where it came from. A measured value carries `origin` of
`certificate` and a `source` pointing at the record and through it at the report.
A value the kernel computed from the empirical route carries `origin` of
`derived` and `derived_from` listing the input entries it was computed from,
which are the masses of the elements meeting there.

The route itself is named in `assumptions`, once per junction it was used on,
with the junction and the affected input identifiers. That is the correct place
for it under the definition already in force: an assumption is a choice the
kernel made that the user did not make, and using an empirical expression where a
measurement was wanted is exactly such a choice.

The consequence is the one this project cares about. `input_basis` counts the
inputs by origin, so a report can say in one line how many junction values were
measured and how many were estimated, and a reader deciding whether to believe
the number sees the difference without reading the whole list. A prediction built
on estimated junction values and one built on measured values are different
claims, and they are visibly different documents.

## What the empirical route covers

The kernel holds the empirical route as a named thing with a declared coverage,
and the coverage is checked before the route runs rather than after. A junction
the route does not cover does not get a number from it.

CLAIM, NOT VERIFIED. The route in the first release covers rigid junctions
between homogeneous single-leaf elements in cross, T and corner topologies. It
does not cover junctions with an elastic interlayer, junctions in framed and
platform-framed timber construction, junctions where one or more of the meeting
elements is a double-leaf construction, junctions of more elements than the
enumerated topologies describe, and junctions whose behaviour is changed by a
lining or a floating layer running through them.

That paragraph is stated from general knowledge of what this kind of expression
is fitted to, not from the clause, because this project holds no copy of the
standard's text and [standard-text.md](standard-text.md) forbids reconstructing
the method from a secondary source. Issue #53, which implements the route, owes
the check against the clause before the route ships, and the coverage predicate
in the code is what the check corrects. Nothing else in this document depends on
which way that check goes: the mechanism is a declared predicate, and the
predicate is data about the route.

## What the kernel does at the boundary

It refuses the path, names the junction, names the topology and type that were
not covered, and does not produce a number for that path.

It does not fall back to a neighbouring junction type, and it does not compute
the route anyway and mark the result incomplete. A fallback would be the same
failure as an invented measurement with an extra step, and the incomplete flag is
not strong enough to carry it: an incomplete result is still a result, and the
number in it gets quoted.

What refusing costs is real. Lightweight and timber construction is a large and
growing part of what gets built, most of it is outside the coverage above, and a
tool that refuses is a tool that answers nothing for those projects. The
alternative costs more. An estimate outside the range an expression was fitted to
is not a worse estimate, it is a number with no basis, and it would be reported
in the same font as everything else.

The route forward for a user at the boundary is a measured value. The refusal
says so, names the directions it wanted, and says that supplying them makes the
calculation run. That is a real answer rather than a dead end, and it is also the
honest description of what the state of knowledge is for those junctions.

Refusing a junction the model cannot carry, and saying which, is issue #54.

## What the junction model cannot describe

Listed so that the limit is on the record and chosen rather than discovered.

- A connection that is not along a single common edge, including point
  connections and any junction where two elements meet over an area.
- A junction where more elements meet than the topologies enumerate, which
  happens at a column, at a shaft wall and at most stair connections.
- A junction whose properties vary along its length, such as an edge that is
  rigid for part of its run and resiliently supported for the rest. The model
  holds one type per junction, so such an edge has to be split into two junctions
  or approximated by one, and the model does not say which.
- A junction whose behaviour depends on a detail below the level of the element
  model: the fixing pattern, the presence of a continuous seal, the compression
  of a resilient strip under load.
- A junction between elements that the element model itself cannot describe,
  which is most of what [element-model.md](element-model.md) lists as its own
  limits and does not need repeating here.

The consequence of that list, stated once and plainly: this model describes rigid
heavyweight construction well and much of modern lightweight construction not at
all. That is a choice about scope and not a temporary state, and any later work
that widens it is widening the model rather than adding a case to it.

## What would reopen this decision

The clause check owed by issue #53 returning a different coverage, which changes
the predicate and nothing else here. A published empirical route for a
construction class this model refuses, arriving with a validation basis, which
would be a second named route beside the first rather than a widening of it. Or
validation evidence that the directional keying is finer than any available data
supports, which would be an argument about the data and not about the model.

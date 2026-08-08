# Harnesses

The ordinary test suite of this repository runs with no display, no elevation, no
network and no hardware beyond a general purpose computer. That rule is
[../docs/decisions/testability.md](../docs/decisions/testability.md) and it holds
from the first test onwards, because a suite acquires this kind of dependency one
convenient test at a time and the loss is only noticed when somebody else tries
to contribute.

A test that cannot meet all four conditions is not written into the ordinary
suite. It goes here, into a harness named for the thing it needs, behind a build
constraint with the same name, so the ordinary run does not build those files at
all. Moving a test into the ordinary suite then means deleting a build
constraint, which is a visible line in a diff rather than an invisible change of
behaviour.

The harness names are fixed by that decision and the convention is that the name
states the requirement rather than the intent:

`requires-display`, `requires-instrument`, `requires-third-party-tool`,
`requires-network`, `requires-elevation`.

A contributor reading one of those names can tell, without opening anything,
whether their machine can run it. A harness called integration, end-to-end or
slow tells them nothing, so they try it, it fails, and the failure looks like a
defect in the code rather than a missing prerequisite.

Each harness carries, beside it, a statement of exactly what has to be present
for it to run and how a contributor would know they have it. A harness that was
not run is never counted as a pass, and counts are reported per harness and never
summed across them, because one total is the shape that lets an unrun harness
disappear into a green figure.

## What is here today

Nothing. No harness exists, so nothing is being skipped, and the test workflow
says exactly that at the end of every run rather than leaving a reader to assume
either way.

The first one this project knows it will need is `requires-third-party-tool`, for
comparing against a second implementation of the standard, which is issue #87.
The programs that implement it are commercial, so that harness needs a licence
and in at least one case a hardware dongle, which is why its name says what it
needs rather than what it does.

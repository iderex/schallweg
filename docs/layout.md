# The tree

This document says what the top-level directories are and what each one may
depend on. It is the layering decision in
[decisions/layering.md](decisions/layering.md) made visible in the tree, so that
a violation of it appears in a diff instead of being discovered when something
that should not have a file system dependency turns out to have one.

Where this document and that decision differ, the decision is right and this
document is what has to be corrected.

## The directories

Each entry names the layer from the layering decision that it holds, and what it
may import from inside this module.

`acoustic/` is the numeric floor. Bands, band sets, spectra, decibel and energy
quantities, and the arithmetic over them. It imports nothing from this module.
The types in [decisions/frequency-bands.md](decisions/frequency-bands.md) and
[decisions/numeric-contract.md](decisions/numeric-contract.md) live here.

`kernel/` is the calculation. Constructions, elements, junctions, transmission
situations, path enumeration and the path arithmetic. It imports `acoustic/` and
nothing else from this module.

`store/` is the data access layer. Reading, validating and serialising records,
project files and reference datasets. It imports `kernel/` and `acoustic/`. It is
the only place in the module that touches the file system.

`cmd/` is composition. One directory per program. It may import everything above.
It is the only place that writes to a terminal.

`harness/` is the tests that need something the ordinary suite is not allowed to
need, each behind a build constraint named for what it needs. It may import
anything. Nothing imports it.

`data/` is the component database: records and nothing executable. No Go package
imports it, and nothing in the module refers to a particular record. `store/`
knows how to read a record; it does not know that any record exists.

`docs/` is documentation. `testdata/` is fixtures. `.github/` is the workflows and
the templates the hosting service reads.

## The rule

A directory may import the directories below it in the list above and none of the
ones above it. There is no sideways dependency, because the list is a chain and a
chain has one answer to "may this import that" rather than a discussion.

Two consequences carry more weight than the rest and they are the reason the
chain is drawn this way:

`acoustic/` and `kernel/` may not reach the file system or the network. Not
through a convenience function, not to load a reference dataset, not once.
Everything they compute on arrives as an argument. That is what makes a
validation case reproduce on any machine, and it is what makes a case that does
not reproduce a defect in the arithmetic rather than in the surroundings.

Nothing below `cmd/` prints. A thing that went wrong is a returned error and a
thing worth saying is a field in the result.

## What enforces the rule

`cmd/gate/architecture.go`, read by the tests beside it, which the `test` leg of
`go run ./cmd/gate` runs. What it refuses is printed by that suite rather than
listed here.

Until issue #109 landed this section said that nothing enforced any of it. What
it refuses now is the import graph the toolchain reports and the syntax of every
non-test source file: a package in `acoustic/` or `kernel/` that reaches the file
system, the network, another program or the process's surroundings; a package
importing a layer above its own; a package in `store/` importing `math`; a print
below `cmd/`; a logarithm or an exponential outside `acoustic/`; and a package of
this module in no layer at all. Each rule has a test that breaks it deliberately
and a near miss beside it that must go on passing.

**What it still does not reach, which is more than the paragraph above may
suggest.** The graph rule reads what a package asks for by name and not what the
linker pulls in, because every package that returns a formatted error has `os` in
its full dependency list, so a rule written over that list would fire on all of
them and say nothing. What follows is that a package can perform input or output
through something it does not name: through a third-party package, which the tree
has none of today and the dependency policy is what keeps that true, or through
the standard library reached by a route this list does not hold. Whether the data
access layer *calculates* is not decided by anything; what is decided is whether
it imports `math`, which is the part of that sentence a machine can read. And no
rule here sees a test file, so a test in any layer may read a fixture and write
its own arithmetic out by hand, which is how a test in this project is supposed
to work.

## Where the database lives, and when it is released

The records live in this repository, under `data/`, and they are released
separately from the program.

In this repository, because every check that would catch a bad record lives here.
The schema is here, the validation of records against it is here, the fixtures
are here and the gate that runs them is here. A database in a second repository
would need either its own copy of that apparatus, which then drifts, or a route
that validates data it does not hold, which is the same apparatus with an extra
failure mode. One contribution route also means a correction to a record and a
correction to the code that reads it can be reviewed against each other.

Released separately, because the two things change for different reasons and at
different speeds. A wrong laboratory value is wrong today, and waiting for a
software release to correct it is the failure that
[decisions/data-format.md](decisions/data-format.md) already names when it says
the tree is the authority and the shipped artefact is derived from it. So the
database artefact carries its own version and is cut when the data changes, and
the program carries its own version and is cut when the code changes. Neither
version number implies anything about the other, and a result records both, which
[decisions/result-contents.md](decisions/result-contents.md) already requires.

The cost is real and is accepted rather than hidden. One record per file means a
few thousand files eventually, all in the history of the repository somebody
clones to get the kernel, and the history will be dominated by data corrections
rather than by code. What would reopen this is that cost becoming measurable
rather than anticipated: if a clone of this repository is dominated by `data/`
to the point where working on the kernel is slower for it, the records move and
the apparatus moves with them. That is a measurement to make later, not a
prediction to act on now.

## What the tree holds today

Every directory above exists at the commit this document lands on, and the
directories that will hold code hold a package statement and a documentation
comment stating the rule for that layer. They hold no implementation, because the
work that fills them is the milestones after this one, and each package comment
names the issues that do.

The reason the empty packages exist rather than being created with their first
file is that this document says what may import what, and a rule about imports is
easier to follow when the thing it constrains is already there with the rule
written at the top of it. The alternative is that the first file in each layer
arrives with the rule somewhere else.

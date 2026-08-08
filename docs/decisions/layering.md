# Decision: the layers, and what each one may depend on

Status: decided for the first release. What would reopen it is at the end.

## The layers

Five, named from the bottom. A layer may depend on the layers below it and on
nothing above it. There is no sideways dependency between two layers at the same
level, because there are none at the same level: the list is a chain on purpose,
and a chain is the only shape in which "what may this import" has one answer
rather than a discussion.

**The numeric floor.** Spectra, bands, decibel and energy quantities, and the
arithmetic over them. It depends on the language's standard library and on
nothing else in this repository. It promises that a quantity cannot be built in
an invalid state and that an operation which is not defined refuses rather than
returning a plausible number.

**The kernel.** Elements, junctions, transmission situations, path enumeration
and the path arithmetic. It depends on the numeric floor. It promises a result
that carries its own breakdown and its own inputs, in the structure fixed by
[result-contents.md](result-contents.md), and it promises that everything it did
on the user's behalf is named in that result rather than left in the code.

**The data access layer.** Reading, validating and serialising records, project
files and reference datasets. It depends on the kernel for the types a record
becomes, and it is the only layer that touches the file system. It promises that
nothing reaches the kernel that has not been validated, and it promises that a
byte sequence it cannot make sense of becomes a refusal naming the file, the
position and what was expected, rather than a zero value.

**The command line.** Composition. It reads arguments, asks the data access
layer for inputs, asks the kernel for a result, and writes output. It depends on
everything below. It promises that it is the only layer that writes to a
terminal, and that every quantity it prints is a quantity the kernel computed
rather than one it computed itself.

**The harnesses.** The tests that need something the ordinary suite is not
allowed to need, kept behind build constraints as
[testability.md](testability.md) requires. They may depend on anything. Nothing
depends on them, and that is what keeps the rest of the tree honest about what it
needs.

The component database is not in the list, and that is deliberate. It is records
and a schema, it contains no code, and nothing in the chain above depends on it
at build time. The data access layer knows how to read a record; it does not know
that any particular record exists.

## The two rules that carry weight

The kernel may not reach the file system or the network. Not through a
convenience function, not to load a reference dataset, not to read a
configuration file, not once. Everything it computes on arrives as an argument.
The reason is not purity. It is that a kernel with no I/O is a kernel whose
behaviour is fully determined by its inputs, which is what makes a validation
corpus mean anything: a case that reproduces is a case that reproduces on any
machine, and a case that does not reproduce is a defect in the arithmetic rather
than in the environment. It is also what makes the testability rule cheap to
keep, because the layer with the most tests is structurally incapable of
acquiring a dependency on the machine.

Only the command line prints. A layer that prints has an opinion about who is
reading, and a library with an opinion about who is reading is a library nobody
can embed. Below the command line, a thing that went wrong is a returned error
and a thing worth saying is a field in the result.

## What enforces this today

The tests beside `cmd/gate/architecture.go`, which the ordinary suite runs. Both
of the two rules above are among what they refuse, and [../layout.md](../layout.md)
is where the whole set and its limits are written, so that this document and that
one do not each carry a list that drifts.

This section said that nothing enforced any of it until issue #109 landed, and
the sentence it replaces was right at the time. The residual it named was that a
change could put a file read inside the kernel and every run would stay green,
and that particular residual is gone: a package in the kernel or the numeric
floor that reaches the file system is refused, whether it reaches it directly or
through another package of this module.

What survives is narrower and is stated rather than left implied. The rule reads
what a package asks for by name, so input or output reached through a package
that does not name it is not seen, and the thing most likely to do that is the
convenience this section already named: loading a reference dataset from inside
the function that needs it, which [standard-text.md](standard-text.md) pushes
towards the data access layer. What makes that convenience visible is that the
dataset has to come from somewhere, and the somewhere is an import.

## Whether the kernel is a public library

Yes, it is importable by other software, and its interface is not stable yet.

The importable half is not a close call. Every implementation of this standard
today is a program somebody sells, and the reason this project exists is that a
kernel building acceptances depend on became unadoptable when its vendor stopped.
A kernel that is an implementation detail of one command line reproduces that
failure with better manners: an integrator who wants the arithmetic and not the
interface gets nothing, and the second tool that wants it reimplements it. So the
kernel is at an import path other software may use, from the first commit, and it
is documented as an interface rather than as internals.

The stability half is a promise about time and is dated rather than implied.
Before the first release the interface may change in any way, and that is stated
at the import path itself rather than left to be inferred from a version number
somebody has to know how to read. From the first release tagged with a major
version of 1, the interface is stable within that major version: a name that is
exported does not disappear, change its meaning or change its type, and what
grows is additions.

That promise is deliberately narrower than it could be, because two other
promises are already in force and they are the ones an integrator actually needs.
The result structure is versioned from today by
[result-contents.md](result-contents.md), and the record schema is versioned from
today by [data-format.md](data-format.md). A consumer that talks to this project
through a result document or a database record is already on a stable contract,
before any Go symbol is. That is the right order: the data contracts outlive the
language the kernel happens to be written in, and the language interface is the
one that should be allowed to move while the shape of the thing is still being
learned.

The cost of importability, paid knowingly. An exported name is harder to change
than an unexported one even before the promise starts, because somebody will have
imported it anyway. The mitigation is to export little: the types in the result,
the types a caller has to build to ask a question, and the functions that answer
one. Everything else stays unexported, and a thing gets exported when a caller
needs it rather than when it looks generally useful.

## Whether the command line is the product

It is the product for the first release, and it is also the reference consumer of
the kernel.

Those are not in tension. It is the product because it is the only thing an
operator can run, and no other interface is planned into the first release.
Whether one is ever built is entry 7 of issue #1 and is not decided here. It is
the reference consumer because it is the one place in this repository that
exercises the kernel the way an outside integrator would, which is how an
interface that is awkward to use gets discovered by its author rather than by a
stranger.

What that means in practice: the command line gets no privileged access. It calls
the same exported interface anyone else would, and where it needs something that
interface does not offer, the repair is to the interface rather than a private
route into the kernel. A test harness that reaches past a public interface to set
up state is the usual way that rule is broken, and here it would also invalidate
the reference-consumer argument, so it is refused in review.

## Whether the database is reachable independently

Yes. Somebody who wants the component values and not the calculation gets them
without installing or running this program.

The reason is the audience. A large part of the people this project is for
compute in a spreadsheet today, and the thing they are missing first is not an
algorithm, it is a set of laboratory values they can look up and cite. Making
those values reachable only through a program would be this project rebuilding
the subscription database it exists as an alternative to, with a command line in
front of it.

What makes it true is already decided rather than added here. The records are
JSON, one per file, in a tree that groups them by element kind, and the shipped
form is one file in the same syntax with a checksum beside it. A consumer needs a
JSON parser and the schema, both of which are ordinary, and needs nothing from
this repository's code. What this document adds is the direction of the
dependency: the database is versioned separately from the program, so a
correction to a value does not wait for a release of the software and a release
of the software does not restate the data.

Publishing that artefact is blocked on the licence of the database, which is an
open maintainer decision on issue #1, entry 2. Independent reachability is
decided here and does not depend on that answer; what the answer decides is what
a recipient may then do with it.

## What an interface added later attaches to

It attaches to the kernel's public interface and to the result structure, and it
sits above the command line in the chain rather than beside it.

Above rather than beside, because the alternative is two consumers of the kernel
that each grow their own composition logic, and the second one is where the
divergence shows up: a quantity computed one way on the command line and another
way in a window, both correct, both different, and no way for a user comparing
them to tell which they are looking at.

What it is forbidden to reach around, stated so that a later reader does not have
to reconstruct it:

- It may not compute an acoustic quantity itself. Not a sum, not a conversion,
  not a rounding of a value into a different one. If it needs a number that the
  kernel does not return, the kernel gains it.
- It may not read the database's files directly. It goes through the data access
  layer, so that validation, schema version refusal and provenance happen once.
  This restriction is on interfaces belonging to this project and does not apply
  to anybody else's software, which is the point of the previous section.
- It may not present a value without what the result says about that value. A
  defaulted input is marked as defaulted wherever it is shown, and an incomplete
  result is not displayed as a complete one because the layout was tidier.
- It may not become the only place a capability exists. Anything it can do, the
  command line can do, so that a script can do it too.

## Where this leaves the tree

The tree layout follows from this document rather than the other way round, and
it is issue #19, which names this document and matches it. The mapping is one
directory per layer, in the chain's order, plus the data tree and the harnesses.
This document does not print the directory names, because the layout issue is
where they are decided and a copy here would be the second authority on a
question that should have one.

## What would reopen this decision

An integrator's report that the exported interface cannot express a question the
command line can ask, which would mean the reference-consumer claim is false.
A validation result that cannot be reproduced across machines, which would mean
something below the command line is reading the environment. Or an answer to
entry 7 of issue #1 that puts a graphical interface in the first release, which
does not change the chain but changes what has to be stable by when.

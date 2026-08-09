# Decision: what a version number promises

Status: decided for the first release. The promises here bind from the first
release onwards, whatever number that release carries. What would reopen the
decision is at the end.

Two things are released from this repository and they are released separately:
the program and the component database. Each carries its own number and neither
implies the other. This record says what each number promises, what it
deliberately does not promise, and how a user holding one of each knows whether
the pair works.

## The shape of the program's number

Three decimal numbers separated by dots, and nothing else. No pre-release
suffix, no build metadata, no leading zero inside a part.

The release tag is `v` followed by that version, and `cmd/release` refuses a tag
of any other shape. It also refuses a tag whose version is not the version the
program in the artefact set says it is when it is run, because the version is a
constant in the source and the tag is typed by a person, and the two disagree
exactly when somebody tagged before raising the constant or raised it and tagged
the commit before. Both mistakes produce a complete artefact set that is wrong
about what it is, and nothing downstream can tell.

## What a change in each position promises

The number governs the interfaces, and not the numbers the tool computes. That
sentence is the decision; the section after this one is what carries the other
half.

An interface here is anything somebody else builds on: the command line and its
flags, the machine-readable output, the project file, and the exit behaviour.

**Major** is a change that takes something away from somebody who was relying on
it. A command or a flag removed or renamed, an existing flag whose meaning
changes, a machine-readable output field removed or changed in meaning, unit or
type, or a project file that a program at the previous major can no longer read.
The result structure carries its own schema major and what it promises is
[result-contents.md](result-contents.md) rather than anything restated here. The
project file will carry one too; it does not exist yet and issue #91 builds it.

**Minor** is a capability added where nothing is taken away. A new command, a
new flag whose default leaves existing behaviour unchanged, a new field in the
machine-readable output, a junction type the model could not describe before, a
situation that was refused and is now computed.

**Patch** is a correction that adds no capability, removes none, and moves no
computed value. A defect in a message, a document, a refusal that named the wrong
input.

## The position on a numerical change

A change that makes a result more correct is not a bug fix to the person who
recorded the old result against a building. It is the number in their file
changing after they signed it, and it is the single most consequential kind of
change this project can make.

**A numerical change is not expressed in the version number. It is a separate
statement, and every release carries one.**

Folding it into the number was the alternative and it is refused for two
reasons. Every correction to the arithmetic would be a major release, so the
major position would stop meaning "your integration breaks" and start meaning
"something moved", after which it answers neither question. And the two
audiences are different people: the person whose script parses the output and
the person whose report quotes a number are not usually the same person and are
never asking the same thing.

What the separate statement has to do:

- Every release carries a numerical-change statement. Where nothing moved it says
  so in those words. A release with no statement at all is a defect in the
  release rather than a release where nothing moved, and those must not look the
  same.
- Where something moved, the statement names the quantity, the direction and the
  size of the movement, and which cases changed, with the command that produced
  the comparison.
- A release that moves any computed value is a minor increment at least, even
  when nothing else in it is a minor change. So a user who upgrades within a
  patch range can rely on getting the same numbers, which is the one promise a
  person holding a signed report actually needs.

The mechanism that would refuse a patch release moving a number is the numerical
regression gate, issue #111, which records every result and fails a change that
moves one. It does not exist yet. Until it does, the rule above is applied by a
person reading a diff, and this sentence is here so that nobody reads the rule
as something the tree enforces.

## The database's number

The same three positions, over a collection of records rather than over code.

**Major** is the record schema major. Tying the two together is what makes the
compatibility rule below readable: a program that says which record schema majors
it reads has said which databases it takes. What a record schema major promises,
and the migration that rewrites every record when it changes, is
[data-format.md](data-format.md).

**Minor** is records added, and any change that moves a value a user could have
already read. A corrected transcription, a superseded measurement replacing the
value a record reports: both change what a project computes, and the argument
above applies unchanged.

**Patch** is a change that moves no value. A correction to a description, a
provenance field filled in, a source pointer repaired.

## The compatibility rule between the two

A program declares the record schema majors it can read. A database whose major
is outside that range is refused, naming the version it saw and the versions the
program understands, and the refusal is the record-level one that
[data-format.md](data-format.md) already decides, reported once for the artefact
rather than once for every record.

Inside one major, any program reads any database. That holds because of the
within-major rule on records, which allows a field to be added and forbids a
field to be removed, retyped or given a new meaning. It is that rule doing the
work, not this one.

Neither number constrains the other. A user may hold program 1.4.0 with database
3.2.0, and the pair is supported when the program reads major 3.

**A database update can change a result with the program unchanged.** This is
the case a user does not expect, because in most software the program is the
thing that changes. A record corrected in the database changes every project that
cites it, on a machine where nothing was upgraded except the data.

That is why the numerical-change statement above is required of a database
release exactly as it is of a program release. A result already records the
version of the software that produced it, in `kernel_version`, and it records no
database version today. Something has to, or the question "which of my projects
moved" has no answer that does not involve remembering. Where that field is added
is the result structure's own work, and stating that it is owed is what this
record does about it.

The documentation that says all of this in the user's language, rather than in
this one, is issue #129.

## What is not decided here

Which number the first release carries. That belongs to the release that is cut,
issue #131. This record says what any number promises, not which one is next.

Whether the promises weaken below 1.0.0. They do not. A promise that begins when
somebody declares 1.0.0 is a promise nobody can rely on at the moment they most
need it, which is the first time they adopt the tool. If the first release
carries a leading zero, everything above still holds.

Signing and publishing the artefact set is issue #127, and this route does
neither. Building the same bytes twice from one tag is issue #126. Releasing the
database beside the program is issue #128, and the number it carries is decided
here while the route that produces the artefact is decided there.

## What would reopen this

A user, or a downstream tool, reporting that they upgraded within a patch range
and a number moved. That is the promise this record makes most loudly, and one
instance of it being false is worth more than any argument here.

An outside consumer of the machine-readable output arriving, at which point the
major position acquires a cost this record has only reasoned about.

# Decision: what identifies a component record, and what keeps it identified

Status: decided for the first release. It answers the first condition of issue
#56 and nothing else on that issue, which needs a record type and a project file
that do not exist yet.

This is written before there is a second record on purpose. An identity scheme is
the one thing in a database that cannot be changed later: every project file that
ever cited a record is a copy of the scheme, held by somebody this project cannot
reach.

## The scheme

A record's identity is a short lower-case name, minted once from the published
document that established the record, and never derived again.

The one record in the tree carries the shape. The first line of that output is
the record; the second is the schema, which the same glob reaches and which is
not a record:

    $ git ls-files 'data/*/*.json'
    data/floor/ift-17-002083-pr01-x01-x02.json
    data/schema/component-record.schema.json

The name is the laboratory in the form it is commonly abbreviated, the report
number, and the measurement numbers within that report, lower-cased with
everything that is not a letter or a digit written as a single hyphen. The schema
constrains the spelling and not the derivation:

    $ git grep -n 'a-z0-9' -- data/schema/component-record.schema.json
    data/schema/component-record.schema.json:307:      "pattern": "^[a-z0-9][a-z0-9-]{2,63}$"

Three properties do the work, and the third is the one that is easy to miss.

**It is derived from the document rather than from the record.** A description, a
product name, a thickness and a rating are all things this project might have
typed wrongly or might restate later. The report number and the measurement
number are printed on a document that exists outside this project and cannot be
edited by anybody here.

**It is minted once and then frozen.** This is the part that makes it an identity
rather than a derivation. If the report number in the identity is later found to
have been mistyped, the identity keeps the mistake and the `provenance` block is
corrected. An identity that is re-derived whenever its inputs are corrected is a
hash of the record wearing a readable spelling, and it fails for the same reason:
a typographical fix silently repoints every citation.

The cost is stated rather than discovered. A record can carry an identity that
misspells the report it came from, forever, and a reader who reconstructs the
report number from the identity rather than from the provenance will get it
wrong. What the identity promises is that it names one record; it does not
promise to be a correct citation, and the provenance is what is.

**It is readable, and that is a convenience rather than a contract.** Nothing may
parse an identity. A route that split one on hyphens to recover a laboratory
would be reading a field that the paragraph above just said may be wrong, and it
would break on the first laboratory whose abbreviation contains one.

## The two schemes the issue rejects

**A sequential number** depends on insertion order, and the database is assembled
by several people from documents they find in an order nobody controls. Two
people entering records at the same time either collide or coordinate, and
coordination is a step that has to happen before every entry forever. It also
carries no information, so a citation in a project file cannot be sanity checked
by a person reading it.

**A hash of the whole record** changes when any byte of the record changes. A
correction to a description, a layer added years later, or a `superseded` entry
appended would each mint a new identity for the same thing, which breaks every
project that cited it and does so silently. The scheme above is the same idea
narrowed to the fields that cannot be corrected, and then frozen so that even
those cannot move it.

## What is one record and what is two

This is where the scheme has to be stated carefully, because two documents in the
tree pull in different directions and both are right about their own case.

A record is one specimen as one report established it. Two laboratories testing
what a catalogue calls the same construction tested two specimens, built by
different people, in different facilities, and the two results differ for reasons
that are the subject of this whole project. They are two records with two
identities, and presenting them as one would average away the thing a user needs
to see.

A later test of the same specimen, reported by the same laboratory under the same
report, is not a second record. The format decision is the authority for what
happens then, and it is supersession inside the record rather than a new one:

    $ git grep -n 'The list is append-only' -- docs/decisions/data-format.md
    docs/decisions/data-format.md:102:list, and puts the new one in its place. The list is append-only.

So the boundary is the report, not the construction and not the measurement. A
new report is a new record. A further measurement inside a report that the record
already names is a superseded value inside it.

One consequence is worth naming because the first record in the tree is already
on the wrong side of the convenient answer. That report carries thirty records
and twenty-nine of them would share every provenance field with the first:

    $ git grep -n '30 records are available' -- docs/first-record-by-hand.md
    docs/first-record-by-hand.md:197:- 30 records are available from this one report, and 29 of them would share every

They are thirty records and not one, because they are thirty specimens. The
measurement numbers in the identity are what keeps them apart, and a scheme
stopping at the report number would have collapsed them.

## Reopening an old project

This is the property the issue calls the expensive one, and it is not a property
of the identity alone. An identity points at a record. It does not point at what
that record said on a particular day, and no scheme that fits in a name can.

What carries the rest is the database version. The database is released as its
own artefact with its own version and its own checksum, separately from the
program:

    $ git grep -n 'carries its own version' -- data/README.md
    data/README.md:26:artefact carries its own version and is cut when the data changes, and the
    data/README.md:27:program carries its own version and is cut when the code changes. Neither version

So a citation that reproduces is an identity together with the database version
the result was computed against, and a result records both. Reopening an old
project resolves its identities against the version it names, which reproduces
the old numbers exactly, and then resolves the same identities against the
current version to say which of them have moved. A record whose value has been
superseded since carries the old value and the date it stopped being current, so
the comparison can be shown rather than merely announced.

Two things follow that the work downstream owes rather than this document.

A project file cites a database version as well as an identity, or it cannot
reproduce anything. That is issue #91.

A resolution against a version that is not present has to refuse rather than fall
back to the current one. Falling back is the failure this whole arrangement
exists against: it produces a plausible number under an old project's name.

## What this does not decide

It does not decide what happens when a report prints no measurement number, which
is the case a second report will probably bring. The scheme needs a stated rule
for minting an identity when the document gives fewer coordinates than this one
did, and inventing that rule from one example would be inventing it from the
example that did not need it.

It does not decide the identity of anything that is not a component record.
Issue #56 speaks of elements, and an element lives in a project rather than in
the database, which is what `data/schema/component-record.schema.json` already
says about why an element is not a record type. What identifies an element inside
a project is a question for the project file.

Nothing in this repository refuses a violation of any of the above. The schema
constrains how an identity is spelled and nothing reads how it was derived,
nothing checks that a minted identity was not re-minted, and nothing compares an
identity against the provenance it was taken from. A route that could check the
last of those would need the report, which this project does not hold. Until a
check exists, this document is held by review.

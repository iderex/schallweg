# The component database

This directory holds the component records: laboratory values for building
elements, linings and floor coverings, entered from published test certificates.
It holds data and nothing executable. No Go package in this module imports it,
and nothing in the code refers to a particular record. The `store` package knows
how to read a record; it does not know that any record exists.

## Why it is here and not in a repository of its own

Every check that would catch a bad record lives here: the schema, the validation
of records against it, the fixtures and the gate that runs them. A database in a
second repository would need either its own copy of that apparatus, which then
drifts against this one, or a route that validates data it does not hold. One
contribution route also means a correction to a record and a correction to the
code that reads it can be reviewed against each other.

The argument in full, including what this costs and what would reverse it, is in
[../docs/layout.md](../docs/layout.md).

## It is released separately from the program

A wrong laboratory value is wrong today, and waiting for a software release to
correct it is the failure the format decision already names when it says the tree
is the authority and the shipped artefact is derived from it. So the database
artefact carries its own version and is cut when the data changes, and the
program carries its own version and is cut when the code changes. Neither version
number implies anything about the other, and a result records both.

## What a record has to carry

The format, the granularity, the versioning rule and how a value is superseded
without being destroyed are
[../docs/decisions/data-format.md](../docs/decisions/data-format.md). What may be
taken from a certificate, what is never reproduced, and the provenance floor a
record has to reach are
[../docs/decisions/certificate-extraction.md](../docs/decisions/certificate-extraction.md).

Two rules from those documents are worth stating where somebody adding a record
will meet them. No certificate document is stored here, in any form: a record
cites its source and a reviewer follows the citation. And a record enters from a
published certificate that this project read, never by being copied out of
somebody else's database.

## What is here today

One record, the machine-readable schema in `schema/component-record.schema.json`,
and this README.

    git ls-files 'data/**' | wc -l
    3

The one record is `floor/ift-17-002083-pr01-x01-x02.json`. It is here because
entering a record by hand before an importer exists is the cheapest test of the
schema, and what that test found is
[../docs/first-record-by-hand.md](../docs/first-record-by-hand.md). Read that
before adding a second record: it lists the fields the schema asks for that a
published report does not print, the facts a report carries that a record cannot
hold, and the ten decisions the entry needed that nothing in the tree records.

Nothing in this repository validates anything against that schema yet. It is a
document that says what a record must carry and no route reads it, so a record
added today would be checked by whoever reviewed it and by nothing else. The
check that closes that is issue #38.

What provenance has to carry as a data model is issue #74, and the set the
database starts with is issue #81.

# Decision: the on-disk format for component data and how it is versioned

Status: decided for the first release.

The format is chosen against four uses that pull in different directions: read by
this tool, corrected by people, reviewed as a diff, and consumed by other
software. Where they conflict, review wins, because a wrong laboratory value that
nobody can see in a diff is the failure this database exists to avoid.

## Editing format

JSON, UTF-8, one record per file, under a directory tree that groups records by
element kind.

JSON rather than a binary store, because a correction has to be readable in a
diff by somebody who is checking it against a test report, and an opaque blob
turns every correction into a claim about a tool's output. JSON rather than a
tabular text format, because a record is not flat: it carries a spectrum, a
provenance block, and a history of superseded values.

JSON rather than YAML or TOML, for one reason each. YAML's implicit typing turns
unquoted tokens into the wrong type without complaining, and a database whose
values are physical quantities cannot afford a parser that decides what a token
means. TOML is pleasant to hand-edit but its schema tooling is thinner, and this
project needs a machine-readable schema more than it needs comments.

The cost of JSON is real and is accepted: no comments, so anything a human would
write in a comment has to be a field with a name, and trailing-comma and quoting
mistakes are a common first contribution failure. The schema check catches both,
and the error message naming the record and the field is part of what the check
owes.

## Shipping format

One artefact containing every record, in the same syntax, with a checksum beside
it.

Same syntax, different packaging. A consumer that has parsed one record can parse
the artefact without a second parser, and the project does not maintain two
serialisations that can disagree. What differs is that the shipped artefact is
one file with a stated build order, so its bytes are reproducible from the tree,
and it carries the schema version and the checksum of the tree it was built from.

The tree is the authority and the artefact is a derived thing. Nobody edits the
artefact, and a correction that appears only there is a defect.

Publishing that artefact is blocked on the licence of the database, which is an
open maintainer decision on issue #1, entry 2. The format decision here does not
depend on that answer and is not waiting on it. Building the artefact and
checksumming it is issue #80.

## Record granularity

One record per file.

Effect on review: a correction is a small diff in a file whose path names what
was corrected, and two people correcting two elements never touch the same file.
That is the whole reason for the choice.

Effect on scale: the tree gets large. A few thousand records is a few thousand
files, which is slow to list, slow to clone on a bad connection, and enough to
make a naive tool that reads every file at startup feel slow. That cost is
carried by the shipping artefact: the tool reads one file, not a few thousand.

The alternative, many records per file, is compact and makes every correction
touch a file that somebody else is also editing, which converts an independent
correction into a merge conflict. At the scale this database is aiming for, that
is the more expensive failure.

## Schema versioning

Every record carries a `schema` field with a major version. The rule has three
parts.

Within a major version, a field may be added and an enumerated value may gain a
member. No field is removed, no field changes its meaning or its unit, and no
field changes type.

A reader that meets a record whose major version is higher than the one it knows
refuses that record, names it, and says which version it saw and which it
understands. It does not skip the record silently and it does not attempt to read
the fields it recognises, because a major version exists precisely to say that
the fields it recognises may no longer mean what it thinks. Refusing one record
does not abort the whole load: the reader reports every such record and continues,
so the user sees the size of the problem rather than the first instance of it.

A major version change requires a migration in the tree that rewrites every
record, so no record is ever left behind at an older major. The database in the
tree is therefore always at one version, and the multi-version case above exists
for a consumer holding an older copy of the software, which is the case that
actually occurs.

## Superseding a value without destroying it

A laboratory value is a fact about a particular test on a particular date. A
later test is a different fact, not a correction of the first, and overwriting
the first destroys evidence somebody may need to explain an old result.

The rule: a value is never edited in place once it has been published. A record
carries the current value and a `superseded` list. Adding a new measurement moves
the previous value, with its provenance and the date it was current, into that
list, and puts the new one in its place. The list is append-only.

Each superseded entry carries why it was superseded, in one of two forms that are
kept apart because they mean different things. A `remeasured` entry means a new
test exists and the old one was correct for its own test. A `corrected` entry
means the earlier value was entered wrongly from its report, and it names what
was wrong. Collapsing these two into one field would make an entry error look
like a physical change, which is the distinction the whole list is for.

A result computed from a superseded value can be explained afterwards, because
the value and its date are still in the record. The mechanics of making the
correction are issue #78.

## The machine-readable schema

The schema is a JSON Schema document in the tree, and writing it is issue #73.
It is the authority for what a record must carry; this document is the authority
for the format and the versioning rule and does not restate the field list,
because a field list in prose drifts against the schema that decides it.

Validating every committed record against it, and refusing a data file that no
schema claims, is issue #38.

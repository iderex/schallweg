# Decision: the bill of materials, its format and where it comes from

Status: decided for the first release.

An engineer installing this in a consultancy is asked what is inside it, and a
public body asks the same question in writing. The answer is produced by the
build, in a format the person asking can hand to their own tooling, and it is
attached to what it describes so it travels with the artefact rather than being
reconstructed from memory afterwards.

## The format

SPDX 2.3, serialised as JSON.

SPDX rather than the alternatives for one reason that outweighs the rest here:
its 2.2.1 revision was published as ISO/IEC 5962:2021, and the audience for this
project is public bodies and the consultants who answer to them, where an
answer that names an ISO publication is worth more than an answer that names a
better tool. The 2.3 revision is the current published version of that line and
is what this repository writes.

JSON rather than the tag-value or spreadsheet serialisations, because the
document is read by machines far more often than by people, and because a
reviewer can already read JSON in this repository.

The version string is written in exactly one place, `spdxVersion` in
`cmd/sbom/spdx.go`, and the build refuses a document that does not declare it.
That refusal exists because a format version that drifts is read wrongly by every
consumer downstream and nothing about the document looks wrong while it happens.

## Two documents, never one

The program and the component database get separate documents.

They are released separately, they carry separate versions, and a user may take
only one. A single document covering both would say that taking the program tells
you something about the data, which it does not. The two also answer different
questions. For the program the question is which modules were linked. For the
database it is which records are in it and what each one's bytes are, because a
database is a collection of facts rather than a graph of libraries.

## Where each one comes from

The program's document is read out of the compiled binary, with `debug/buildinfo`
from the standard library. The toolchain records the module graph and its own
version inside the executable, so the document describes what was linked rather
than what the source declared. Those two can differ, and the artefact is the one
the question is about.

The toolchain is listed as a component. With no dependencies declared, the
standard library and the toolchain that supplied it are the entire supply chain,
and a document that left them out would read as if nothing were inside the
binary.

The database's document lists every file under the record tree with its checksum,
not only the files a route recognised as records. A file nobody listed, sitting
inside an artefact somebody downloaded, is the shape this document exists to
refuse.

## The means

Written in this repository, in Go, against the standard library.

The alternative was an external generator, and it was refused on two counts. It
would be the first dependency this module takes, and its licence could not be
cleared against the licence of this repository, because that is an open
maintainer decision on issue #1 and there is nothing yet to judge compatibility
against. It would also put a download in the middle of the one job whose whole
claim is that it built the tree from a clean checkout. Against that, an SPDX
document is JSON, and `encoding/json` and `debug/buildinfo` are both in the
standard library, so what is being avoided costs less to own than to import.

The module graph is unchanged by this decision:

    go list -m all
    github.com/iderex/schallweg

The cost is real and is the next section.

## What the document does not say, and what nothing here checks

No licence. Every licence field is `NOASSERTION`, which is the specification's
word for "this document does not say". That is the honest value rather than a
placeholder: this repository has no licence file, the choice is an open
maintainer decision on issue #1, and a document asserting a licence would be
asserting something nobody has decided.

No vulnerabilities. A bill of materials is a list of what is there. Whether any
of it has a known advisory is a different question and a different route.

No conformance to the published SPDX JSON schema. Nothing in this repository
validates a produced document against it. Doing that needs a schema validator,
which is a dependency, and the dependency policy refuses one that cannot be
cleared today. What the suite checks instead is the field list written into
`cmd/sbom/sbom_test.go`, which is narrower than the schema. A document that
passes every test here can still be refused by a reader that validates properly,
and that is the residual rather than a thing that is about to be fixed.

No database version. The database carries its own version, cut when the data
changes, and no release has happened, so the field is left out rather than filled
with a number this repository invented.

## What makes it evidence

The document is reproducible from the tree. The creation timestamp and the
document namespace are arguments rather than things the command reads from the
clock or invents, and the file list is sorted rather than left in the order a
file system happened to walk it. So the same commit produces the same bytes
twice, and a document that differs means the artefact differs.

The build compares the document against the toolchain's own reader: every module
`go version -m` prints for the binary has to appear in the document, or the check
fails. That is a second route to the same fact rather than the same route twice.

Stated exactly, because a green run has to be read for what it is: the module
graph is empty today, so that comparison compares an empty list and passes
without having refused anything. It begins to bite on the first dependency this
project takes.

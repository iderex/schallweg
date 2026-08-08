# Threat model

A desk-side calculation tool has a small threat model, and most of the value in
writing it down is in saying what is not a threat, because that is what stops
effort going to the wrong places.

This document is read alongside [SECURITY.md](../SECURITY.md), which says how to
report something and what is in and out of scope for a report. Where the two
overlap, SECURITY.md is the one a reporter reads and this one is the reasoning
behind it.

Every control named below points at the issue that delivers it. Where a control
is already in the tree the entry says so and names it; where it is not, the entry
names the issue and nothing more, because an issue is a plan and a plan is not a
control.

## What there is to protect

The result. A number this software produces gets copied into a letter, a report
or an acceptance document, and somebody makes a decision about a building with
it. It is the asset with the most value and the least protection, because
nothing about a wrong number looks wrong.

The operator's project data. Rooms, elements, addresses, client names, and
whatever else a project file has been made to carry. It sits on the operator's
own machine. What it may contain and what the operator owes their client for it
is issue #115, and this document does not answer the data protection question.

The component database and its provenance. The records are the reason to trust a
result, and a record whose certificate reference is wrong is worse than a record
that is missing, because it carries the appearance of a source.

The release artefact and the database artefact, once they exist. A user
downloads them and runs them, which makes them the shortest path from anybody
who can alter them to a machine inside a consultancy or a public body.

The build pipeline. It runs with tokens, on inputs that changes bring with them,
and it is what produces everything in the paragraph above.

## Who might do something about it

Nobody in this model is a targeted attacker with a budget. That is a statement
about what this project is, and it is worth being explicit that it is an
assumption rather than a finding: nothing here has been tested against one.

The sender of a file. A colleague, a client or a manufacturer sends a project
file, a spectrum or a certificate. They are usually not hostile and the file is
usually just malformed, which is why the parser has to treat both the same way.

Somebody who can alter what a user downloads. A mirror, a compromised release, a
substituted database file, a network in between.

Somebody who can get code into the build. A pull request, a dependency, an
action referenced by a tag that later moves.

The user, unintentionally. Somebody supplying laboratory data for a construction
that is not the one on site, or reading a prediction as though it were a
measurement. This is the likeliest route to real harm in the whole document and
it is not an attack.

## The entry points, and what addresses each one

A project file from an untrusted sender. This is the main one, because
colleagues send each other project files and one of them opens it. What
addresses it: a reader that refuses a file from a newer version rather than
guessing at it, issue #91; errors that name the input that caused them so a
refusal is actionable rather than a dead end, issue #94; and fuzzing of every
parser that takes a file from a stranger, issue #106. The layering already in
the tree keeps the parsing in one layer: `docs/decisions/layering.md` puts every
question about untrusted bytes in the data access layer, and
`cmd/gate/architecture.go` refuses a file system dependency in the two layers
below it.

A spectrum in an exchange format. The same shape as above and a smaller surface.
What addresses it today: the reader in `store/spectrum.go` refuses a spectrum
with a missing band, an unknown quantity, a unit that is not the quantity's and a
declared version it does not know, and its fixtures are in
`store/testdata/`. What is still owed: the fuzzing in issue #106.

A component record. The records are data in the tree, edited by people and read
by the program. What addresses it: validation that refuses a record that cannot
be trusted, issue #77; a provenance definition that says what a record has to
carry before it counts, issue #74; and a correction route that does not destroy
what a record said before, issue #78.

A database file downloaded from this project. What addresses it: publication as
an artefact with a checksum, issue #80; release of the database beside the
program and versioned separately from it, issue #128.

The release artefact. What addresses it: signed and published artefacts, issue
#127; a release reproducible from a tag so that a second party can rebuild it and
compare, issue #126; and a bill of materials a user can read, issue #124.

The build pipeline. What addresses it today: the workflow security audit
(`.github/workflows/zizmor.yml`), code scanning
(`.github/workflows/codeql.yml`), the dependency review
(`.github/workflows/dependency-review.yml`), and actions pinned to a commit
rather than to a tag, which is visible in every workflow file in the tree. What
is still owed: pinned inputs and a lock file that cannot drift, issue #107; a
second analyser with a different lens, issue #101; verified signatures on the
protected branch, issue #112; and the required status checks that would make any
of the above a precondition of a merge rather than a report beside it, issue #36.

Until issue #36 lands, every check in this document is a report and not a gate.
That sentence is the most important one in this section and it is not softened
anywhere else in this document.

## The harm that is not a security failure

A wrong number that survives to a building acceptance is the largest harm this
software can do, and it belongs in this model even though it is not a security
threat in the usual sense. Nothing about it involves an adversary. It is the
ordinary failure mode of a calculation tool and it is the reason the rest of the
plan is arranged the way it is.

What addresses it: the validation corpus, which is the published comparison
exercises collected in issue #83, the agreement criterion decided before anything
is measured in issue #84, the first case encoded end to end in issue #85, the
whole corpus run in the gate in issue #86, and the record of every case where
this implementation disagrees and why in issue #89. Beside it, the numerical
regression gate over that corpus, issue #111, which is what catches a change that
moves a number nobody was looking at.

Underneath both, the proofs at the arithmetic level: the rating procedures proved
against worked examples with their sources, issue #45; hand-worked airborne and
impact cases, issues #66 and #72; and the refusals that stop a calculation
running on data that cannot support it, issues #46, #54 and #64.

Beside all of that, the statement that a prediction is not a measurement, issue
#117, which is a control on how a result is read rather than on how it is
computed, and is the only control here that reaches the last step.

The boundary this shares with SECURITY.md is stated there and repeated here in
one sentence because it is the thing people get wrong: a wrong number that a
chosen input produces on purpose, in order to make an acceptance pass that should
have failed, is still a wrong number, reported in the open.

## What is out of scope

There is no network service, so there is no authentication, no session, no
access control and no server-side anything. Nothing in this project sends
anything anywhere unless an operator switches it on, which is its own decision
record in `docs/decisions/nothing-leaves-the-host.md`, and the test that proves
it is issue #116.

There is no multi-user model. The program runs as the operator, reads what the
operator can read and writes what the operator can write. Separating one user of
one machine from another is the operating system's job and this software does
not attempt it.

No secrets are held. There is no credential, no token and no key in anything the
program reads or writes. The build pipeline holds tokens and that is the entry
point above, not this one.

A denial of service from a very large input is out of scope as a vulnerability
and in scope as a defect, for the reason SECURITY.md gives.

Whether this project ever hosts anything is entry 3 of issue #1 and is
unanswered. This document models the answer that is true today, which is that it
hosts nothing. An answer that changes that reopens most of this document rather
than adding to it.

## What was not analysed

This section exists so the document cannot be read as complete.

No adversary with resources was modelled, and nothing here has been tested
against one.

The toolchain below this project was not analysed. The Go distribution, the
runner images, and the hosting service itself are trusted here without argument.

No dependency was analysed individually. The module has no third-party
dependency in it today, which is a fact about the tree rather than a control, and
the controls that would hold it are issues #107 and #124.

Nothing was analysed by running it. No fuzzing has been done, no scanner output
was read as part of writing this, and no attack was attempted against any parser.
The parser entry points above name issue #106 for exactly that reason.

The import route from a structured certificate, issue #76, was not analysed
because it does not exist. It will be the largest untrusted-input surface this
project has when it does, and this document is owed a revision at that point
rather than a line added to it.

No graphical interface was analysed. Whether the first release has one is entry 7
of issue #1 and is unanswered.

The data protection position is not here. It is issue #115 and it is a different
question answered for a different reader.

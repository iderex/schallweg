# Data protection: what this software holds, where it stays, what it sends

Written for the person who has to answer a data protection question in writing.
It states what personal data this software can hold, where that data lives, what
the program transmits, how long anything is kept, and what evidence stands behind
each of those answers. Where the answer is that something has not been built or
not been measured, it says so in the same sentence rather than in a footnote.

The operator of this software is a consultancy or an engineering office running
it on their own machine for their own client. That operator is responsible for
the personal data in their projects. This document exists so that the
responsibility can be discharged from stated facts instead of from assumptions
about what a program might be doing.

## What personal data this software can hold

Two answers, because they are different in kind. One is about the operator's own
work, which does not exist in this tree yet. The other is about the component
data this project publishes, which does.

### The operator's project data

None today. The program does one thing:

    go run ./cmd/schallweg
    schallweg 0.0.0

It reads no file, creates no file and holds nothing. It writes one line to
standard output and exits. There is no project file format yet, and issue #91
builds it.

When it exists, a project file describes a building and the rooms in it, and the
categories of personal data it can carry follow from that. An address, which
identifies a building and in a single-occupancy building identifies a household.
The name of the client the work is done for. The name of the person who
commissioned the work, where that is a person rather than an organisation. Room
and occupancy descriptions, which in a residential assessment can identify a
household and sometimes name occupants. And whatever free text the operator
writes into the file, which is not a category anybody can bound in advance.

That list is the constraint #91 works under rather than a description of a format
that exists. It is stated here now because a format decided without it is a
format somebody has to unpick afterwards.

### The component data this project publishes

The component database ships with this project. It is not the operator's data,
and an operator asked what personal data is in this software has to be able to
answer for it as well.

A record carries the laboratory that performed the test and the client the test
was performed for, which are organisations, though a sole trader is a natural
person and the field does not distinguish. It carries the report number and the
specimen designation, which name documents and objects rather than people. And it
carries the person who entered the record, by name or by account name, which is
personal data about a contributor and is a required field:

    git grep -n 'entered_by' -- data/
    data/floor/ift-17-002083-pr01-x01-x02.json:36:    "entered_by": "iderex",
    data/schema/component-record.schema.json:242:              "entered_by",
    data/schema/component-record.schema.json:251:            "required": ["described_from", "entered_by", "entered_on"]
    data/schema/component-record.schema.json:359:        "entered_by": {
    data/schema/component-record.schema.json:421:        "provenance": { "$ref": "#/$defs/provenance", "required": ["entered_by", "entered_on"] },

The field is there because entry from a document is the step where a mistake is
invisible, and a record nobody is named against is a record nobody can be asked
about. That is a deliberate trade and this document names it rather than leaving
an operator to find it in a schema.

## Where the data is stored

On the operator's own machine, in files the operator names and puts where they
choose, and nowhere else.

The program creates no application directory of its own, no cache, no index, no
log file and no temporary store. That is verifiable today in the strongest way
available, which is that it creates nothing at all: the program's own source is
one file, and printing the line above is the whole of what it does.

When the project file lands under #91, the same position has to hold, and the
requirement is that the file the operator names is the only place their project
data is written. A program that keeps a recent-files list, a crash-recovery copy
or a working directory beside the file has created a second copy the operator did
not ask for and does not know to delete.

## What is transmitted

Nothing.

The program is built from this module and the language's standard library, with
no third-party package that could reach anything on its own:

    go list -m all
    github.com/iderex/schallweg

And no networking package of that standard library is in the program's transitive
dependency set at all, so there is nothing in the built program that could open a
connection:

    go list -deps ./cmd/schallweg | grep -E '^net(/|$)'
    (no output, exit status 1)

The program also does not reach the network at one remove by starting another
program, because the package that runs one is not in that set either:

    go list -deps ./cmd/schallweg | grep -E '^os/exec$'
    (no output, exit status 1)

Both commands were run at the commit this document is being added on. What they
are and are not evidence for is the last section of this document, and it should
be read before either is quoted.

## Everything that could ever transmit, and what turns it on

Three features, and a feature not among them is not permitted to connect at all.
The list is fixed by [decisions/nothing-leaves-the-host.md](decisions/nothing-leaves-the-host.md),
which is the authority for it. None of the three is written yet.

Downloading a component database update sends the version of the database
currently held and nothing else, and receives the published artefact and its
checksum. It happens when the operator runs the update command, and at no other
time: there is no background check, no reminder and no timer.

Submitting a component record back to this project sends the record the operator
chose to submit and the provenance it carries, with the exact payload shown
before it goes. It happens when the operator runs the submit command and confirms
it. Whether such submissions are accepted at all is an open maintainer decision,
entry 6 of issue #1.

Fetching a published comparison exercise is not part of the shipped program. It
belongs to the tooling that maintains the validation corpus, it runs on a
maintainer's machine, and the corpus is committed so that the ordinary suite
reads files rather than the network.

None of the three is on by default, and none has a default that could be turned
on by a configuration file, an environment variable or an installer choice. What
would have to be true for one of them to be on by default is written in the
decision record, and one of the three conditions it names is a statement in this
document, so a default that is not described here is a default nobody was told
about.

If this document and the decision record ever say different things, the decision
record decides and the difference is a defect here.

## How long anything is kept

Until the operator deletes it.

The program has no retention period, no expiry and no housekeeping, because it
holds nothing of its own to keep. The operator's project files are ordinary files
in the operator's file system, under the operator's own backup and retention
arrangements, and deleting a project file is the whole of deleting the project.

Nothing is retained by this project, because nothing reaches this project. That
follows from the transmission position above rather than from any promise about
handling data on arrival, and it is the shorter and stronger of the two.

## What a report contains

The report is what leaves the operator's control, because it is the thing handed
to a client, a contractor or an authority. What a report carries is therefore the
part of this document an operator will need most.

It cannot be stated yet. There is no report generator in this tree, and issue #96
builds it. When it does, it owes this section the list of what a report carries
and, separately, a short statement of what a prediction is not, which is the
condition issue #117 is still open on.

Until then this document states no report contents, and an operator handing over
a document produced from this software is handing over something they assembled
themselves.

## What proves this and what does not

The transmission claim above is the one somebody will rely on, so what stands
behind it is set out rather than implied.

The layering rule in the local gate refuses the networking packages by name, but
only in the numeric floor and the kernel. It does not reach the data access layer
or the command line, which are the layers a connection would be written in, so it
is a rule about internal purity and not a proof about the program.

The suite rules refuse a test that imports a networking package. That says the
tests do not connect. It says nothing about the program, and treating one as
evidence for the other is the substitution the decision record names as the one
to avoid.

The two commands in the transmission section are a measurement of the program's
dependency set at one commit. They are not a check. Nothing re-runs them, and a
later change that adds a networking package to the command line would not make
anything go red.

No run of this program with outbound network denied at the operating system level
has been made, and no check over the import graph refuses a package that reaches
the network from a code path outside the three features above. Issue #116 owes
both, and it is the test this document would otherwise name.

So the correct description of the transmission claim is that it is a commitment,
supported by two commands over the dependency set at one commit, and not a
verified property. No later edit of this document may soften that sentence.

## What this document is not about

It is about what this program does, not about the machine it runs on. That
machine has an operating system, very likely a backup agent, a file synchroniser
and a virus scanner, and any of them may move a project file somewhere without
consulting this program. An operator holding other people's building data has
that problem whatever tool they use, and this project neither solves it nor
pretends to.

It also makes no statement about the legal roles between the operator and their
client, which depend on the engagement rather than on the software. What it does
establish is that this project is not a recipient of the operator's project data,
because it receives none of it. Whether this project ever hosts anything is an
open maintainer decision, entry 3 of issue #1, and an answer that it does would
put a second position beside this one.

## What this document still owes

One thing, and it is the one that would make the transmission section evidence
rather than a commitment. Issue #115 asks that this document name the test that
proves nothing leaves the host. There is no such test in this tree, and issue
#116 is where it is built. Until it exists this document names the absence
instead, in the section above, and #115 stays open on that condition.

# Decision: nothing leaves the host

Status: decided. This is a birth requirement and it is a default rather than a
setting somebody is expected to find.

## The position, in one paragraph

This program runs on your computer and keeps your work there. When you install it
and use it normally it makes no connection to anything: it does not check for
updates, it does not send usage figures, it does not report crashes, it does not
fetch a font or a map or a picture, it does not ask a server whether your licence
is valid, and it does not upload your project. Everything it needs to compute a
result is already on the machine, and every result it produces stays on the
machine until you move it yourself. A few things this program could do would need
a connection, downloading a newer component database is one of them, and each of
those is switched off until you switch it on, tells you where it is connecting
to and what it is sending, and stays switched off for everybody else.

That paragraph is the promise. The rest of this document is what makes it true
and what would make it false.

## Why this is a design decision and not a privacy notice

A building acoustics project file names a building. It usually carries an
address, often a client, sometimes a floor plan, and sometimes the names of the
people who commissioned the work. Under European data protection law that is
personal data about people who are not the operator and who have no relationship
with this project at all. The operator is a consultancy acting for a client, and
every byte of that file which leaves the operator's machine turns a desk tool
into a processing arrangement with contracts, records and a legal basis behind
it.

The reason to fix this now is that the position is cheap to hold and expensive to
recover. A tool that has never sent anything has nothing to disclose, no
agreement to sign, no retention period to state and no transfer to justify. A
tool that sends one thing has all of that, forever, for one feature. There is no
proportionate version of the second state, so the first one is defended at the
boundary rather than argued about per feature.

## Every feature that may ever make an outbound connection

The list is exhaustive as of this decision, in the sense that a feature not on it
is not permitted to connect at all. Adding a feature that connects means adding
it here first, in the same change, and that is the mechanism the list exists for:
it turns "we should think about the network" into an edit that shows up in a
diff.

**Downloading a component database update.** Sends: the version of the database
currently held, and nothing else. Not the project, not a record identifier, not a
count of anything. Receives: the published database artefact and its checksum.
Turned on by: the operator running the update command explicitly. There is no
background check, no reminder, and no timer, so the program never contacts the
publication point except in the second after somebody asked it to.

**Submitting a component record back to this project.** Sends: the record the
operator chose to submit and the provenance it carries, which is a laboratory
value and a citation to a published test report. Turned on by: the operator
running the submit command and confirming what is being sent, with the exact
payload shown before it goes. Whether such submissions are accepted at all is
entry 6 of issue #1 and is not decided here; the entry is listed so that the
feature cannot arrive later without passing through this document.

**Fetching a published comparison exercise for the validation corpus.** Sends: an
ordinary request for a document at an address written in the tree. Turned on by:
a maintainer running the collection tooling. This one is not part of the shipped
program at all: it belongs to the harness that maintains the corpus, it runs on a
maintainer's machine, and the corpus itself is committed so that the ordinary
suite reads files rather than the network. It is on the list because it is the
one place in this repository where an outbound connection is legitimate, and a
list that only holds shipped features would leave it unaccounted for.

Nothing else. Specifically not, and each of these is a thing tools of this kind
routinely do: telemetry of any shape including anonymous counters, an update
check for the program itself, crash or error reporting, remote configuration,
feature flags, licence or entitlement checks, a documentation link that resolves
at run time, an analytics or crash library pulled in by a dependency, a font or
stylesheet fetched by a report renderer, and a map or geocoding lookup for a
building address. The last two are worth naming because they arrive as a
side effect of doing something else useful, which is how they get past a review
that is looking at the useful thing.

## None of them is on by default

No feature on that list is enabled by default, and none of them has a default
that could be changed by a configuration file, an environment variable, an
installer choice or a first-run wizard. The distinction matters: a feature that
is off by default but can be turned on by a file the operator did not write is
off by default in the release notes and on in the field.

What would have to change for one of them to be on by default is written here so
that the question has an answer before somebody asks it in a hurry. It would take
all three of the following, and no two of them are sufficient:

- A decision recorded on issue #1, by the maintainer, naming the feature and
  saying that it is on by default. This document is not the place that decision
  is made and a pull request is not either.
- A statement, in the operator-facing data protection documentation, of what is
  sent, to where, on what legal basis and for how long it is kept. The
  documentation is issue #115 and it is where the operator finds out, so a
  default that is not in it is a default nobody was told about.
- A route for the operator to turn it off that does not require them to have read
  this document, and that survives an upgrade.

Until all three exist, the answer to "can we make it the default" is no, and it
is no by construction rather than by argument.

## The dependency policy that keeps a connection from arriving by accident

The threat this section is about is not a contributor adding a telemetry call.
That is visible, it is one line, and any review catches it. The threat is a
library added for an unrelated and entirely good reason that contacts something
on its own, in an initialiser, on first use, or in a code path nobody exercised
during review. The reviewer of that change was reading the reason the library was
added, and there was nothing in the diff to read.

The policy has four parts and they are ordered by how much they cost.

**Have almost no dependencies.** This is the whole of the defence and everything
else is a fallback. This project's computation is arithmetic over small arrays of
numbers, its data is JSON, and its interface is a command line. The language's
standard library covers all three. A dependency is therefore an exception that
has to argue for itself against a maintained standard library, and the argument
has to say why the thing it does is not something this project should own.

**A new dependency is a decision with a written reason.** The change that adds
one carries, in its pull request body, what it does, what it replaces, its own
transitive dependency count, and whether it makes any outbound connection in any
code path. That last question is asked explicitly and answered in words, so an
unexamined dependency is visible as an unanswered question rather than as
silence. Where the answer is that it does connect, the dependency is refused
unless the feature it serves is on the list above.

**The transitive graph is pinned and reviewed.** Locking the dependency graph and
making the restore reproducible is issue #27, and a scan over the resolved graph
is part of it. Pinning does not stop a dependency from connecting; what it stops
is the graph changing under a build without anybody seeing it, which is the state
in which the review above would have been performed on a different set of
packages than the one that ships.

**A test proves the default rather than asserting it.** Proving that nothing
leaves the host unless the operator switches it on is issue #116, and this
document names it because a promise in a paragraph is worth what the thing
checking it is worth. Two shapes are worth having and they catch different
failures. A run of the program over a representative workload with outbound
network denied at the operating system level, which fails if anything tried and
catches a connection whatever brought it in. And a check over the transitive
import graph refusing a package that reaches the network from any code path
reachable outside the features listed above, which catches a dependency before it
ships rather than when the workload happens to reach it.

The ordinary test suite already forbids the network to itself, by
[testability.md](testability.md), and that is not the same guarantee. It says the
tests do not connect. It says nothing about the program, and treating one as
evidence for the other is exactly the substitution that must not be made.

## What is not claimed

This document is about what this program does. It is not a claim about the
machine it runs on, which has an operating system, a backup agent, a file
synchroniser and a virus scanner, any of which may send a project file somewhere
without asking this program's opinion. An operator handling other people's
building data has that problem whatever tool they use, and this project neither
solves it nor pretends to.

Nothing here has been measured yet, because there is nothing to measure: the
program prints its own version and exits. The list above is a design constraint
on work that has not been written, and the first measurement is issue #116. Until
then the correct description of the promise in this document is that it is a
commitment, not a verified property, and no later edit of this document may
soften that sentence into a claim that it was checked.

## What would reopen this decision

A maintainer decision on entry 3 of issue #1 that this project hosts something,
which would not change what the program does on the operator's machine but would
put a second position beside this one and require this document to say which
applies where. Or a legal obligation on the operator, rather than on this
project, that requires a connection, which would be a feature on the list like
any other and would still not be a default.

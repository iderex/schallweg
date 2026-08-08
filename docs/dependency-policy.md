# The dependency policy

The cheapest way to keep a dependency graph honest is to have almost none, and
that is the policy rather than a consequence of it.

## What this project is made of

At the commit this document lands on, the module declares no dependencies at all:

    go list -m all
    github.com/iderex/schallweg

That is not an accident of being early. The computation here is arithmetic over
small arrays of numbers, the data is JSON, and the interface is a command line.
The standard library covers all three, it is maintained by the people who
maintain the language, and it does not disappear. So a dependency is an exception
that has to argue for itself, and the argument has to say why the thing it does
is not something this project should own.

## What has to be true for a new dependency to be added

All of it, in the pull request that adds it, in words rather than by implication.

**What it does, and what it replaces.** If it replaces nothing, the change is
adding a capability rather than removing work, and that is a different argument
and usually a weaker one.

**Its own transitive graph, counted.** Not "a few". The number, with the command
that produced it. A library with one useful function and forty transitive modules
behind it is forty modules of review nobody will do.

**Whether it makes an outbound connection in any code path.** Asked explicitly
and answered in words, so an unexamined dependency shows up as an unanswered
question rather than as silence. Where the answer is yes, the dependency is
refused unless the feature it serves is on the list in
[decisions/nothing-leaves-the-host.md](decisions/nothing-leaves-the-host.md). That
is the threat this whole policy is shaped around: not a contributor adding
telemetry, which any review catches, but a library added for an entirely good
reason that contacts something on its own, in an initialiser or on first use,
where the reviewer of that change had nothing to read.

**Whether it is maintained, and what happens if it stops.** This project exists
because a calculation kernel that building acceptances depend on was orphaned by
its vendor. A dependency that goes the same way is the same problem one level
down, so the answer to "what if this stops" is part of the decision rather than a
thing to find out.

**Its licence, and whether it is compatible with the licence of this repository.**
That licence is an open maintainer decision on issue #1, entry 1, so today the
honest answer is that no dependency can be cleared on this point and the question
is recorded rather than answered.

A dependency added for a test is a dependency. It ships in the module graph, it
is resolved by every build, and the reviewer of a later change sees it as
something the project already accepted.

## What the tree does about it

Four things, on every pull request, on every push to the default branch, and
again every week.

**The graph is exactly what the tree declares.** The build runs with the module
in read-only mode, so a build that would have to change `go.mod` or `go.sum`
fails instead of updating them. `go mod verify` then checks that every module in
the module cache still hashes to what the lock records, which is the check that
notices a tampered or substituted module rather than a changed declaration.

**The declaration is exactly what the source needs.** `go mod tidy` is run and
the tree is refused if it changed anything. That catches the two drifts in
opposite directions: a dependency used in the source but not declared, and a
dependency declared but no longer used, which is the one that quietly stays in
the graph for years after the code that wanted it was deleted.

**The resolved graph is scanned for known vulnerabilities, transitive included.**
`govulncheck` reads the advisory database and the actual call graph, so it
reports what is reachable from this code rather than everything that is anywhere
in the module list. It also covers the standard library, which matters here more
than usual: with no dependencies, the standard library and the toolchain are the
entire supply chain. The scan runs weekly as well as on every change, because an
advisory published tomorrow is about code that was written yesterday and nothing
in this repository will have changed.

**The declaration is not read for its licence, and this is the gap.** No check
here reads the licence of any dependency, because the licence of this repository
is an open maintainer decision on issue #1 and there is nothing yet to judge
compatibility against. Until that entry is answered the question above is
recorded in the pull request and answered by nobody, which is a weaker position
than the other three and is stated here rather than left to be discovered.

Beside those, the dependency review on a pull request reads what a change adds
against the advisory database, and fails on any known vulnerability at any
severity.

## The lock file, and what it is on an empty graph

`go.sum` is this module's lock file. It records a cryptographic hash for every
module version in the transitive graph, and the build refuses a module whose
content does not match.

There is no `go.sum` in this tree:

    git ls-files go.sum | wc -l
    0

That is what an empty graph produces, and it is worth being exact about what it
does and does not mean. It means there is nothing to lock, so the property "the
build resolved the versions the tree declares" is held by the graph being empty
rather than by a file. It does not mean the mechanism is absent: read-only mode,
`go mod verify` and the tidy check are all in force today and all of them would
fail on the first dependency that arrived without its hashes. What is true is
that none of them has anything to refuse yet, and a green run today is therefore
weaker evidence than the same green run will be after the first dependency.

That the tampering refusal is real rather than assumed was measured on a branch
that added one dependency and then altered its recorded hash. What that run
showed is on issue #27.

## Pinning what runs, not only what is imported

An action used by a workflow and a tool installed by one are both inputs to a
build, and both are pinned. Every action is referenced by commit hash rather than
by tag, because a tag can be moved to point at different code and a hash cannot.
Every tool a workflow installs is installed at a named version rather than at
whatever is newest, for the same reason.

This costs something and the cost is stated: a pinned input does not receive
security fixes on its own, so pinning converts a silent update into a change
somebody has to make. That is the trade being taken deliberately, and the weekly
vulnerability scan is what keeps the pin from becoming a hiding place.

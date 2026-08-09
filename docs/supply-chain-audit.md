# The supply chain, end to end, and where it stops holding

Every piece of this chain has an issue behind it and most of them are closed. A
closed issue says a piece exists. It does not say the pieces meet, and a chain
holds at its joins. This document reads the whole of it in one pass and says
where it holds, where it holds for a weaker reason than it looks like, and where
it does not reach at all.

Read at commit `3c8d8a8`. Every command below was run there, in a checkout of
that commit, and nothing in this document was taken from a working tree.

## What comes into this repository from outside

Four kinds of thing, and they are pinned by four different mechanisms, which is
why an audit that checks one of them proves nothing about the rest.

Actions, which run with this repository's own credentials. Tools a step
downloads and executes. The language toolchain the runner installs. And the
module graph the code declares.

## The actions

Every one is pinned to a commit, with the version in a comment beside it:

    $ git grep -hoE 'uses: [^ ]+' -- .github/workflows | wc -l
    36
    $ git grep -hoE 'uses: [^ ]+@[0-9a-f]{40}' -- .github/workflows | wc -l
    36

The two counts being equal is the whole of the claim. What the distinct
references are is printed by the first command with `sort -u` rather than
written here, because a list in a document drifts against the workflows that
decide it.

What refuses an unpinned one is zizmor's `unpinned-uses` rule, in the gating
step of the workflow security analysis, which runs at `--min-severity=low` and
fails the job on any actionable finding.

That rule has not been demonstrated to bite here. No fixture holds it, the tool
is installed from the network on each run rather than being present on a
machine an audit can drive, and this repository's own standard for a check is
the one `pattern-scanning-proof` meets: a rule is held to a case it has to find.
So the correct statement is that a check is configured to refuse an unpinned
action and that nothing in this tree proves it does.

## The tools a step downloads and runs

Four, pinned three different ways, and the difference matters more than the
count.

Two are pinned by the SHA-256 of the exact asset, and the step that fetches them
refuses anything else before it runs it:

    $ git grep -nE '^\s+[A-Z_]*SHA256: ' -- .github/workflows
    .github/workflows/mutation-testing.yml:51:  MUTATOR_SHA256: "b02a42e47935f891c9a411d68c07e211c7082609e79c2435b67c85ee9658c538"
    .github/workflows/pattern-scanning.yml:64:  ANALYSER_SHA256: "40c21299eeddabf743b856daa843d24f9d4a027130671cd45b3b21776fd9ab26"

Both workflows record how the hash was obtained and both compare it with
`sha256sum --check --strict` rather than with a shell comparison, so a truncated
value cannot match a prefix. This is the strongest form in the tree and it is
the form the other two do not have.

One is pinned by version at a Python package index. The wheel for a released
version is not replaced there, so the tool itself is fixed by the version, and
`--no-build` stops a source distribution's build script from running at all.
What is not fixed is the resolution of that tool's own dependencies, which
happens at run time and can pick up a newer compatible release between two runs
of the same workflow file. That is the one place in this repository where a step
executes content that can change under a fixed reference.

One is pinned by module version and verified by the language's checksum
database. `go install` of a tool at a version resolves that tool's build list
from its own module file and verifies every module in it against the public
checksum database, so the bytes are fixed by the version even though no digest
appears in the workflow. It is weaker than a written digest in one specific way:
it trusts the checksum database rather than a value somebody in this repository
read and wrote down.

Nothing refuses an unpinned tool. All four pins are written by hand in the
workflow that uses them, and a fifth tool added with a bare version and no
checksum would pass every check here.

## The language toolchain

The version comes from the module file, in every job that runs a Go command:

    $ git grep -c 'go-version-file: go.mod' -- .github/workflows | wc -l
    6
    $ git grep -c 'GOTOOLCHAIN: local' -- .github/workflows | wc -l
    6

The module file names a language version rather than a patch release, so the
runner installs whichever patch of that line is current when the job runs, and
two runs a month apart can compile with different compilers. `GOTOOLCHAIN:
local` is set in the same six jobs, which stops a module graph from pulling a
different toolchain in mid-build; it does not fix which patch the runner
installed in the first place.

For a check that is closer to a feature than a defect: a newer compiler's vet
finds things the older one did not, and a check is supposed to move. For a
release it is not, because reproducing an artefact means using the compiler that
made it. There is no release route in this repository, so this is an open
question rather than an answered one, and it belongs to #125 and #126.

## The module graph

Refused three ways in the dependency job: the build runs with the graph the tree
declares and fails if it would change it, every module is verified against the
recorded hashes, and a declaration that drifted from the source fails.

The graph is empty:

    $ go list -m all
    github.com/iderex/schallweg
    $ git ls-files go.sum | wc -l
    0

So all three legs pass over nothing today. They are worth having for the day
there is something in the graph, and a green run of them now is not evidence
that they would refuse anything, in the same way and for the same reason the
build job already says of its own comparison.

## The bill of materials against the artefact

This is the condition most projects do not have, and it is met. The build job
writes the document for the program, then reads the modules recorded inside the
binary it built and fails if the document does not name one of them. The
comparison is between the command's own reader and the toolchain's printer,
which is a second route to the same fact rather than the same route twice.

Its evidence today is weak for the reason the workflow states itself: with an
empty module graph the comparison compares an empty list and passes without
having refused anything. The mechanism exists and it has not yet had anything to
find. Both halves are true and neither should be quoted without the other.

## Where the chain stops

At publication, because there is no publication.

Nothing is released from this repository. No artefact is published, so none is
signed, so no verification command can be documented, and the two conditions on
#107 that ask for those cannot be met by anything written here. #125 builds the
release route and #127 signs and publishes what it produces. Until both exist,
the honest summary of this audit is that the chain holds from the inputs to the
built artefact and has no end.

## What this audit examined, and what it did not

It read every workflow file, the module file, and the dependency policy, at the
commit named at the top.

It ran no workflow. Everything said above about what a job does at run time is
read from what the job says it does, and a workflow that behaves differently on
a runner than its text describes would not be caught by this.

It did not read the source of any action at the commit it is pinned to. Pinning
fixes what runs; it says nothing about what that code does, and this audit makes
no claim about any of it.

It did not judge any dependency's licence, and the dependency job says the same
of itself. There is nothing to judge compatibility against while the licence of
this repository is an open maintainer decision, entry 1 of #1.

It did not scan the actions or the downloaded tools for known vulnerabilities.
The vulnerability scan covers the module graph of this repository's own code and
nothing that arrives through a workflow.

It did not verify any signature, because nothing here publishes one.

It is a reading rather than a check. Nothing re-runs it, and a change that
weakens any of the pins above would not make it go red. What it is for is that
the next person to read this chain starts from what was found rather than from
the beginning.

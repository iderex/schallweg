# The reference gate, and every place this one differs

The gate this repository is measured against runs on the single sign-on plugin at
`github.com/Flowfin/jellyfin-plugin-sso`. It is public, its branch rule can be
read by anybody, and it was arrived at by somebody solving the same problem
rather than invented here.

Parity is not copying. That gate is for a plugin loaded into another program's
process, in another language, shipped through a package manifest, whose job is to
decide whether a stranger may log in. This is a calculation tool that runs on an
engineer's machine, produces numbers a building acceptance may rest on, and reads
only files its user chose. Some of that gate does not apply here, some applies
differently, and some of what this project needs is not there at all. Each of
those is a deviation and each owes a reason.

## What each branch rule requires

The reference, read at the commit this document lands on:

    $ gh api repos/Flowfin/jellyfin-plugin-sso/rulesets --jq '.[] | "\(.id) \(.name)"'
    18802863 Protect main and 5.0

    $ gh api repos/Flowfin/jellyfin-plugin-sso/rulesets/18802863 --jq '{enforcement, bypass: .bypass_actors, required: [.rules[] | select(.type=="required_status_checks") | .parameters.required_status_checks[].context]}'
    {"bypass":[],"enforcement":"active","required":["build","ABI floor build","Package (JPRM) / Build package","Package (JPRM) / Generate SBOM","CodeQL","Analyze (csharp)","DCO sign-off","Deterministic PR-hygiene checks","Enforce greppable invariants","Reject Trojan Source Unicode","Audit workflows (zizmor)","prettier","dependency-review"]}

This repository, at the same moment:

    $ gh api repos/iderex/schallweg/rulesets/20527526 --jq '{enforcement, bypass: .bypass_actors, types: [.rules[].type]}'
    {"bypass":[],"enforcement":"active","types":["deletion","non_fast_forward","pull_request"]}

**Thirteen required checks against none.** Both rules are active and neither has a
bypass actor, so what a merge is refused for is the whole of the distance: over
there, thirteen named check runs; here, that a pull request exists and that the
branch is neither deleted nor force-pushed. Asking for the required list is issue
#36, and until it lands every check this repository runs is a check a merge can be
made without.

That is the largest deviation in this document and it is not an argument about
scope. It is a setting nobody has changed yet.

## What this repository runs today

    $ ls .github/workflows
    build.yml
    codeql.yml
    dco.yml
    dependencies.yml
    dependency-review.yml
    format-and-vet.yml
    pattern-scanning.yml
    pr-hygiene.yml
    scorecard.yml
    test.yml
    text-rules.yml
    unicode-guard.yml
    zizmor.yml

    $ grep -h '^    name:' .github/workflows/*.yml | sort -u | sed 's/^ *name: //'
    Audit workflows (zizmor)
    build
    code-scanning
    code-scanning-suppressions
    DCO sign-off
    dependencies
    Deterministic PR-hygiene checks
    format-and-vet
    pattern-scanning
    pattern-scanning-proof
    PR-hygiene rules bite
    Reject Trojan Source Unicode
    Scorecard analysis
    test
    text-rules

## Element by element

Adopted means this repository runs the same check for the same reason. Adapted
means the reason survives and the subject changes. Dropped means the reason does
not reach this project at all.

### The required set

| Reference element | Here | Reasoning | Delivered by |
| --- | --- | --- | --- |
| `build` | Adopted | A tree that does not build is a tree no other check's result means anything about. | #20 |
| `ABI floor build` | **Dropped** | It exists because a plugin has to keep loading into host versions it does not control. Nothing here loads into anybody's process, so there is no floor to hold. | none |
| `Package (JPRM) / Build package` | Adapted | The reason is that the shipped thing is built by the gate rather than by hand. The subject is a signed release artefact rather than a plugin package. | #125, #126 |
| `Package (JPRM) / Generate SBOM` | Adopted | A bill of materials produced by the build rather than written by hand. It already runs here and covers the program and the database. | #28 |
| `CodeQL`, `Analyze (csharp)` | Adapted | Same analyser, same reason, the language it reads is Go. It also fails on any finding not deliberately accepted, and a suppression naming no rule and no reason is refused by a second check. | #25 |
| `DCO sign-off` | Adopted | Identical, down to the check-run name. | #33 |
| `Deterministic PR-hygiene checks` | Adapted | The reason is that some things are properties of the change rather than of the code. The rules differ: a record whose value changes without its provenance changing is the failure this database cannot afford, and there is no package manifest to keep fresh. | #103 |
| `Enforce greppable invariants` | Adapted | The fail-closed `git grep` shape is already here twice, in the text rules and in the suppression scan. Which invariants this project owes is its own list. | #102 |
| `Reject Trojan Source Unicode` | Adopted | Identical, down to the check-run name. Bidirectional and invisible control characters are a problem about what a reviewer is shown, and that is language-independent. | inherited scaffolding, comments rewritten by #32 |
| `Audit workflows (zizmor)` | Adopted | Identical. The workflows are the most privileged thing in either repository. | inherited scaffolding, comments rewritten by #32 |
| `prettier` | Adapted | A formatter for the language this project chose, plus vet and a documentation check, in one check run. | #24 |
| `dependency-review` | Adopted | Identical, and it now sits beside a lock file it can be read against. | #27 |

### Beyond the required set

The reference runs these on their own schedules without gating a merge.

| Reference element | Here | Reasoning | Delivered by |
| --- | --- | --- | --- |
| `Scorecard supply-chain security` | Adopted | Already running. | inherited scaffolding |
| `Repo Invariant Lint (Opengrep)` | Adopted | A second analyser with a different engine, because one analyser's blind spot is not a property anybody can enumerate in advance. The same engine, pinned by version and by the checksum of the asset, over rules of this repository's own. It runs on every pull request rather than on a schedule, as `pattern-scanning`, with `pattern-scanning-proof` beside it holding every rule to a fixture it has to find. | #101 |
| `Stryker mutation testing` | Adapted | The reason is stronger here than there. Most of this codebase is small pure functions over numbers, which is exactly where a surviving mutant means something, and a test that computes a value and asserts it is a number covers the line and proves nothing. | #105 |
| `Fuzz (SharpFuzz)` | Adapted | The subject is the parsers that take files from strangers, which here is the spectrum exchange format, the record reader and the project file. | #106 |
| `E2E Login Harness` | Adapted | The reason is that something has to exercise the whole thing the way a user does. There is no login; the equivalent is the worked example that runs from a clean checkout. | #97 |
| `Manifest freshness`, manifest regeneration | Adapted | There is no package manifest. The equivalent obligation is that the published database artefact matches the tree it was built from, and that a user is told what an upgrade means. | #80, #129 |
| The coverage bar, pinned to the modules that decide security outcomes | Adapted | The subject changes completely. There is no authentication surface here. The equivalent is the arithmetic that produces a number somebody may rely on. | #104 |
| `Wiki Lint` | **Dropped** | There is no wiki, and a documentation tree in the repository is read by the checks that already read the tree. | none |
| The publish and beta workflows | Adapted | The reason survives as a release route that signs what it publishes. The cadence does not: there is nothing to publish nightly until there is something to publish at all. | #125, #127, #131 |
| `.NET` | Adapted | The language's own build and test verbs, which here are the `build` and `test` checks. | #20, #22 |

The reference's coverage bar is pinned to a number in its own tree rather than to
a percentage in a workflow file:

    $ gh api repos/Flowfin/jellyfin-plugin-sso/contents/scripts/check-coverage.py --jq .content | base64 -d | grep -n 'SECURITY_LINE_BAR'
    68:SECURITY_LINE_BAR = 92.0

That shape is worth adopting whatever the number becomes, because a bar pinned to
the surfaces that decide the outcome says something a bar over the whole tree does
not. It is adopted: the number and the list of surfaces sit in
`cmd/gate/coverage.go`, beside the argument for both, and the check reports the
whole-tree figure without gating on it.

## What this repository adds that the reference has none of

| Element | Why it exists here | Delivered by |
| --- | --- | --- |
| `test`, held to four conditions: no display, no elevation, no network, no hardware | A suite that needs any of them is a suite somebody eventually stops running, and a validation corpus is worth what its reproducibility is worth. | #22 |
| The suite's own rules, read out of the source: no floating point equality, no clock, no randomness, no changed working directory, no fixture path leaving the package | A numerical project grows flaky tests in a particular way and it is not timing. All of these pass on the machine they were written on. | #37 |
| `text-rules` | Line endings recorded in the index rather than observed in a working tree, plus a byte-exact exemption that is a route rather than a hole. A fixture whose carriage returns are the point is a thing this project has and a plugin does not. | #30 |
| `dependencies`: a read-only build, `go mod verify`, `go mod tidy` refused if it changes anything, and a vulnerability scan over the resolved graph | The reference's dependency review reads what a change adds. This reads what the build resolved. | #27 |
| The architecture rules, read from the import graph and from the syntax | The kernel may not reach the file system or the network, nothing below the command line prints, and the decibel arithmetic lives in one place. A kernel determined only by its arguments is what makes a validation case mean anything. | #109 |
| The numerical regression gate over the validation corpus | The deviation upward with no counterpart at all: a plugin deciding whether somebody may log in has no numbers to regress. A refactor that moves a result by half a decibel has to be a deliberate act with an argument attached. | #111 |
| `code-scanning-suppressions` | Failing on every analyser finding is only livable where a false positive has a cheap and honest exit, and an exit that has to name the rule and the reason is the difference between accepting a finding and hiding it. | #25 |

## What would make this gate stronger rather than merely different

Three things, and only one of them is being pursued as a matter of course.

**The numerical regression gate, #111, is the real one.** Every check in the
reference set answers a question about the code. None of them answers whether the
number changed, because over there no number is the product. Here the number is
the product, and a check that refuses a moved result unless the move is recorded
with a reason is a control the reference has no analogue for and no need of. It is
planned and it is in the quality milestone.

**The coverage bar's subject, #104, is the second, and the number is now
decided.** Pinning a high bar to the surfaces that decide a number argued for a
higher bar than the reference's, on a smaller surface. The bar is 93.0 against
the reference's 92.0, the surface it applies to is the arithmetic rather than the
tree, and the reasoning for both is at the constant in `cmd/gate/coverage.go`
rather than restated here. What is still open is the surface's size: one package
carries it today, and the rating procedures, the in situ correction and the path
evaluation each join the list in the change that writes them.

**The third is the one this document must not claim.** A gate is stronger than
another gate when it refuses more of what matters, and this one currently refuses
nothing at merge time, because no status check is required. Everything above
describes checks that run and report. Until #36 lands, the correct statement about
this gate's strength relative to the reference is that it is weaker, by the whole
list, and every adopted and adapted row above is a check somebody can merge past.

## What this document does not do

It does not evaluate whether the reference gate is right. It is a standard this
project is measured against and it is treated as one.

It does not describe what any reference workflow does beyond its own declared
name and the check-run names in the branch rule above. Where a row reasons about
purpose, the purpose is this project's reading of it and not a statement about
somebody else's intent.

It records no measurement of either repository's checks passing. What each check
here refuses is argued in the pull request that added it and printed by the check
itself.

It goes stale the moment either branch rule changes, and nothing detects that.
Both commands are in it so that a reader can re-run them rather than trust the
output pasted here.

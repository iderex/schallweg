# Decision: the testability rule

Status: decided. This is a birth requirement, written before there is code to
break it.

## The rule

Every automated test in the ordinary suite of this repository runs on a machine
with no display attached, with no elevated privileges, with no network access,
and with no hardware beyond a general purpose computer. Those four conditions are
the whole rule. A test that cannot meet all four is not written as a test in the
ordinary suite: it goes into a separate harness that is named for the thing it
needs, and its results are never added to the ordinary suite's numbers. The
ordinary suite is the one a contributor runs on a laptop with the network cable
out, and it has to stay that way from the first test onwards, because a suite
acquires this kind of dependency one convenient test at a time and the loss is
only noticed when somebody else tries to contribute.

The four conditions, explicitly:

1. No display. The suite runs on a machine with no screen, no window server and
   no virtual framebuffer standing in for one.
2. No elevation. The suite runs as an ordinary user. Nothing in it installs a
   service, registers a scheduled task, writes outside the working tree and the
   temporary directory, or triggers an administrative consent prompt.
3. No network. The suite makes no outbound connection. Not a fast one, not a
   cached one, not one to a local mock that happens to listen on a port.
4. No special hardware. No measurement instrument, no dongle, no licence server,
   no accelerator, no specific processor feature.

## What that excludes

The list below is what this project can already see will not meet the rule. It
is not asserted to be complete, and a category discovered later is added here
rather than argued into the ordinary suite.

**Anything that drives a graphical interface.** Needs a display or a substitute
for one. Whether this project ever has a graphical interface is an open
maintainer decision on issue #1, entry 7, so this category may stay empty. It is
listed anyway, because the harness has to exist before the first such test rather
than after it.

**Anything that reads from a measurement instrument.** Needs the instrument, its
driver, and usually a calibration. Relevant here if this project ever imports
measured spectra directly from equipment.

**Anything that compares this implementation against another program.** Needs
that program installed, and the programs that implement this standard are
commercial, so it also needs a licence and in at least one case a hardware
dongle. This category is real and is planned work: comparing against a second
implementation where one is reachable is issue #87.

**Anything that fetches a published case from its source at run time.** Needs the
network. The corpus of published comparison exercises is collected once and
committed, so the ordinary validation tests read files in the tree. A test that
re-fetches a source to check it has not changed is useful and belongs in a
network harness, not in the ordinary suite.

**Anything that verifies behaviour requiring administrative rights.** Needs an
elevated account. Nothing in the planned design should need this, and the
category is listed so that a test which does need it has an obvious destination
other than the ordinary suite.

## The harnesses and how they are named

One harness per category above. The convention is that the name states the
requirement, not the intent:

- `requires-display`
- `requires-instrument`
- `requires-third-party-tool`
- `requires-network`
- `requires-elevation`

The reason the name says what it needs is that a contributor reading the name has
to be able to tell, without opening anything, whether their machine can run it. A
harness called integration, end-to-end or slow tells the reader nothing about the
four conditions, so the reader tries it, it fails, and the failure looks like a
defect in the code rather than a missing prerequisite. `requires-third-party-tool`
carries one further fact in its name: what it needs cannot be obtained by
installing something free, so nobody spends an afternoon trying.

Each harness carries, beside it, a statement of exactly what has to be present
for it to run and how a contributor would know they have it.

## Preventing the ordinary suite from acquiring a dependency

Three mechanisms, in order of how early they catch it.

The harnesses are compiled out of the ordinary run. A harness lives behind a
build constraint named for the harness, so the ordinary command does not build
those files at all. A test moved into the ordinary suite by accident therefore
has to be moved deliberately, by deleting the constraint, which is a visible line
in a diff rather than an invisible change of behaviour.

The gate job denies what the rule denies. The unit test check, issue #22, runs
the ordinary suite with no display, as an unprivileged user, and with outbound
network denied or, where the runner cannot deny it, with an outbound connection
during the suite failing the job. A test that acquires one of the four
dependencies fails there rather than being skipped there, and that difference is
the point: a skip would let the dependency in quietly.

The suite refuses to be empty. The same check fails when zero tests ran, which is
the failure that looks exactly like success and is how a broken discovery turns
the whole gate into decoration.

What would detect a slow drift rather than a single bad test is the first
mechanism combined with review: the only way into the ordinary suite is a diff
that deletes a build constraint, and that line is short and unmistakable.

## Reporting

A run says what it examined and what it did not.

The ordinary suite's output names every harness that exists, states that it was
not run, and states what running it would need. A run that covered less than
everything must not read like one that covered everything and found nothing.

A skipped harness is never counted as a pass. Counts are reported per harness and
are never summed into one number across harnesses, because a single total is
exactly the shape that lets an unrun harness disappear into a green figure.

Where a harness was run, its output says on what and with what present, so that a
later reader can tell whether the third-party tool it compared against was the
version the record claims.

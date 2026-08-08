# Security policy

## Reporting

Report privately, through this repository's own private reporting route:
[open a security advisory](https://github.com/iderex/schallweg/security/advisories/new).
It is enabled, it goes to the maintainer and to nobody else, and it does not
become public until an advisory is published.

Do not open a public issue for a security report. A calculation tool is installed
inside consultancies and public bodies, and a report describing how to make one
of those installations execute something is a report that should not be readable
before there is a fix.

Expect an acknowledgement within seven days. If none arrives, the report has not
been seen: open a public issue saying that a security report is waiting, with no
detail about it in the issue, and it will be picked up.

## Supported versions

None yet, and that is the honest answer rather than a table with one row in it:

    git tag | wc -l
    0

There is no release, so there is no released version to support. What is
supported today is the default branch, and a fix lands there. When the first
release exists this section says which versions get fixes and for how long, and
that is part of the release work rather than a promise made here in advance.

## What is in scope

The surfaces where this project reads something a stranger produced, or produces
something somebody else trusts:

- The file parsers. Anything that reads a project file, a spectrum in an exchange
  format, or a certificate extraction.
- The component database reader. The records are data in the tree, edited by
  people and read by the program.
- The release route and its artefacts, including the checksums and signatures a
  user checks a download against, and the bill of materials that says what is
  inside it.
- The workflows in this repository, which run with tokens and can be made to run
  what a change puts in front of them.

If you are not sure whether something is in scope, report it. Deciding is not
your job and getting it wrong in the cautious direction costs nothing.

## What is out of scope

- A wrong number. See below, because it is the common case and it matters that
  the difference is clear.
- A denial of service caused by handing the program a very large input. This is a
  desk-side tool that runs on a file the operator chose. A calculation that takes
  a long time on a huge project file is a defect worth fixing and it is not a
  vulnerability.
- Anything about a hosted service. There is none. Nothing in this project sends
  anything anywhere unless an operator switches it on, and that rule is
  [its own decision record](docs/decisions/nothing-leaves-the-host.md). A report
  that this project contacts something on its own is very much in scope, and is
  the opposite of this entry.
- A finding produced by a scanner, pasted in with no reading of what the code
  does. Say what an attacker gets and how.

## A wrong result is not a vulnerability

The most common report a calculation tool receives is that a number is wrong. It
is the most valuable kind of defect report this project can get and it is not a
security report.

A wrong number is a defect. Open a public issue for it, with the inputs, the
number you got, the number you expected and where your expectation comes from. A
wrong number is only fixable in the open, because fixing it means somebody else
checking the arithmetic against a source.

A vulnerability is when the program does something other than compute: reads or
writes a file it was not given, runs something, leaks a credential, or lets one
input decide what happens to another user's data. That goes through the private
route above.

The boundary case, stated so it does not have to be guessed: a wrong number that
a chosen input produces on purpose, in order to make an acceptance pass that
should have failed, is still a wrong number. It is reported in the open, at the
top of the queue, because everybody using the tool needs to know about it and the
fix is arithmetic rather than a patch to withhold.

## What happens after a fix

The fix lands with a test that fails without it. The advisory is published and
says what the failure was, what it allowed, and which versions carry the fix. A
reporter who wants credit gets it and a reporter who wants none is not named.

# Contributing

## Before anything else

```
go run ./cmd/gate
```

That is the whole local gate, and it is one command rather than a list. Its legs
run in order and it stops at the first failure. If it does not pass, nothing else
matters yet.

The run says what it examined. A leg that this command does not ask for prints
that it was not asked for and what asking would cost, so a run that covered less
than everything cannot be read as one that covered everything and found nothing.

What the legs are is printed by the command and is not listed here. A list in a
document drifts against the thing it describes, and this document would be the
worst place for that drift, because it is the one somebody follows exactly and
then meets a red check naming something no document mentioned.

A green run here is not a green run on the server, and the command says so at the
end. Some checks read this repository through the hosting service and have no
local equivalent.

## No work without an issue

Every change starts as an issue and lands as a pull request.

An issue says what is wrong, what the evidence is, and what done means. If the
evidence is a number, it carries the command that produced it.

Nothing in this repository reads any of that. There is no check that refuses an
issue with no evidence or a pull request body with none, so this section is a
rule a person applies.

## Every asserted fact carries the command that produced it

This matters more here than in most projects, because the subject is numbers. A
sentence saying a value is right, a deviation is small, or a check passes is
worth what the command behind it is worth, and the command belongs beside the
claim, run at the commit being pushed rather than in a working tree nobody else
has.

Where a claim cannot be backed by a command, write it as a claim and say so.
"Verified", "not measured" and "not evaluated on this route" are different words
for different things and they are not interchangeable.

## Sign off every commit

Every commit needs a `Signed-off-by` line matching its author. It is how you
assert the Developer Certificate of Origin, and the sign-off check refuses a pull
request where any commit lacks one.

```
git commit -s
```

If you have already committed without it:

```
git rebase --signoff <base>
```

The line has to match the author name and email of the commit it is on, so
`git commit -s` after changing your git identity produces a line that does not
match and is refused.

## Text in this repository

Tracked text is UTF-8 with LF line endings, in the repository and in every
working tree including Windows. `.gitattributes` states the rule and git applies
it when you stage a file, so you do not have to do anything about it.

### Adding a fixture whose bytes are the point

Some fixtures exist to carry exactly the bytes a parser has to cope with: a
carriage return, an unusual encoding, a missing final newline. If those bytes are
normalised on the way into the repository, the fixture stops testing what it was
written for.

Put such a file under a `testdata/byte-exact/` directory. `.gitattributes` marks
everything there `-text`, so nothing rewrites it in either direction. Two things
follow and both are your problem to weigh:

- The file reviews as a binary blob. A reviewer cannot read the change, so keep
  it small and say in the pull request body what its bytes are and why.
- The exemption is only recognised in that directory. A `-text` file anywhere
  else fails the text rules check, which is deliberate: it keeps the exemption a
  route rather than a hole.

There is one such fixture in the tree already, and reading it is the fastest way
to see the shape.

## What a pull request body carries

The body is where the change is argued. It carries what changed and what failure
it prevents, the commands behind every number in it, and what the change does not
do.

It also carries one sentence naming the means the change is made of, the
language, the format or the tool, and why that fits what this repository has to
do. That question is asked every time rather than carried over from the last
change.

A negative disclosure stays negative. If something was not run, not measured or
not checked, say so, and let the admission survive every later edit of the body.

## What a commit message carries

What changed and what failure it prevents. Where a correction is being made, what
was wrong and how it was found. One topic per commit and one per pull request.

## No copy of the standard, and what you can still do

Most people reading this do not have EN ISO 12354 on their desk. It is sold
rather than published, and this repository will never contain its text, its
tables or its figures.

That excludes less than it sounds like. The work that does not need the text
includes: the component database and everything that reads, validates and
corrects it; the command line, its output formats and its error messages; the
documentation and the worked examples; the tooling, the gate and the fixtures;
collecting published comparison exercises, which are papers rather than the
standard; and reviewing any of the above.

What does need it is the arithmetic of the kernel itself and the wording of what
each symbol means. If you want to work on that and do not have access, say so on
the issue rather than reconstructing the method from a secondary source, because
a wrong reconstruction is expensive to find later.

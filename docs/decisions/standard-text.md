# Decision: what may be written here from the standard

Status: decided. This is a birth requirement. It is written before the first
formula is committed, because the cheapest hour to get it wrong is the first one
and the most expensive correction is the one that has to walk back through every
file that was written under the wrong rule.

## The line

The method may be implemented, described, referenced and reviewed here. The
document that describes it may not be reproduced here.

That sentence is the whole decision and everything below is its application.
A method is a way of getting from inputs to a number. It is not owned by the body
that wrote it down, and implementing it is the intended and lawful use of a
published standard. The document is a work: its wording, its tables, its figures
and its arrangement belong to whoever produced and sells it, and none of them
comes into this repository.

Both directions of this cost something, and both costs are real. Reproducing
normative text or a table of reference values would infringe, and it would also
make this repository undistributable under any licence, which would take the
project down with it. Refusing to reference the standard at all would make the
code unreviewable, because a reader who cannot see which clause a function claims
to implement cannot check it against anything. The line above is where those two
costs are smallest at once.

## Three tiers for a number

Every number this project needs comes from one of three places, and which tier a
number is in decides what happens to it. The tier is a property of the number,
not of how convenient it would be.

**Derivable.** The value follows from a rule that can be written in a sentence.
It is computed in the source from that rule, and the rule is written beside it.
Nothing is transcribed, so nothing is copied, and the code carries a statement a
reviewer can check rather than a list they can only trust. Third-octave band
centre frequencies are the standing example: they are a geometric series and the
naming convention that rounds them for display is itself a stated rule, so the
tree computes them and states the rule instead of pasting a column of numbers.

**Not derivable, independently published.** The value is a fact this project
cannot compute, and it appears in a source that is free to read and free to cite:
a national regulation, a public authority document, a paper, a manufacturer's own
published certificate. It ships as a data file, and every value in that file
carries the source it came from, individually rather than as a preamble. The
source is what makes the value reviewable, and a value whose source is the
convenience of whoever typed it is not reviewable at all.

**Not derivable, only in the standard.** The value exists as a table printed in a
document that is sold. It is not in this repository, in any form, in any layout,
under any name. The program that needs it reads it from a data file the operator
supplies, and where that file is absent the program refuses to compute the
quantity and says exactly which file it wanted, which clause of which part the
values come from, and what the operator has to do. It does not guess, it does not
substitute a default, and it does not compute a nearby quantity and present it as
the one that was asked for.

The third tier has a cost and it is stated here rather than discovered by the
first user. A fresh installation of this program, with no data supplied, cannot
produce the weighted single-number ratings, because their reference curves are
tier three. That is a worse first experience than a tool that pastes sixteen
numbers into a source file and says nothing. It is the price of a repository that
can be distributed under a licence, and it is paid deliberately.

Nothing in the tiers is a judgement about whether a short table of numbers is
protected in any particular jurisdiction. Reasonable lawyers differ on that, this
project has not asked one, and a project that depends on winning that argument
has a licence position that only holds while nobody tests it. The tiers are built
so the argument never has to be had.

## Clause references, symbol names, quoted text

Three separate rules, because they sit at different distances from the document.

A clause reference is a pointer and a pointer is not a copy. Reference the
standard by its number, its part, its edition year and the clause or annex, and
say in this project's own words what that clause is being relied on for. This is
required rather than merely permitted: a function that computes a quantity the
standard defines and does not say which clause it claims to implement cannot be
reviewed against anything, and it cannot be checked when the standard is revised.

Allowed:

    // Computes the apparent sound reduction index by summing path
    // contributions, following EN ISO 12354-1:2017, clause 4.1.

Not allowed, because the pointer has been replaced by the thing it points at:

    // Computes the apparent sound reduction index. The standard defines this
    // in clause 4.1 as follows: "<the clause's own sentences>".

A symbol name is vocabulary and vocabulary is how a field talks to itself.
Symbols and quantity names from the standard are used freely in code, in
identifiers, in documentation and in output. `R_w`, `D_nT_w`, `L_n_w`, `K_ij`,
`Delta_R_w` and the rest are what the reader already knows, and inventing private
names for them would make this implementation harder to check while protecting
nobody. Where the code writes a symbol, it also writes what this project means by
it, in this project's own sentence rather than the standard's.

Allowed:

    // K_ij is the vibration reduction index for the path from element i to
    // element j across a junction: the direction-averaged drop in vibration
    // level across that junction, normalised so that it does not depend on
    // the size of the elements or of the junction.

Not allowed, because the definition has stopped being this project's sentence:

    // K_ij: "<the standard's own definition, transcribed>".

Quoted text is the one that has no safe general rule, so it has a narrow one. No
normative sentence of the standard is quoted in this repository, on the issue
tracker, in a commit message, in a pull request body or in a wiki page. Not one
sentence, not under a fair-dealing or quotation exception, not with an
attribution attached. The exception those arguments rely on is real in most
European jurisdictions and this project still does not use it, for two reasons.
It is a judgement that has to be made per quotation and nobody here is qualified
to make it fifty times. And a repository containing no quoted normative text at
all is a claim that can be checked by looking, while a repository containing some
is a claim about every one of them.

What replaces a quotation is a reference plus this project's own description, and
where a description would be so close to the original that it is a paraphrase in
name only, the answer is a shorter description and a clause reference, not a
longer one.

The same rule applies to figures and to tables of reference values, and it needs
no separate wording: a figure is not derivable and is not independently
published, so it is tier three, and so is a table the standard prints.

This rule covers every surface this project writes on, not only the tree. The
issue tracker is public, the commit log is public, and a pull request body is
public. A quotation is no less a copy for sitting in a comment box, and a
repository whose tree is clean and whose issues are not has the same problem with
a longer path to it.

## Whether a contributor needs a copy

No, for most of the work, and this is a deliberate design of the project rather
than an accident of who turned up.

The work that needs a copy is the arithmetic of the kernel and the wording of
what each symbol means: writing the path equations, deciding what a quantity is,
and reviewing either. A contributor who does not hold the standard should say so
on the issue rather than reconstructing the method from a textbook, a lecture
slide or another implementation. A reconstruction that is wrong is expensive to
find, because it produces a number rather than an error, and the number is
plausible.

The work that does not need a copy is most of the project, and it is listed in
the contribution guide rather than restated here so that the two do not drift.
It covers the component database and everything that reads, validates and
corrects it, the command line and its formats and its error messages, the
documentation and the worked examples, the tooling and the gate and the fixtures,
collecting published comparison exercises, and reviewing any of that.

A consequence worth stating: because tier three values are supplied rather than
shipped, a contributor without a copy can write the code that reads a reference
dataset, the schema it is validated against, and the refusal that fires when it
is missing, without ever seeing the numbers. That is not a workaround. It is the
tier boundary doing what it was drawn to do.

## What downstream work owes this decision

An issue whose done-condition would put a tier three value into this repository
is planned wrong, and the repair is to the issue rather than to this rule. What
it becomes instead is an issue that defines the dataset, its schema, its
validation and the refusal, and leaves the values to the operator.

Where such an issue names this document, a reader can see the constraint before
the work starts. Where it does not, the constraint still applies, because a rule
that only binds the issues that remembered to cite it is not a rule. This
document is the authority either way.

## Whether any of this is machine-checkable

Mostly not, and this section says so rather than implying a gate that does not
exist.

What is checkable, and is planned rather than present: every value in a shipped
data file carries a source, and a value whose source is absent or names the
standard is refused. That is an ordinary schema and provenance check over files
this project controls, it belongs with the provenance work in the database
milestone, and it is the one part of this decision a machine can decide, because
it reduces to a field being present and not matching a pattern.

What is not checkable is everything that matters most. No check can tell a
paraphrase from a quotation, recognise a table of reference values pasted into a
Go source file as a slice of numbers, or judge whether a comment has stopped
being this project's own sentence. A scan for a suspicious block of numeric
literals would fire on every fixture and every test in a numerical project, which
is a check that gets switched off in a week.

So this rule is held by review and by this document, and the residual is that a
determined or careless contributor can break it and no run will go red. The way
that residual gets smaller is the provenance check above plus the habit of asking
where a number came from in every review, and the way it gets larger is a hurry.

## What would reopen this decision

An answer to entry 1 of issue #1 that chooses a licence with a compatibility
constraint this document has not accounted for. A change in what the publishing
bodies allow, stated by them rather than inferred. Or a tier three dataset
becoming independently published, which moves those values to tier two and is a
change to the data rather than to the rule.

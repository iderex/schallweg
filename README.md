# schallweg

EN 12354 is the European calculation standard for airborne and impact sound insulation and the basis of acoustic compliance across most of Europe, including flanking transmission, which in practice decides whether a building meets its requirement. Every tool for it is commercial, and BASTIAN, in use in the Nordic countries since 2000 and still used by many consultancies, is no longer maintained by its vendor: a calculation kernel that building acceptances depend on is orphaned and nobody can adopt it. Two halves, both doable at a desk: the kernel is algebraic SEA relations validatable against published round-robin trials, and the data half structures manufacturers published ISO 10140 test certificates, which today sit behind subscription databases.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is written down there with its reasons before the code
that depends on it exists.

See [NOTICE.md](NOTICE.md) for the intended-use notice.

## The standard, and what is not here

This project implements a published method. It does not reproduce the document
that describes it, and it never will.

A method is a way of getting from inputs to a number, and implementing one is the
intended use of a published standard. The document is a different thing: its
wording, its tables, its figures and its arrangement belong to the bodies that
produce and sell it. None of them is here, in any form, under any name. Where the
code needs to say which part of the method it implements, it cites the clause and
then says in this project's own words what it is relying on that clause for. A
citation is a pointer, not a copy.

Some values the method uses exist only as tables printed in the sold document.
Those are not in this repository either. The program reads them from a file the
operator supplies, and where that file is absent it refuses to compute the
quantity and says which file it wanted and what the operator has to do. The cost
is real and is not hidden: a fresh installation with nothing supplied cannot
produce the weighted single-number ratings.

This project is not a substitute for the standard. If you need the normative text
you buy it from your national standards body, and anyone reviewing or certifying
work done with this tool needs the document itself.

The full decision, including the three tiers a number can fall into and what is
allowed in a comment, is [docs/decisions/standard-text.md](docs/decisions/standard-text.md).
This section and that document say the same thing, and where they ever differ,
that document is the one that is right and this section is what has to be
corrected.

### If you think something here should not be

If you believe this repository contains text, a table or a figure from a standard,
or anything else somebody holds rights in, report it and it will be dealt with
rather than argued about.

Open an issue if you are content for the report to be public, or use the private
route in [SECURITY.md](SECURITY.md) if you would rather it were not. Either
reaches the maintainer.

What happens then: the material is taken out of the default branch first and
assessed afterwards, because leaving it in place while the argument runs is the
one option with a real cost to somebody else. You get an answer saying what was
removed, and what replaced it, since something usually can replace it. Where the
report turns out to be about something this project may lawfully do, the answer
says that, in public, and the material comes back with the reasoning beside it.
Removal first is not an admission and the record says so either way.

# Decision: what may be taken from a published test certificate

Status: decided. Written before the first record is entered, because the
correction is one record's worth of work today and a thousand records' worth
after the database has been seeded.

## The line

A measured value is a fact about a physical object. The certificate is a
document. The values are extracted; the document is not reproduced.

That distinction is not this project's invention and it is not a clever reading.
A laboratory measured a wall and wrote down what it found. What it found is what
happened, and what happened is not authored. The report that says so is authored:
its sentences, its layout, its figures, its photographs and the way it arranges
what it presents all belong to whoever produced it.

Everything below is the application of that line to the cases that actually
arise, and the cases that arise are all in the middle.

## What is extracted

The quantities and the facts needed to use them, and nothing that is there to
describe rather than to measure:

- The measured spectra and single-number quantities, band by band.
- The physical description of the specimen reduced to the fields this project's
  element model defines: mass per unit area, thickness, layer composition,
  dimensions of the tested specimen, and the mounting and loss factor conditions
  the measurement was made under. Reduced to fields, which is the point of the
  next section.
- The identifying facts about the test itself, which are the provenance below.

That is the whole of it. A record is a structured object whose fields were
decided by this project, filled in from a document. It is not a transcription of
that document into a different file format, and the difference between those two
things is visible: a transcription would carry whatever the certificate happened
to say, and a record carries only what the schema asked for.

## What is never reproduced

The certificate's own sentences, in any quantity. Its layout, its tables as
tables, its figures, its diagrams, its photographs and its logotypes.

Descriptive text is the case worth stating separately, because it is the one
somebody will want to copy in good faith. A certificate usually describes the
construction in a paragraph: what the leaves are, what sits between them, how it
was fixed. That paragraph is useful, it is the fastest way to convey what was
tested, and it is written by the laboratory.

It is reduced to structured fields, not reproduced and not paraphrased. Reduced
means the paragraph is read and the fields it implies are filled in, so what
lands in the record is a layer list with materials and thicknesses rather than a
sentence about them. Paraphrase is refused as a separate route, for two reasons.
A close paraphrase of a short factual description is not reliably outside the
original's protection, and it is worse than useless technically: it carries the
same words in a different order and is still not searchable, comparable or
checkable, so it takes the legal risk and returns none of the benefit that made
structuring the data worth doing.

Where the construction cannot be reduced to the fields the model has, that is a
finding about the element model rather than a licence to paste the paragraph. It
goes on the issue for the model, and until it is answered the record is not
entered.

## The provenance a record carries

The fields below are what the extraction rule above forces, and each is here
because a record without it cannot be checked or is not lawfully extractable.
This document is not the schema. The schema is issue #73 and what provenance has
to carry as a data model is issue #74, and this section states the floor those
have to reach rather than the shape they take.

- The laboratory that performed the test, by its name. Necessary rather than
  useful: a measured value has no meaning apart from who measured it, and a
  reader deciding whether to rely on the number is in part deciding whether to
  rely on the laboratory.
- The report or certificate number and its date. Necessary because it is the only
  way anybody can go and read the original. A record whose value cannot be traced
  back to a specific document is an assertion by whoever typed it, and this
  project would then be exactly the unauditable source it exists to replace.
- The standard the test was performed to, with its part and edition. Necessary
  because the same symbol means different things under different editions and
  different test methods, and a value compared across editions without knowing it
  is a silent error.
- The client the test was performed for, usually the manufacturer, and the
  product designation as the manufacturer names it. Necessary because a value is
  a value of a product somebody sells, and a user who cannot map the record onto
  something they can buy cannot use it.
- Who entered the record and when. Necessary because entry from a document is
  the step where a mistake is invisible: the number is plausible, nothing refuses
  it, and the only route back is to the person who read the page.
- Where the certificate was obtained, as a citation to the published location.
  Necessary because it is the difference between a record extracted from a
  published document and a record whose origin nobody can state, and the second
  kind is the one that turns into a removal request.

Every one of those is an identifying fact about a document, which is what a
citation is, and a citation is not a reproduction. That is why the list stops
where it does. It carries enough to find the original and nothing that would let
somebody reconstruct it.

## Whether a copy of a certificate is stored here

No. No certificate document is stored in this repository, in any form: not the
original file, not a rendering of it, not an extract of its pages, not an image
of a table from it.

The convenience being given up is real and worth naming, because somebody will
propose this again. Holding the document would make a record checkable without
leaving the repository, it would survive the manufacturer taking the page down,
and it would make review of a data correction a matter of opening one file. All
true, and all of it is redistributing somebody else's work, which is the one
thing the line at the top of this document says not to do.

What replaces it is the citation. A reviewer checking a record follows the
citation to the published document and reads it there.

The cost of that, stated: published certificates move and disappear, so a
citation will sometimes not resolve, and then a record's value cannot be
re-checked against anything. That is a worse position than holding the document
and it is accepted. What softens it is that the provenance still names the
laboratory, the report number and the date, which is enough to ask the
laboratory or the manufacturer directly. What does not soften it is a private
archive kept somewhere else and referred to from here, which would be the same
act with an extra step.

Applied rather than merely stated. The tree holds no certificate document today:

    git ls-files | grep -icE '\.(pdf|docx?|jpe?g|png|tiff?)$'
    0

That command is what the sentence rests on. It is a statement about the tracked
tree at the commit it was run on, not a guarantee about later ones, and what
would make it a standing property is the data validation work in issue #38
refusing a file the schema does not claim.

## A removal request

Somebody will ask for a record to be removed. Sometimes it will be a
manufacturer who does not want a value in public, sometimes a laboratory with a
view about its report, and occasionally there will be a real defect behind the
request. The procedure below is written now so that the answer is not improvised
under pressure, and the order of the steps is the point: the technical question
is answered before the position is taken.

1. The request is recorded as an issue, with what was asked and by whom, and the
   record it names is identified.
2. The record is checked against its own citation. If the value was entered
   wrongly, that is a correction and it happens immediately, by the superseding
   route in [data-format.md](data-format.md), whatever else is decided. A wrong
   number is removed because it is wrong, not because somebody asked.
3. If the value is correct, the position is that a measured fact with a citation
   is not a thing this project is obliged to withdraw, and this project says so
   plainly rather than complying by default. A request that comes with a legal
   basis stated is answered on that basis and with advice; a request that comes
   without one is answered with this paragraph.
4. Where a record is withdrawn, it is withdrawn visibly. The record identity
   remains, it is marked withdrawn, and it carries the date and the reason in the
   same append-only history that carries superseded values. It is not deleted
   from the tree and it is not deleted from history, because a database that can
   quietly lose a record is a database whose absences mean nothing.

What happens to results already computed from a withdrawn record. Nothing, and
that is a design property rather than a difficulty. A result carries every input
value it used, its origin, and the record it came from, by
[result-contents.md](result-contents.md), so a result computed last year is still
readable and still explains itself after the record is gone. What a later reader
gains is that the record identity now resolves to a withdrawn entry with a date
and a reason, so they can tell that the input behind an old number is no longer
current and why. A result is a document on the operator's machine and this
project has no route to it, which is the correct consequence of
[nothing-leaves-the-host.md](nothing-leaves-the-host.md) and is stated here so
that nobody promises a recall this project cannot perform.

## Whether this project extracts from any existing database

It does not, and this is the sentence a reader should be able to point at.

There is a real legal difference between reading individual published
certificates and systematically extracting a substantial part of somebody else's
collection. The second can infringe a database right even where every individual
value in it is a fact that nobody owns, because what is protected is the
investment in assembling the collection rather than the contents. Subscription
databases for this field exist and are named in [../gap.md](../gap.md), and they
are exactly the kind of collection that right is for.

So the rule is that a record enters this database from a published certificate
that this project read, and the citation names that certificate. A record does
not enter by being copied out of another database, and it does not enter by being
copied out of another database and then re-cited to the underlying certificate,
which is the same act with the evidence removed. Where a value is known only
because a subscription database published it, it is not entered here.

The cost is the whole shape of the data half of this project: entry is manual,
slow, and bounded by what laboratories and manufacturers actually publish, and
this database will be smaller than the commercial ones for a long time. That is
the position, and the alternative is a collection that cannot be defended and a
project that cannot say what it is.

The tree currently contains no records at all, so nothing above has been applied
to anything yet:

    git ls-files 'data/**' | wc -l
    0

Which record set the database starts with is issue #81, and it is where this rule
first has to hold against the temptation to fill a database quickly.

## What would reopen this decision

An answer to entry 2 of issue #1 that puts the database under a licence with an
attribution requirement the provenance list above does not satisfy. A
manufacturer or laboratory publishing certificates under terms that permit
redistribution, which would move their documents out of the section above for
those documents only. Or a removal request that produces a legal argument this
document's step 3 has no answer to, which is a thing to record in the issue and
answer once, not to settle by editing this file quietly.

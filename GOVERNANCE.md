# Governance

Who decides what here, and how to argue with a decision. It is written before the
first disagreement rather than during one, which is the only time it can be
written fairly.

## The shape today

There is one maintainer, and that is the whole of it. Saying so is more useful
than a structure with empty seats in it: a document describing a committee that
does not exist tells a contributor nothing about who will answer them.

Everything below is written so that it still reads correctly when there is more
than one maintainer, and the one thing that would have to change is the sentence
above.

## How a decision gets made

A change starts as an issue and lands as a pull request. That is in
[CONTRIBUTING.md](CONTRIBUTING.md) and it is the ordinary route for anything that
touches the tree.

A decision that shapes what comes after it is written down before the code that
depends on it, as a record in `docs/decisions/`. Those records are the authority.
Where a record and any other document disagree, the record is right and the other
document is what has to be corrected, which several of them say about themselves
in their own words.

Some decisions are the maintainer's and are parked rather than taken. They are
entries in issue #1, each one stating the options and what each costs and none
stating a recommendation. An entry is answered by a comment on that issue naming
the option chosen. Work blocked on an entry names the entry and stops there
rather than assuming an answer, and nothing else in the plan quietly decides one
of them.

## How to argue with a decision

In the open, on the issue or the record that carries it.

Where the decision is in a record, argue with the record. The records state their
reasoning and most of them state what would reopen them, which is the shortest
route: show that the reopening condition has been met and the argument is already
half made.

Where a pull request is refused, the reason is in the body of that pull request
rather than underneath it, and that is where the disagreement continues. If the
reason turns out to be wrong, the body says so afterwards. A refusal that a
reader cannot find the reason for is itself a defect worth reporting.

There is no appeal to anybody else, because there is nobody else. What there is
instead is a public record: every decision, its reasoning and every argument
against it stay readable, and a decision that cannot survive being read is one
worth reopening.

## Who decides whether a record is accurate

This is not a code review question and it is not settled by the same instincts. A
disputed component record is a technical judgement about somebody else's
measurement, quite often somebody who sells the product being measured, and it
needs a route that is specific and that does not become personal.

The rule the route rests on: a value in the component database is accurate if it
traces to a published test certificate, and the certificate is the evidence
rather than anyone's opinion of the number. That is the whole standard, and it is
what makes a disagreement about a record answerable rather than a matter of
authority.

How to dispute one. Open a data correction issue, which is a template in this
repository and which asks for the record, the field, the value it holds, the
value it should hold and the certificate behind it. A correction without a
certificate cannot be checked, which is why the template requires one.

What happens then. The certificate is read against the record. One of four things
follows, and which one is recorded on the issue:

- The record was read wrongly from its source. It is corrected, and the
  correction does not destroy what the record said before, which is issue #78.
- A newer test exists for the same product. Both are kept, because a prediction
  made against the older value has to remain reproducible, and which one is
  current is a field rather than a deletion.
- The certificate does not say what the report says it says. The record stands
  and the issue says why, quoting neither party's summary but the certificate.
- The certificate cannot be obtained. The record is withdrawn rather than kept
  with a caveat beside it. A value nobody can check is worse than an absent one,
  because an absent value stops a calculation and an unbacked one does not.

Where the dispute comes from the manufacturer of the product. It goes through the
same route and gets no shortcut and no delay. What changes is only that the
certificate is usually easier to obtain, which helps everybody. A request to
remove an unfavourable but correctly transcribed value is refused, in public, with
the certificate cited.

Where the dispute is about whether a record should exist at all rather than about
its numbers. That is not this route. Whether component data submitted by users is
accepted, and on what terms, is entry 6 of issue #1 and is unanswered, so today
every record traces to a certificate this project read itself.

Where somebody believes a record reproduces something they hold rights in, the
route is the one in [README.md](README.md) under "If you think something here
should not be", which removes first and assesses afterwards. That is deliberately
not this route, because the costs run the other way.

## Conduct

[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) covers behaviour and says how to report
it. A conduct report is not a technical dispute and the two do not share a route:
a refused data correction is not a conduct matter, and a conduct report is not
answered by re-reading a certificate.

# The committed fuzz corpus

Each directory here is named for a fuzz target and holds inputs that target has
to keep surviving. The toolchain reads them two ways, and both matter.

Under `go test`, with no fuzzing asked for, every file here runs as an ordinary
test case. That is what makes an entry a regression test rather than a note: an
input committed here is executed on every pull request, by the same command
everybody already runs.

Under `go test -fuzz`, the same files seed the search, so a run starts from the
inputs that were interesting last time instead of from nothing.

## What is in here and how it got here

An entry arrives one of two ways. Either a fuzz run found it and the toolchain
wrote it out, or somebody added an input that ought to keep being tried. Both
are welcome; neither is a place to put a fixture that a readable test could hold
instead. A corpus entry is an escaped byte string and a reader cannot see what it
is for, so the name carries that and the fixtures a person is meant to read live
in this package's `testdata` directory as documents.

The two entries under `FuzzReadDocument` today are both findings, made against a
parser that had been deliberately broken, and both are kept because the input
that found a defect once is the input most likely to find its return:

`value-whose-shortest-spelling-needs-an-exponent` is a document whose smallest
band value is far below a decibel. It is the case where a writer that chose the
wrong floating point format verb produces a spelling this format does not have,
so the document comes out of the writer unable to go back into the reader.

`header-line-with-one-field` is the format's own name on a line by itself and
nothing after it. It is the case where a line is split into fewer fields than
the reader then reads, which is an index past the end of the slice rather than a
refusal.

## What the corpus is not

It is not evidence that the parser is correct. It is the set of inputs somebody
happened to try, and a parser survives every one of them the moment it is written
to. What the corpus gives is that no defect these inputs found can come back
without a red run.

# Decision: what a result carries besides the number

Status: decided for the first release. The versioning promise at the end binds
every later change to this structure.

## The shape

A result is a record, not a number. It carries the number, the breakdown that
explains the number, and everything needed to reproduce it. Nothing in the list
below is optional, and none of it lives in a log.

The reason is that the number alone is the least useful part. An engineer needs
the dominant path to know what to change, and a reviewer needs the inputs to know
whether to believe it. A tool whose output cannot be checked by somebody who did
not run it is a tool whose output is an assertion.

## The structure, field by field

### Top level

- `schema` is the identifier and major version of this structure. A consumer
  reads this first and refuses what it does not understand.
- `kernel_version` is the version of the software that produced the result. Two
  results with different values here are not comparable without checking what
  changed.
- `standard_parts` names the parts of the standard whose method was applied, so
  a reader knows which method produced this and does not have to infer it.
- `model` is `detailed` or `simplified`. It is never absent and never defaulted.
  The reason it exists and why it is mandatory is
  [model-shape.md](model-shape.md).
- `quantity` names what was computed, in the standard's own symbol, for example
  `R'w`. The unit is stated beside it rather than assumed from the symbol.
- `value` is the number, with its unit.
- `bands` is present only for the detailed model, and carries the computed
  spectrum band by band before rating.
- `situation` identifies the source room, the receiving room and the separating
  element, each by the identity defined for elements and rooms, so the result
  points at what it was computed for.
- `paths` is the breakdown, described below.
- `inputs` is every input value that entered the calculation, described below.
- `assumptions` is every choice the kernel made on the user's behalf, each
  named, described below.
- `completeness` is `complete` or `incomplete`, described below.
- `input_basis` is the count of inputs by how they were obtained, described
  below. It is not an uncertainty and it is never presented as one. Why this
  project ships no numeric uncertainty in the first release is
  [uncertainty.md](uncertainty.md).

### `paths`

The breakdown is a field of the result, on the same footing as the value. It is
not an option, not a verbosity level and not a second command. A result without
it is not a result this kernel produces.

Each entry carries:

- `id`, stable within the result so a report can refer to it.
- `kind`, `direct` or `flanking`.
- `route`, the standard's own two-letter route notation, for example `Dd`, `Df`,
  `Fd`, `Ff`.
- `source_element` and `receiving_element`, by element identity.
- `junction`, by junction identity, absent on the direct path.
- `contribution`, in the same unit as `value`, so the dominant path is visible
  by reading the list and no rerun is needed.
- `share`, the fraction of the total energy this path accounts for, which is
  derived from `contribution` and included because the arithmetic that gets it
  wrong is exactly the arithmetic a reader wants to check.

### `inputs`

Every value that entered the calculation appears once here, and every reference
to a value elsewhere in the result is by the `id` of an entry in this list. There
is no value in a result that cannot be traced to an entry here.

Each entry carries:

- `id`, referenced from `paths` and from `assumptions`.
- `element` or `junction`, the thing the value belongs to.
- `quantity` and `value` with its unit.
- `origin`, one of `certificate`, `estimate`, `derived` or `default`.
  `certificate` means a measured value from a test report. `estimate` means a
  value somebody supplied without a test behind it. `derived` means this kernel
  computed it from other inputs, and then `derived_from` lists their ids.
  `default` means the kernel supplied it because nothing was given, and then it
  also appears in `assumptions` and forces `completeness` to `incomplete`.
- `source`, present when `origin` is `certificate`, pointing at the database
  record and, through it, at the test report the record was entered from. The
  record identity and what provenance a record has to carry are the subject of
  the component database work, and this structure only requires that the pointer
  resolves.

### `assumptions`

Each entry names one choice the kernel made that the user did not make, in
words, with the input ids it affected. A default value, a fallback junction
description, or a laboratory-to-situ correction applied without a measured basis
are all assumptions and all appear here. An empty list means the kernel made
none, and that is a claim the kernel has to be able to make truthfully.

### `completeness`

`complete` means every input was supplied or measured and nothing was defaulted.
`incomplete` means at least one input has `origin` of `default`, and the
`missing` list names each one and what was used instead.

A result may be `incomplete` and still be produced. What it may not be is
`incomplete` and indistinguishable from `complete`, which is the failure this
field exists to prevent.

### `input_basis`

Counts of the entries in `inputs` by `origin`. It is a summary of a list the
reader already has, and it exists so that a report can say in one line that
eleven of fourteen inputs were measured and three were estimated.

## Worked example

The situation is two rooms with one separating wall and two flanking edges.
A real situation usually has four flanking edges and twelve flanking paths; two
are shown so the example fits on a page, and the shape is the same.

THE NUMBERS BELOW ARE ILLUSTRATIVE AND ARE NOT A VALIDATED CALCULATION. They are
here to show the structure in full, they were not produced by this kernel, which
does not exist yet, and they must not be quoted as a result of this project.

```json
{
  "schema": "schallweg.result/1",
  "kernel_version": "0.0.0",
  "standard_parts": ["EN ISO 12354-1"],
  "model": "simplified",
  "quantity": "R'w",
  "unit": "dB",
  "value": 52.4,
  "situation": {
    "source_room": "room:living-a",
    "receiving_room": "room:living-b",
    "separating_element": "el:wall-sep-240"
  },
  "paths": [
    {
      "id": "p1",
      "kind": "direct",
      "route": "Dd",
      "source_element": "el:wall-sep-240",
      "receiving_element": "el:wall-sep-240",
      "junction": null,
      "contribution": 55.0,
      "share": 0.436
    },
    {
      "id": "p2",
      "kind": "flanking",
      "route": "Ff",
      "source_element": "el:floor-cast-200",
      "receiving_element": "el:floor-cast-200",
      "junction": "jc:t-floor-sep",
      "contribution": 60.1,
      "share": 0.134
    },
    {
      "id": "p3",
      "kind": "flanking",
      "route": "Fd",
      "source_element": "el:floor-cast-200",
      "receiving_element": "el:wall-sep-240",
      "junction": "jc:t-floor-sep",
      "contribution": 62.5,
      "share": 0.077
    },
    {
      "id": "p4",
      "kind": "flanking",
      "route": "Df",
      "source_element": "el:wall-sep-240",
      "receiving_element": "el:floor-cast-200",
      "junction": "jc:t-floor-sep",
      "contribution": 62.5,
      "share": 0.077
    },
    {
      "id": "p5",
      "kind": "flanking",
      "route": "Ff",
      "source_element": "el:ceiling-cast-200",
      "receiving_element": "el:ceiling-cast-200",
      "junction": "jc:t-ceiling-sep",
      "contribution": 59.4,
      "share": 0.157
    },
    {
      "id": "p6",
      "kind": "flanking",
      "route": "Fd",
      "source_element": "el:ceiling-cast-200",
      "receiving_element": "el:wall-sep-240",
      "junction": "jc:t-ceiling-sep",
      "contribution": 63.0,
      "share": 0.069
    },
    {
      "id": "p7",
      "kind": "flanking",
      "route": "Df",
      "source_element": "el:wall-sep-240",
      "receiving_element": "el:ceiling-cast-200",
      "junction": "jc:t-ceiling-sep",
      "contribution": 63.9,
      "share": 0.056
    }
  ],
  "inputs": [
    {
      "id": "i1",
      "element": "el:wall-sep-240",
      "quantity": "Rw",
      "value": 57.0,
      "unit": "dB",
      "origin": "certificate",
      "source": {
        "record": "db:wall/cast-concrete-240",
        "report": "lab-report-2019-0413"
      }
    },
    {
      "id": "i2",
      "element": "el:floor-cast-200",
      "quantity": "Rw",
      "value": 55.0,
      "unit": "dB",
      "origin": "certificate",
      "source": {
        "record": "db:floor/cast-concrete-200",
        "report": "lab-report-2019-0287"
      }
    },
    {
      "id": "i3",
      "element": "el:ceiling-cast-200",
      "quantity": "Rw",
      "value": 55.0,
      "unit": "dB",
      "origin": "certificate",
      "source": {
        "record": "db:floor/cast-concrete-200",
        "report": "lab-report-2019-0287"
      }
    },
    {
      "id": "i4",
      "junction": "jc:t-floor-sep",
      "quantity": "Kij",
      "value": 8.7,
      "unit": "dB",
      "origin": "derived",
      "derived_from": ["i1", "i2"]
    },
    {
      "id": "i5",
      "junction": "jc:t-ceiling-sep",
      "quantity": "Kij",
      "value": 8.7,
      "unit": "dB",
      "origin": "derived",
      "derived_from": ["i1", "i3"]
    },
    {
      "id": "i6",
      "element": "el:wall-sep-240",
      "quantity": "Rw_situ",
      "value": 57.0,
      "unit": "dB",
      "origin": "default",
      "derived_from": ["i1"]
    }
  ],
  "assumptions": [
    {
      "name": "no laboratory-to-situ correction applied",
      "detail": "No in situ correction was supplied for the separating element, so the laboratory value was used unchanged.",
      "affects": ["i6"]
    }
  ],
  "completeness": "incomplete",
  "missing": [
    {
      "input": "i6",
      "wanted": "in situ correction for el:wall-sep-240",
      "used_instead": "laboratory value unchanged"
    }
  ],
  "input_basis": {
    "certificate": 3,
    "estimate": 0,
    "derived": 2,
    "default": 1
  }
}
```

## The versioning promise

`schema` carries a major version. Within one major version:

- No field is removed and no field changes its meaning or its unit.
- No field changes type.
- New fields may be added, and a consumer must ignore fields it does not know.
- An enumerated value may gain a member, so a consumer must handle an unknown
  member of `origin`, `kind` or `model` by refusing rather than by guessing.

A consumer may rely on: `schema`, `model`, `quantity`, `unit`, `value`, the
presence and completeness of `paths`, the ability to resolve every reference in
`paths` and `assumptions` to an entry in `inputs`, and the meaning of
`completeness`.

A consumer may not rely on: the order of `paths` or `inputs`, the exact spelling
of the free text in `assumptions`, the numeric precision of `share`, or the
absence of fields not listed here.

Any change that breaks the list above is a new major version, and the two are
allowed to exist side by side rather than one silently replacing the other.

## Cost, paid deliberately

This structure is large, it has to serialise, and once anything reads it the
promise above binds. That is the price of a result somebody else can check, and
it is being paid here rather than discovered later, when the reporting work would
otherwise be the first place the question is asked.

# Decision: what a construction is, what an element is, and what a layer is

Status: decided for the first release.

Everything this kernel computes is about building elements, and the whole data
half of this project is a collection of them, so this is the second most
load-bearing decision after the spectrum type. It is also the decision that fixes
how large the database can get, which is the part that is easy to get wrong
cheaply and expensive to correct later.

## Three words that are not the same thing

The first decision is vocabulary, because the failure that follows from getting
it wrong is a database that stores every combination of everything.

A **construction** is a type. It is what a laboratory tested or what somebody
described: a two hundred and forty millimetre cast concrete wall, a double leaf
partition on steel studs. It has no position, it is not in a building, and one
construction appears in many buildings. This is what a database record holds.

An **element** is one instance of a construction in one building: this wall,
between these two rooms, of this area, with this lining on this side. It has a
position, it has as-built dimensions, and it exists only inside a project. This
is never a database record.

A **layer** is a part of a construction's makeup. It is a material, a thickness
and the way it is attached. A layer has no acoustic values of its own and is
never an object anybody looks up. It exists so that a construction's description
is structured rather than a sentence.

The distinction between construction and element does real work, and it does it
at the in situ correction. The correction between laboratory and building
conditions needs the dimensions of the specimen that was tested and the
dimensions of the thing that was built, and those are different numbers that
live on different objects. A model with one object has to hold both on it and
then decide, per field, which one is meant.

## Measured against described

Both are expressible. Only one produces a number in the first release, and the
line between them is stated here rather than discovered at the first surprising
result.

A **measured construction** carries laboratory values from a test, with the
provenance that [certificate-extraction.md](certificate-extraction.md) requires.
This is what the method wants and what certificates publish.

A **described construction** carries the makeup and the physical properties and
no laboratory values. It is what a user has when the exact thing they are
building was never tested, which is most of the time, so the model has to be able
to hold it rather than pretending those cases do not exist.

The first release computes from measured constructions only. Asking for a result
from a described construction refuses, names the construction and the quantity
that was missing, and says that no measured value is held for it.

The reason is that turning a description into a sound reduction spectrum is a
different method from the one this project implements. It is element-level
prediction, it has its own assumptions, its own accuracy and its own validation
corpus, and none of that exists here. Adding an unvalidated estimate would put a
number into a result beside measured values, where the whole design of
[result-contents.md](result-contents.md) is built to keep those two apart. So the
model carries descriptions, and the estimation route is a later decision with its
own validation rather than a convenience added inside this one.

The precedence rule, for when a construction has both. The measured value wins,
per quantity, always. An estimate is never used to fill in a quantity that was
measured, and it is never used to fill in bands that a measurement did not cover:
[frequency-bands.md](frequency-bands.md) refuses a partial spectrum outright, and
completing one from a different source would be the same invention with two
origins. Where an estimate is used, the result records the input as an estimate
rather than as a certificate value, using the origin field that already exists.

## An element, a lining and a covering

A lining is a separate object that attaches to a construction. So is a floor
covering. Neither is a property of the thing it attaches to.

The reason is arithmetic about the size of the database. If a lining is a
property, then a base wall with twelve possible linings is twelve records, and
adding a thirteenth lining means entering twelve more. With forty base walls and
twenty linings that is eight hundred records for sixty facts. If a lining is its
own object with its own improvement, it is sixty records, and the combination is
computed. Certificates publish them separately too, because that is how they are
tested: a base element is measured, then measured again with the lining, and the
improvement is the difference.

The two are kept apart from each other because they act on different quantities.
A lining improves the airborne sound reduction of the element it is on. A floor
covering improves impact sound transmission through the floor it is on. They have
different improvement quantities, they enter the calculation at different points,
and a model that called them both an improvement would let one be applied where
the other belongs.

An improvement carries the base construction it was measured on. This is not
optional and it is not a provenance nicety: an improvement measured on a heavy
base and applied to a lightweight one is a known way to be several decibels
wrong, and the only way anything downstream can warn about it is if the base is
in the record. Where a lining is applied to a base that is not the one it was
measured on, the kernel computes and records an assumption naming both. Where
the covering improvement belongs in the impact calculation is issue #69, and that
issue is the one place it may be applied.

## The fields

This section defines the objects as the kernel holds them. The record schema is
issue #73 and derives from this; where the two ever disagree the schema is what
runs and this document is what has to be corrected.

Units follow [numeric-contract.md](numeric-contract.md): SI throughout, no
alternates, and the formatter is where a thickness becomes millimetres for
reading.

### Construction

- `id`, required. Stable identity, which is issue #56.
- `kind`, required. One of an enumerated set: wall, floor, roof, window, door,
  and what else the situation model needs. Not free text, because it decides
  which calculations the construction may enter.
- `basis`, required. `measured` or `described`. Never inferred from whether the
  laboratory values happen to be present, because an absent spectrum and a
  description are different statements.
- `mass_per_area`, kg/m2, required. Needed by the in situ correction and by the
  junction estimation route, and there is no calculation in this project that can
  proceed without it.
- `thickness`, m, required. The total, derived from the layers where layers are
  given and stated directly where they are not.
- `layers`, ordered, required for `described`, optional for `measured`. A
  measured construction whose makeup was not published carries an empty list, and
  that emptiness is recorded rather than filled with a guess.
- `airborne_lab`, a spectrum, required for `measured` when the construction is
  used in an airborne calculation. The laboratory sound reduction index.
- `impact_lab`, a spectrum, required for `measured` floors used in an impact
  calculation. The laboratory normalised impact sound pressure level.
- `specimen_area`, m2, required with any laboratory spectrum. The area of the
  tested specimen.
- `specimen_edge_length`, m, required with any laboratory spectrum, for the edge
  the correction is referenced to.
- `lab_loss_factor`, dimensionless, required with any laboratory spectrum. The
  total loss factor of the specimen as tested. The in situ correction is a
  correction from this value to the building's, so a laboratory value without it
  cannot be corrected and can only be used uncorrected, which is an assumption
  the result then has to carry.
- `provenance`, required for `measured`. What it has to carry is
  [certificate-extraction.md](certificate-extraction.md) and issue #74.

### Layer

- `material`, required. A name from a controlled vocabulary, not free text, so
  two records of the same thing are the same thing.
- `thickness`, m, required.
- `density`, kg/m3, or `mass_per_area`, kg/m2, one of the two required. A sheet
  is naturally described by mass per area and a solid by density, and forcing one
  form makes half the records carry a converted number nobody can check against
  its source.
- `attachment`, required. How this layer is connected to the one below:
  bonded, mechanically fixed, resiliently fixed, or free. This is what
  distinguishes constructions that are identical on a materials list and tens of
  decibels apart in practice.

### Lining and covering

- `id`, `kind` and `provenance` as for a construction, with `kind` being lining
  or covering.
- `improvement`, a spectrum, required. The airborne improvement for a lining, the
  impact improvement for a covering. Two different quantities, and a lining
  record cannot carry the impact one or the reverse.
- `base_construction`, required. The construction it was measured on, by
  identity, for the reason in the previous section.
- `layers` and `mass_per_area` as for a construction, both required, because a
  lining has mass and that mass matters at the junction it sits beside.

### Element

- `id`, required. Unique within the project.
- `construction`, required. By identity.
- `area`, m2, required. As built, not as tested.
- `lining`, optional, and a separate entry per side, because a partition lined on
  one side is not a partition lined on both.
- `covering`, optional, floors only.
- `situ_loss_factor`, dimensionless, optional. The building's total loss factor
  where it is known. Where it is absent the kernel applies the default named in
  the next section.

## When a required property is absent

The kernel refuses, and the refusal names the element, the construction, the
quantity and what the calculation wanted it for. It does not substitute a
plausible value, and it does not compute a result over the properties it does
have.

That is the rule for a required property. There is exactly one documented
exception and it is not an exception to the rule so much as a different case: an
optional property with a defensible neutral value, where the neutral value is
named, recorded as an assumption, and forces the result to `incomplete`. The
building's loss factor is the standing example. Where it is unknown the
calculation can proceed on the laboratory value uncorrected, which is a stated
and conservative choice rather than an invented measurement, and the result says
so in the words the user reads.

The line between the two is whether the value is a measurement of a physical
object. A missing measurement is refused. A missing correction has a neutral
value and gets one, marked. That is the same line
[frequency-bands.md](frequency-bands.md) draws for a missing band and it is drawn
in the same place for the same reason.

## The model against four real constructions

Four cases the model has to hold, worked through to show which fields carry them.

THE VALUES BELOW ARE ILLUSTRATIVE STRUCTURAL FIGURES CHOSEN TO EXERCISE THE
MODEL. They are not from any certificate and no laboratory spectrum is shown,
because inventing one would be exactly what
[certificate-extraction.md](certificate-extraction.md) and
[result-contents.md](result-contents.md) exist to prevent. Where a spectrum
belongs, the field is named and its presence is stated.

**A heavy wall.** Cast concrete, 240 mm.

    kind                 wall
    basis                measured
    mass_per_area        552 kg/m2
    thickness            0.240 m
    layers               one: concrete, 0.240 m, density 2300 kg/m3, bonded
    airborne_lab         present, on the core band set
    specimen_area        10.0 m2
    specimen_edge_length 13.0 m
    lab_loss_factor      present
    provenance           laboratory, report number, date, standard, client

One layer, one mass, one spectrum. This is the case every model handles and it is
here as the baseline the other three are compared against.

**A lightweight wall.** Two gypsum board leaves on separate steel studs with a
cavity.

    kind                 wall
    basis                measured
    mass_per_area        45 kg/m2
    thickness            0.155 m
    layers               gypsum board, 0.0125 m, 10 kg/m2, mechanically fixed
                         gypsum board, 0.0125 m, 10 kg/m2, mechanically fixed
                         cavity with absorbent, 0.105 m, resiliently fixed
                         gypsum board, 0.0125 m, 10 kg/m2, mechanically fixed
                         gypsum board, 0.0125 m, 10 kg/m2, mechanically fixed
    airborne_lab         present, on the extended band set
    lab_loss_factor      present

What this case tests is whether `attachment` carries enough. Two constructions
with this identical materials list, one on shared studs and one on separate
studs, differ by a large margin, and the only field that says which is the
attachment of the cavity. It is why that field is required rather than
descriptive.

It also shows why the extended band set exists. Lightweight double leaf
constructions have their resonance in the low bands, which is where they fail and
where the core set says nothing.

**A floating floor.** A concrete base with a floating screed on it.

    the base:
    kind                 floor
    basis                measured
    mass_per_area        460 kg/m2
    thickness            0.200 m
    layers               concrete, 0.200 m, density 2300 kg/m3, bonded
    airborne_lab         present
    impact_lab           present

    the floating screed, a separate record:
    kind                 covering
    mass_per_area        90 kg/m2
    layers               screed, 0.045 m, density 1950 kg/m3, free
                         resilient layer, 0.030 m, 2 kg/m2, resiliently fixed
    improvement          present, the impact improvement
    base_construction    the concrete floor above, by identity

This is the case the separate-object decision was made for. The screed is one
record and it applies to every concrete base in the database, and the record says
which base it was measured on so that applying it to a lightweight floor produces
an assumption in the result rather than silence.

**A window.** Double glazing in a frame.

    kind                 window
    basis                measured
    mass_per_area        30 kg/m2
    thickness            0.024 m
    layers               glass, 0.006 m, density 2500 kg/m3, free
                         cavity, 0.012 m, free
                         glass, 0.006 m, density 2500 kg/m3, free
    airborne_lab         present
    specimen_area        1.44 m2
    lab_loss_factor      present

The window is the case that tests whether the model insists on things a window
does not have. It has no impact spectrum, which is expected for its kind rather
than a missing required field, and its specimen area is much smaller than a
wall's, which is why the specimen dimensions are per record and not a constant.

The frame is inside the tested specimen and not separately modelled. That is a
limit of this model and it is stated rather than discovered: a window whose frame
differs from the tested one is a different construction, and this model has no
way to say how different.

## What the model cannot describe

Stated so that the limit is chosen rather than found.

An element whose performance depends on its size in a way the specimen dimensions
do not capture. A construction with a deliberate opening in it, such as a
ventilation path or a door within a partition, which is a composite and belongs
to the situation rather than to one construction. An element that is not planar.
And any construction whose behaviour depends on its edge condition in the
building in a way the in situ correction does not model, which includes much of
what lightweight and timber construction actually does.

The last one is the one that matters most, and it is the same boundary the
junction model runs into. Where a case falls outside, the kernel refuses and says
so rather than producing a number that looks like the others.

## What would reopen this decision

An estimation route for described constructions arriving with its own validation
corpus, which would change the first release restriction and nothing else in this
document. A certificate source that publishes linings without naming the base
they were measured on, which would force a decision about whether such a record
may be entered at all. Or the in situ correction work in issue #52 needing a
specimen property this field list does not carry.

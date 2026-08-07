# The gap this project exists to fill

Every entry below carries the date it was checked and the address that was read
on that date. Where a claim could not be sourced, the sentence carrying it is
marked `UNSUPPORTED` in place. A marked sentence is not a weaker version of a
sourced one, it is a different kind of sentence, and it is left in because the
reader should be able to see where the record is thin.

All web addresses in this document were read on 2026-08-08. Vendor pages change
without notice, so a later reader should re-read them rather than trust the
quotations here.

## What the standard is

EN 12354 and its ISO revision ISO 12354 estimate the acoustic performance of a
building from the measured performance of its elements. Part 1 covers airborne
sound insulation between rooms, part 2 impact sound insulation, part 3 airborne
sound insulation against outdoor sound, part 4 sound transmitted from a building
to the outside, and part 6 sound absorption in enclosed spaces. The part numbers
above are named by two independent vendor pages read on 2026-08-08 and quoted
under the tool entries below.

The text of the standard is sold rather than published. `https://www.iso.org/`
returned HTTP 403 to the fetch attempted on 2026-08-08, so no price for any part
is quoted here and none is asserted. `UNSUPPORTED`: that the price of the
standard is a barrier to a small office is a plausible claim that this record
does not evidence.

## Tools that implement the standard

Six entries. Each names what was read and when. The list is what was found on
2026-08-08 and it is not asserted to be complete.

### CadnaB

Publisher: DataKustik GmbH.
Licensing model: commercial. No price is shown on the product page.
Checked 2026-08-08 at `https://www.datakustik.com/products/`, which lists CadnaA,
CadnaB and CadnaR, and describes CadnaB as "the software to calculate airborne
and impact sound transmission between rooms".

### BASTIAN

Publisher: DataKustik GmbH.
Licensing model: commercial. No price found.
Maintenance status, observed on 2026-08-08: BASTIAN does not appear anywhere on
DataKustik's own product page (`https://www.datakustik.com/products/`), which
lists CadnaA, CadnaB and CadnaR. On the same date, Sonusoft AB writes at
`https://www.sonusoft.com/` that "BASTIAN has been used in the Nordic countries
since 2000 to calculate sound insulation in buildings. The software is still used
by many consultants, but it is no longer maintained by the supplier DataKustik
GmbH."

That is the whole evidence for the unmaintained claim: an absence from the
publisher's product list, and a statement by a third party that sells a database
for it. The publisher has not been observed saying it, and no end-of-life notice
was found.

### INSUL

Publisher: Marshall Day Acoustics Ltd.
Licensing model: commercial. No price on the publisher's page, checked
2026-08-08 at `https://www.insul.co.nz/`, and no price on the distributor page
of Akustikbuero K5 GmbH, checked the same day at
`https://www.k5-akustik.de/en/software/insul/`.
Relation to the standard: the publisher's own page names no part of EN 12354 or
ISO 12354 at all; it describes prediction of transmission loss in third-octave
bands. The distributor page names DIN EN 12354-3 among the standards it works
to, alongside ISO 140-18, EN 673, DIN 4109 and DIN EN ISO 3382. So INSUL is
primarily an element-level prediction tool and its coverage of the
building-level parts is not established here.

### AcouBAT by CYPE

Publisher: CYPE Ingenieros, S.A., Alicante, developed with CSTB.
Licensing model: commercial. The page carries a purchase link to `shop.cype.com`
and states no price. Checked 2026-08-08 at
`https://info.cype.com/en/software/acoubat-by-cype/`.
What that page does name is national regulation sets rather than parts of the
standard: CTE DB HR in Spain, the Nouvelle Reglementation Acoustique in France,
RRAE in Portugal, Italian requirements, and NBR 15575 in Brazil.
`UNSUPPORTED`: that AcouBAT implements parts 1 to 6 of the standard is stated in
vendor and press material found in search results that this record did not read
directly, and it is not asserted here.

A separate CSTB shop address for the same product,
`https://boutique.cstb.fr/Detail/Logiciels/BIM-et-maquette-numerique/AcouBAT-by-CYPE`,
returned a not-found page on 2026-08-08.

### SONarchitect ISO

Publisher: Sound of Numbers SL.
Licensing model: commercial. No price stated. Checked 2026-08-08 at
`https://www.soundofnumbers.net/sonarchitect/index.php/en/sonarchitect-iso/isoen12354`.
That page names its coverage part by part: "Airborne noise insulation values
according to ISO 12354-1:2017", "Impact noise levels according to ISO
12354-2:2017", "Facade acoustic insulation values according to ISO 12354-3:2017",
"Noise emission levels from noisy enclosures according to ISO 12354-4:2017" and
"Reverberation times in rooms according to EN 12354-6". This is the clearest
public statement of scope found on 2026-08-08 and it is what the part list at the
top of this document rests on.

### Acoulatis

Publisher: Sonusoft AB.
Licensing model: commercial. No price found on the page read.
Checked 2026-08-08 at `https://www.sonusoft.com/`, which describes it as a
modular building acoustics prediction tool for airborne and impact sound
insulation.

### What this list does not contain

No implementation was found on 2026-08-08 that is free of charge, open source, or
otherwise inspectable by the person relying on its numbers. That is a negative
result over the six entries above and over the search that produced them, not a
proof that none exists.

## What the input data costs

A prediction under this standard needs measured element performance, which comes
from laboratory tests under the ISO 10140 series. Two costs are separable: buying
access to a collection somebody else assembled, and having an element tested.

Access to a collection, with a public price. Sonusoft AB sells the SOAB database
at "420 EUR / year" for "Use of the SOAB database during 1 year, including
regular updates by download or e-mail about 2 times/year, per simultaneous user
coded on the USB-dongle(s) or in the cloud, per software", checked 2026-08-08 at
`https://www.sonusoft.com/subscribe`. The same page states the subscription is
per concurrent user and per software platform, and names BASTIAN, CadnaB and
SONarchitect as the platforms. Two of those three are the products in the list
above, and the price therefore sits on top of whatever the tool itself costs.

The cost of a laboratory test. `UNSUPPORTED`: no price for an ISO 10140 test from
any laboratory was found on 2026-08-08, and none is asserted. What is established
is only that the tests exist as a standardised procedure, from the ISO catalogue
entries for ISO 10140-1 and ISO 10140-2 that appeared in the same search.

The prices of the tools themselves. Of the six entries above, none published a
price on any page read on 2026-08-08. Every one of them routes a prospective
buyer to a quotation, a shop that was not read, or a distributor. So the honest
statement is that tool prices in this field are not public, which is a finding
about the market rather than a gap in this record.

## What a practice without a licence does today

This is the claim the readme leans on hardest and it is the one with the least
evidence here.

`UNSUPPORTED`: that architecture practices and small engineering offices work
from spreadsheets, from national component catalogues, or by delegating the
calculation to a consultant is not evidenced by anything read on 2026-08-08. It
is the sentence a reader should challenge first, and the way to answer it is a
survey or a set of named interviews, neither of which this record contains.

What can be said without a survey is narrower and follows from the entries above.
A practice that wants to compute a flanking-inclusive result under this standard
has to buy a commercial tool at a price it cannot look up in advance, and, if it
does not have its own certificates, also buy a data subscription at 420 EUR per
year per concurrent user per platform. There is no free or inspectable route in
the list above. What practices actually do when faced with that is unevidenced.

## Who this is for

The audience is the engineer or acoustician who has to produce a defensible
number for a building and be able to show how it was produced.

That audience is assumed to know what a sound reduction index is, what a
third-octave spectrum is, that flanking transmission exists and usually decides
the outcome, and that a prediction is not a measurement.

That audience is not assumed to have a copy of the standard, to know the symbol
conventions of any particular national annex, to have any element data of its
own, or to administer the machine it works on.

A second audience is named in the readme and is deliberately not the primary one
here: the architect who knows the requirement and not the method. The difference
matters for later decisions, because a tool that assumes the first audience may
present a result the second cannot check, and a tool that assumes the second has
to make choices on the user's behalf and disclose every one of them.

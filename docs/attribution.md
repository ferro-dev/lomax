# Attribution & Naming

This document is the longer-form companion to the [`NOTICE`](../NOTICE) file and
the "About the name" section of the [`README`](../README.md). It exists to (1)
state the attribution in full, and (2) track the naming risk and the project's
commitment around it.

## Who Alan Lomax was

Alan Lomax (1915–2002) was an American ethnomusicologist, folklorist, and field
recordist. Over a career spanning more than six decades he travelled across the
United States and beyond — the American South, Appalachia, the Caribbean,
Britain, Italy, Spain — recording folk songs, work songs, blues, ballads, and
oral traditions that would otherwise have gone undocumented. Much of that work
fed the Library of Congress's Archive of American Folk Song. In 1983 he founded
the Association for Cultural Equity (ACE) to repatriate his recordings to the
communities they came from.

His life's work — finding, cataloguing, organising, and preserving a body of
music so it endures — is, in spirit, what this tool tries to be in software.
That is the entire reason for the name.

## Statement of non-affiliation

`lomax` (the software project) is **independent and unaffiliated**. It is not
produced by, endorsed by, sponsored by, or otherwise connected to:

- the **Alan Lomax estate**;
- the **Association for Cultural Equity** (ACE), the nonprofit Lomax founded in
  1983 — <https://www.culturalequity.org/>; or
- the **Library of Congress American Folklife Center**, which holds the Alan
  Lomax Collection.

For Alan Lomax's actual archive, recordings, and legacy, the canonical source is
the Association for Cultural Equity's research archive:
<https://research.culturalequity.org/>.

The name is used as an homage, not as a claim of association.

## Where attribution appears

Per the project's [planning document](music-cli-plan.md#13-community-infrastructure),
attribution is a first-class requirement and must appear on these surfaces:

| Surface | Form |
|---------|------|
| `README.md` | "About the name" section — who Lomax was, why the name, non-affiliation note |
| `NOTICE` | Named-after credit + non-affiliation, shipped in every distribution |
| `lomax about` / `lomax --version` | One-line credit (see below) |
| Documentation site landing page | Equivalent paragraph to the README |
| This file (`docs/attribution.md`) | Full statement + risk tracking |

The required one-line credit for `--version` / `about` output:

```
Named after Alan Lomax (1915–2002). Independent project, not affiliated with the Lomax estate or ACE.
```

## Naming-risk commitment

The homage framing makes an objection unlikely, but it is possible. The
project's position:

> If the Lomax estate, the Association for Cultural Equity, or the Library of
> Congress objects to the use of this name, the project commits in good faith
> to renaming.

### Objection log

No objections received. If one arrives, record it here with the date, the
objecting party, the nature of the objection, and the resolution, so the history
is auditable.

| Date | Party | Summary | Resolution |
|------|-------|---------|------------|
| —    | —     | (none)  | —          |

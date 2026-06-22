# INC-001  A normalization gap let a duplicate uppercase email slip through

> 日本語版は [INC-001-email-normalization.ja.md](INC-001-email-normalization.ja.md) を参照。

- status: resolved
- date: 2026-06-22
- severity: high
- related-ac: AC-001
- regression-test: TestDuplicateEmailRejected (the "A@Example.com" and surrounding-whitespace cases)

## Symptom
In production, the same user could register twice with the case-different emails "A@Example.com" and
"a@example.com".

## Reproduction condition
With "a@example.com" already registered, attempting to register with "A@Example.com" or with leading/trailing
whitespace as in " a@example.com" is not judged a duplicate and succeeds.

## Root cause
The duplicate check was done before normalizing the email. Because the string comparison happened before
folding case and surrounding whitespace, the same email with different notation was judged as distinct.

## Permanent countermeasure
Changed the duplicate check to run after normalize(email) (lowercasing, trimming surrounding whitespace). Also
added the "A@Example.com" case and the surrounding-whitespace case to AC-001's examples table in acceptance.md,
and froze them as regression tests.

## Wiring
- acceptance.md: added the uppercase case and the surrounding-whitespace case to AC-001's examples table (the trace is frozen).
- test: TestDuplicateEmailRejected confirms those cases red→green.
- lesson: in any capability involving an identity judgment, state in the spec's intent "what is normalized
  before comparison", and always place that normalization before the judgment. Comparing before normalization
  produces a hole of the same kind as this incident.

## What this incident demonstrates as a model
A failure is always wired to one regression test. The back-flow of incident → adding a criterion to acceptance.md
→ freezing the test guarantees, by a test, that the same problem is never hit twice. An incident with an empty
regression-test field is incomplete.

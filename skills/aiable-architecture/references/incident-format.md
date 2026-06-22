# The incident format

An incident is "a searchable memory of one's own failures". It has two purposes. One is to never hit the
same problem twice. The other is to let the AI load this record into context during related work, so it
can work in light of past failures. So an incident is not a poem for humans but a structured record.

> 日本語版は [incident-format.ja.md](incident-format.ja.md) を参照。

## The most important principle: an incident is always wired to a regression test

Each incident is always tied to one regression test. With this wiring, "never hit the same problem twice"
is guaranteed by a test. Enforce, as a mechanism, the back-flow of incident → test (turning a failure into
a new test and freezing it). An incident with no wiring is incomplete and is not considered to have a
countermeasure in place.

The regression test is also reflected into acceptance.md. Add the new acceptance criterion derived from the
incident (often example-based; for a PBT counterexample, the minimized concrete value) to acceptance.md, add
a row to the trace table, and run it through the normal approval-and-freeze flow (phase 2). An incident is
the entry point that returns from phase 4 to phase 2.

## The format

Each incident is one file `incident/INC-<serial>-<short-name>.md`.

```markdown
# INC-001  A normalization gap let a duplicate uppercase email slip through

- status: resolved
- date: 2026-06-22
- severity: high
- related-ac: AC-001
- regression-test: TestDuplicateEmailRejected (the A@Example.com case)

## Symptom
In production, the same user could register twice with the case-different emails "A@Example.com" and
"a@example.com".

## Reproduction condition
With "a@example.com" already registered, attempting to register with "A@Example.com" succeeds.

## Root cause
The duplicate check was done before normalizing the email. Because the comparison happened before folding
case and surrounding whitespace, the same email with different notation was judged as distinct.

## Permanent countermeasure
Changed the duplicate check to run after normalization. Also added the uppercase case and the
surrounding-whitespace case to AC-001 in acceptance.md as examples, and froze them as regression tests.

## Wiring
- acceptance.md: added the "A@Example.com" and surrounding-whitespace cases to AC-001's examples table
- test: TestDuplicateEmailRejected confirms those cases red→green
```

## The meaning of the fields

- **status** — open / investigating / resolved. resolved means the regression-test is wired and green.
- **related-ac** — the ID of the acceptance criterion this incident relates to. If it spawned a new
  criterion, that ID.
- **regression-test** — the name of the test that guarantees this problem never happens again. The most
  important field. An incident with this empty is considered incomplete.
- **Symptom / Reproduction condition / Root cause / Permanent countermeasure** — information the AI loads
  into context in future related work to avoid the same kind of failure. Write it concisely, but concretely
  enough to reproduce.

## When building an incident from a PBT counterexample

When a counterexample appears from random exploration or fuzzing, record the concrete input value that rapid
minimized in the "reproduction condition", and fix that input as an example-based regression test. PBT is the
net of exploration; the hole it finds is plugged with a deterministic example test. This way, a problem found
by probabilistic exploration is left in the record in a form that can be deterministically reproduced and
prevented.

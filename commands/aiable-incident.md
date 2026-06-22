---
name: aiable-incident
description: Phase 4 of Aiable Architecture — record a production failure, wire it to a regression test, add the criterion to acceptance.md, and return to phase 2.
argument-hint: <capability>
---

Drive **phase 4 (incident)** of the Aiable Architecture flow for the capability: `$ARGUMENTS`.

Follow the `aiable-architecture` skill as the governing discipline (read its `SKILL.md` and
`references/incident-format.md`). In the **current project**:

1. Create `capabilities/$ARGUMENTS/incident/INC-<serial>-<short-name>.md` with the structured fields: status,
   date, severity, related-ac, **regression-test**, symptom, reproduction condition, root cause, permanent
   countermeasure, and wiring. The regression-test field must not be empty.
2. Wire the failure to a regression test: add the new criterion (often example-based; for a PBT counterexample,
   the minimized concrete value) to `capabilities/$ARGUMENTS/spec/acceptance.md` and add a trace row.
3. Hand the new/updated criteria back through phase 2 (`/aiable-test`) for approval and re-freezing — an incident
   is the entry point that returns from phase 4 to phase 2.
4. Confirm the regression test reproduces the failure (red) and then passes after the countermeasure (green).

If no capability name was given, ask the user which capability the incident belongs to.

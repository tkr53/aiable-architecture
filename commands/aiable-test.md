---
name: aiable-test
description: Phase 2 of Aiable Architecture — generate the three test layers from the frozen acceptance.md, run mutation checks, then hand to the human for approval and freezing.
argument-hint: <capability>
---

Drive **phase 2 (test)** of the Aiable Architecture flow for the capability: `$ARGUMENTS`.

Follow the `aiable-architecture` skill as the governing discipline (read its `SKILL.md`, plus
`references/property-types.md` for property conversion). In the **current project**:

1. Require that `capabilities/$ARGUMENTS/spec/acceptance.md` has passed phase 1 (every AC has a normative block,
   open questions empty). If not, stop and tell the user to run `/aiable-spec` first.
2. From the normative blocks only (never from any implementation), generate the three test layers under
   `capabilities/$ARGUMENTS/test/`: example-based, property-based (per `references/property-types.md`), and a
   mutation check (a few deliberately wrong impls the tests must turn red).
3. Keep each test traceable 1:1 with an AC via the trace table.
4. Present the acceptance criteria list for the human to judge completeness and approve. On approval, mark the
   trace rows **frozen**.
5. After freezing, do **not** edit the tests. If you believe a test is wrong, raise it as a spec question, not
   an edit.

If no capability name was given, ask the user which capability to work on.

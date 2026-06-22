---
name: aiable-impl
description: Phase 3 of Aiable Architecture — generate the disposable implementation that turns the frozen tests green, without loosening any test.
argument-hint: <capability>
---

Drive **phase 3 (impl)** of the Aiable Architecture flow for the capability: `$ARGUMENTS`.

Follow the `aiable-architecture` skill as the governing discipline (read its `SKILL.md`). In the
**current project**:

1. Require that the tests under `capabilities/$ARGUMENTS/test/` are frozen (phase 2 complete). If not, stop and
   tell the user to run `/aiable-test` first.
2. Generate or regenerate `capabilities/$ARGUMENTS/impl/` until all three layers of frozen tests are green. The
   impl is disposable and AI-owned.
3. Do **not** loosen, skip, or edit any frozen test to make it pass. Making tests green by weakening them is the
   worst failure mode and is forbidden.
4. If a frozen test (or a PBT counterexample) seems wrong, do not fix it yourself — escalate to the user as a
   possible spec problem.
5. Report the final test result honestly (which layers pass, any remaining failures).

If no capability name was given, ask the user which capability to work on.

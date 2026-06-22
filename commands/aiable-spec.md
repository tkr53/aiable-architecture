---
name: aiable-spec
description: Phase 1 of Aiable Architecture — discuss and nail down a capability's spec (intent.md + acceptance.md) until the strict stop condition is met.
argument-hint: <capability>
---

Drive **phase 1 (spec)** of the Aiable Architecture flow for the capability: `$ARGUMENTS`.

Follow the `aiable-architecture` skill as the governing discipline (read its `SKILL.md` and, for the
acceptance format, its `references/acceptance-format.md`). In the **current project**:

1. Open or create `capabilities/$ARGUMENTS/spec/intent.md` and `capabilities/$ARGUMENTS/spec/acceptance.md`.
2. Discuss with the user and exhaustively eliminate questions and uncertainty. Hold open items in the
   open-questions section and strike them off as they resolve.
3. Write each acceptance criterion in the two-layer format (prose intent + normative block). Classify each as
   `kind: example` or `kind: property`. Keep machine-used values only inside the normative block.
4. Do **not** proceed to approval until BOTH strict stop conditions hold: (a) every AC has a normative block,
   and (b) the open-questions section is empty.
5. When the stop condition is reached, summarize the criteria and prompt the user for approval. Do not generate
   tests here — that is `/aiable-test`.

If no capability name was given, ask the user which capability to work on.

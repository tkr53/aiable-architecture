---
name: aiable-architecture
description: >-
  Discipline for building and maintaining software with Aiable Architecture, in a form where AI
  writes the code and humans own only the spec and the tests. Runs the canonical flow of
  spec → test → impl → incident per capability. Always use it when the user wants to nail down
  the spec of a new capability, generate and approve tests, build an implementation that satisfies
  the frozen tests, or record a failure to prevent recurrence, and whenever any of
  /aiable-spec /aiable-test /aiable-impl /aiable-incident is invoked.
  It also activates without an explicit request in contexts such as "architecture that assumes AI
  writes the code", "spec-driven", "make tests the sole acceptance gate", or "humans do not write code".
---

# Aiable Architecture

This is a discipline for software architecture designed on the premise that "AI writes the
implementation, and humans own only the spec and the tests". This file defines the first
principles, trust model, physical layout, canonical flow, test strategy, incident practice, and
command entry points that must always be followed when working under that premise. Individual
formats and conversion rules are split out into `references/`; read the relevant file when you
enter the phase that needs it.

> 日本語版は [SKILL.ja.md](SKILL.ja.md) を参照。

To state the whole architecture in one sentence: **source code is an artifact that can be
regenerated from the spec and the tests, and the assets truly worth preserving are the three of
spec, test, and incident.** Almost every discipline that follows is deduced from this single point.
Operate it having understood *why* it is so. If you understand the reason, you can fall on the
right side even in situations not spelled out here.

---

## 1. First principle: code is an artifact, spec/test/incident are the source

Traditional architecture treats source code as the primary asset and tests and specs as its
supporting material. Aiable Architecture inverts this. The implementation (impl) is a disposable
artifact that AI generates to satisfy the spec, and it can be thrown away and regenerated at any
time. The true source is the following three:

- **spec** — what must be satisfied. The primary material that humans discuss and settle.
- **test** — the anchor of trust that automatically verifies the spec is satisfied.
- **incident** — past failures and their permanent countermeasures. A learning asset for
  preventing recurrence.

Once you accept this inversion, several consequences follow automatically. Humans do not edit impl
directly (editing it breaks regenerability and lets the spec and impl drift apart). The amount of
impl code or the beauty of its internal structure is not itself an asset. As long as the spec and
tests are correct, the impl can be rebuilt any number of times.

> Why invert it? If AI takes on the implementation, human involvement should be concentrated where
> it has the highest value. The highest value is in deciding "what should be built" (spec) and
> "how to guarantee correctness" (test), not "how to write it" (impl). The latter is delegated to AI.

---

## 2. Trust model: AI-generated, human-approved, freeze tests before the implementation

The heart of this architecture is to always drive the anchor of trust outside the AI. If AI writes
the implementation and AI also writes the tests, then nothing is proven even when both are green
(the tests may have been written to suit the implementation). The disciplines below avoid this.

**Separation of author and approver.** AI may generate the tests and properties, but a human judges
their completeness and approves them. What is approved is not the generated test code but the list
of acceptance criteria and properties laid out in human language in `acceptance.md`. The human does
not read the test code; they check the list of acceptance criteria for gaps or excess.

**Freeze the tests before the implementation.** This is the most important ordering in the trust
model. If the implementation exists first, the tests the AI writes are reverse-engineered to make
that implementation green, and human approval degrades from "do the tests capture the spec?" to
"do the tests match the implementation?". That proves nothing. So the tests are approved and frozen
based on the spec alone, while no implementation exists. Only the moment an implementation later
turns them green does green carry meaning. Approval being independent of the implementation is the
very condition that anchors trust outside.

**Authority boundary after freezing.** AI cannot edit approved and frozen tests and properties.
Loosening a test to make the implementation green is the worst failure mode. If the AI judges a test
to be wrong, it does not fix it on its own but escalates to a human: "isn't this a spec problem?".
Against frozen verification, all the AI may do is file an objection, not edit; the authority to move
the verification always rests with the spec layer (= the human). The same holds when PBT finds a
counterexample: it is forbidden for the AI to weaken the property or narrow the generation range to
make it green. A counterexample is either "a bug in the implementation" or "an error in the property";
if you believe it is the latter, raise it to a human.

> The mechanism that prevents approval from becoming hollow is defined in
> `references/acceptance-format.md`. The key point: trace acceptance criteria and tests 1:1, so that
> reading acceptance.md alone is enough for the human.

---

## 3. Physical layout: split by capability, not by layer

Do not split directories by technical layer (controller / service / repository, etc.). Splitting by
layer scatters a single feature across multiple directories and inflates the context the AI must load
to fix one spot. Instead, split by **capability (a self-contained unit of behavior as a single spec)**
and place spec, test, impl, incident, and contract adjacent inside it. Everything needed for one
concern gathers physically in one place, so the AI completes its work by loading only that slice into
context.

```
capabilities/
  <capability-name>/
    spec/
      intent.md       # purpose, background, non-functional requirements, trade-off decision record, open questions
      acceptance.md   # acceptance criteria (1:1 with tests). Has normative blocks
    test/             # the frozen anchor of trust. Generated from and traceable to acceptance.md
    impl/             # AI-owned, disposable, regenerable
    incident/         # cause of failures and permanent countermeasures. Wired to regression tests
    contract.md       # the input/output boundary this capability promises to the outside
shared/               # minimal. Purely technical elements with no behavior only
```

**References between capabilities go through the contract only, in one direction.** When one
capability depends on another, the only thing it may reference is the other's `contract.md`; it must
not reference the other's impl. Since impl is disposable, depending on impl breaks the dependency
every time it is regenerated. Only the contract is the stable promise between capabilities.

**Keep shared thin.** What may live in shared is limited to purely technical, behavior-free elements
that belong to no capability's spec (cross-cutting type definitions, format constants, etc.). The test
is "should this have a spec/test?". If it should, it is behavior, and it must belong to some capability.
The only things you may offload to shared are those that do not need to be tested.

> Why squeeze shared this hard? shared is where the temptation to "just put it here for now" operates,
> and once you loosen it, locality collapses and the AI ends up reading all of shared to fix one feature.

---

## 4. The canonical flow and each phase's completion condition

A capability is built and maintained in the following order. Each phase has a clear completion
condition, and you do not advance to the next until it is met.

1. **Nail down the spec.** Humans and AI discuss and settle intent.md and acceptance.md. Exhaustively
   eliminate questions and uncertainty.
   **Completion condition (strict, both required)**: (a) every acceptance criterion has a normative
   block (inspectable by the lint in `references/acceptance-format.md`), and (b) the open-questions
   section of intent.md is empty. Do not proceed to approval until both are met. If either is missing,
   keep discussing.

2. **Generate, approve, and freeze the tests.** From the frozen acceptance.md, the AI generates the
   three test layers. The human judges completeness on acceptance.md, approves, and freezes the tests.
   See §5 for details.
   **Completion condition**: every row of the trace table is in the "frozen" state and human approval
   has been obtained.

3. **Generate the impl.** The AI generates the implementation until all frozen tests are green. impl is
   disposable.
   **Completion condition**: all three layers of frozen tests are green. Making them green by loosening
   tests is forbidden (the authority boundary of §2).

4. **Operate incidents.** When a failure occurs in production, record it in incident, wire it to a
   regression test, add the property/example to acceptance.md, and return to phase 2. See §6 for details.

> The reason phase 1's stop condition is made mechanically observable is to keep the principle
> "minimize questions" from being left to human intuition. If the stop condition is expressed as a
> state on a file (lint passes, the questions section is empty), then "this is enough" becomes an
> inspection rather than a subjective judgment.

---

## 5. Test strategy: the three layers of example, property, and mutation

Tests are composed of three layers, each with a different role. All three take acceptance.md as their
sole basis and must not be derived from the implementation (deriving from the implementation makes the
oracle problem recur).

- **Example-based (example)** — confirms the concrete inputs and outputs the human has in mind.
  Corresponds to `kind: example` criteria in acceptance.md.
- **Property-based (property, PBT)** — declares the input space itself and lets the machine explore it,
  verifying invariants on inputs that are hard for humans to imagine (boundaries, empty, maximal,
  Unicode, broken ordering). Corresponds to `kind: property` criteria. The human writes the property in
  words in acceptance.md, and the AI merely converts it into a generator and assertions. The property
  types (round-trip, invariant, idempotence, order-independence, test-oracle) follow the vocabulary in
  `references/property-types.md`.
- **Mutation testing** — a means of confirming, before approval, that the tests above "actually bite".
  Have the AI produce a few deliberately wrong implementations and see whether the tests to be frozen
  can turn them red. This prevents the oversight of tests so loose that any implementation passes.

To organize the roles: mutation inspects the **strength of the tests** (doubts the tests), while PBT
inspects the **completeness of the spec** (doubts the implementation). The two are complementary.

**Operational constraints on PBT.** Because counterexamples appear probabilistically, frozen tests must
be deterministically reproducible with a fixed seed (anything that passes or fails run to run is
disqualified as an anchor of trust). In CI, state the fixed seed and the number of explorations, run a
random-seed exploration job on a separate track, and turn any new counterexample into an incident. When
you find a counterexample, distill the minimized input into an example-based regression test wired to an
incident and fix it.

The detailed conversion templates (how to lower each property type into test code) live in
`references/property-types.md`. Read it in the test-generation phase.

---

## 6. Incident practice: always wire a failure to a regression test

An incident is "a searchable memory of one's own failures". It is not a poem for humans but a structured
record the AI can load into context during related work. At minimum, each incident has a symptom,
reproduction condition, root cause, permanent countermeasure, and **the name of the regression test that
guards it**.

That last item is the essence. If an incident is always wired to one regression test, then "never hit the
same problem twice" is guaranteed by a test. Enforce, as a mechanism, the back-flow of incident → test
(turning a failure into a new test and freezing it). The detailed format lives in
`references/incident-format.md`.

---

## 7. Command entry points

The user can drive each phase with an explicit command. The commands are merely entry points; the body
of the discipline is in this SKILL.md and the references. Each command operates by the following
conventions.

- **/aiable-spec `<capability>`** — phase 1. Start or continue the discussion of intent.md and
  acceptance.md, eliminating questions until the strict stop condition of §4 is met. Once the stop
  condition is reached, prompt for approval.
- **/aiable-test `<capability>`** — phase 2. From the frozen acceptance.md, generate the three test
  layers, pass mutation testing, and send them to human approval and freezing. Do not edit tests after
  freezing (§2).
- **/aiable-impl `<capability>`** — phase 3. Generate the impl that turns the frozen tests green. Do not
  loosen tests. If you believe a test is wrong, raise an objection (§2).
- **/aiable-incident `<capability>`** — phase 4. Record a failure, wire it to a regression test, add the
  criterion to acceptance.md, and return to phase 2.

> No command may break the authority boundary of §2. A command only provides a starting point for the
> work; the discipline of freezing, approval, and escalation is always held outside the command (this
> SKILL.md).

---

## references

Read the relevant file when you need it in each phase.

- `references/acceptance-format.md` — the format of acceptance.md, normative blocks, the trace table,
  the open-questions section, and the conventions for the second-layer lint and third-layer hash approval
  (added later).
- `references/property-types.md` — the 5 property types and the conversion template into test code for
  each type.
- `references/incident-format.md` — the structured incident format and how to wire it to regression tests.

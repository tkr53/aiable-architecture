# The format of acceptance.md

`acceptance.md` is the primary material that defines a capability's acceptance criteria. test/ is
generated from this file, and each test corresponds 1:1 with an acceptance criterion (AC) in the
trace table at the end. The purpose of this file is to create a state where a human can approve by
looking only at "are there any gaps or excess in the acceptance criteria?" without reading the test
code.

> 日本語版は [acceptance-format.ja.md](acceptance-format.ja.md) を参照。

## The four design requirements

This file must satisfy four things. First, give each AC a unique ID so it can be mechanically matched
1:1 with a test. Second, classify explicitly whether each AC is "confirmed by example (example)" or
"confirmed by property (property)". Third, what the human reads at approval time is contained in this
file alone, so they need not read the generated test code. Fourth, the AI can deterministically
generate tests from this file, and conversely can trace from a test back to this file.

To satisfy these, each AC has two layers: "prose for humans to read" and "a structural block for
machines to read". Prose alone makes the conversion non-deterministic; structure alone makes human
approval hollow. By splitting into two layers — prose handles "why", structure handles "what" — and not
overlapping their responsibilities, you prevent drift.

## Separation of prose and structure (the most important drift countermeasure)

Do not write the same fact twice. "Values that machines use", such as given/when/then or an examples
table, are written only in the structural block. The prose side (headings and intent lines) carries no
values, only the intent of why that criterion is needed. A value leaking into the prose is a red flag;
move the value into the structure.

The structural block is surrounded by `<!-- normative:begin AC-xxx -->` and
`<!-- normative:end AC-xxx -->`. Inside the surround is the region that "carries meaning = re-approval if
it changes"; outside the surround (intent and heading prose) is the region that is "supporting = no
re-approval even if it changes". When the AI generates tests, the only thing it may use as the basis for
conversion is the inside of this surround. It may read the intent, but must not derive tests from it.

## The AC format

Each AC takes the following form.

```markdown
### AC-001  Cannot register with an existing email
**intent:** Prevent duplicate registration of the same email. The root of data integrity.
            Decide after normalizing case and surrounding whitespace. Comparing before
            normalization lets duplicates slip through, so make explicit here that the
            comparison is post-normalization.
**kind:** example

<!-- normative:begin AC-001 -->
- given: email "a@example.com" is already registered
- when:  registration is attempted with an email that is identical after normalization
- then:  registration fails, with reason EMAIL_TAKEN
- and:   the count of the users store does not increase

examples:
| email          | prior state | expected result |
|----------------|-------------|-----------------|
| a@example.com  | registered  | EMAIL_TAKEN     |
| A@Example.com  | registered  | EMAIL_TAKEN     |
|  a@example.com | registered  | EMAIL_TAKEN     |
<!-- normative:end AC-001 -->
```

For a property, set kind to property and write the property-type, generator, and invariant in the
normative block. The format per type follows `property-types.md`.

```markdown
### AC-002  Registration preserves the count
**intent:** For any input sequence, the number of successful registrations matches the store count.
            Failures have no effect.
**kind:** property

<!-- normative:begin AC-002 -->
- property-type: invariant
- generator: a sequence (length 0..N) of registration requests mixing valid and invalid
- invariant: users.count == number of successful registrations
- note: failed requests (EMAIL_TAKEN, VALIDATION_ERROR) do not modify the store
<!-- normative:end AC-002 -->
```

## The open-questions section (one half of the spec phase's stop condition)

Place an open-questions section at the top of the file. Hold the questions the AI raised during the spec
phase here in structured form, and strike them off as discussion resolves each one. This section is
outside the normative region (on the prose side, not subject to re-approval), but it functions as the
gatekeeper of approval.

```markdown
## Open questions
(Must be empty. As long as any remain, you cannot proceed to approval.)
- [ ] e.g. Is the minimum password length defined in this spec, or delegated to another capability?
```

The spec phase's completion condition is strictly two things, and both must be met. First, every AC has a
normative block (inspected by the lint below). Second, this open-questions section is empty. If either is
missing, keep discussing.

## The trace table (the physical core of approval)

Place a trace table at the end of the file. The human does not read the test code; they judge gaps and
excess by the number of rows in this table and each row's AC-ID. A row with no AC section, or an AC with no
row, is an inconsistency in either case.

```markdown
## trace
| AC-ID  | kind     | test name                                 | state  |
|--------|----------|-------------------------------------------|--------|
| AC-001 | example  | TestDuplicateEmailRejected                | frozen |
| AC-002 | property | TestCountInvariant                        | frozen |
```

Only humans can add or remove rows. The AI cannot. This is a restatement, at the level of the table, of
"the AI cannot edit frozen tests".

## Completeness check (the last device against hollow approval)

Below the trace table, place a checklist the human answers at approval time.

```markdown
## Completeness check
- [ ] Have all failure cases been enumerated (any missing rejection reasons)?
- [ ] Is each property's generator wider than "the inputs the implementation plans to handle"?
- [ ] Are there any criteria written in later after looking at the impl (if so, a red flag)?
```

The last item is a question for self-checking that the oracle problem has not recurred. The human
periodically confirms that tests or criteria have not been written to match the implementation.

## Second-layer lint (introduce right away)

An existence check against acceptance.md itself. Run it as a pre-stage of test generation. The checks are
as follows.

- Is the normative block non-empty where an AC section exists?
- When kind is property, is property-type / generator missing inside the normative block?
- Do the AC-IDs disagree between the heading and the trace table?
- Does an AC section exist corresponding to each trace-table row?
- Is the open-questions section empty at approval time?

These are existence-level consistency rather than value matching, so they can be detected mechanically.
Value-matching checks are unnecessary (by the policy of not writing values twice across prose and
structure, no disagreeing values exist in the first place).

## Third-layer hash approval (added later)

The final means of preventing hollow approval and "silently rewritten" content. Record a hash of **the
concatenation of the insides of the normative blocks** of acceptance.md at approval time, and leave a wiring
that test generation was done against that hash. Anything outside the normative region (prose such as
intent) is excluded from the hash, so wording tweaks do not trigger re-approval; re-approval is required
only when given/when/then, an examples table, or a property definition changes.

Do not introduce this feature from the start; add it after the format and generation flow have stabilized.
The setup that makes the add-on a pure addition is the `normative:begin/end` delimiters above. You can mount
fixed approval just by adding code that concatenates the insides of the delimiters and takes a hash, without
touching the format at all.

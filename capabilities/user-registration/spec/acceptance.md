# Acceptance: user-registration

> Primary material. Owned and edited by humans. test/ is generated from this file, and each test
> corresponds 1:1 with an AC-ID in the trace table at the end. impl does not use this file as its basis.

> 日本語版は [acceptance.ja.md](acceptance.ja.md) を参照。

## Open questions
(Must be empty. As long as any remain, you cannot proceed to approval.)

## Terms and premises
- user: a registration subject that has an email and a password
- registered: a state where exactly one record whose post-normalization email matches exists in the users store
- normalization: lowercasing the email and trimming surrounding whitespace

---

### AC-001  Cannot register with an existing email
**intent:** Prevent double registration of the same email. The root of data integrity. Judge after
            normalization. Comparing before normalization lets duplicates with case or whitespace
            differences slip through, so make explicit as intent that the comparison is post-normalization.
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

---

### AC-002  A valid new email can be registered
**intent:** The happy path. An unregistered valid email registers successfully, and the count increases by 1.
**kind:** example

<!-- normative:begin AC-002 -->
- given: the users store is empty
- when:  registering with a valid email "new@example.com" and a non-empty password
- then:  registration succeeds
- and:   the count of the users store becomes 1

examples:
| email            | password | expected result |
|------------------|----------|-----------------|
| new@example.com  | secret1  | success         |
| User@Example.com | pw123456 | success         |
<!-- normative:end AC-002 -->

---

### AC-003  Registration preserves the count (failures do not change it)
**intent:** For any input sequence, the store count matches the number of successful registrations. Failed
            requests (duplicate, invalid) do not modify the store at all.
**kind:** property

<!-- normative:begin AC-003 -->
- property-type: invariant
- generator: a sequence (length 0..N) of registration requests mixing valid / invalid / duplicate
- invariant: store.Count() == (the number of times Register returned a nil error)
- note: failures (EMAIL_TAKEN, VALIDATION_ERROR) do not modify the store
<!-- normative:end AC-003 -->

---

### AC-004  Registration order does not affect the final state
**intent:** Registering distinct emails yields the same final store regardless of the order given.
**kind:** property

<!-- normative:begin AC-004 -->
- property-type: order-independence (commutativity)
- generator: a set of distinct valid emails, and two random permutations of it
- property: the store after registering all in permutation A == the store after registering all in permutation B
            (store identity is judged by the equality of the set of normalized emails)
<!-- normative:end AC-004 -->

---

### AC-005  A stored password cannot be restored
**intent:** The prohibition of plaintext storage and non-restorability. The stored value differs every time
            even for the same password (salt).
**kind:** property

<!-- normative:begin AC-005 -->
- property-type: negation of round-trip (one-way)
- generator: an arbitrary password string (including empty, maximal length, Unicode, control characters)
- property: after a successful registration, the plaintext password (as raw bytes) does not appear in the decoded stored value (salt+digest). The stored value is hex-decoded before the check; comparing against the hex text gives false positives because hex uses only 0-9a-f (see incident INC-002)
- property: registering the same password twice yields stored hashes that differ from each other
- note: a coincidental byte match for a very short password is not a plaintext-storage failure; the non-appearance check applies to passwords of >= 6 bytes, while the differing-hash property above covers all lengths
<!-- normative:end AC-005 -->

---

## trace
| AC-ID  | kind     | test name                       | state  |
|--------|----------|---------------------------------|--------|
| AC-001 | example  | TestDuplicateEmailRejected      | frozen |
| AC-002 | example  | TestValidRegistrationSucceeds   | frozen |
| AC-003 | property | TestCountInvariant              | frozen |
| AC-004 | property | TestOrderIndependence           | frozen |
| AC-005 | property | TestPasswordOneWay              | frozen |

## Completeness check
- [x] Have all failure cases been enumerated (any missing rejection reasons for EMAIL_TAKEN / VALIDATION_ERROR)?
- [x] Is each property's generator wider than "the inputs the implementation plans to handle"?
- [x] Are there any criteria written in later after looking at the impl (if so, a red flag)?

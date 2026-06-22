# INC-002  The one-way password test falsely flagged plaintext via a hex coincidence

> 日本語版は [INC-002-password-hex-false-positive.ja.md](INC-002-password-hex-false-positive.ja.md) を参照。

- status: resolved
- date: 2026-06-22
- severity: medium
- related-ac: AC-005
- regression-test: TestPasswordOneWay (decoded raw-byte comparison; e.g. the "a" case)

## Symptom
TestPasswordOneWay failed intermittently (depending on the rapid seed) with "plaintext password appeared in
stored value", even though the implementation hashes the password correctly and stores no plaintext.

## Reproduction condition
A short password composed of hex characters, e.g. "a", "f", or "0". It reproduces with `-count` over a few
random seeds, or with the fixed seed 3490257832917523786.

## Root cause
The test checked `strings.Contains(hash, password)` against the hex-encoded stored value. The stored form is
`hex(salt) + ":" + hex(digest)`, and hex uses only the 16 symbols 0-9a-f. A 96-hex-character string is
virtually certain to contain any given single hex character, so a password like "a" was a substring purely by
encoding coincidence — not because plaintext was stored. The check was a flawed proxy for "plaintext is not
stored".

## Permanent countermeasure
Decode the stored value to its raw bytes (salt+digest) before the substring check, removing the hex-alphabet
coincidence. Because the digest is effectively random, even raw bytes can coincidentally contain a very short
password, so the non-appearance assertion is applied only to passwords of >= 6 bytes (constant
`minMeaningfulPasswordLen`); the salt property (same password → different stored value) continues to cover all
lengths. AC-005 in acceptance.md was updated to state the decoded raw-byte comparison and the length scope, and
re-frozen through the normal approval flow (phase 2).

## Wiring
- acceptance.md: AC-005's normative block now specifies the decoded raw-byte comparison and the >= 6-byte scope.
- test: TestPasswordOneWay decodes via `decodeStored` and asserts on raw bytes; the previously failing case
  (e.g. "a") now passes deterministically.

## Lesson
When a property says "X does not appear in the stored value", verify it against the value's decoded/canonical
form, not an encoded text representation, and bound the claim to the range where appearance implies real storage
rather than coincidence against random data.

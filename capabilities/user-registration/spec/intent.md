# Intent: user-registration

> 日本語版は [intent.ja.md](intent.ja.md) を参照。

## Purpose
Accept new registration of users that have an email and a password. Prevent duplicate registration
and guarantee that the stored credentials are safe.

## Background
A subject with minimal realism as a model for this architecture. It naturally includes an example
criterion (duplicate rejection) and multiple property types (invariant, order-independence, the
negation of round-trip).

## Non-functional requirements
- Passwords must not be stored in plaintext. Restoration from the stored value must be impossible.
- Prevent double registration of the same email, including notational variation (case, surrounding
  whitespace).

## Trade-off decision record
- Email identity is judged "after normalization (lowercasing, trimming surrounding whitespace)". Store
  the notation as-is, but normalize only for the judgment. The reason is to prevent duplicates while
  respecting the user's input notation.
- The specifics of the password hashing scheme (bcrypt, etc.) are not fixed in this capability's spec but
  delegated to the impl. What the spec requires is only the observable properties of "not stored in
  plaintext", "not restorable", and "the stored value differs every time even for the same input (the
  existence of a salt)".

## Open questions
(Must be empty. As long as any remain, you cannot proceed to approval.)

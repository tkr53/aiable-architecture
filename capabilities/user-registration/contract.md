# Contract: user-registration

The boundary this capability promises to the outside. Other capabilities may reference only this
contract and must not reference the impl directly.

> 日本語版は [contract.ja.md](contract.ja.md) を参照。

## Provided operations

### Register(store, request) error
- input: Request{ Email string, Password string }
- output: error
  - nil: registration succeeded
  - ErrEmailTaken ("EMAIL_TAKEN"): an email identical after normalization is already registered
  - ErrValidationFailed ("VALIDATION_ERROR"): the email is invalid (empty / no @) or the password is empty

## Guaranteed properties
- Identity is judged after email normalization (lowercasing, trimming surrounding whitespace).
- On failure, the store is not modified.
- Passwords are not stored in plaintext and cannot be restored from the stored value.

## Public accessors (for tests and adjacent capabilities)
- Store.Count() int — number of registered entries
- Store.NormalizedEmails() map[string]struct{} — the set of normalized emails
- Store.StoredHash(normalizedEmail) (string, bool) — the stored hash (with an existence check)

## What is not promised
- The specifics of the hashing scheme (may change; do not depend on it).
- The internal representation of a record (a private matter of the impl).

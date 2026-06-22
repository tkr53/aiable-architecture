# Property types and conversion into rapid

Property-based testing (PBT) declares the input space itself and lets the machine explore it, verifying
invariants on inputs that are hard for humans to imagine (boundaries, empty, maximal, Unicode, broken
ordering). This file defines, per type, the template for deterministically converting each property written
in acceptance.md into Go test code for `pgregory.net/rapid`.

> 日本語版は [property-types.ja.md](property-types.ja.md) を参照。

## The grand principle: do not derive properties from the implementation

What is essentially hard in PBT is not generating inputs but deciding what the property (invariant) should
be. Defining a property is essentially a restatement of the spec. So you must not write properties by looking
at the implementation. Making the implementation's behavior into the invariant declares "correct" together
with the implementation's bugs, and the oracle problem recurs inside PBT. The human writes the property in
words in acceptance.md, and the AI merely converts it into a generator and assertions.

## Why rapid

rapid is a declarative PBT library close to Hypothesis that automatically minimizes counterexamples. A
minimized counterexample can be distilled directly into a regression test for an incident. Furthermore, with
`rapid.MakeFuzz` you can make the same property definition a target of Go's standard fuzzer. This way one
definition serves two purposes.

- **Frozen test**: run `rapid.Check` with a fixed number of trials. Deterministic, and an anchor of trust.
- **Exploration job**: fuzz the same function with `rapid.MakeFuzz` and run it long and randomized on a
  separate track. Turn any new counterexample into an incident.

A frozen test must be deterministic. Fix the number of trials and make it reproducible. Separate the long
randomized exploration into a job distinct from the frozen test.

## The five types

Each property in acceptance.md is classified into one of the following 5 types. The human's approval work
reduces to the finite choice of "which type is this criterion", and the AI's conversion is also deterministic
because the template is fixed per type.

### 1. Round-trip

A type where encoding then decoding returns the original, or storing then retrieving matches. Its negation
(one-way: cannot be restored) is also included here.

```go
// acceptance: encode → decode returns the original
func TestRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		original := rapid.String().Draw(t, "original")
		restored := Decode(Encode(original))
		if restored != original {
			t.Fatalf("round-trip mismatch: got %q want %q", restored, original)
		}
	})
}
```

### 2. Invariant

A type where some quantity (sum, count, balance) is preserved before and after an operation.

```go
// acceptance: after a sequence of registrations, store count == number of successful registrations
func TestCountInvariant(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		store := NewStore()
		reqs := rapid.SliceOf(genRequest()).Draw(t, "reqs")
		succeeded := 0
		for _, r := range reqs {
			if Register(store, r) == nil {
				succeeded++
			}
		}
		if store.Count() != succeeded {
			t.Fatalf("invariant broken: count=%d succeeded=%d", store.Count(), succeeded)
		}
	})
}
```

### 3. Idempotence

A type where applying the same operation twice gives the same result as applying it once.

```go
// acceptance: applying the same normalization twice equals applying it once
func TestIdempotent(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "s")
		once := Normalize(s)
		twice := Normalize(Normalize(s))
		if once != twice {
			t.Fatalf("not idempotent: once=%q twice=%q", once, twice)
		}
	})
}
```

### 4. Order-independence (commutativity)

A type where changing the order of inputs gives the same result.

```go
// acceptance: registering distinct emails yields the same final state regardless of order
func TestOrderIndependence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		emails := rapid.SliceOfDistinct(genEmail(), func(e string) string { return e }).Draw(t, "emails")
		permA := append([]string(nil), emails...)
		permB := shuffle(t, emails)

		a, b := NewStore(), NewStore()
		for _, e := range permA {
			Register(a, Request{Email: e, Password: "x"})
		}
		for _, e := range permB {
			Register(b, Request{Email: e, Password: "x"})
		}
		if !a.Equal(b) {
			t.Fatalf("order changed final state")
		}
	})
}
```

### 5. Test oracle

A type where a slow but obviously correct naive implementation and the real implementation produce matching
results. Guarantees that the real implementation's optimization has not changed behavior.

```go
// acceptance: the real implementation always agrees with the naive one
func TestAgainstOracle(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := genInput().Draw(t, "input")
		if Fast(input) != Naive(input) {
			t.Fatalf("disagrees with oracle on %v", input)
		}
	})
}
```

## Conversion into an exploration job (MakeFuzz)

Extract the body of the `rapid.Check` you wrote as a frozen test into a function, and mount it on the standard
fuzzer with `rapid.MakeFuzz`. This lets you explore the same property long and randomized. In CI, make the
frozen test (deterministic) the required gate and run the fuzz on a separate track and schedule.

```go
func propCountInvariant(t *rapid.T) {
	// the same logic as the body of TestCountInvariant above
}

func TestCountInvariant(t *testing.T) { rapid.Check(t, propCountInvariant) }

func FuzzCountInvariant(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propCountInvariant)) }
```

## Distilling counterexamples

When a counterexample appears from fuzzing or random exploration, take the input rapid minimized and fix it as
an **example-based regression test** wired to an incident. The flow is: explore with PBT → minimize the
counterexample → distill into an example test → wire to an incident. This way "never hit the same problem twice"
is guaranteed by a deterministic test. The concrete format of distillation follows `incident-format.md`.

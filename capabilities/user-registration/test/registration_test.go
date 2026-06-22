// The tests in this file are generated from spec/acceptance.md and correspond 1:1 with the trace table.
// They are frozen tests and must not be loosened to make the implementation green. If you judge a test
// to be wrong, go through a correction of acceptance.md (human approval).
// このファイルのテストは spec/acceptance.md から生成され、trace 表と 1:1 で対応する。
// 凍結されたテストであり、実装を緑にするために緩めてはならない。テストが誤っていると
// 判断した場合は acceptance.md の修正（人間の承認）を通すこと。
package registration_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	reg "example.com/aiable/user-registration/impl"

	"pgregory.net/rapid"
)

// ---- AC-001 (example): cannot register with an existing email / 既存メールでは登録できない ----
func TestDuplicateEmailRejected(t *testing.T) {
	cases := []string{
		"a@example.com",  // same notation / 同一表記
		"A@Example.com",  // case difference / 大文字違い
		" a@example.com", // surrounding-whitespace difference / 前後空白違い
	}
	for _, dup := range cases {
		s := reg.NewStore()
		if err := reg.Register(s, reg.Request{Email: "a@example.com", Password: "secret1"}); err != nil {
			t.Fatalf("setup registration failed / setup の登録に失敗: %v", err)
		}
		before := s.Count()
		err := reg.Register(s, reg.Request{Email: dup, Password: "another"})
		if err != reg.ErrEmailTaken {
			t.Fatalf("dup=%q: want EMAIL_TAKEN / EMAIL_TAKEN を期待したが %v", dup, err)
		}
		if s.Count() != before {
			t.Fatalf("dup=%q: count increased on failure / 失敗時に件数が増えた (%d→%d)", dup, before, s.Count())
		}
	}
}

// ---- AC-002 (example): a valid new email can be registered / 有効な新規メールは登録できる ----
func TestValidRegistrationSucceeds(t *testing.T) {
	cases := []struct {
		email, password string
	}{
		{"new@example.com", "secret1"},
		{"User@Example.com", "pw123456"},
	}
	for _, c := range cases {
		s := reg.NewStore()
		if err := reg.Register(s, reg.Request{Email: c.email, Password: c.password}); err != nil {
			t.Fatalf("email=%q: want success / 成功を期待したが %v", c.email, err)
		}
		if s.Count() != 1 {
			t.Fatalf("email=%q: want count 1 / 件数 1 を期待したが %d", c.email, s.Count())
		}
	}
}

// ---- generators ----

// genEmail generates a valid email (always contains @).
// genEmail は有効な email を生成する（@ を必ず含む）。
func genEmail() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		local := rapid.StringMatching(`[a-zA-Z0-9]{1,8}`).Draw(t, "local")
		domain := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "domain")
		return local + "@" + domain + ".com"
	})
}

// genRequest generates a registration request that may mix valid, invalid, and duplicate.
// genRequest は有効・無効・重複が混在しうる登録リクエストを生成する。
func genRequest() *rapid.Generator[reg.Request] {
	return rapid.Custom(func(t *rapid.T) reg.Request {
		// Mix valid emails with invalid ones (containing no @ or empty).
		// email は有効なものか、無効なもの（@ なしや空文字を含む）かを混ぜる。
		email := rapid.OneOf(
			genEmail(),
			rapid.StringMatching(`[a-z]{0,5}`), // no @ (tends to be invalid); also includes empty / @ を含まない（無効になりやすい）。空文字も含む
		).Draw(t, "email")
		// Mix empty (invalid) and non-empty passwords. Empty may also be generated.
		// password は空（無効）か非空かを混ぜる。空も生成しうる。
		password := rapid.StringMatching(`.{0,12}`).Draw(t, "password")
		return reg.Request{Email: email, Password: password}
	})
}

// ---- AC-003 (property, invariant): the count matches the number of successful registrations / 件数は成功登録数に一致する ----
func propCountInvariant(t *rapid.T) {
	s := reg.NewStore()
	reqs := rapid.SliceOfN(genRequest(), 0, 30).Draw(t, "reqs")
	succeeded := 0
	for _, r := range reqs {
		if reg.Register(s, r) == nil {
			succeeded++
		}
	}
	if s.Count() != succeeded {
		t.Fatalf("invariant broken / 不変量違反: count=%d succeeded=%d", s.Count(), succeeded)
	}
}

func TestCountInvariant(t *testing.T) { rapid.Check(t, propCountInvariant) }

// Exploration job: run the same property on the standard fuzzer long-running (separate track).
// 探索ジョブ: 同じプロパティを標準ファザーで長時間回す（別枠）。
func FuzzCountInvariant(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propCountInvariant)) }

// ---- AC-004 (property, order-independence): registration order does not affect the final state / 登録順序は最終状態に影響しない ----
func propOrderIndependence(t *rapid.T) {
	emails := rapid.SliceOfNDistinct(genEmail(), 0, 20,
		func(e string) string { return strings.ToLower(strings.TrimSpace(e)) },
	).Draw(t, "emails")

	// Build two orderings from the same set (Fisher-Yates, drawing randomness deterministically from rapid).
	// 同じ集合から 2 通りの順序を作る（Fisher-Yates、乱数は rapid から決定的に引く）。
	permA := shuffleDraw(t, emails, "permA")
	permB := shuffleDraw(t, emails, "permB")

	a := reg.NewStore()
	for _, e := range permA {
		reg.Register(a, reg.Request{Email: e, Password: "x"})
	}
	b := reg.NewStore()
	for _, e := range permB {
		reg.Register(b, reg.Request{Email: e, Password: "x"})
	}

	if !sameKeySet(a.NormalizedEmails(), b.NormalizedEmails()) {
		t.Fatalf("final state changed by order / 順序によって最終状態が変わった")
	}
}

func TestOrderIndependence(t *testing.T) { rapid.Check(t, propOrderIndependence) }

func FuzzOrderIndependence(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propOrderIndependence)) }

// ---- AC-005 (property, negation of round-trip): the password is non-restorable and differs every time / パスワードは復元不能・毎回異なる ----
func propPasswordOneWay(t *rapid.T) {
	password := rapid.String().Draw(t, "password")
	email := genEmail().Draw(t, "email")
	key := strings.ToLower(strings.TrimSpace(email))

	s := reg.NewStore()
	if err := reg.Register(s, reg.Request{Email: email, Password: password}); err != nil {
		// If the password is empty or otherwise invalid, it simply is not registered. Out of scope.
		// password が空など無効なら登録されないだけ。検査対象外。
		return
	}
	hash, ok := s.StoredHash(key)
	if !ok {
		t.Fatalf("registration succeeded but no hash found / 登録成功したのにハッシュが見つからない")
	}
	// The plaintext password must not appear in the decoded stored value (raw salt+digest bytes).
	// Comparing against the hex text gives false positives because hex uses only 0-9a-f (INC-002),
	// so decode first. A coincidental byte match for a very short password is not a real plaintext
	// leak, so only assert for passwords long enough that a match implies actual storage.
	// 平文 password が復号後の保存値（生の salt+digest バイト列）に出現しないこと。hex 文字列との
	// 比較は hex が 0-9a-f しか使わないため偽陽性になる（INC-002）。先に復号する。極端に短い
	// password の偶然一致は実際の平文保存ではないので、一致が保存を意味する長さ以上でのみ検査する。
	raw := decodeStored(t, hash)
	if len(password) >= minMeaningfulPasswordLen && bytes.Contains(raw, []byte(password)) {
		t.Fatalf("plaintext password appeared in stored value / 平文 password が保存値に出現した")
	}

	// Re-registering the same password with a different email yields a different hash (salt).
	// 同じ password を別 email で再登録すると、ハッシュは異なる（ソルト）。
	email2 := "x" + email
	key2 := strings.ToLower(strings.TrimSpace(email2))
	if err := reg.Register(s, reg.Request{Email: email2, Password: password}); err != nil {
		return
	}
	hash2, _ := s.StoredHash(key2)
	if hash == hash2 {
		t.Fatalf("same hash for same password: salt not effective / 同じ password で同一ハッシュ: ソルトが効いていない")
	}
}

func TestPasswordOneWay(t *testing.T) { rapid.Check(t, propPasswordOneWay) }

func FuzzPasswordOneWay(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propPasswordOneWay)) }

// ---- helpers ----

// minMeaningfulPasswordLen is the byte length at or above which a password appearing in the
// decoded stored value implies real plaintext storage rather than a coincidental match against
// the random salt+digest bytes (see INC-002).
// minMeaningfulPasswordLen は、復号後の保存値に password が現れたとき偶然一致ではなく
// 実際の平文保存を意味すると見なせる最小バイト長（INC-002 参照）。
const minMeaningfulPasswordLen = 6

// decodeStored decodes the "hex(salt):hex(digest)" stored form into its raw bytes.
// decodeStored は "hex(salt):hex(digest)" 形式の保存値を生バイト列に復号する。
func decodeStored(t *rapid.T, stored string) []byte {
	parts := strings.SplitN(stored, ":", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected stored format / 想定外の保存形式: %q", stored)
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("salt decode failed / salt のデコード失敗: %v", err)
	}
	digest, err := hex.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("digest decode failed / digest のデコード失敗: %v", err)
	}
	return append(salt, digest...)
}

func shuffleDraw(t *rapid.T, in []string, label string) []string {
	out := append([]string(nil), in...)
	// Fisher-Yates. Drawing each swap position from rapid explores deterministically and exhaustively.
	// Fisher-Yates。各スワップ位置を rapid から引くことで決定的かつ網羅的に探索する。
	for i := len(out) - 1; i > 0; i-- {
		j := rapid.IntRange(0, i).Draw(t, label)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func sameKeySet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

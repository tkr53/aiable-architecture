// このファイルのテストは spec/acceptance.md から生成され、trace 表と 1:1 で対応する。
// 凍結されたテストであり、実装を緑にするために緩めてはならない。テストが誤っていると
// 判断した場合は acceptance.md の修正（人間の承認）を通すこと。
package registration_test

import (
	"strings"
	"testing"

	reg "example.com/aiable/user-registration/impl"

	"pgregory.net/rapid"
)

// ---- AC-001 (example): 既存メールでは登録できない ----
func TestDuplicateEmailRejected(t *testing.T) {
	cases := []string{
		"a@example.com",  // 同一表記
		"A@Example.com",  // 大文字違い
		" a@example.com", // 前後空白違い
	}
	for _, dup := range cases {
		s := reg.NewStore()
		if err := reg.Register(s, reg.Request{Email: "a@example.com", Password: "secret1"}); err != nil {
			t.Fatalf("setup の登録に失敗: %v", err)
		}
		before := s.Count()
		err := reg.Register(s, reg.Request{Email: dup, Password: "another"})
		if err != reg.ErrEmailTaken {
			t.Fatalf("dup=%q: EMAIL_TAKEN を期待したが %v", dup, err)
		}
		if s.Count() != before {
			t.Fatalf("dup=%q: 失敗時に件数が増えた (%d→%d)", dup, before, s.Count())
		}
	}
}

// ---- AC-002 (example): 有効な新規メールは登録できる ----
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
			t.Fatalf("email=%q: 成功を期待したが %v", c.email, err)
		}
		if s.Count() != 1 {
			t.Fatalf("email=%q: 件数 1 を期待したが %d", c.email, s.Count())
		}
	}
}

// ---- generators ----

// genEmail は有効な email を生成する（@ を必ず含む）。
func genEmail() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		local := rapid.StringMatching(`[a-zA-Z0-9]{1,8}`).Draw(t, "local")
		domain := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "domain")
		return local + "@" + domain + ".com"
	})
}

// genRequest は有効・無効・重複が混在しうる登録リクエストを生成する。
func genRequest() *rapid.Generator[reg.Request] {
	return rapid.Custom(func(t *rapid.T) reg.Request {
		// email は有効なものか、無効なもの（@ なしや空文字を含む）かを混ぜる。
		email := rapid.OneOf(
			genEmail(),
			rapid.StringMatching(`[a-z]{0,5}`), // @ を含まない（無効になりやすい）。空文字も含む
		).Draw(t, "email")
		// password は空（無効）か非空かを混ぜる。空も生成しうる。
		password := rapid.StringMatching(`.{0,12}`).Draw(t, "password")
		return reg.Request{Email: email, Password: password}
	})
}

// ---- AC-003 (property, 不変量): 件数は成功登録数に一致する ----
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
		t.Fatalf("不変量違反: count=%d succeeded=%d", s.Count(), succeeded)
	}
}

func TestCountInvariant(t *testing.T) { rapid.Check(t, propCountInvariant) }

// 探索ジョブ: 同じプロパティを標準ファザーで長時間回す（別枠）。
func FuzzCountInvariant(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propCountInvariant)) }

// ---- AC-004 (property, 順序非依存): 登録順序は最終状態に影響しない ----
func propOrderIndependence(t *rapid.T) {
	emails := rapid.SliceOfNDistinct(genEmail(), 0, 20,
		func(e string) string { return strings.ToLower(strings.TrimSpace(e)) },
	).Draw(t, "emails")

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
		t.Fatalf("順序によって最終状態が変わった")
	}
}

func TestOrderIndependence(t *testing.T) { rapid.Check(t, propOrderIndependence) }

func FuzzOrderIndependence(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propOrderIndependence)) }

// ---- AC-005 (property, 往復性の否定): パスワードは復元不能・毎回異なる ----
func propPasswordOneWay(t *rapid.T) {
	password := rapid.String().Draw(t, "password")
	email := genEmail().Draw(t, "email")
	key := strings.ToLower(strings.TrimSpace(email))

	s := reg.NewStore()
	if err := reg.Register(s, reg.Request{Email: email, Password: password}); err != nil {
		// password が空など無効なら登録されないだけ。検査対象外。
		return
	}
	hash, ok := s.StoredHash(key)
	if !ok {
		t.Fatalf("登録成功したのにハッシュが見つからない")
	}
	// 平文 password が保存値に出現しない（空文字は部分文字列として常に含まれるため除外）。
	if password != "" && strings.Contains(hash, password) {
		t.Fatalf("平文 password が保存値に出現した")
	}

	// 同じ password を別 email で再登録すると、ハッシュは異なる（ソルト）。
	email2 := "x" + email
	key2 := strings.ToLower(strings.TrimSpace(email2))
	if err := reg.Register(s, reg.Request{Email: email2, Password: password}); err != nil {
		return
	}
	hash2, _ := s.StoredHash(key2)
	if hash == hash2 {
		t.Fatalf("同じ password で同一ハッシュ: ソルトが効いていない")
	}
}

func TestPasswordOneWay(t *testing.T) { rapid.Check(t, propPasswordOneWay) }

func FuzzPasswordOneWay(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propPasswordOneWay)) }

// ---- helpers ----

func shuffleDraw(t *rapid.T, in []string, label string) []string {
	out := append([]string(nil), in...)
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

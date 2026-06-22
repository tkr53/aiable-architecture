# プロパティの型と rapid への変換

プロパティベーステスト (PBT) は入力空間そのものを宣言して機械に探索させ、人間が想像しにくい
入力（境界・空・極大・Unicode・順序崩れ）での不変条件を検証する。このファイルは、acceptance.md
に書かれた各プロパティを Go の `pgregory.net/rapid` のテストコードへ決定的に変換するための
テンプレートを型ごとに定める。

## 大原則：プロパティは実装から導出しない

PBT で本質的に難しいのは入力生成ではなく、プロパティ（不変条件）を何にするかである。プロパティの
定義は実質的に仕様の再記述に等しい。だから実装を見てプロパティを書いてはならない。実装の挙動を
そのまま不変条件にすると、実装のバグごと「正しい」と宣言してしまい、オラクル問題が PBT の中で
再発する。プロパティは人間が acceptance.md に言葉で書き、AI はそれを generator とアサーションへ
変換するだけである。

## なぜ rapid か

rapid は Hypothesis に近い宣言的 PBT ライブラリで、反例を自動で最小化する。最小化された反例は
そのまま incident の回帰テストへ蒸留できる。さらに `rapid.MakeFuzz` で、同じプロパティ定義を
Go 標準ファザーのターゲットにできる。これにより一つの定義で二つの用途を兼ねられる。

- **凍結テスト**：`rapid.Check` を固定の試行回数で回す。決定的で、信頼の錨になる。
- **探索ジョブ**：同じ関数を `rapid.MakeFuzz` でファズ化し、別枠で長時間・乱択で回す。新しい
  反例が出たら incident 化する。

凍結テストは決定的でなければならない。試行回数を固定し、再現可能にする。乱択の長時間探索は
凍結テストとは別のジョブに分離する。

## 5 つの型

acceptance.md の各プロパティは、次の 5 型のいずれかに分類する。人間の承認作業は「この条件は
どの型か」という有限の選択に落ち、AI の変換も型ごとにテンプレートが決まるので決定的になる。

### 1. 往復性 (round-trip)

encode してから decode すると元に戻る、保存してから取得すると一致する、という型。その否定形
（one-way: 復元できないこと）もここに含む。

```go
// acceptance: encode → decode で元に戻る
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

### 2. 不変量 (invariant)

操作の前後で、ある量（合計・件数・残高）が保存される、という型。

```go
// acceptance: 一連の登録後、ストア件数 == 成功した登録数
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

### 3. 冪等性 (idempotence)

同じ操作を 2 回適用しても 1 回と同じ結果になる、という型。

```go
// acceptance: 同じ正規化を 2 回かけても 1 回と同じ
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

### 4. 順序非依存 (commutativity / order-independence)

入力の順序を変えても結果が同じになる、という型。

```go
// acceptance: 相異なる email の登録は順序によらず同じ最終状態
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

### 5. 対オラクル (test oracle)

遅いが自明に正しい素朴実装と、本実装の結果が一致する、という型。本実装の最適化が挙動を
変えていないことを保証する。

```go
// acceptance: 本実装は素朴実装と常に一致する
func TestAgainstOracle(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		input := genInput().Draw(t, "input")
		if Fast(input) != Naive(input) {
			t.Fatalf("disagrees with oracle on %v", input)
		}
	})
}
```

## 探索ジョブへの変換（MakeFuzz）

凍結テストとして書いた `rapid.Check` の中身を関数に切り出し、`rapid.MakeFuzz` で標準ファザーに
載せる。これで同じプロパティを長時間・乱択で探索できる。CI では凍結テスト（決定的）を必須
ゲートにし、ファズは別枠・別スケジュールで回す。

```go
func propCountInvariant(t *rapid.T) {
	// 上の TestCountInvariant の中身と同一のロジック
}

func TestCountInvariant(t *testing.T) { rapid.Check(t, propCountInvariant) }

func FuzzCountInvariant(f *testing.F) { f.Fuzz(rapid.MakeFuzz(propCountInvariant)) }
```

## 反例の蒸留

ファズや乱択探索で反例が出たら、rapid が最小化した入力を取り、incident に紐づく**例示ベースの
回帰テスト**として固定する。PBT で探索 → 反例を最小化 → 例示テストへ蒸留 → incident に結線、
という流れである。これにより「同じ問題を二度踏まない」が決定的なテストで保証される。蒸留の
具体的な書式は `incident-format.md` に従う。

# Acceptance: user-registration

> 一次資料。人間が所有・編集する。test/ はこのファイルから生成され、各テストは末尾の
> trace 表で AC-ID と 1:1 に対応する。impl はこのファイルを根拠にしない。

## 未解決質問
（空であること。残っている限り承認に進めない）

## 用語と前提
- ユーザー: email と password を持つ登録主体
- 登録済み: users ストアに、正規化後の email が一致するレコードが 1 件存在する状態
- 正規化: email を小文字化し前後空白を除去すること

---

### AC-001  既存メールでは登録できない
**intent:** 同一 email の二重登録を防ぐ。データ整合性の根幹。正規化後に判定する。
            正規化前に比較すると大文字違いや空白違いの重複をすり抜けるため、正規化後の
            比較であることを意図として明示する。
**kind:** example

<!-- normative:begin AC-001 -->
- given: email "a@example.com" が既に登録済み
- when:  正規化後に同一となる email で登録を試みる
- then:  登録は失敗し、理由は EMAIL_TAKEN
- and:   users ストアの件数は増えない

examples:
| email          | 事前状態 | 期待結果     |
|----------------|---------|-------------|
| a@example.com  | 登録済み | EMAIL_TAKEN |
| A@Example.com  | 登録済み | EMAIL_TAKEN |
|  a@example.com | 登録済み | EMAIL_TAKEN |
<!-- normative:end AC-001 -->

---

### AC-002  有効な新規メールは登録できる
**intent:** 正常系。未登録の有効な email は登録に成功し、件数が 1 増える。
**kind:** example

<!-- normative:begin AC-002 -->
- given: users ストアが空
- when:  有効な email "new@example.com" と非空の password で登録する
- then:  登録は成功する
- and:   users ストアの件数は 1 になる

examples:
| email            | password | 期待結果 |
|------------------|----------|---------|
| new@example.com  | secret1  | 成功     |
| User@Example.com | pw123456 | 成功     |
<!-- normative:end AC-002 -->

---

### AC-003  登録は件数を保存する（失敗は件数を変えない）
**intent:** どんな入力列を与えても、ストア件数は成功した登録の回数に一致する。失敗
            リクエスト（重複・無効）はストアを一切変更しない。
**kind:** property

<!-- normative:begin AC-003 -->
- property-type: 不変量 (invariant)
- generator: 有効/無効/重複が混在する登録リクエストの列（長さ 0〜N）
- invariant: store.Count() == （Register が nil error を返した回数）
- note: 失敗（EMAIL_TAKEN, VALIDATION_ERROR）はストアを変更しない
<!-- normative:end AC-003 -->

---

### AC-004  登録順序は最終状態に影響しない
**intent:** 相異なる email の登録は、与える順序によらず同じ最終ストアを生む。
**kind:** property

<!-- normative:begin AC-004 -->
- property-type: 順序非依存 (commutativity)
- generator: 相異なる有効 email の集合と、その 2 つのランダム順列
- property: 順列 A で全登録した後のストア == 順列 B で全登録した後のストア
            （ストアの同一性は、正規化 email の集合の一致で判定する）
<!-- normative:end AC-004 -->

---

### AC-005  保存したパスワードは元に戻せない
**intent:** 平文保存の禁止と復元不能性。同一 password でも保存値は毎回異なる（ソルト）。
**kind:** property

<!-- normative:begin AC-005 -->
- property-type: 往復性の否定 (one-way)
- generator: 任意の password 文字列（空・極大長・Unicode・制御文字を含む）
- property: 登録成功後のレコードに平文 password が出現しない
- property: 同じ password を 2 回登録すると、保存されたハッシュは互いに異なる
<!-- normative:end AC-005 -->

---

## trace
| AC-ID  | kind     | test 名                          | 状態   |
|--------|----------|---------------------------------|--------|
| AC-001 | example  | TestDuplicateEmailRejected      | 凍結   |
| AC-002 | example  | TestValidRegistrationSucceeds   | 凍結   |
| AC-003 | property | TestCountInvariant              | 凍結   |
| AC-004 | property | TestOrderIndependence           | 凍結   |
| AC-005 | property | TestPasswordOneWay              | 凍結   |

## 網羅性チェック
- [x] 失敗系は全て列挙したか（EMAIL_TAKEN / VALIDATION_ERROR の拒否理由に漏れはないか）
- [x] 各 property の generator は「実装で対応予定の入力」より広いか
- [x] impl を見て後から書き足した条件はないか（あれば red flag）

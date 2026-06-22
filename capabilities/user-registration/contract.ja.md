# Contract: user-registration

この capability が外部に約束する境界。他の capability はこの contract のみを参照してよく、
impl を直接参照してはならない。

## 提供する操作

### Register(store, request) error
- 入力: Request{ Email string, Password string }
- 出力: error
  - nil: 登録成功
  - ErrEmailTaken ("EMAIL_TAKEN"): 正規化後に同一 email が登録済み
  - ErrValidationFailed ("VALIDATION_ERROR"): email が無効（空・@ なし）または password が空

## 保証する性質
- 同一性は email の正規化（小文字化・前後空白除去）後で判定する。
- 失敗時はストアを変更しない。
- password は平文で保存されず、保存値から復元できない。

## 公開アクセサ（テスト・隣接 capability 向け）
- Store.Count() int — 登録済み件数
- Store.NormalizedEmails() map[string]struct{} — 正規化 email の集合
- Store.StoredHash(normalizedEmail) (string, bool) — 保存ハッシュ（存在検査つき）

## 約束しないこと
- ハッシュ方式の具体（変更されうる。依存しないこと）。
- レコードの内部表現（impl の私的事項）。

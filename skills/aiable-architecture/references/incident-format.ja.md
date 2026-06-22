# incident の書式

incident は「自分の失敗に対する検索可能な記憶」である。目的は二つ。一つは、同じ問題を二度
踏まないこと。もう一つは、AI が関連作業時にこの記録をコンテキストへ載せ、過去の失敗を踏まえて
作業できるようにすること。だから incident は人間向けのポエムではなく、構造化された記録にする。

## 最重要の原則：incident は必ず回帰テストへ結線する

各 incident は必ず一本の回帰テストに紐づく。この結線があれば「同じ問題を二度踏まない」が
テストによって保証される。incident → test の逆流（失敗を新しいテストに変えて凍結する）を
仕組みとして強制する。結線のない incident は未完成であり、対策が入ったとは見なさない。

回帰テストは acceptance.md にも反映する。incident から導かれた新しい受け入れ条件（多くは
例示ベース、PBT の反例なら最小化された具体値）を acceptance.md に追加し、trace 表に行を足し、
通常の承認・凍結フロー（フェーズ 2）を通す。incident はフェーズ 4 からフェーズ 2 へ戻る入口
である。

## 書式

各 incident は 1 ファイル `incident/INC-<連番>-<短い名前>.md` とする。

```markdown
# INC-001  正規化漏れで大文字 email の重複登録をすり抜けた

- status: resolved
- date: 2026-06-22
- severity: high
- related-ac: AC-001
- regression-test: TestDuplicateEmailRejected (A@Example.com のケース)

## 症状
本番で同一ユーザーが大文字違いの email "A@Example.com" と "a@example.com" で
二重に登録できてしまった。

## 再現条件
既に "a@example.com" が登録済みの状態で、"A@Example.com" で登録を試みると成功してしまう。

## 根本原因
重複判定を email の正規化前に行っていた。大文字・前後空白を畳む前に比較したため、
表記が異なる同一 email を別物と判定した。

## 恒久対策
重複判定を正規化後に行うよう変更。あわせて acceptance.md の AC-001 に大文字ケースと
前後空白ケースを examples として追加し、回帰テストとして凍結した。

## 結線
- acceptance.md: AC-001 の examples 表に "A@Example.com" と前後空白のケースを追加
- test: TestDuplicateEmailRejected が当該ケースを赤→緑で確認
```

## フィールドの意味

- **status** — open / investigating / resolved。resolved は regression-test が結線され、緑で
  あることを意味する。
- **related-ac** — この incident が関係する受け入れ条件の ID。新規条件を生んだ場合はその ID。
- **regression-test** — この問題を二度と起こさないことを保証するテストの名前。最重要フィールド。
  ここが空の incident は未完成とみなす。
- **症状 / 再現条件 / 根本原因 / 恒久対策** — AI が次回の関連作業でコンテキストに載せ、同種の
  失敗を避けるための情報。簡潔に、しかし再現できる程度に具体的に書く。

## PBT の反例から incident を作る場合

乱択探索やファズで反例が出たら、rapid が最小化した具体的な入力値を「再現条件」に記し、その
入力を例示ベースの回帰テストとして固定する。PBT は探索の網であり、見つかった穴は決定的な
例示テストで塞ぐ。これにより、確率的な探索で見つけた問題が、決定的に再現・防止できる形で
記録に残る。

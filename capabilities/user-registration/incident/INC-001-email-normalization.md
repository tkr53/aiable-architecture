# INC-001  正規化漏れで大文字 email の重複登録をすり抜けた

- status: resolved
- date: 2026-06-22
- severity: high
- related-ac: AC-001
- regression-test: TestDuplicateEmailRejected（"A@Example.com" と前後空白のケース）

## 症状
本番で同一ユーザーが大文字違いの email "A@Example.com" と "a@example.com" で二重に
登録できてしまった。

## 再現条件
既に "a@example.com" が登録済みの状態で "A@Example.com" あるいは前後に空白を含む
" a@example.com" で登録を試みると、重複と判定されず成功してしまう。

## 根本原因
重複判定を email の正規化前に行っていた。大文字・前後空白を畳む前に文字列比較したため、
表記の異なる同一 email を別物と判定した。

## 恒久対策
重複判定を normalize(email)（小文字化・前後空白除去）後に行うよう変更した。あわせて
acceptance.md の AC-001 の examples 表に "A@Example.com" と前後空白のケースを追加し、
回帰テストとして凍結した。

## 結線
- acceptance.md: AC-001 の examples 表に大文字ケースと前後空白ケースを追加（trace は凍結）。
- test: TestDuplicateEmailRejected が当該ケースを赤→緑で確認する。
- 教訓: 同一性判定を伴う capability では、「何を正規化してから比較するか」を spec の intent に
  明記し、その正規化を必ず判定の前段に置く。正規化前の比較は本 incident と同型の穴を生む。

## この incident が手本として示していること
失敗は必ず一本の回帰テストへ結線する。incident → acceptance.md への条件追加 → test の凍結、
という逆流が、同じ問題を二度踏まないことをテストで保証する。regression-test フィールドが
空の incident は未完成である。

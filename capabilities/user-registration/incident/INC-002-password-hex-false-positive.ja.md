# INC-002  hex の偶然一致で one-way パスワードテストが平文を誤検知した

> English version: see [INC-002-password-hex-false-positive.md](INC-002-password-hex-false-positive.md).

- status: resolved
- date: 2026-06-22
- severity: medium
- related-ac: AC-005
- regression-test: TestPasswordOneWay（復号後の生バイト比較。例: "a" のケース）

## 症状
実装は password を正しくハッシュ化し平文を保存していないにもかかわらず、TestPasswordOneWay が
rapid のシードに依存して断続的に「平文 password が保存値に出現した」で失敗した。

## 再現条件
hex 文字だけからなる短い password、例えば "a"・"f"・"0"。`-count` で数個の乱択シードを回すか、
固定シード 3490257832917523786 で再現する。

## 根本原因
テストが hex エンコードされた保存値に対して `strings.Contains(hash, password)` で検査していた。保存形式は
`hex(salt) + ":" + hex(digest)` であり、hex は 0-9a-f の 16 文字しか使わない。96 文字の hex 文字列は
任意の 1 文字をほぼ確実に含むため、"a" のような password はエンコードの偶然で部分一致しただけで、
平文が保存されたからではない。この検査は「平文が保存されていない」ことの不完全な代理だった。

## 恒久対策
部分一致検査の前に保存値を生バイト列（salt+digest）へ復号し、hex アルファベット由来の偶然一致を取り除いた。
digest は実質ランダムなので生バイトでも極端に短い password は偶然含まれうる。そこで非出現の検査は
6 バイト以上の password にのみ適用し（定数 `minMeaningfulPasswordLen`）、全長のカバーはソルト性質
（同一 password → 異なる保存値）に委ねる。acceptance.md の AC-005 を復号後の生バイト比較と長さの適用範囲を
明記する形に更新し、通常の承認フロー（フェーズ 2）で再凍結した。

## 結線
- acceptance.md: AC-005 の normative ブロックが復号後の生バイト比較と 6 バイト以上の適用範囲を規定する。
- test: TestPasswordOneWay は `decodeStored` で復号し生バイトで検査する。以前失敗していたケース（例: "a"）は
  決定的にパスするようになった。

## 教訓
property が「X が保存値に出現しない」と述べるときは、エンコードされたテキスト表現ではなく、値の復号後・
正準形に対して検証する。さらに、出現が偶然ではなく実際の保存を意味する範囲に主張を限定する。

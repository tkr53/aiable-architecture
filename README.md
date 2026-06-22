# Aiable Architecture

生成 AI 時代のためのソフトウェアアーキテクチャ。**AI が実装を書き、人間は仕様とテストだけを
所有する**ことを前提に設計されている。

## 第一原理

ソースコードは spec とテストから再生成可能な生成物であり、真に保存すべき資産は
**spec・test・incident** の三点である。実装（impl）は使い捨てで、いつでも作り直せる。
この逆転からアーキテクチャ全体が演繹される。

## 5 つの柱

1. **人間は impl を直接書かない。** 人間が触るのは spec と test だけ。実装は AI が生成する。
2. **層ではなく capability で割る。** 一つの関心事の spec/test/impl/incident を物理的に隣接させ、
   AI が読み込むコンテキストを最小化する。
3. **テストを信頼の錨にする。** ただしテストは spec へトレース可能でなければならない。テストは
   実装より先に承認・凍結される。
4. **仕様は徹底的に議論し、疑問と不確定要素を最低化する。** spec フェーズには厳格な停止条件が
   ある（全 AC に normative ブロック ＋ 未解決質問が空）。
5. **各 capability に spec / test / incident を持たせる。** incident は障害の原因と対策を記録し、
   必ず回帰テストへ結線して、同じ問題を二度踏まないことをテストで保証する。

## 信頼モデル

テストとプロパティは AI が生成し、人間が acceptance.md 上で過不足を判断して承認する。承認は
実装より先に行い、テストを凍結する。凍結後、AI はテストを編集できない（実装を緑にするために
テストを緩めるのを防ぐため）。テストが誤っていると思えば、AI は勝手に直さず人間にエスカレー
ションする。検証を動かす権限は常に仕様レイヤー（人間）にある。

## テスト戦略：三層

- **例示ベース** — 人間が想定する具体的な入出力。
- **プロパティベース (PBT)** — 入力空間を宣言して機械に探索させる。`pgregory.net/rapid` を使用。
- **ミューテーション検査** — テストがちゃんと噛むかを承認前に確認する。

## ディレクトリ構成

```
capabilities/
  <capability>/
    spec/        intent.md, acceptance.md（normative ブロック・trace 表・未解決質問）
    test/        凍結される信頼の錨。acceptance.md から生成
    impl/        AI 所有・使い捨て
    incident/    障害記録。回帰テストへ結線
    contract.md  外部への約束（capability 間はこれだけを参照）
shared/          最小限。振る舞いを持たない技術的要素のみ
```

## この Skill の使い方

`SKILL.md` がアーキテクチャの規律本体。`references/` に書式と変換テンプレートを置く。

- `references/acceptance-format.md` — acceptance.md の書式
- `references/property-types.md` — プロパティ 5 類型の rapid への変換
- `references/incident-format.md` — incident の書式と回帰テスト結線

コマンド入口（Claude Code 等で `/コマンド名 <capability>` として、または自律起動で）:

- `/aiable-spec` — 仕様を詰める
- `/aiable-test` — テストを生成・承認・凍結
- `/aiable-impl` — 凍結テストを満たす実装を生成
- `/aiable-incident` — 障害を記録し回帰テストへ結線

## 実例

`capabilities/user-registration/` がすべての規約に準拠した手本。手元で動かすには:

```bash
cd capabilities/user-registration
go mod tidy
go test ./...
```

example テストは標準ライブラリのみ、property テストは rapid を使用する。`go test -fuzz` で
探索ジョブ（FuzzCountInvariant 等）を別枠で回せる。

## ステータス

これは設計の参照実装であり、思想を伝えるための手本である。実プロジェクトに適用する際は、
言語・テスト基盤・CI 構成を自分の環境に合わせて読み替えること。

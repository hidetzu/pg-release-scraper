# pg-release-scraper

[![CI](https://github.com/hidetzu/pg-release-scraper/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/pg-release-scraper/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidetzu/pg-release-scraper)](https://goreportcard.com/report/github.com/hidetzu/pg-release-scraper)
[![Go Reference](https://pkg.go.dev/badge/github.com/hidetzu/pg-release-scraper.svg)](https://pkg.go.dev/github.com/hidetzu/pg-release-scraper)
[![Release](https://img.shields.io/github/v/release/hidetzu/pg-release-scraper)](https://github.com/hidetzu/pg-release-scraper/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

> English: [README.md](./README.md)

[PostgreSQL のリリースノート](https://www.postgresql.org/docs/release/) を指定したバージョン範囲でスクレイピングし、Excelワークブック（`.xlsx`）に出力するCLIツールです。

PostgreSQL のメジャーバージョンアップ時の **影響調査ワークシート** として使うことを想定しています。出力Excelには「翻訳」「調査キーワード」「確認結果」など、SI現場でよく使われる列があらかじめ用意されており、項目ごとの調査結果をそのまま書き込めます。

## 想定ユースケース

メジャーバージョンアップを行う前に、対象バージョン間のリリースノート（修正項目）を全件洗い出して影響範囲を調査する用途。例えば PostgreSQL 14.5 → 15.6 へのアップグレード時、14.5 から 15.6 までのすべてのリリースノート項目を1ファイルにまとめ、項目ごとに調査内容を記録できる形式で出力します。

## インストール

```bash
go install github.com/hidetzu/pg-release-scraper/cmd/pg-release-scraper@latest
```

ソースからビルドする場合:

```bash
git clone https://github.com/hidetzu/pg-release-scraper.git
cd pg-release-scraper
go build ./cmd/pg-release-scraper
```

## 使い方

```bash
pg-release-scraper --start 14.0 --end 15.6
```

### フラグ

| フラグ | 説明 | デフォルト |
|---|---|---|
| `--start <version>` | 開始バージョン（含む）例: `14.0` | — （必須） |
| `--end <version>` | 終了バージョン（含む）例: `15.6` | — （必須） |
| `--format <md\|xlsx\|both>` | 出力形式 | `both` |
| `--stdout` | Markdown を `.md` ファイルではなく標準出力に書く（xlsx が含まれる場合は引き続きファイル出力） | `false` |
| `--output <dir>` | 出力先ディレクトリ | `./output` |
| `--quiet` | エラー以外の出力を抑制 | `false` |
| `--version` | ツールのバージョンを表示して終了 | — |

### LLM へのパイプ

```bash
pg-release-scraper --start 14.5 --end 15.6 --stdout | claude -p "アプリへの影響を要約して"
```

### 出力

デフォルトでは2種類のファイルが出力ディレクトリに生成されます:

- `postgresql-release-notes_YYYYMMDD-HHMM.xlsx` — 日本のSI/DBA向けの調査ワークシート
- `postgresql-release-notes_YYYYMMDD-HHMM.md` — LLM入力／PR・Issue貼付け／grep用のMarkdown

片方だけ欲しい場合は `--format` を、Markdown を直接パイプしたい場合は `--stdout` を使ってください。

#### Excelワークブック

ワークブックには2つのシートが含まれます:

1. **`PostgreSQLリリースノート`** — 下記の列を持つメインのワークシート。
2. **`Attribution`** — データソース・著作権・ライセンス情報を埋め込んだシート。生成ファイル単体で再配布された際にも帰属情報が残るようにしています。

メインシートの列構成:

| 列 | ヘッダー | 説明 |
|---|---|---|
| A | `Ver` | リリースノートの対象 PostgreSQL バージョン |
| B | `No` | 連番 |
| C | `原文` | リリースノート原文（英語） |
| D | `翻訳(意味)` | 翻訳・意味（空欄、自分で記入） |
| E | `調査キーワード` | 調査キーワード（空欄、自分で記入） |
| F | `確認結果` | 確認結果（空欄、自分で記入） |
| G | `調査対象` | 調査対象（空欄、自分で記入） |
| H | `備考` | 備考（空欄、自分で記入） |

#### Markdown の構造

```markdown
# PostgreSQL Release Notes (14.5 → 15.6)

Generated: ...
Source: https://www.postgresql.org/docs/release/

---

## 14.6

1. **Tighten security restrictions within REFRESH MATERIALIZED VIEW CONCURRENTLY (Heikki Linnakangas)**

   One step of a concurrent refresh command...

2. **Fix memory leak when performing JIT inlining (Andres Freund)**

   There have been multiple reports...

## 14.7
...

---

## Attribution
- Source: https://www.postgresql.org/docs/release/
- Copyright (c) The PostgreSQL Global Development Group
- ...
```

## 動作の補足

- バージョン比較は数値タプルとして行います（`14` == `14.0`、`9.6` < `9.6.24`）
- リリースノート取得前に postgresql.org の `robots.txt` を取得して許可状態を確認します。許可されていないパスがあればエラーで終了します
- スクレイピング時はリクエスト間に 500ms 待機します（公式サイトへの負荷を考慮）
- 該当するリリースノートが0件だった場合（レンジ指定ミス等）は、標準エラーに警告を出して exit 0 で終了します

## データ提供元と帰属

リリースノートは <https://www.postgresql.org/docs/release/> から取得しています。

PostgreSQL のドキュメントは PostgreSQL Global Development Group の著作物であり、[PostgreSQL License](https://www.postgresql.org/about/licence/)（BSD-2-Clause類似のパーミッシブライセンス）の下で配布されています。生成された `.xlsx` を再配布する場合は、原著作権表示と免責文を保持してください。

本ツールは PostgreSQL Global Development Group とは無関係です。

## ライセンス

MIT - [LICENSE](./LICENSE) を参照してください。MITライセンスは本ツールのソースコードにのみ適用されます。スクレイピングで取得したリリースノート本文は引き続き PostgreSQL License の下にあります（上記参照）。

## コントリビュート

Issue / PR は歓迎です。PR作成前に [CONTRIBUTING.md](./CONTRIBUTING.md) を確認してください。

## 変更履歴

[CHANGELOG.md](./CHANGELOG.md) を参照してください。

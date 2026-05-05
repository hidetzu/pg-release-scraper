# pg-release-scraper

[![CI](https://github.com/hidetzu/pg-release-scraper/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/pg-release-scraper/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidetzu/pg-release-scraper)](https://goreportcard.com/report/github.com/hidetzu/pg-release-scraper)
[![Go Reference](https://pkg.go.dev/badge/github.com/hidetzu/pg-release-scraper.svg)](https://pkg.go.dev/github.com/hidetzu/pg-release-scraper)
[![Release](https://img.shields.io/github/v/release/hidetzu/pg-release-scraper)](https://github.com/hidetzu/pg-release-scraper/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

> 日本語版: [README.ja.md](./README.ja.md)

A CLI tool that scrapes [PostgreSQL release notes](https://www.postgresql.org/docs/release/) for a given version range and exports them to an Excel workbook (`.xlsx`).

Designed as an **upgrade-impact investigation worksheet** for PostgreSQL major-version upgrades. The output Excel sheet contains pre-defined columns (in **Japanese**) for tracking translations, investigation keywords, and findings — a workflow common in Japanese DBA / SI environments.

## Use case

Compiling release-note diffs across many minor versions during an upgrade impact assessment. For example, when upgrading PostgreSQL 14.5 → 15.6, this tool produces a single workbook listing every release-note item from 14.5 through 15.6, ready for line-by-line investigation.

## Installation

```bash
go install github.com/hidetzu/pg-release-scraper/cmd/pg-release-scraper@latest
```

Or build from source:

```bash
git clone https://github.com/hidetzu/pg-release-scraper.git
cd pg-release-scraper
go build ./cmd/pg-release-scraper
```

## Usage

```bash
pg-release-scraper --start 14.0 --end 15.6
```

### Flags

| Flag | Description | Default |
|---|---|---|
| `--start <version>` | Start version, inclusive (e.g. `14.0`) | — (required) |
| `--end <version>` | End version, inclusive (e.g. `15.6`) | — (required) |
| `--format <md\|xlsx\|both>` | Output format(s) to generate | `both` |
| `--stdout` | Write Markdown to stdout instead of an `.md` file (xlsx still goes to file when included) | `false` |
| `--output <dir>` | Output directory | `./output` |
| `--quiet` | Suppress non-error output | `false` |
| `--rules <path>` | YAML rules file for exclude filtering — see [Filtering with rules](#filtering-with-rules) | — |
| `--version` | Print tool version and exit | — |

### Pipe to an LLM

```bash
pg-release-scraper --start 14.5 --end 15.6 --stdout | claude -p "Summarize the impact for our application"
```

### Output

By default the tool generates two parallel files in the output directory:

- `postgresql-release-notes_YYYYMMDD-HHMM.xlsx` — Excel workbook for the Japanese investigation worksheet workflow
- `postgresql-release-notes_YYYYMMDD-HHMM.md` — Markdown for LLMs, PRs, issues, and grep-friendly review

Use `--format` to limit to one, or `--stdout` to pipe Markdown directly.

#### Excel workbook

The workbook contains two sheets:

1. **`PostgreSQLリリースノート`** — the main worksheet with the columns below.
2. **`Attribution`** — data-source, copyright, and license notes embedded so the file carries attribution even when redistributed standalone.

Main sheet columns:

| Column | Header (JP) | Description |
|---|---|---|
| A | `Ver` | Source PostgreSQL version |
| B | `No` | Sequential row number |
| C | `原文` | Original release-note text (English) |
| D | `翻訳(意味)` | Translation / meaning (blank, fill-in) |
| E | `調査キーワード` | Investigation keyword (blank, fill-in) |
| F | `確認結果` | Verification result (blank, fill-in) |
| G | `調査対象` | Investigation target (blank, fill-in) |
| H | `備考` | Notes (blank, fill-in) |

#### Markdown structure

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

## Filtering with rules

When investigating an upgrade, many release-note items (build-system tweaks, platform-specific fixes, documentation-only changes) are not relevant to application behaviour. Pass `--rules <path>` to drop them up front:

```bash
pg-release-scraper --start 14.5 --end 15.6 --rules examples/rules/app-impact.yaml
```

A starter file is provided at [`examples/rules/app-impact.yaml`](./examples/rules/app-impact.yaml). Tune it for your environment — text matching produces false positives/negatives, so the final call always belongs to a human reviewer.

### Rules file format (v1)

```yaml
version: 1
rules:
  - id: exclude-build
    action: exclude         # only "exclude" is supported in v0.2.0
    match:
      kind: regex           # keyword | regex
      target: detail        # only "detail" is supported in v0.2.0
      value: '(?i)\b(meson|autoconf)\b'
    rationale: |
      Optional free-form note, recorded on the Excel Attribution sheet for traceability.
```

- `keyword` matches case-insensitively as a substring.
- `regex` uses Go's [RE2](https://pkg.go.dev/regexp/syntax) — no PCRE lookarounds. Add `(?i)` inside the pattern for case-insensitive regex matching.
- `id` must be unique within the file.
- `rationale` is informational; it does not affect matching.
- `include` actions are **not** supported in v0.2.0; specifying one fails fast with `exit 2`.

### Evaluation

Each release-note item is checked against every `exclude` rule. If at least one rule matches (OR), the item is treated as out-of-scope. Rule order within the file does not affect the result.

### Output integration

When `--rules` is used:

- **stderr** (suppressed by `--quiet`): kept/removed counts and per-rule match counts
- **Markdown** (kept items only): filter metadata in the header; excluded items are not rendered
- **Excel main sheet** (all items): excluded items are still listed, with `対象外 (rule: <id>[, <id>...])` automatically filled in column F (`確認結果`) so reviewers can audit and override the auto-judgement
- **Excel `Attribution` sheet**: rule list with ID/Action/Kind/Value/Matched/Rationale

### Validation errors

The tool exits with status `2` on any rule-file error (missing fields, unknown enum value, duplicate `id`, regex compile failure, etc.). Error messages include the rule index and id:

```
rules.yaml: rule[2] (id=exclude-docs): invalid regex: error parsing regexp: ...
```

## Behavior notes

- Versions are compared as numeric tuples; `14` == `14.0` and `9.6` < `9.6.24`
- `robots.txt` on postgresql.org is fetched and honored before any release page is requested; if a path is disallowed the tool aborts with an error
- The scraper waits 500ms between requests to be polite to postgresql.org
- 0 results (e.g. an out-of-range version pair) prints a stderr warning and exits 0

## Data source & attribution

Release notes are scraped from <https://www.postgresql.org/docs/release/>.

PostgreSQL documentation is © The PostgreSQL Global Development Group and distributed under the [PostgreSQL License](https://www.postgresql.org/about/licence/) (a BSD-2-Clause-style permissive license). When redistributing the generated `.xlsx` file, retain the original copyright and disclaimer notices.

This tool is not affiliated with the PostgreSQL Global Development Group.

## License

MIT — see [LICENSE](./LICENSE). The MIT license applies to this tool's source code only; scraped release-note text remains under the PostgreSQL License (see above).

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before opening a pull request.

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

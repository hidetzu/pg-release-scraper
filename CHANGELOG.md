# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-05

### Added
- `--rules <path>` flag for YAML-based filtering of release-note items, aimed at upgrade-impact investigations.
- YAML rules schema v1 (`exclude` action × `keyword` / `regex` matching on `detail`).
- Filter summary on stderr (suppressed by `--quiet`) and embedded metadata in the Markdown header and the Excel `Attribution` sheet.
- Auto-marking of excluded items in the Excel `確認結果` (column F) with `対象外 (rule: <id>[, <id>...])`, so reviewers can audit and override the auto-judgement.
- Sample rules file at `examples/rules/app-impact.yaml` (build / platform / docs-only / translation excludes).

### Changed
- Markdown layout restructured to mirror the Excel worksheet: each item is now a `### Title` heading with `Ver` / `No` metadata bullets, separated by `---`. Item N in the Markdown corresponds to row N in the Excel sheet.
- Excluded items are now retained in the Markdown output with the same `確認結果: 対象外 (rule: ...)` marker as the Excel sheet (previously dropped from Markdown).
- Output filenames now embed the version range: `postgresql-release-notes_<start>_<end>_YYYYMMDD-HHMM.{md,xlsx}` (previously timestamp-only), so runs against different ranges in the same directory don't get confused.

[0.2.0]: https://github.com/hidetzu/pg-release-scraper/releases/tag/v0.2.0

## [0.1.0] - 2026-05-05

### Added
- Initial public release of `pg-release-scraper` CLI.
- Scraping PostgreSQL release notes by version range.
- `robots.txt` enforcement before fetching pages.
- Excel (`.xlsx`) export with investigation worksheet columns.
- Embedded `Attribution` sheet in Excel outputs.
- Markdown (`.md`) export.
- `--format` (`md|xlsx|both`) and `--stdout` support.
- Basic CI (`go vet`, `go test`, `go build`).
- OSS hygiene: `CONTRIBUTING.md`, `CLAUDE.md`, `CODEOWNERS`, pull request and issue templates.

[0.1.0]: https://github.com/hidetzu/pg-release-scraper/releases/tag/v0.1.0

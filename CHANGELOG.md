# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

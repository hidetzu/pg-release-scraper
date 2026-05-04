# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pg-release-scraper is a Go CLI that scrapes [PostgreSQL release notes](https://www.postgresql.org/docs/release/) over a version range and exports them to Excel (`.xlsx`) and Markdown (`.md`). It's designed as an upgrade-impact investigation worksheet, primarily for Japanese DBA / SI environments — the Excel headers are intentionally in Japanese (`原文`, `翻訳(意味)`, etc.).

## Build & Test Commands

```bash
make build              # Build to bin/pg-release-scraper
make test               # go test ./... -v -race
make vet                # go vet ./...
make fmt                # go fmt ./...
make lint               # golangci-lint run (requires installation)

go run ./cmd/pg-release-scraper --start 14.5 --end 15.6
go run ./cmd/pg-release-scraper --start 14.5 --end 15.6 --stdout --quiet
```

## Project Structure

- **cmd/pg-release-scraper/** — CLI entry point (flag parsing, orchestration).
- **internal/scraper/** — HTTP fetching, robots.txt validation, retry, HTML parsing (goquery). `Version`/`Release` types and `Fetch()`.
- **internal/excel/** — `.xlsx` rendering (excelize). Two sheets: main worksheet + Attribution.
- **internal/markdown/** — `.md` rendering. Per-version `##` headings + numbered list. `Render()` writes to `io.Writer`, used for both file output and `--stdout`.

## Key Design Conventions

- **Conventional Commits** — `<type>(<scope>): <subject>`. Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`, `revert`.
- **Bilingual README** — `README.md` (English) and `README.ja.md` (Japanese) must stay in sync when CLI behavior changes.
- **robots.txt is enforced** — every fetch path is checked via `temoto/robotstxt` before requesting. Do not bypass this; it's a deliberate operational guardrail.
- **Attribution must be preserved** — both `.xlsx` (Attribution sheet) and `.md` (footer) embed the PostgreSQL License attribution. The PostgreSQL documentation is © PostgreSQL Global Development Group, used under the PostgreSQL License.
- **Polite scraping** — 500ms delay between requests, exponential-backoff retry (1s/2s/4s, max 3) on network errors / 429 / 5xx.
- **stderr for progress, stdout for data** — `fetched N items` / `wrote ...` messages go to stderr so `--stdout` Markdown output isn't polluted.

## Adding a new CLI flag

1. Add `flag.String/Bool` in `cmd/pg-release-scraper/main.go`.
2. Update both READMEs (flags table).
3. Ensure progress messages go to stderr if the flag affects stdout.

## Adding a new output format

1. Create `internal/<format>/` mirroring `internal/markdown` (`Render(io.Writer, ...)` + `Write(..., outDir) (path, error)`).
2. Wire into `main.go`'s `--format` switch.
3. Add tests for `Render` (table-driven) and `Write` (temp dir).
4. Update both READMEs.

# Contributing

## Prerequisites
- Go 1.26+

## Setup
```bash
git clone https://github.com/hidetzu/pg-release-scraper.git
cd pg-release-scraper
go test ./...
```

## Development Rules
- Keep changes focused and small.
- Add or update tests for behavior changes.
- Update `README.md` and `README.ja.md` when CLI behavior changes.
- Preserve attribution and license notices for scraped PostgreSQL documentation.

## Pull Request Checklist
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] Documentation updated if needed
- [ ] No unrelated file changes

## Reporting Bugs
Please use the bug report template and include:
- CLI command used
- Expected behavior
- Actual behavior
- Environment details (OS, Go version)

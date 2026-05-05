package markdown

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hidetzu/pg-release-scraper/internal/filter"
)

func Write(items []filter.Annotated, start, end, outDir string, summary *filter.Summary) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}
	filename := fmt.Sprintf("postgresql-release-notes_%s_%s_%s.md", start, end, time.Now().Format("20060102-1504"))
	path := filepath.Join(outDir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := Render(f, items, start, end, summary); err != nil {
		return "", err
	}
	return path, nil
}

// Render writes a structured Markdown report. Each release-note item gets a
// `### Title` heading with metadata (Ver / No / 確認結果) and a horizontal
// rule between items, mirroring the Excel worksheet's column structure
// without forcing the body into a Markdown table.
//
// Excluded items (those with non-empty ExcludedBy) are still rendered, with
// "対象外 (rule: <id>[, <id>...])" in the 確認結果 metadata so the Markdown
// matches the Excel sheet 1:1.
func Render(w io.Writer, items []filter.Annotated, start, end string, summary *filter.Summary) error {
	ew := &errWriter{w: w}
	ew.printf("# PostgreSQL Release Notes (%s → %s)\n\n", start, end)
	ew.printf("Generated: %s\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"))
	ew.println("Source: https://www.postgresql.org/docs/release/")
	if summary != nil {
		ew.println()
		ew.printf("Filter: rules=%s  kept=%d/%d\n", summary.RulesPath, len(summary.Result.Kept()), summary.Result.Total)
		for _, r := range summary.Rules {
			ew.printf("- %s (%s, %s): matched %d\n", r.ID, r.Action, r.Kind, summary.Result.Hits[r.ID])
		}
	}
	ew.println()
	ew.println("---")
	ew.println()

	var currentVersion string
	for i, it := range items {
		detail := strings.TrimSpace(it.Release.Detail)
		if detail == "" {
			continue
		}
		if it.Release.Version != currentVersion {
			currentVersion = it.Release.Version
			ew.printf("## %s\n\n", currentVersion)
		}

		paras := strings.Split(detail, "\n")
		title := strings.TrimSpace(paras[0])
		ew.printf("### %s\n\n", title)
		ew.printf("- Ver: %s\n", it.Release.Version)
		ew.printf("- No: %d\n", i+1)
		if len(it.ExcludedBy) > 0 {
			ew.printf("- 確認結果: 対象外 (rule: %s)\n", strings.Join(it.ExcludedBy, ", "))
		}
		ew.println()

		for _, p := range paras[1:] {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			ew.printf("%s\n\n", p)
		}

		ew.println("---")
		ew.println()
	}

	ew.println("## Attribution")
	ew.println()
	ew.println("- Source: https://www.postgresql.org/docs/release/")
	ew.println("- Copyright (c) The PostgreSQL Global Development Group")
	ew.println("- License: [PostgreSQL License](https://www.postgresql.org/about/licence/)")
	ew.println("- Redistributors should retain original copyright and disclaimer notices.")
	ew.println("- This tool is not affiliated with the PostgreSQL Global Development Group.")
	return ew.err
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) printf(format string, args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintf(ew.w, format, args...)
}

func (ew *errWriter) println(args ...any) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprintln(ew.w, args...)
}

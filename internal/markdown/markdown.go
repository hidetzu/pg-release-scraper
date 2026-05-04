package markdown

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

func Write(releases []scraper.Release, start, end, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}
	filename := fmt.Sprintf("postgresql-release-notes_%s.md", time.Now().Format("20060102-1504"))
	path := filepath.Join(outDir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := Render(f, releases, start, end); err != nil {
		return "", err
	}
	return path, nil
}

func Render(w io.Writer, releases []scraper.Release, start, end string) error {
	ew := &errWriter{w: w}
	ew.printf("# PostgreSQL Release Notes (%s → %s)\n\n", start, end)
	ew.printf("Generated: %s\n", time.Now().UTC().Format("2006-01-02 15:04 UTC"))
	ew.println("Source: https://www.postgresql.org/docs/release/")
	ew.println()
	ew.println("---")
	ew.println()

	for _, g := range groupByVersion(releases) {
		ew.printf("## %s\n\n", g.version)
		for i, item := range g.items {
			paras := strings.Split(strings.TrimSpace(item.Detail), "\n")
			ew.printf("%d. **%s**\n", i+1, paras[0])
			for _, p := range paras[1:] {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				ew.println()
				ew.printf("   %s\n", p)
			}
			ew.println()
		}
	}

	ew.println("---")
	ew.println()
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

type versionGroup struct {
	version string
	items   []scraper.Release
}

// groupByVersion preserves the input ordering and skips items whose
// Detail is empty (which would otherwise render as "1. ****" rows).
func groupByVersion(releases []scraper.Release) []versionGroup {
	var groups []versionGroup
	for _, r := range releases {
		if strings.TrimSpace(r.Detail) == "" {
			continue
		}
		if len(groups) == 0 || groups[len(groups)-1].version != r.Version {
			groups = append(groups, versionGroup{version: r.Version})
		}
		groups[len(groups)-1].items = append(groups[len(groups)-1].items, r)
	}
	return groups
}

package markdown

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hidetzu/pg-release-scraper/internal/filter"
	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

func annotated(rs ...scraper.Release) []filter.Annotated {
	return filter.AnnotateAll(rs)
}

func TestRender(t *testing.T) {
	items := annotated(
		scraper.Release{Version: "15.6", Detail: "First note title\nFirst note body."},
		scraper.Release{Version: "15.6", Detail: "Second note title\nMulti-line\nbody here."},
		scraper.Release{Version: "15.7", Detail: "Later version note\nWith body."},
	)

	var buf bytes.Buffer
	if err := Render(&buf, items, "15.6", "15.7", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()

	mustContain := []string{
		"# PostgreSQL Release Notes (15.6 → 15.7)",
		"## 15.6",
		"## 15.7",
		"### First note title",
		"- Ver: 15.6",
		"- No: 1",
		"First note body.",
		"### Second note title",
		"- No: 2",
		"Multi-line",
		"body here.",
		"### Later version note",
		"- Ver: 15.7",
		"## Attribution",
		"PostgreSQL Global Development Group",
		"PostgreSQL License",
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("Render output missing %q\n--- output ---\n%s", want, out)
		}
	}

	// kept items (no rules applied) must NOT carry the 対象外 marker.
	if strings.Contains(out, "対象外") {
		t.Errorf("output should not contain 対象外 when no rules are applied:\n%s", out)
	}
}

func TestRender_ExcludedItemsCarryMarker(t *testing.T) {
	items := []filter.Annotated{
		{Release: scraper.Release{Version: "15.6", Detail: "Kept item\nbody"}},
		{
			Release:    scraper.Release{Version: "15.6", Detail: "Excluded item\nbody"},
			ExcludedBy: []string{"ex-build", "ex-docs"},
		},
	}
	var buf bytes.Buffer
	if err := Render(&buf, items, "15.6", "15.6", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "### Kept item") {
		t.Errorf("kept item heading missing:\n%s", out)
	}
	if !strings.Contains(out, "### Excluded item") {
		t.Errorf("excluded item should still be rendered:\n%s", out)
	}
	if !strings.Contains(out, "- 確認結果: 対象外 (rule: ex-build, ex-docs)") {
		t.Errorf("excluded item should carry 対象外 metadata:\n%s", out)
	}
}

func TestRender_SkipsEmptyDetail(t *testing.T) {
	items := annotated(
		scraper.Release{Version: "15.6", Detail: "First valid item\nbody"},
		scraper.Release{Version: "15.6", Detail: ""},
		scraper.Release{Version: "15.6", Detail: "   \n  "},
		scraper.Release{Version: "15.6", Detail: "Third valid item\nbody"},
	)
	var buf bytes.Buffer
	if err := Render(&buf, items, "15.6", "15.6", nil); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "### \n") || strings.Contains(out, "###  \n") {
		t.Errorf("expected no empty heading; got:\n%s", out)
	}
	if !strings.Contains(out, "### First valid item") {
		t.Errorf("missing first item; got:\n%s", out)
	}
	if !strings.Contains(out, "### Third valid item") {
		t.Errorf("missing third item; got:\n%s", out)
	}
}

func TestRender_PropagatesWriteError(t *testing.T) {
	items := annotated(scraper.Release{Version: "15.6", Detail: "x\nbody"})

	t.Run("first write fails", func(t *testing.T) {
		err := Render(&failingWriter{failAfter: 0}, items, "15.6", "15.6", nil)
		if err == nil {
			t.Fatal("expected error from failing writer, got nil")
		}
	})

	t.Run("mid-stream write fails", func(t *testing.T) {
		err := Render(&failingWriter{failAfter: 3}, items, "15.6", "15.6", nil)
		if err == nil {
			t.Fatal("expected error from mid-stream failure, got nil")
		}
	})
}

type failingWriter struct {
	failAfter int
	calls     int
}

func (fw *failingWriter) Write(p []byte) (int, error) {
	fw.calls++
	if fw.calls > fw.failAfter {
		return 0, errors.New("simulated write failure")
	}
	return len(p), nil
}

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	items := annotated(
		scraper.Release{Version: "15.6", Detail: "Sample\nWith body."},
	)
	path, err := Write(items, "15.6", "15.6", dir, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("output not in temp dir: %s", path)
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md extension, got %s", path)
	}
	if !strings.Contains(filepath.Base(path), "15.6_15.6") {
		t.Errorf("filename should contain version range: %s", path)
	}
}

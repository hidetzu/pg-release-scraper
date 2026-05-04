package markdown

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

func TestRender(t *testing.T) {
	releases := []scraper.Release{
		{Version: "15.6", Detail: "First note title\nFirst note body."},
		{Version: "15.6", Detail: "Second note title\nMulti-line\nbody here."},
		{Version: "15.7", Detail: "Later version note\nWith body."},
	}

	var buf bytes.Buffer
	if err := Render(&buf, releases, "15.6", "15.7"); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()

	mustContain := []string{
		"# PostgreSQL Release Notes (15.6 → 15.7)",
		"## 15.6",
		"## 15.7",
		"1. **First note title**",
		"   First note body.",
		"2. **Second note title**",
		"   Multi-line",
		"   body here.",
		"1. **Later version note**",
		"## Attribution",
		"PostgreSQL Global Development Group",
		"PostgreSQL License",
	}
	for _, want := range mustContain {
		if !strings.Contains(out, want) {
			t.Errorf("Render output missing %q\n--- output ---\n%s", want, out)
		}
	}
}

func TestRender_SkipsEmptyDetail(t *testing.T) {
	releases := []scraper.Release{
		{Version: "15.6", Detail: "First valid item\nbody"},
		{Version: "15.6", Detail: ""},
		{Version: "15.6", Detail: "   \n  "},
		{Version: "15.6", Detail: "Third valid item\nbody"},
	}
	var buf bytes.Buffer
	if err := Render(&buf, releases, "15.6", "15.6"); err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "****") {
		t.Errorf("expected no empty bold; got:\n%s", out)
	}
	if !strings.Contains(out, "1. **First valid item**") {
		t.Errorf("missing first item; got:\n%s", out)
	}
	if !strings.Contains(out, "2. **Third valid item**") {
		t.Errorf("expected number 2 (gap-free) for third item; got:\n%s", out)
	}
}

func TestRender_PropagatesWriteError(t *testing.T) {
	releases := []scraper.Release{{Version: "15.6", Detail: "x\nbody"}}

	t.Run("first write fails", func(t *testing.T) {
		err := Render(&failingWriter{failAfter: 0}, releases, "15.6", "15.6")
		if err == nil {
			t.Fatal("expected error from failing writer, got nil")
		}
	})

	t.Run("mid-stream write fails", func(t *testing.T) {
		err := Render(&failingWriter{failAfter: 3}, releases, "15.6", "15.6")
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
	releases := []scraper.Release{
		{Version: "15.6", Detail: "Sample\nWith body."},
	}
	path, err := Write(releases, "15.6", "15.6", dir)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("output not in temp dir: %s", path)
	}
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("expected .md extension, got %s", path)
	}
}

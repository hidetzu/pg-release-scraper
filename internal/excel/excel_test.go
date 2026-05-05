package excel

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hidetzu/pg-release-scraper/internal/filter"
	"github.com/hidetzu/pg-release-scraper/internal/scraper"
	"github.com/xuri/excelize/v2"
)

func TestWrite(t *testing.T) {
	dir := t.TempDir()
	items := filter.AnnotateAll([]scraper.Release{
		{Version: "15.6", Detail: "First release note item"},
		{Version: "15.6", Detail: "Second release note item\nwith newline"},
		{Version: "15.7", Detail: "Item from a later version"},
	})

	path, err := Write(items, "15.6", "15.7", dir, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("output not in temp dir: got %s, want under %s", path, dir)
	}
	if !strings.Contains(filepath.Base(path), "15.6_15.7") {
		t.Errorf("filename should contain version range: %s", path)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	t.Run("sheet names", func(t *testing.T) {
		got := f.GetSheetList()
		want := []string{"PostgreSQLリリースノート", "Attribution"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("sheets = %v, want %v", got, want)
		}
	})

	t.Run("main sheet headers", func(t *testing.T) {
		want := []string{"Ver", "No", "原文", "翻訳(意味)", "調査キーワード", "確認結果", "調査対象", "備考"}
		for i, w := range want {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			got, err := f.GetCellValue("PostgreSQLリリースノート", cell)
			if err != nil {
				t.Fatalf("GetCellValue %s: %v", cell, err)
			}
			if got != w {
				t.Errorf("header at %s = %q, want %q", cell, got, w)
			}
		}
	})

	t.Run("main sheet data rows", func(t *testing.T) {
		cases := []struct{ cell, want string }{
			{"A2", "15.6"},
			{"B2", "1"},
			{"C2", "First release note item"},
			{"A3", "15.6"},
			{"B3", "2"},
			{"C3", "Second release note item\nwith newline"},
			{"A4", "15.7"},
			{"B4", "3"},
			{"C4", "Item from a later version"},
		}
		for _, c := range cases {
			got, err := f.GetCellValue("PostgreSQLリリースノート", c.cell)
			if err != nil {
				t.Fatalf("GetCellValue %s: %v", c.cell, err)
			}
			if got != c.want {
				t.Errorf("%s = %q, want %q", c.cell, got, c.want)
			}
		}
	})

	t.Run("attribution sheet content", func(t *testing.T) {
		mustContain := []string{
			"postgresql.org",
			"PostgreSQL Global Development Group",
			"PostgreSQL License",
		}
		var combined strings.Builder
		for row := 1; row <= 10; row++ {
			cell, _ := excelize.CoordinatesToCellName(1, row)
			v, _ := f.GetCellValue("Attribution", cell)
			combined.WriteString(v)
			combined.WriteString("\n")
		}
		text := combined.String()
		for _, want := range mustContain {
			if !strings.Contains(text, want) {
				t.Errorf("Attribution sheet missing %q; got:\n%s", want, text)
			}
		}
	})
}

func TestWrite_ExcludedItemsMarked(t *testing.T) {
	dir := t.TempDir()
	items := []filter.Annotated{
		{Release: scraper.Release{Version: "15.6", Detail: "Kept item"}},
		{
			Release:    scraper.Release{Version: "15.6", Detail: "Excluded by one rule"},
			ExcludedBy: []string{"ex-build"},
		},
		{
			Release:    scraper.Release{Version: "15.7", Detail: "Excluded by two rules"},
			ExcludedBy: []string{"ex-build", "ex-docs"},
		},
	}

	path, err := Write(items, "15.6", "15.7", dir, nil)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	// All 3 items should be present in the main sheet (excluded ones are NOT
	// dropped — they get a marker in the 確認結果 column instead).
	cases := []struct{ cell, want string }{
		{"C2", "Kept item"},
		{"F2", ""},
		{"C3", "Excluded by one rule"},
		{"F3", "対象外 (rule: ex-build)"},
		{"C4", "Excluded by two rules"},
		{"F4", "対象外 (rule: ex-build, ex-docs)"},
	}
	for _, c := range cases {
		got, err := f.GetCellValue("PostgreSQLリリースノート", c.cell)
		if err != nil {
			t.Fatalf("GetCellValue %s: %v", c.cell, err)
		}
		if got != c.want {
			t.Errorf("%s = %q, want %q", c.cell, got, c.want)
		}
	}
}

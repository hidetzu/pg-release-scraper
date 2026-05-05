package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hidetzu/pg-release-scraper/internal/filter"
	"github.com/xuri/excelize/v2"
)

const sheetName = "PostgreSQLリリースノート"
const attributionSheetName = "Attribution"

var headers = []string{"Ver", "No", "原文", "翻訳(意味)", "調査キーワード", "確認結果", "調査対象", "備考"}

// Write renders all items (including ones excluded by rules) to the main
// worksheet. Excluded items get an automatic mark in the 確認結果 (F)
// column so reviewers can audit and override the auto-judgement.
func Write(items []filter.Annotated, outDir string, summary *filter.Summary) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}

	f := excelize.NewFile()
	defer f.Close()

	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return "", err
	}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return "", err
		}
	}

	for i, it := range items {
		row := i + 2
		if err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), it.Release.Version); err != nil {
			return "", err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), i+1); err != nil {
			return "", err
		}
		if err := f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), it.Release.Detail); err != nil {
			return "", err
		}
		if len(it.ExcludedBy) > 0 {
			mark := fmt.Sprintf("対象外 (rule: %s)", strings.Join(it.ExcludedBy, ", "))
			if err := f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), mark); err != nil {
				return "", err
			}
		}
	}

	if err := addAttributionSheet(f, summary); err != nil {
		return "", err
	}

	if err := f.SetColWidth(sheetName, "A", "B", 5); err != nil {
		return "", err
	}
	if err := f.SetColWidth(sheetName, "C", "C", 100); err != nil {
		return "", err
	}
	if err := f.SetColWidth(sheetName, "D", "D", 50); err != nil {
		return "", err
	}
	if err := f.SetColWidth(sheetName, "E", "H", 20); err != nil {
		return "", err
	}

	border := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1},
		{Type: "right", Color: "000000", Style: 1},
		{Type: "top", Color: "000000", Style: 1},
		{Type: "bottom", Color: "000000", Style: 1},
	}
	borderStyle, err := f.NewStyle(&excelize.Style{Border: border})
	if err != nil {
		return "", err
	}
	wrapStyle, err := f.NewStyle(&excelize.Style{
		Border:    border,
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
	})
	if err != nil {
		return "", err
	}

	lastRow := len(items) + 1
	if err := f.SetCellStyle(sheetName, "A1", fmt.Sprintf("H%d", lastRow), borderStyle); err != nil {
		return "", err
	}
	if err := f.SetCellStyle(sheetName, "C1", fmt.Sprintf("C%d", lastRow), wrapStyle); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("postgresql-release-notes_%s.xlsx", time.Now().Format("20060102-1504"))
	path := filepath.Join(outDir, filename)
	if err := f.SaveAs(path); err != nil {
		return "", fmt.Errorf("save xlsx: %w", err)
	}
	return path, nil
}

func addAttributionSheet(f *excelize.File, summary *filter.Summary) error {
	if _, err := f.NewSheet(attributionSheetName); err != nil {
		return err
	}

	rows := []string{
		"Data source: https://www.postgresql.org/docs/release/",
		"Copyright (c) The PostgreSQL Global Development Group",
		"License: PostgreSQL License (https://www.postgresql.org/about/licence/)",
		"Redistributors should retain original copyright and disclaimer notices.",
		"This tool is not affiliated with the PostgreSQL Global Development Group.",
	}
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetCellValue(attributionSheetName, cell, row); err != nil {
			return err
		}
	}
	if err := f.SetColWidth(attributionSheetName, "A", "A", 120); err != nil {
		return err
	}
	style, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{WrapText: true, Vertical: "top"},
	})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(attributionSheetName, "A1", fmt.Sprintf("A%d", len(rows)), style); err != nil {
		return err
	}

	if summary != nil {
		if err := writeFilterSection(f, len(rows)+2, summary); err != nil {
			return err
		}
	}
	return nil
}

func writeFilterSection(f *excelize.File, startRow int, s *filter.Summary) error {
	row := startRow
	header := fmt.Sprintf("Filter rules: %s  (kept %d / %d)", s.RulesPath, len(s.Result.Kept()), s.Result.Total)
	cell, _ := excelize.CoordinatesToCellName(1, row)
	if err := f.SetCellValue(attributionSheetName, cell, header); err != nil {
		return err
	}
	row += 2

	cols := []string{"ID", "Action", "Kind", "Value", "Matched", "Rationale"}
	for i, c := range cols {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		if err := f.SetCellValue(attributionSheetName, cell, c); err != nil {
			return err
		}
	}
	row++

	for _, r := range s.Rules {
		vals := []any{r.ID, string(r.Action), string(r.Kind), r.Value, s.Result.Hits[r.ID], r.Rationale}
		for i, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			if err := f.SetCellValue(attributionSheetName, cell, v); err != nil {
				return err
			}
		}
		row++
	}
	return nil
}

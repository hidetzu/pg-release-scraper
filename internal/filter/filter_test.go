package filter

import (
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

func mustKeyword(t *testing.T, id string, action Action, value string) Rule {
	t.Helper()
	return Rule{
		ID:      id,
		Action:  action,
		Kind:    KindKeyword,
		Target:  TargetDetail,
		Value:   value,
		Matcher: keywordMatcher{needle: lower(value)},
	}
}

func mustRegex(t *testing.T, id string, action Action, pattern string) Rule {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("compile %q: %v", pattern, err)
	}
	return Rule{
		ID:      id,
		Action:  action,
		Kind:    KindRegex,
		Target:  TargetDetail,
		Value:   pattern,
		Matcher: regexMatcher{re: re},
	}
}

func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func releases(items ...string) []scraper.Release {
	out := make([]scraper.Release, len(items))
	for i, d := range items {
		out[i] = scraper.Release{Version: "15.6", Detail: d}
	}
	return out
}

func keptDetails(rs []scraper.Release) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Detail
	}
	return out
}

func TestApply(t *testing.T) {
	all := releases(
		"Update meson build scripts",
		"Improve pg_dump performance",
		"Fix the documentation for COPY",
		"Add new SQL function jsonb_path_exists",
	)

	t.Run("no rules: pass through", func(t *testing.T) {
		got := Apply(nil, all)
		if len(got.Kept) != len(all) {
			t.Fatalf("kept=%d, want %d", len(got.Kept), len(all))
		}
		if got.Total != len(all) {
			t.Fatalf("total=%d, want %d", got.Total, len(all))
		}
	})

	t.Run("exclude only", func(t *testing.T) {
		rules := []Rule{
			mustKeyword(t, "no-build", ActionExclude, "meson"),
			mustKeyword(t, "no-docs", ActionExclude, "documentation"),
		}
		got := Apply(rules, all)
		want := []string{
			"Improve pg_dump performance",
			"Add new SQL function jsonb_path_exists",
		}
		if !reflect.DeepEqual(keptDetails(got.Kept), want) {
			t.Fatalf("kept=%v, want %v", keptDetails(got.Kept), want)
		}
		if got.Hits["no-build"] != 1 || got.Hits["no-docs"] != 1 {
			t.Fatalf("hits=%v", got.Hits)
		}
	})

	t.Run("include only narrows", func(t *testing.T) {
		rules := []Rule{
			mustKeyword(t, "only-pgdump", ActionInclude, "pg_dump"),
		}
		got := Apply(rules, all)
		want := []string{"Improve pg_dump performance"}
		if !reflect.DeepEqual(keptDetails(got.Kept), want) {
			t.Fatalf("kept=%v, want %v", keptDetails(got.Kept), want)
		}
		if got.Hits["only-pgdump"] != 1 {
			t.Fatalf("hits=%v", got.Hits)
		}
	})

	t.Run("include then exclude", func(t *testing.T) {
		// include matches 3 (anything mentioning a verb 'Update' or function/tooling),
		// then exclude knocks out documentation entries.
		rules := []Rule{
			mustRegex(t, "in-anything", ActionInclude, `.+`),
			mustKeyword(t, "ex-docs", ActionExclude, "documentation"),
		}
		got := Apply(rules, all)
		want := []string{
			"Update meson build scripts",
			"Improve pg_dump performance",
			"Add new SQL function jsonb_path_exists",
		}
		if !reflect.DeepEqual(keptDetails(got.Kept), want) {
			t.Fatalf("kept=%v, want %v", keptDetails(got.Kept), want)
		}
		if got.Hits["in-anything"] != 4 {
			t.Fatalf("include hits=%d, want 4", got.Hits["in-anything"])
		}
		if got.Hits["ex-docs"] != 1 {
			t.Fatalf("exclude hits=%d, want 1", got.Hits["ex-docs"])
		}
	})

	t.Run("multiple include rules: union (OR)", func(t *testing.T) {
		rules := []Rule{
			mustKeyword(t, "in-pgdump", ActionInclude, "pg_dump"),
			mustKeyword(t, "in-jsonb", ActionInclude, "jsonb"),
		}
		got := Apply(rules, all)
		want := []string{
			"Improve pg_dump performance",
			"Add new SQL function jsonb_path_exists",
		}
		if !reflect.DeepEqual(keptDetails(got.Kept), want) {
			t.Fatalf("kept=%v, want %v", keptDetails(got.Kept), want)
		}
	})

	t.Run("rule order does not affect result", func(t *testing.T) {
		rulesA := []Rule{
			mustKeyword(t, "ex-docs", ActionExclude, "documentation"),
			mustKeyword(t, "ex-meson", ActionExclude, "meson"),
		}
		rulesB := []Rule{
			mustKeyword(t, "ex-meson", ActionExclude, "meson"),
			mustKeyword(t, "ex-docs", ActionExclude, "documentation"),
		}
		gotA := Apply(rulesA, all)
		gotB := Apply(rulesB, all)
		if !reflect.DeepEqual(keptDetails(gotA.Kept), keptDetails(gotB.Kept)) {
			t.Fatalf("order matters: A=%v B=%v", keptDetails(gotA.Kept), keptDetails(gotB.Kept))
		}
	})

	t.Run("hits map includes zero-count rules", func(t *testing.T) {
		rules := []Rule{
			mustKeyword(t, "ex-no-match", ActionExclude, "this-string-never-appears"),
		}
		got := Apply(rules, all)
		if v, ok := got.Hits["ex-no-match"]; !ok || v != 0 {
			t.Fatalf("expected ex-no-match=0 in hits, got %v", got.Hits)
		}
		// All releases should survive.
		got2 := keptDetails(got.Kept)
		want := keptDetails(all)
		sort.Strings(got2)
		sort.Strings(want)
		if !reflect.DeepEqual(got2, want) {
			t.Fatalf("kept=%v, want %v", got2, want)
		}
	})

	t.Run("keyword match is case-insensitive", func(t *testing.T) {
		rs := releases("Refactor MESON build helpers")
		rules := []Rule{
			mustKeyword(t, "ex-build", ActionExclude, "meson"),
		}
		got := Apply(rules, rs)
		if len(got.Kept) != 0 {
			t.Fatalf("expected all excluded, got %v", keptDetails(got.Kept))
		}
	})
}

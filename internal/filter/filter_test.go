package filter

import (
	"reflect"
	"regexp"
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

	t.Run("no rules: pass through, no annotations", func(t *testing.T) {
		got := Apply(nil, all)
		if got.Total != len(all) {
			t.Fatalf("total=%d, want %d", got.Total, len(all))
		}
		if len(got.Items) != len(all) {
			t.Fatalf("items=%d, want %d", len(got.Items), len(all))
		}
		for i, it := range got.Items {
			if len(it.ExcludedBy) != 0 {
				t.Errorf("items[%d] should have no exclusions, got %v", i, it.ExcludedBy)
			}
		}
		if !reflect.DeepEqual(keptDetails(got.Kept()), keptDetails(all)) {
			t.Fatalf("Kept()=%v, want %v", keptDetails(got.Kept()), keptDetails(all))
		}
	})

	t.Run("exclude rule annotates and drops from Kept", func(t *testing.T) {
		rules := []Rule{
			mustKeyword(t, "no-build", ActionExclude, "meson"),
			mustKeyword(t, "no-docs", ActionExclude, "documentation"),
		}
		got := Apply(rules, all)

		wantKept := []string{
			"Improve pg_dump performance",
			"Add new SQL function jsonb_path_exists",
		}
		if !reflect.DeepEqual(keptDetails(got.Kept()), wantKept) {
			t.Fatalf("Kept()=%v, want %v", keptDetails(got.Kept()), wantKept)
		}
		if got.Hits["no-build"] != 1 || got.Hits["no-docs"] != 1 {
			t.Fatalf("hits=%v", got.Hits)
		}

		// Items must include all releases, with annotations on excluded ones.
		if len(got.Items) != len(all) {
			t.Fatalf("items=%d, want %d", len(got.Items), len(all))
		}
		wantExcl := map[string][]string{
			"Update meson build scripts":            {"no-build"},
			"Improve pg_dump performance":           nil,
			"Fix the documentation for COPY":        {"no-docs"},
			"Add new SQL function jsonb_path_exists": nil,
		}
		for _, it := range got.Items {
			if !reflect.DeepEqual(it.ExcludedBy, wantExcl[it.Release.Detail]) {
				t.Errorf("ExcludedBy for %q = %v, want %v",
					it.Release.Detail, it.ExcludedBy, wantExcl[it.Release.Detail])
			}
		}
	})

	t.Run("multiple rules match same release: all are recorded", func(t *testing.T) {
		rs := releases("Update documentation for the meson build")
		rules := []Rule{
			mustKeyword(t, "no-build", ActionExclude, "meson"),
			mustKeyword(t, "no-docs", ActionExclude, "documentation"),
		}
		got := Apply(rules, rs)
		if len(got.Items) != 1 {
			t.Fatalf("items=%d, want 1", len(got.Items))
		}
		want := []string{"no-build", "no-docs"}
		if !reflect.DeepEqual(got.Items[0].ExcludedBy, want) {
			t.Fatalf("ExcludedBy=%v, want %v", got.Items[0].ExcludedBy, want)
		}
		if got.Hits["no-build"] != 1 || got.Hits["no-docs"] != 1 {
			t.Fatalf("hits=%v", got.Hits)
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
		if !reflect.DeepEqual(keptDetails(gotA.Kept()), keptDetails(gotB.Kept())) {
			t.Fatalf("order matters: A=%v B=%v", keptDetails(gotA.Kept()), keptDetails(gotB.Kept()))
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
		if !reflect.DeepEqual(keptDetails(got.Kept()), keptDetails(all)) {
			t.Fatalf("expected all kept, got %v", keptDetails(got.Kept()))
		}
	})

	t.Run("keyword match is case-insensitive", func(t *testing.T) {
		rs := releases("Refactor MESON build helpers")
		rules := []Rule{
			mustKeyword(t, "ex-build", ActionExclude, "meson"),
		}
		got := Apply(rules, rs)
		if len(got.Kept()) != 0 {
			t.Fatalf("expected all excluded, got %v", keptDetails(got.Kept()))
		}
		if got.Items[0].ExcludedBy[0] != "ex-build" {
			t.Fatalf("ExcludedBy=%v", got.Items[0].ExcludedBy)
		}
	})

	t.Run("regex matcher", func(t *testing.T) {
		rs := releases("On Windows, fix path separator", "On Linux, fix locale")
		rules := []Rule{
			mustRegex(t, "ex-windows", ActionExclude, `(?i)\bon Windows\b`),
		}
		got := Apply(rules, rs)
		if len(got.Kept()) != 1 || got.Kept()[0].Detail != "On Linux, fix locale" {
			t.Fatalf("Kept()=%v", keptDetails(got.Kept()))
		}
	})
}

func TestAnnotateAll(t *testing.T) {
	rs := releases("a", "b", "c")
	got := AnnotateAll(rs)
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}
	for i, it := range got {
		if it.Release.Detail != rs[i].Detail {
			t.Errorf("[%d].Release=%v, want %v", i, it.Release, rs[i])
		}
		if it.ExcludedBy != nil {
			t.Errorf("[%d].ExcludedBy should be nil, got %v", i, it.ExcludedBy)
		}
	}
}

package filter

import "github.com/hidetzu/pg-release-scraper/internal/scraper"

// Annotated wraps a scraper.Release with the IDs of any rules that matched
// it. ExcludedBy is empty for releases that survived filtering.
type Annotated struct {
	Release    scraper.Release
	ExcludedBy []string
}

type Result struct {
	Items []Annotated
	Hits  map[string]int // rule ID -> matched count
	Total int
}

// Kept returns just the releases that were not excluded by any rule, in the
// original input order.
func (r Result) Kept() []scraper.Release {
	out := make([]scraper.Release, 0, len(r.Items))
	for _, it := range r.Items {
		if len(it.ExcludedBy) == 0 {
			out = append(out, it.Release)
		}
	}
	return out
}

// Summary bundles the rules-file path, the loaded rules, and the apply
// result. Output renderers consume this to embed filter metadata; pass nil
// when no filtering was applied.
type Summary struct {
	RulesPath string
	Rules     []Rule
	Result    Result
}

// Apply evaluates exclude rules against releases. A release that matches at
// least one rule is annotated with the matching rule IDs (and dropped from
// Kept()); other releases pass through unchanged.
//
// Hit counts are incremented for every matching rule (a release that
// matches multiple rules counts once per rule), so totals are useful for
// tuning.
func Apply(rules []Rule, releases []scraper.Release) Result {
	hits := make(map[string]int, len(rules))
	for _, r := range rules {
		hits[r.ID] = 0
	}

	items := make([]Annotated, 0, len(releases))
	for _, rel := range releases {
		var excludedBy []string
		for _, r := range rules {
			if r.Matcher.Matches(rel) {
				hits[r.ID]++
				excludedBy = append(excludedBy, r.ID)
			}
		}
		items = append(items, Annotated{Release: rel, ExcludedBy: excludedBy})
	}

	return Result{Items: items, Hits: hits, Total: len(releases)}
}

// AnnotateAll wraps releases as pass-through Annotated values (no rules
// applied). Useful when callers want a uniform []Annotated regardless of
// whether filtering ran.
func AnnotateAll(releases []scraper.Release) []Annotated {
	out := make([]Annotated, len(releases))
	for i, r := range releases {
		out[i] = Annotated{Release: r}
	}
	return out
}

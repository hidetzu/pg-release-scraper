package filter

import "github.com/hidetzu/pg-release-scraper/internal/scraper"

type Result struct {
	Kept  []scraper.Release
	Hits  map[string]int // rule ID -> matched count
	Total int
}

// Summary bundles the rules-file path, the loaded rules, and the apply
// result. Output renderers consume this to embed filter metadata; pass nil
// when no filtering was applied.
type Summary struct {
	RulesPath string
	Rules     []Rule
	Result    Result
}

// Apply evaluates rules against releases using include-then-exclude semantics:
//   - If include rules exist, a release must match at least one to survive.
//   - Exclude rules then drop any surviving release that matches at least one.
//
// Hit counts are incremented for every matching rule (a release that matches
// multiple rules counts once per rule), so the totals are useful for tuning.
func Apply(rules []Rule, releases []scraper.Release) Result {
	hits := make(map[string]int, len(rules))
	for _, r := range rules {
		hits[r.ID] = 0
	}

	var includes, excludes []Rule
	for _, r := range rules {
		switch r.Action {
		case ActionInclude:
			includes = append(includes, r)
		case ActionExclude:
			excludes = append(excludes, r)
		}
	}

	kept := make([]scraper.Release, 0, len(releases))
	for _, rel := range releases {
		if len(includes) > 0 {
			matchedInclude := false
			for _, r := range includes {
				if r.Matcher.Matches(rel) {
					hits[r.ID]++
					matchedInclude = true
				}
			}
			if !matchedInclude {
				continue
			}
		}

		excluded := false
		for _, r := range excludes {
			if r.Matcher.Matches(rel) {
				hits[r.ID]++
				excluded = true
			}
		}
		if !excluded {
			kept = append(kept, rel)
		}
	}

	return Result{
		Kept:  kept,
		Hits:  hits,
		Total: len(releases),
	}
}

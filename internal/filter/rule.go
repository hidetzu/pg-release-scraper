package filter

import (
	"regexp"
	"strings"

	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

type Action string

const (
	ActionInclude Action = "include"
	ActionExclude Action = "exclude"
)

type Kind string

const (
	KindKeyword Kind = "keyword"
	KindRegex   Kind = "regex"
)

type Target string

const (
	TargetDetail Target = "detail"
)

type Matcher interface {
	Matches(scraper.Release) bool
}

type Rule struct {
	ID        string
	Action    Action
	Kind      Kind
	Target    Target
	Value     string
	Rationale string
	Matcher   Matcher
}

type keywordMatcher struct {
	needle string // pre-lowered
}

func (m keywordMatcher) Matches(r scraper.Release) bool {
	return strings.Contains(strings.ToLower(r.Detail), m.needle)
}

type regexMatcher struct {
	re *regexp.Regexp
}

func (m regexMatcher) Matches(r scraper.Release) bool {
	return m.re.MatchString(r.Detail)
}

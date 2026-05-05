package filter

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type rawConfig struct {
	Version int       `yaml:"version"`
	Rules   []rawRule `yaml:"rules"`
}

type rawRule struct {
	ID        string   `yaml:"id"`
	Action    string   `yaml:"action"`
	Match     rawMatch `yaml:"match"`
	Rationale string   `yaml:"rationale,omitempty"`
}

type rawMatch struct {
	Kind   string `yaml:"kind"`
	Target string `yaml:"target"`
	Value  string `yaml:"value"`
}

// LoadRulesFile reads a YAML rules file and returns compiled Rules.
// On any validation error it returns a descriptive error suitable for
// presentation to the user (the caller should exit with code 2).
func LoadRulesFile(path string) ([]Rule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	return loadRules(f, path)
}

func loadRules(r io.Reader, name string) ([]Rule, error) {
	var raw rawConfig
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%s: yaml parse: %w", name, err)
	}
	if raw.Version != 1 {
		return nil, fmt.Errorf("%s: version must be 1, got %d", name, raw.Version)
	}

	seen := make(map[string]struct{}, len(raw.Rules))
	rules := make([]Rule, 0, len(raw.Rules))
	for i, rr := range raw.Rules {
		rule, err := compileRule(rr, i)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if _, dup := seen[rule.ID]; dup {
			return nil, fmt.Errorf("%s: rule[%d] (id=%s): duplicate id", name, i, rule.ID)
		}
		seen[rule.ID] = struct{}{}
		rules = append(rules, rule)
	}
	return rules, nil
}

func compileRule(rr rawRule, idx int) (Rule, error) {
	if rr.ID == "" {
		return Rule{}, fmt.Errorf("rule[%d]: id is required", idx)
	}
	where := fmt.Sprintf("rule[%d] (id=%s)", idx, rr.ID)

	var action Action
	switch rr.Action {
	case string(ActionExclude):
		action = ActionExclude
	case "":
		return Rule{}, fmt.Errorf("%s: action is required", where)
	case "include":
		return Rule{}, fmt.Errorf("%s: action=%q is not supported in v0.2.0 (only %q is supported)", where, rr.Action, ActionExclude)
	default:
		return Rule{}, fmt.Errorf("%s: invalid action %q (must be exclude)", where, rr.Action)
	}

	var kind Kind
	switch rr.Match.Kind {
	case string(KindKeyword):
		kind = KindKeyword
	case string(KindRegex):
		kind = KindRegex
	case "":
		return Rule{}, fmt.Errorf("%s: match.kind is required", where)
	default:
		return Rule{}, fmt.Errorf("%s: invalid match.kind %q (must be keyword or regex)", where, rr.Match.Kind)
	}

	var target Target
	switch rr.Match.Target {
	case string(TargetDetail):
		target = TargetDetail
	case "":
		return Rule{}, fmt.Errorf("%s: match.target is required", where)
	case "version":
		return Rule{}, fmt.Errorf("%s: match.target=%q is not supported in v0.2.0 (only %q is supported)", where, rr.Match.Target, TargetDetail)
	default:
		return Rule{}, fmt.Errorf("%s: invalid match.target %q", where, rr.Match.Target)
	}

	if rr.Match.Value == "" {
		return Rule{}, fmt.Errorf("%s: match.value must not be empty", where)
	}

	var matcher Matcher
	switch kind {
	case KindKeyword:
		matcher = keywordMatcher{needle: strings.ToLower(rr.Match.Value)}
	case KindRegex:
		re, err := regexp.Compile(rr.Match.Value)
		if err != nil {
			return Rule{}, fmt.Errorf("%s: invalid regex: %w", where, err)
		}
		matcher = regexMatcher{re: re}
	}

	return Rule{
		ID:        rr.ID,
		Action:    action,
		Kind:      kind,
		Target:    target,
		Value:     rr.Match.Value,
		Rationale: rr.Rationale,
		Matcher:   matcher,
	}, nil
}

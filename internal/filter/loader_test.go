package filter

import (
	"strings"
	"testing"
)

func loadString(t *testing.T, body string) ([]Rule, error) {
	t.Helper()
	return loadRules(strings.NewReader(body), "test.yaml")
}

func TestLoadRules_OK(t *testing.T) {
	body := `
version: 1
rules:
  - id: ex-build
    action: exclude
    match:
      kind: regex
      target: detail
      value: '(?i)\bmeson\b'
    rationale: |
      Build system noise.
  - id: ex-pgdump
    action: exclude
    match:
      kind: keyword
      target: detail
      value: pg_dump
`
	rules, err := loadString(t, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("rules=%d, want 2", len(rules))
	}
	if rules[0].ID != "ex-build" || rules[0].Action != ActionExclude || rules[0].Kind != KindRegex {
		t.Fatalf("rule[0] = %+v", rules[0])
	}
	if rules[1].ID != "ex-pgdump" || rules[1].Action != ActionExclude || rules[1].Kind != KindKeyword {
		t.Fatalf("rule[1] = %+v", rules[1])
	}
	if rules[0].Rationale == "" {
		t.Fatalf("rationale should be preserved")
	}
}

func TestLoadRules_Errors(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantSub string
	}{
		{
			name: "wrong version",
			body: `
version: 2
rules: []
`,
			wantSub: "version must be 1",
		},
		{
			name: "duplicate id",
			body: `
version: 1
rules:
  - id: dup
    action: exclude
    match: {kind: keyword, target: detail, value: x}
  - id: dup
    action: exclude
    match: {kind: keyword, target: detail, value: y}
`,
			wantSub: "duplicate id",
		},
		{
			name: "empty id",
			body: `
version: 1
rules:
  - action: exclude
    match: {kind: keyword, target: detail, value: x}
`,
			wantSub: "id is required",
		},
		{
			name: "invalid action",
			body: `
version: 1
rules:
  - id: r1
    action: maybe
    match: {kind: keyword, target: detail, value: x}
`,
			wantSub: `invalid action "maybe"`,
		},
		{
			name: "include action not supported",
			body: `
version: 1
rules:
  - id: r1
    action: include
    match: {kind: keyword, target: detail, value: x}
`,
			wantSub: `not supported in v0.2.0`,
		},
		{
			name: "missing action",
			body: `
version: 1
rules:
  - id: r1
    match: {kind: keyword, target: detail, value: x}
`,
			wantSub: "action is required",
		},
		{
			name: "invalid kind",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: glob, target: detail, value: x}
`,
			wantSub: `invalid match.kind "glob"`,
		},
		{
			name: "missing kind",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {target: detail, value: x}
`,
			wantSub: "match.kind is required",
		},
		{
			name: "target=version not supported in v0.2.0",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: keyword, target: version, value: "15.6"}
`,
			wantSub: `not supported in v0.2.0`,
		},
		{
			name: "invalid target",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: keyword, target: title, value: x}
`,
			wantSub: `invalid match.target "title"`,
		},
		{
			name: "missing target",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: keyword, value: x}
`,
			wantSub: "match.target is required",
		},
		{
			name: "empty value",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: keyword, target: detail, value: ""}
`,
			wantSub: "match.value must not be empty",
		},
		{
			name: "invalid regex",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: regex, target: detail, value: "["}
`,
			wantSub: "invalid regex",
		},
		{
			name: "unknown field rejected",
			body: `
version: 1
rules:
  - id: r1
    action: exclude
    match: {kind: keyword, target: detail, value: x}
    extra: oops
`,
			wantSub: "yaml parse",
		},
		{
			name:    "broken yaml",
			body:    "version: 1\nrules: [: ]",
			wantSub: "yaml parse",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadString(t, tc.body)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestLoadRules_EmptyRules(t *testing.T) {
	rules, err := loadString(t, "version: 1\nrules: []\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("expected zero rules, got %d", len(rules))
	}
}

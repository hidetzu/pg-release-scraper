package scraper

import (
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"two parts", "14.0", []int{14, 0}, false},
		{"three parts", "9.6.24", []int{9, 6, 24}, false},
		{"single part", "15", []int{15}, false},
		{"surrounding whitespace", " 14.0 ", []int{14, 0}, false},
		{"empty", "", nil, true},
		{"non-numeric", "abc", nil, true},
		{"non-numeric part", "14.x", nil, true},
		{"trailing dot", "14.", nil, true},
		{"leading dot", ".14", nil, true},
		{"empty middle part", "14..0", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got.Parts, tt.want) {
				t.Errorf("ParseVersion(%q).Parts = %v, want %v", tt.input, got.Parts, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"14.0", "14.0", 0},
		{"14.0", "15.0", -1},
		{"15.0", "14.0", 1},
		{"14.0", "14.5", -1},
		{"14.5", "14.0", 1},
		{"9.6", "9.6.24", -1},
		{"9.6.24", "9.6", 1},
		{"14", "14.5", -1},
		{"14", "14.0", 0},
		{"14.0.0", "14", 0},
		{"15.6", "15.6.0", 0},
		{"10.0", "9.6.24", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a, err := ParseVersion(tt.a)
			if err != nil {
				t.Fatalf("ParseVersion(%q): %v", tt.a, err)
			}
			b, err := ParseVersion(tt.b)
			if err != nil {
				t.Fatalf("ParseVersion(%q): %v", tt.b, err)
			}
			got := sign(a.Compare(b))
			if got != tt.want {
				t.Errorf("(%q).Compare(%q) sign = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

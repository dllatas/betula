package lib

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLayoutForUnit(t *testing.T) {
	type LayoutTest struct {
		Unit     TimeUnit
		Expected string
		Sample   time.Time
		Want     string
	}

	sample := time.Date(2025, 5, 19, 13, 45, 12, 123456789, time.UTC)

	tests := []LayoutTest{
		{Unit: UnitYear, Expected: "2006", Sample: sample, Want: "2025"},
		{Unit: UnitMonth, Expected: "2006-01", Sample: sample, Want: "2025-05"},
		{Unit: UnitDay, Expected: "2006-01-02", Sample: sample, Want: "2025-05-19"},
		{Unit: UnitHour, Expected: "2006-01-02 15", Sample: sample, Want: "2025-05-19 13"},
		{Unit: UnitMinute, Expected: "2006-01-02 15:04", Sample: sample, Want: "2025-05-19 13:45"},
		{Unit: UnitSecond, Expected: "2006-01-02 15:04:05", Sample: sample, Want: "2025-05-19 13:45:12"},
		{Unit: UnitMilli, Expected: "2006-01-02 15:04:05.000", Sample: sample, Want: "2025-05-19 13:45:12.123"},
		{Unit: UnitMicro, Expected: "2006-01-02 15:04:05.000000", Sample: sample, Want: "2025-05-19 13:45:12.123456"},
		{Unit: UnitNano, Expected: "2006-01-02 15:04:05.000000000", Sample: sample, Want: "2025-05-19 13:45:12.123456789"},
	}

	for _, tt := range tests {
		layout := layoutForUnit(tt.Unit)

		if layout != tt.Expected {
			t.Errorf("Unit %s: expected layout %q, got %q", tt.Unit, tt.Expected, layout)
		}

		formatted := tt.Sample.Format(layout)
		if formatted != tt.Want {
			t.Errorf("Unit %s: expected formatted %q, got %q", tt.Unit, tt.Want, formatted)
		}

		parsed, err := time.Parse(layout, formatted)
		if err != nil {
			t.Errorf("Unit %s: failed to parse back %q: %v", tt.Unit, formatted, err)
			continue
		}

		roundTrip := parsed.Format(layout)
		if roundTrip != formatted {
			t.Errorf("Unit %s: round-trip mismatch: got %q, want %q", tt.Unit, roundTrip, formatted)
		}
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		units    []TimeUnit
		element  TimeUnit
		expected int
	}{
		{
			name:     "element present at start",
			units:    []TimeUnit{UnitYear, UnitMonth, UnitDay},
			element:  UnitYear,
			expected: 0,
		},
		{
			name:     "element present in middle",
			units:    []TimeUnit{UnitYear, UnitMonth, UnitDay},
			element:  UnitMonth,
			expected: 1,
		},
		{
			name:     "element present at end",
			units:    []TimeUnit{UnitYear, UnitMonth, UnitDay},
			element:  UnitDay,
			expected: 2,
		},
		{
			name:     "element not present",
			units:    []TimeUnit{UnitYear, UnitMonth},
			element:  UnitDay,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexOf(tt.units, tt.element)
			if got != tt.expected {
				t.Fatalf("indexOf() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestIsUnitCompatible(t *testing.T) {
	tests := []struct {
		name      string
		viewUnit  TimeUnit
		queryUnit TimeUnit
		want      bool
	}{
		{
			name:      "same unit is compatible",
			viewUnit:  UnitSecond,
			queryUnit: UnitSecond,
			want:      true,
		},
		{
			name:      "coarser query unit is compatible",
			viewUnit:  UnitSecond,
			queryUnit: UnitMinute,
			want:      true,
		},
		{
			name:      "finer query unit is not compatible",
			viewUnit:  UnitMinute,
			queryUnit: UnitSecond,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnitCompatible(tt.viewUnit, tt.queryUnit)
			if got != tt.want {
				t.Fatalf("IsUnitCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewViewDefinitionAndString(t *testing.T) {
	keys := []string{"club", "player", "goals"}
	v := NewViewDefinition("test-view", keys, UnitHour)

	if v.Name != "test-view" {
		t.Fatalf("expected name %q, got %q", "test-view", v.Name)
	}
	if len(v.KeyOrder) != len(keys) {
		t.Fatalf("expected %d keys, got %d", len(keys), len(v.KeyOrder))
	}
	if v.Granularity != UnitHour {
		t.Fatalf("expected granularity %q, got %q", UnitHour, v.Granularity)
	}

	s := v.String()
	expectedFragments := []string{
		"View[test-view]",
		"club",
		"player",
		"goals",
		string(UnitHour),
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(s, fragment) {
			t.Fatalf("String() output %q does not contain %q", s, fragment)
		}
	}
}

func TestParseUnit(t *testing.T) {
	tests := []struct {
		input       string
		want        TimeUnit
		expectError bool
	}{
		{input: "year", want: UnitYear},
		{input: "MONTH", want: UnitMonth},
		{input: "Day", want: UnitDay},
		{input: "hour", want: UnitHour},
		{input: "minute", want: UnitMinute},
		{input: "second", want: UnitSecond},
		{input: "millisecond", want: UnitMilli},
		{input: "microsecond", want: UnitMicro},
		{input: "nanosecond", want: UnitNano},
		{input: "", expectError: true},
		{input: "invalid", expectError: true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parse_%s", tt.input), func(t *testing.T) {
			got, err := ParseUnit(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseUnit(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

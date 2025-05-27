package lib

import (
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

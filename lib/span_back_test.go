package lib

import (
	"fmt"
	"testing"
	"time"
)

func TestTruncateBack(t *testing.T) {
	type Case struct {
		Ref     time.Time
		Unit    TimeUnit
		Count   int
		Want    time.Time
		Message string
	}

	ref := time.Date(2025, 5, 19, 13, 45, 12, 123456789, time.UTC)

	cases := []Case{
		{
			Ref:     ref,
			Unit:    UnitYear,
			Count:   2,
			Want:    time.Date(2023, 5, 19, 13, 45, 12, 123456789, time.UTC),
			Message: "2 years back",
		},
		{
			Ref:     ref,
			Unit:    UnitMonth,
			Count:   3,
			Want:    time.Date(2025, 2, 19, 13, 45, 12, 123456789, time.UTC),
			Message: "3 months back",
		},
		{
			Ref:     ref,
			Unit:    UnitDay,
			Count:   5,
			Want:    time.Date(2025, 5, 14, 13, 45, 12, 123456789, time.UTC),
			Message: "5 days back",
		},
		{
			Ref:     ref,
			Unit:    UnitHour,
			Count:   1,
			Want:    ref.Add(-1 * time.Hour),
			Message: "1 hour back",
		},
		{
			Ref:     ref,
			Unit:    UnitMinute,
			Count:   30,
			Want:    ref.Add(-30 * time.Minute),
			Message: "30 minutes back",
		},
		{
			Ref:     ref,
			Unit:    UnitSecond,
			Count:   10,
			Want:    ref.Add(-10 * time.Second),
			Message: "10 seconds back",
		},
		{
			Ref:     ref,
			Unit:    UnitMilli,
			Count:   500,
			Want:    ref.Add(-500 * time.Millisecond),
			Message: "500 milliseconds back",
		},
		{
			Ref:     ref,
			Unit:    UnitMicro,
			Count:   1000,
			Want:    ref.Add(-1000 * time.Microsecond),
			Message: "1000 microseconds back",
		},
		{
			Ref:     ref,
			Unit:    UnitNano,
			Count:   42,
			Want:    ref.Add(-42 * time.Nanosecond),
			Message: "42 nanoseconds back",
		},
	}

	for _, c := range cases {
		got, err := truncateBack(c.Ref, c.Unit, c.Count)
		if err != nil {
			t.Fatal(err.Error())
		}

		if !got.Equal(c.Want) {
			t.Errorf("%s: expected %v, got %v", c.Message, c.Want, got)
		}
	}
}

func TestTruncateBack_UnsupportedUnit(t *testing.T) {
	ref := time.Now()
	_, err := truncateBack(ref, TimeUnit("unsupported"), 1)
	if err == nil {
		t.Fatal("expected error for unsupported unit, got nil")
	}
}

func TestSpanBack(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	// Add 3 events over 3 days
	now := time.Date(2025, 5, 22, 9, 0, 0, 0, time.UTC)
	for i := range 3 {
		ts := now.AddDate(0, 0, -i)
		err := instance.Append(Event{
			Timestamp: ts,
			Labels:    map[string]string{"userid": "u1"},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Span back 2 days from now (inclusive window: today and yesterday)
	shards, err := instance.SpanBack(now, UnitDay, 2, verbose)
	if err != nil {
		t.Fatal(err)
	}

	if len(shards) != 2 {
		instance.Print("")
		for _, shard := range shards {
			fmt.Printf("(%+v)\n", shard)
		}
		t.Errorf("expected 2 shards, got %d", len(shards))
	}
}

func TestSpanBack_LastSevenDays_InclusiveWindow(t *testing.T) {
	view := NewViewDefinition("habito-user-habit", []string{"userid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	// Reference: Oct 24, 2025 at 10:00 UTC
	now := time.Date(2025, 10, 24, 10, 0, 0, 0, time.UTC)

	// Add one event for each day from Oct 17..Oct 24
	for i := 0; i <= 7; i++ {
		ts := time.Date(2025, 10, 24-i, 9, 0, 0, 0, time.UTC)
		if err := instance.Append(Event{Timestamp: ts, Labels: map[string]string{"userid": "u1"}}); err != nil {
			t.Fatalf("append failed for day offset %d: %v", i, err)
		}
	}

	// Ask for last 7 days (inclusive): expect 7 shards covering Oct 18..Oct 24
	shards, err := instance.SpanBack(now, UnitDay, 7, verbose)
	if err != nil {
		t.Fatalf("spanback failed: %v", err)
	}

	if len(shards) != 7 {
		instance.Print("spanback-7d")
		for _, shard := range shards {
			fmt.Printf("(%+v)\n", shard)
		}
		t.Fatalf("expected 7 shards, got %d", len(shards))
	}

	// Verify the first shard is Oct 18 and last is Oct 24
	expectedStart := time.Date(2025, 10, 18, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2025, 10, 24, 0, 0, 0, 0, time.UTC)
	if !shards[0].Key.Equal(expectedStart) {
		t.Errorf("unexpected first shard: got %s want %s", shards[0].Key, expectedStart)
	}
	if !shards[len(shards)-1].Key.Equal(expectedEnd) {
		t.Errorf("unexpected last shard: got %s want %s", shards[len(shards)-1].Key, expectedEnd)
	}
}

func TestSpanBack_OneDayOnHourlyView(t *testing.T) {
	view := NewViewDefinition("hourly-view", []string{"userid"}, UnitHour)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	// Start of day in UTC (to avoid time zone hell)
	now := time.Date(2025, 5, 22, 23, 59, 0, 0, time.UTC)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	user := "alice"

	// Add 24 events, one per hour
	for i := range 24 {
		ts := startOfDay.Add(time.Duration(i) * time.Hour)
		err := instance.Append(Event{
			Timestamp: ts,
			Labels: map[string]string{
				"userid": user,
			},
		})
		if err != nil {
			t.Fatalf("append failed at hour %d: %+v", i, err)
		}
	}

	// Perform SpanBack with UnitDay
	shards, err := instance.SpanBack(now, UnitDay, 1, verbose)
	if err != nil {
		t.Fatalf("spanback failed: %+v", err)
	}

	if len(shards) != 24 {
		instance.Print("spanback")
		for _, shard := range shards {
			fmt.Printf("(%+v)\n", shard)
		}
		t.Errorf("expected 24 shards (1 per hour), got %d", len(shards))
	}

	// Optional: verify shard keys are hour-aligned
	for i, shard := range shards {
		expected := startOfDay.Add(time.Duration(i) * time.Hour)
		if shard.Key != expected {
			t.Errorf("shard[%d] key mismatch: got %s, want %s", i, shard.Key, expected)
		}
	}
}

func TestSpanBack_Errors(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	ref := time.Date(2025, 5, 22, 9, 0, 0, 0, time.UTC)

	// incompatible units: view granularity day, query hour (finer)
	if _, err := instance.SpanBack(ref, UnitHour, 1, verbose); err == nil {
		t.Fatal("expected error for incompatible units, got nil")
	}

	// zero duration
	if _, err := instance.SpanBack(ref, UnitDay, 0, verbose); err == nil {
		t.Fatal("expected error for zero duration, got nil")
	}

	// unsupported unit flows into truncateToUnit/truncateBack error
	if _, err := instance.SpanBack(ref, TimeUnit("unsupported"), 1, verbose); err == nil {
		t.Fatal("expected error for unsupported unit, got nil")
	}
}

func TestTruncateToUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		unit     TimeUnit
		expected time.Time
	}{
		{
			name:     "Year",
			unit:     UnitYear,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Month",
			unit:     UnitMonth,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Day",
			unit:     UnitDay,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 27, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Hour",
			unit:     UnitHour,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 0, 0, 0, time.UTC),
		},
		{
			name:     "Minute",
			unit:     UnitMinute,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 8, 0, 0, time.UTC),
		},
		{
			name:     "Second",
			unit:     UnitSecond,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 8, 55, 0, time.UTC),
		},
		{
			name:     "Milli",
			unit:     UnitMilli,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 8, 55, 123000000, time.UTC),
		},
		{
			name:     "Micro",
			unit:     UnitMicro,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456789, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 8, 55, 123456000, time.UTC),
		},
		{
			name:     "Nano",
			unit:     UnitNano,
			input:    time.Date(2025, 5, 27, 11, 8, 55, 123456789, time.UTC),
			expected: time.Date(2025, 5, 27, 11, 8, 55, 123456789, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := truncateToUnit(tt.input, tt.unit)
			if err != nil {
				t.Fatal(err.Error())
			}
			if !got.Equal(tt.expected) {
				t.Errorf("%s failed:\nwant %s\ngot  %s", tt.name, tt.expected, got)
			}
		})
	}
}

func TestTruncateToUnit_UnsupportedUnit(t *testing.T) {
	now := time.Now()
	if _, err := truncateToUnit(now, TimeUnit("unsupported")); err == nil {
		t.Fatal("expected error for unsupported unit, got nil")
	}
}

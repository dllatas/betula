package lib

import (
	"log/slog"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	view := NewViewDefinition("habit", []string{"userid", "habitid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)

	// Common time for all shards
	now := time.Date(2025, 5, 27, 0, 0, 0, 0, time.UTC)

	// Insert events
	events := []Event{
		{Timestamp: now, Labels: map[string]string{"userid": "alice", "habitid": "run"}},
		{Timestamp: now, Labels: map[string]string{"userid": "alice", "habitid": "read"}},
		{Timestamp: now, Labels: map[string]string{"userid": "bob", "habitid": "run"}},
		{Timestamp: now, Labels: map[string]string{"userid": "carol", "habitid": "write"}},
	}

	for _, e := range events {
		if err := instance.Append(e); err != nil {
			t.Fatalf("append failed: %+v", err)
		}
	}

	tests := []struct {
		name     string
		match    map[string][]string
		expected int
	}{
		{
			name:     "Multiple key match",
			match:    map[string][]string{"userid": {"alice"}, "habitid": {"run"}},
			expected: 1,
		},
		{
			name:     "Single key match",
			match:    map[string][]string{"userid": {"alice"}},
			expected: 1, // One shard, matching entries inside
		},
		{
			name:     "Multiple values per key",
			match:    map[string][]string{"userid": {"alice", "bob"}},
			expected: 1,
		},
		{
			name:     "No match",
			match:    map[string][]string{"userid": {"nonexistent"}},
			expected: 0,
		},
		{
			name:     "Empty match input",
			match:    map[string][]string{},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := instance.Filter(tt.match)
			if err != nil {
				t.Fatalf("filter failed: %+v", err)
			}

			if len(filtered) != tt.expected {
				// instance.Print(tt.name)
				slog.Warn("not same length", "filtered", filtered, "expected", tt.expected, "match", tt.match)
				t.Errorf("expected %d shards, got %d", tt.expected, len(filtered))
			}

			t.Log(tt.match)
			printShards(filtered, "")
		})
	}
}

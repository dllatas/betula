package lib

import (
	"fmt"
	"testing"
	"time"
)

type RangeTest struct {
	Desc      string
	From      time.Time
	To        time.Time
	Expected  []string
	ExpectErr bool
}

func TestViewInstanceRange(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, "2006-01-02")
	mapper := NewViewMapper(view)
	instance := NewViewInstance(view, mapper)

	// Insert 5 days of events
	base := time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		ts := base.AddDate(0, 0, i)
		instance.Append(Event{
			Timestamp: ts,
			Labels:    map[string]string{"userid": fmt.Sprintf("u%d", i)},
		})
	}

	tests := []RangeTest{
		{
			Desc:     "Exact match",
			From:     base.AddDate(0, 0, 1),
			To:       base.AddDate(0, 0, 3),
			Expected: []string{"2025-05-21", "2025-05-22", "2025-05-23"},
		},
		{
			Desc:     "Full range",
			From:     base,
			To:       base.AddDate(0, 0, 4),
			Expected: []string{"2025-05-20", "2025-05-21", "2025-05-22", "2025-05-23", "2025-05-24"},
		},
		{
			Desc:     "Single day",
			From:     base.AddDate(0, 0, 2),
			To:       base.AddDate(0, 0, 2),
			Expected: []string{"2025-05-22"},
		},
		{
			Desc:     "Before range",
			From:     base.AddDate(0, 0, -5),
			To:       base.AddDate(0, 0, -1),
			Expected: []string{},
		},
		{
			Desc:     "After range",
			From:     base.AddDate(0, 0, 5),
			To:       base.AddDate(0, 0, 10),
			Expected: []string{},
		},
		{
			Desc:      "Invalid range",
			From:      base.AddDate(0, 0, 3),
			To:        base.AddDate(0, 0, 1),
			ExpectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Desc, func(t *testing.T) {
			result, err := instance.Range(test.From, test.To)
			if test.ExpectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(test.Expected) {
				t.Fatalf("len: expected %d shards, got %d", len(test.Expected), len(result))
			}

			for i, shard := range result {
				if shard.Key != test.Expected[i] {
					t.Errorf("value: expected shard %s, got %s", test.Expected[i], shard.Key)
				}
			}
		})
	}
}

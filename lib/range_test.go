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
	Expected  []time.Time
	ExpectErr bool
}

func TestViewInstanceRange(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, "2006-01-02")
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

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
			Desc: "Exact match",
			From: base.AddDate(0, 0, 1),
			To:   base.AddDate(0, 0, 3),
			Expected: []time.Time{
				time.Date(2025, 5, 21, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			Desc: "Full range",
			From: base,
			To:   base.AddDate(0, 0, 4),
			Expected: []time.Time{
				time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 21, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 23, 0, 0, 0, 0, time.UTC),
				time.Date(2025, 5, 24, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			Desc: "Single day",
			From: base.AddDate(0, 0, 2),
			To:   base.AddDate(0, 0, 2),
			Expected: []time.Time{
				time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			Desc:     "Before range",
			From:     base.AddDate(0, 0, -5),
			To:       base.AddDate(0, 0, -1),
			Expected: []time.Time{},
		},
		{
			Desc:     "After range",
			From:     base.AddDate(0, 0, 5),
			To:       base.AddDate(0, 0, 10),
			Expected: []time.Time{},
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
			result, err := instance.Range(test.From, test.To, verbose)
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
				instance.Print(test.Desc)
				t.Fatalf("len: expected %d shards, got %d. wanted %+v got %+v. From %+v To %+v", len(test.Expected), len(result), test.Expected, result, test.From, test.To)
			}

			for i, shard := range result {
				if shard.Key != test.Expected[i] {
					t.Errorf("value: expected shard %s, got %s", test.Expected[i], shard.Key)
				}
			}
		})
	}
}

func TestFindClosestLaterIndex(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	// Insert 5 days of events
	base := time.Date(2025, 5, 20, 0, 0, 0, 0, time.UTC)
	for i := range 5 {
		ts := base.AddDate(0, 0, i)
		instance.Append(Event{
			Timestamp: ts,
			Labels:    map[string]string{"userid": fmt.Sprintf("u%d", i)},
		})
	}

	yesterday := time.Date(2025, 5, 19, 0, 0, 0, 0, time.UTC)

	if verbose {
		instance.Print("findClosestLaterIndex")
	}

	wanted := 0
	got := instance.findCloserIndex(yesterday, ">", verbose)

	if got != wanted {
		t.Errorf("[yesterday] found wrong index wanted %d got %d", wanted, got)
	}

	daysLater := time.Date(2025, 5, 26, 0, 0, 0, 0, time.UTC)

	wanted = 4
	got = instance.findCloserIndex(daysLater, "<", verbose)

	if got != wanted {
		t.Errorf("[daysLater] found wrong index wanted %d got %d", wanted, got)
	}
}

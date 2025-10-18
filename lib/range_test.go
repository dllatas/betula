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
	for i := range 5 {
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

	// Between existing shards: pick next greater for ">" and previous lesser for "<"
	mid := time.Date(2025, 5, 21, 12, 0, 0, 0, time.UTC) // between 21 and 22

	// expect index 2 (2025-05-22) for ">"
	wanted = 2
	got = instance.findCloserIndex(mid, ">", verbose)
	if got != wanted {
		t.Errorf("[between, >] wrong index wanted %d got %d", wanted, got)
	}

	// expect index 1 (2025-05-21) for "<"
	wanted = 1
	got = instance.findCloserIndex(mid, "<", verbose)
	if got != wanted {
		t.Errorf("[between, <] wrong index wanted %d got %d", wanted, got)
	}
}

func TestRange_MissingFromAlignsToNextGreater(t *testing.T) {
	view := NewViewDefinition("test", []string{"userid"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)
	verbose := false

	// Create shards on 9,10,12,13 (missing 11)
	base := time.Date(2025, 10, 9, 0, 0, 0, 0, time.UTC)
	for _, offset := range []int{0, 1, 3, 4} { // 9,10,12,13
		_ = instance.Append(Event{
			Timestamp: base.AddDate(0, 0, offset),
			Labels:    map[string]string{"userid": "u"},
		})
	}

	from := time.Date(2025, 10, 11, 22, 0, 0, 0, time.UTC) // aligns to 2025-10-11 (missing)
	to := time.Date(2025, 10, 13, 10, 0, 0, 0, time.UTC)   // aligns to 2025-10-13 (present)

	shards, err := instance.Range(from, to, verbose)
	if err != nil {
		t.Fatalf("range failed: %+v", err)
	}

	if len(shards) != 2 {
		instance.Print("TestRange_MissingFromAlignsToNextGreater")
		t.Fatalf("expected 2 shards (12 and 13), got %d", len(shards))
	}

	if shards[0].Key != time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC) {
		t.Errorf("first shard mismatch: got %s", shards[0].Key)
	}
	if shards[1].Key != time.Date(2025, 10, 13, 0, 0, 0, 0, time.UTC) {
		t.Errorf("second shard mismatch: got %s", shards[1].Key)
	}
}

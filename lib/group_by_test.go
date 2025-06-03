package lib

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIsValidGroupBy(t *testing.T) {
	v := &ViewDefinition{
		KeyOrder: []string{"userid", "habitid", "locationid"},
	}

	type groupByKey struct {
		groupByKeys []string
		valid       bool
		desc        string
	}

	groupByKeys := []groupByKey{
		{
			[]string{},
			false,
			"empty keys: no grouping",
		},
		{
			[]string{"userid"},
			true,
			"valid: top-level key",
		},
		{
			[]string{"userid", "habitid"},
			true,
			"valid: key order",
		},
		{
			[]string{"habitid"},
			true,
			"valid: lower-level key in correct order",
		},
		{
			[]string{"habitid", "locationid"},
			true,
			"valid: mid and lower keys in correct order",
		},
		{
			[]string{"locationid", "habitid"},
			false,
			"invalid: out of order",
		},
		{
			[]string{"userid", "locationid"},
			true,
			"valid: skips habitid but maintains order",
		},
		{
			[]string{"shard"},
			true,
			"valid: group by virtual shard",
		},
		{
			[]string{"shard", "userid"},
			true,
			"valid: shard plus top level",
		},
		{
			[]string{"shard", "habitid"},
			true,
			"valid: shard plus mid level",
		},
		{
			[]string{"habitid", "shard"},
			false,
			"invalid: shard must always be first if present",
		},
		{
			[]string{"shard", "locationid"},
			true,
			"valid: shard plus lowest-level",
		},
		{
			[]string{"other"},
			false,
			"invalid: key not in definition",
		},
	}

	for _, groupByKey := range groupByKeys {
		t.Run(fmt.Sprintf("groupBy(%v) -> %s", groupByKey.groupByKeys, groupByKey.desc), func(t *testing.T) {
			got := v.IsValidGroupBy(groupByKey.groupByKeys)
			if got != groupByKey.valid {
				t.Errorf("expected %v for groupByKeys=%v (%s), got %v", groupByKey.valid, groupByKey.groupByKeys, groupByKey.desc, got)
			}
		})
	}
}

func TestGroupByBasic(t *testing.T) {
	d := NewViewDefinition("basic", []string{"userid", "habitid"}, "day")

	m := NewViewMapper(d)

	i := NewViewInstance(m)

	ev := Event{
		Timestamp: time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		Labels: map[string]string{
			"userid":  "alice",
			"habitid": "read",
		},
	}

	bobEv := Event{
		Timestamp: time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		Labels: map[string]string{
			"userid":  "bob",
			"habitid": "run",
		},
	}

	err := i.Append(ev)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = i.Append(ev)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = i.Append(bobEv)
	if err != nil {
		t.Fatal(err.Error())
	}

	type GroupByBasicTest struct {
		desc    string
		groupBy []string
		want    map[string]int
	}

	basicTests := []GroupByBasicTest{
		{
			"group by userid only",
			[]string{"userid"},
			map[string]int{
				"alice": 2,
				"bob":   1,
			},
		},
		{
			"group by user and habit",
			[]string{"userid", "habitid"},
			map[string]int{
				"alice.read": 2,
				"bob.run":    1,
			},
		},
		{
			"group by habit only",
			[]string{"habitid"},
			map[string]int{
				"read": 2,
				"run":  1,
			},
		},
		{
			"group by shard and user",
			[]string{"shard", "userid"},
			map[string]int{
				"2025-06-02.alice": 2,
				"2025-06-02.bob":   1,
			},
		},
		{
			"group by full path with shard",
			[]string{"shard", "userid", "habitid"},
			map[string]int{
				"2025-06-02.alice.read": 2,
				"2025-06-02.bob.run":    1,
			},
		},
		{
			"invalid group by key",
			[]string{"invalid"},
			nil,
		},
	}

	for _, basicTest := range basicTests {
		t.Run(basicTest.desc, func(t *testing.T) {
			got, err := i.GroupBy(basicTest.groupBy)

			if basicTest.want == nil {
				if err == nil {
					t.Fatalf("expected error for groupBy %v, got nil", basicTest.groupBy)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, basicTest.want) {
				t.Errorf("unexpected result:\ngot:  %v\nwant: %v", got, basicTest.want)
			}
		})
	}
}

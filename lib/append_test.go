package lib

import (
	"fmt"
	"testing"
	"time"
)

func TestAppendBasic(t *testing.T) {
	view := NewViewDefinition(
		"habit_view",
		[]string{"user", "habit", "event"},
		"2006-01-02",
	)
	mapper := NewViewMapper(view)

	instance := NewViewInstance(mapper)

	ev := Event{
		Timestamp: time.Date(2024, 5, 19, 12, 0, 0, 0, time.UTC),
		Labels: map[string]string{
			"user":  "alice",
			"habit": "run",
			"event": "done",
		},
	}

	if err := instance.Append(ev); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Check that shard was created
	partitionKey := "2024-05-19"
	idx, ok := instance.ShardIndex[partitionKey]
	if !ok {
		t.Fatalf("shard for %s not found", partitionKey)
	}
	shard := instance.Shards[idx]

	// Walk the tree using composite keys
	rootKey := "user:alice"
	hKey := "habit:run"
	eKey := "event:done"

	userNode, ok := shard.Roots[rootKey]
	if !ok {
		t.Fatal("missing root node for user:alice")
	}
	if userNode.Count != 1 {
		t.Errorf("expected user node count 0, got %d", userNode.Count)
	}

	habitNode, ok := userNode.Nodes[hKey]
	if !ok {
		t.Fatal("missing habit node for habit:run")
	}
	if habitNode.Count != 1 {
		t.Errorf("expected habit node count 0, got %d", habitNode.Count)
	}

	eventNode, ok := habitNode.Nodes[eKey]
	if !ok {
		t.Fatal("missing event node for event:done")
	}
	if eventNode.Count != 1 {
		t.Errorf("expected event node count 1, got %d", eventNode.Count)
	}
}

type TestStep struct {
	GetEvents func() []*Event
	Wanted    *ViewInstance
}

type TestAppendRun struct {
	Desc  string
	Steps []TestStep
}

type PartitionTime struct {
	Time       time.Time // original
	Label      string    // formatted
	Normalized time.Time // parsed version of label (stripped time-of-day)
}

func NewPartitionTime(t time.Time, layout string) (PartitionTime, error) {
	label := t.Format(layout)
	normalized, err := time.Parse(layout, label)
	if err != nil {
		return PartitionTime{}, fmt.Errorf("failed to parse layout: %s for value %s: %w", layout, label, err)
	}

	return PartitionTime{
		Time:       t,
		Label:      label,
		Normalized: normalized,
	}, nil
}

func TestAppend(t *testing.T) {
	userID := "129e7245-17c8-4abf-a235-8b49e534046f"
	habitID := "a71e1998-30b6-4e6b-a9ec-779e00b022d4"
	timeLayout := "2006-01-02"

	now, err := NewPartitionTime(time.Now(), timeLayout)
	if err != nil {
		t.Error(err.Error())
	}

	tomorrow, err := NewPartitionTime(now.Time.Add(time.Hour*24), timeLayout)
	if err != nil {
		t.Error(err.Error())
	}

	yesterday, err := NewPartitionTime(now.Time.Add(time.Hour*-24), timeLayout)
	if err != nil {
		t.Error(err.Error())
	}

	view := NewViewDefinition(
		"test_view",
		[]string{"userid", "habitid"},
		UnitDay,
	)
	mapper := NewViewMapper(view)

	runs := []TestAppendRun{
		{
			Desc: "Append a single event",
			Steps: []TestStep{
				{
					GetEvents: func() []*Event {
						return []*Event{
							{
								Timestamp: now.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
						}
					},
					Wanted: &ViewInstance{
						Mapper: mapper,
						ShardIndex: map[string]int{
							now.Label: 0,
						},
						Shards: []*Shard{
							{
								Key:       now.Label,
								ParsedKey: now.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Desc: "Append two events in different dates",
			Steps: []TestStep{
				{
					GetEvents: func() []*Event {
						return []*Event{
							{
								Timestamp: now.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
						}
					},
					Wanted: &ViewInstance{
						Mapper: mapper,
						ShardIndex: map[string]int{
							now.Label: 0,
						},
						Shards: []*Shard{
							{
								Key:       now.Label,
								ParsedKey: now.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					GetEvents: func() []*Event {
						return []*Event{
							{
								Timestamp: tomorrow.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
						}
					},
					Wanted: &ViewInstance{
						Mapper: mapper,
						ShardIndex: map[string]int{
							now.Label:      0,
							tomorrow.Label: 1,
						},
						Shards: []*Shard{
							{
								Key:       now.Label,
								ParsedKey: now.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
							{
								Key:       tomorrow.Label,
								ParsedKey: tomorrow.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Desc: "Append two events in a single date",
			Steps: []TestStep{
				{
					GetEvents: func() []*Event {
						return []*Event{
							{
								Timestamp: now.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
							{
								Timestamp: now.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
						}
					},
					Wanted: &ViewInstance{
						Mapper: mapper,
						ShardIndex: map[string]int{
							now.Label: 0,
						},
						Shards: []*Shard{
							{
								Key:       now.Label,
								ParsedKey: now.Normalized,
								Count:     2,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 2,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 2,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Desc: "Append events not in chronological order",
			Steps: []TestStep{
				{
					GetEvents: func() []*Event {
						return []*Event{
							{
								Timestamp: tomorrow.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
							{
								Timestamp: yesterday.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
							{
								Timestamp: now.Time,
								Labels: map[string]string{
									"userid":  userID,
									"habitid": habitID,
								},
							},
						}
					},
					Wanted: &ViewInstance{
						Mapper: mapper,
						ShardIndex: map[string]int{
							yesterday.Label: 0,
							now.Label:       1,
							tomorrow.Label:  2,
						},
						Shards: []*Shard{
							{
								Key:       yesterday.Label,
								ParsedKey: yesterday.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
							{
								Key:       now.Label,
								ParsedKey: now.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
							{
								Key:       tomorrow.Label,
								ParsedKey: tomorrow.Normalized,
								Count:     1,
								Roots: map[string]*Node{
									"userid:" + userID: {
										Key:   "userid",
										Value: userID,
										Count: 1,
										Nodes: map[string]*Node{
											"habitid:" + habitID: {
												Key:   "habitid",
												Value: habitID,
												Count: 1,
												Nodes: map[string]*Node{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, run := range runs {
		instance := NewViewInstance(mapper)

		for stepIdx, step := range run.Steps {
			events := step.GetEvents()

			for _, event := range events {
				err := instance.Append(*event)
				if err != nil {
					t.Fatalf("(%s) (%d) append failed for event %+v: %+v", run.Desc, stepIdx, event, err)
				}
			}

			if !compareViewInstances(instance, step.Wanted) {
				instance.Print("got")
				step.Wanted.Print("wanted")
				t.Fatalf("%s failed:\ngot  %#v\nwant %#v\n", run.Desc, instance, step.Wanted)

			}
		}
	}
}

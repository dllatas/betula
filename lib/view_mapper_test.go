package lib

import (
	"testing"
	"time"
)

func TestViewMapper_Basic(t *testing.T) {
	view := NewViewDefinition(
		"habit",
		[]string{"user", "habit", "event"},
		"2006-01-02",
	)
	mapper := NewViewMapper(view)

	ev := Event{
		Timestamp: time.Date(2024, 5, 19, 13, 45, 0, 0, time.UTC),
		Labels: map[string]string{
			"user":  "alice",
			"habit": "run",
			"event": "start",
		},
	}

	wantPartition := "2024-05-19"
	gotPartition, err := mapper.PartitionKey(ev)
	if err != nil {
		t.Errorf("partition failed %s", err.Error())
	}

	if gotPartition != wantPartition {
		t.Errorf("expected partition %s, got %s", wantPartition, gotPartition)
	}

	wantPath := []string{"alice", "run", "start"}
	gotPath, err := mapper.TreePath(ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, segment := range wantPath {
		if gotPath[i] != segment {
			t.Errorf("expected path[%d] = %s, got %s", i, segment, gotPath[i])
		}
	}
}

func TestViewMapper_MissingLabel(t *testing.T) {
	view := NewViewDefinition(
		"habit",
		[]string{"user", "habit", "event"},
		"2006-01-02",
	)

	mapper := NewViewMapper(view)

	ev := Event{
		Timestamp: time.Date(2024, 5, 19, 13, 45, 0, 0, time.UTC),
		Labels: map[string]string{
			"user":  "alice",
			"event": "start", // "habit" missing
		},
	}

	_, err := mapper.TreePath(ev)
	if err == nil {
		t.Fatal("expected error for missing label key, got nil")
	}
}

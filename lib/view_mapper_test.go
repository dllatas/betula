package lib

import (
	"testing"
	"time"
)

func TestViewMapper_Basic(t *testing.T) {
	view := NewViewDefinition("habit", []string{"user", "habit", "event"}, UnitDay)
	mapper := NewViewMapper(view)

	ev := Event{
		Timestamp: time.Date(2024, 5, 19, 13, 45, 0, 0, time.UTC),
		Labels: map[string]string{
			"user":  "alice",
			"habit": "run",
			"event": "start",
		},
	}

	wantPartition := time.Date(2024, 5, 19, 0, 0, 0, 0, time.UTC)
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
	view := NewViewDefinition("habit", []string{"user", "habit", "event"}, UnitDay)

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

func TestViewMapper_AccessorsAndAlignErrors(t *testing.T) {
	view := NewViewDefinition("habit", []string{"user"}, UnitDay)
	mapper := NewViewMapper(view)

	if mapper.ViewLayout() == "" {
		t.Fatal("expected non-empty layout from ViewLayout()")
	}
	if mapper.Granularity() != UnitDay {
		t.Fatalf("expected Granularity() %q, got %q", UnitDay, mapper.Granularity())
	}

	// alignToPartition should fail on zero time
	zero := time.Time{}
	if _, err := mapper.alignToPartition(zero); err == nil {
		t.Fatal("expected error when aligning zero time, got nil")
	}
}

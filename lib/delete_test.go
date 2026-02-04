package lib

import (
	"testing"
	"time"
)

func TestViewInstanceDelete(t *testing.T) {
	view := NewViewDefinition("test", []string{"country", "user"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)

	ts := time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC)
	ev := Event{
		Timestamp: ts,
		Labels: map[string]string{
			"country": "fr",
			"user":    "alice",
		},
	}

	if err := instance.Append(ev); err != nil {
		t.Fatalf("append: %v", err)
	}

	shards, err := instance.Range(ts, ts, false)
	if err != nil {
		t.Fatalf("range: %v", err)
	}
	if len(shards) != 1 {
		t.Fatalf("expected 1 shard, got %d", len(shards))
	}
	shard := shards[0]

	countryNode := shard.Roots[shardMapKey("country", "fr")]
	if countryNode == nil || countryNode.Count != 1 {
		t.Fatalf("expected country node count=1, got %#v", countryNode)
	}
	userNode := countryNode.Nodes[shardMapKey("user", "alice")]
	if userNode == nil || userNode.Count != 1 {
		t.Fatalf("expected user node count=1, got %#v", userNode)
	}
	if shard.Count != 1 {
		t.Fatalf("expected shard count=1, got %d", shard.Count)
	}

	if err := instance.Delete(ev); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if shard.Count != 0 {
		t.Fatalf("expected shard count=0 after delete, got %d", shard.Count)
	}
	if countryNode.Count != 0 {
		t.Fatalf("expected country node count=0 after delete, got %d", countryNode.Count)
	}
	if userNode.Count != 0 {
		t.Fatalf("expected user node count=0 after delete, got %d", userNode.Count)
	}
}

func TestViewInstanceDelete_NoShard(t *testing.T) {
	view := NewViewDefinition("test", []string{"user"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)

	ev := Event{
		Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
		Labels:    map[string]string{"user": "alice"},
	}

	// Deleting from an empty instance should be a no-op and return nil.
	if err := instance.Delete(ev); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestViewInstanceDelete_ValidationErrors(t *testing.T) {
	view := NewViewDefinition("test", []string{"user"}, UnitDay)
	mapper := NewViewMapper(view)
	instance := NewViewInstance(mapper)

	t.Run("partition key error", func(t *testing.T) {
		ev := Event{
			Timestamp: time.Time{},
			Labels:    map[string]string{"user": "alice"},
		}
		if err := instance.Delete(ev); err == nil {
			t.Fatal("expected error for zero timestamp, got nil")
		}
	})

	t.Run("tree path error", func(t *testing.T) {
		ev := Event{
			Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
			Labels:    map[string]string{}, // missing user label
		}
		if err := instance.Delete(ev); err == nil {
			t.Fatal("expected error for missing label, got nil")
		}
	})
}

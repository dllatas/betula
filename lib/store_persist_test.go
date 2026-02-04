package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_SaveToFile_EmptyViewsNoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.gob")

	s := NewInMemoryStore()
	s.DataPath = path

	if err := s.SaveToFile(); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no file to be created when store is empty, but %q exists", path)
	}
}

func TestStore_SaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.gob")

	s := NewInMemoryStore()
	s.DataPath = path

	def := NewViewDefinition("sessions", []string{"user"}, UnitDay)
	mapper := NewViewMapper(def)
	view := NewViewInstance(mapper)
	if err := view.Append(Event{
		Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
		Labels:    map[string]string{"user": "alice"},
	}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := s.Register(view); err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := s.SaveToFile(); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded, err := LoadStoreFromDataFileIfExists(path)
	if err != nil {
		t.Fatalf("LoadStoreFromDataFileIfExists: %v", err)
	}

	got, err := loaded.Get("sessions")
	if err != nil {
		t.Fatalf("loaded.Get: %v", err)
	}

	if !compareViewInstances(view, got) {
		t.Fatalf("expected loaded view to match original")
	}
}

func TestLoadStoreFromDataFileIfExists_EmptyFileReturnsNewStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.gob")

	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	s, err := LoadStoreFromDataFileIfExists(path)
	if err != nil {
		t.Fatalf("LoadStoreFromDataFileIfExists: %v", err)
	}
	if s == nil || s.Views == nil {
		t.Fatalf("expected non-nil store with views map, got %#v", s)
	}
	if len(s.Views) != 0 {
		t.Fatalf("expected empty store, got %d views", len(s.Views))
	}
}

func TestStore_LoadWALEntries(t *testing.T) {
	store := NewInMemoryStore()

	entries := []WALEntry{
		{
			Op:       "create-view",
			ViewName: "v",
			Keys:     []string{"user"},
			Unit:     string(UnitDay),
		},
		{
			Op:       "create-view", // idempotent replay
			ViewName: "v",
			Keys:     []string{"user"},
			Unit:     string(UnitDay),
		},
		{
			Op:       "append",
			ViewName: "v",
			Payload: Event{
				Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
				Labels:    map[string]string{"user": "alice"},
			},
		},
		{
			Op:       "delete",
			ViewName: "v",
			Payload: Event{
				Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
				Labels:    map[string]string{"user": "alice"},
			},
		},
	}

	if err := store.LoadWALEntries(entries); err != nil {
		t.Fatalf("LoadWALEntries: %v", err)
	}

	v, err := store.Get("v")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(v.Shards) != 1 {
		t.Fatalf("expected 1 shard, got %d", len(v.Shards))
	}
	if v.Shards[0].Count != 0 {
		t.Fatalf("expected count=0 after append+delete, got %d", v.Shards[0].Count)
	}

	t.Run("unsupported op", func(t *testing.T) {
		if err := store.LoadWALEntries([]WALEntry{{Op: "nope", ViewName: "v"}}); err == nil {
			t.Fatal("expected error for unsupported op, got nil")
		}
	})

	t.Run("append to missing view", func(t *testing.T) {
		s := NewInMemoryStore()
		err := s.LoadWALEntries([]WALEntry{{Op: "append", ViewName: "missing", Payload: Event{Timestamp: time.Now()}}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

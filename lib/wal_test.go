package lib

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDeriveWALPath(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "store.data")
	got := deriveWALPath(dataPath)
	want := filepath.Join(dir, "store.wal")
	if got != want {
		t.Fatalf("deriveWALPath() = %q, want %q", got, want)
	}
}

func TestNewWAL_EmptyPath(t *testing.T) {
	if _, err := NewWAL(""); err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestWAL_WriteReplayTruncateRollback(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "store.data")

	w, err := NewWAL(dataPath)
	if err != nil {
		t.Fatalf("NewWAL: %v", err)
	}
	defer w.Close()

	entry1 := WALEntry{
		Timestamp: time.Date(2026, 2, 1, 15, 0, 0, 0, time.UTC),
		Op:        "create-view",
		ViewName:  "v",
		Keys:      []string{"user"},
		Unit:      string(UnitDay),
	}
	entry2 := WALEntry{
		Timestamp: time.Date(2026, 2, 1, 15, 1, 0, 0, time.UTC),
		Op:        "append",
		ViewName:  "v",
		Payload: Event{
			Timestamp: time.Date(2026, 2, 1, 15, 1, 0, 0, time.UTC),
			Labels:    map[string]string{"user": "alice"},
		},
	}

	if err := w.Write(entry1); err != nil {
		t.Fatalf("Write entry1: %v", err)
	}
	if err := w.Write(entry2); err != nil {
		t.Fatalf("Write entry2: %v", err)
	}

	entries, err := w.Replay()
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Rollback removes the last entry.
	if err := w.RollbackLast(); err != nil {
		t.Fatalf("RollbackLast: %v", err)
	}
	entries, err = w.Replay()
	if err != nil {
		t.Fatalf("Replay after rollback: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after rollback, got %d", len(entries))
	}
	if entries[0].Op != "create-view" {
		t.Fatalf("expected remaining op to be create-view, got %q", entries[0].Op)
	}

	// Truncate clears the file.
	if err := w.Truncate(); err != nil {
		t.Fatalf("Truncate: %v", err)
	}
	entries, err = w.Replay()
	if err != nil {
		t.Fatalf("Replay after truncate: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries after truncate, got %d", len(entries))
	}

	// Rollback on an empty WAL should return an error.
	if err := w.RollbackLast(); err == nil {
		t.Fatal("expected error on rollback of empty wal, got nil")
	}
}

func TestWAL_Replay_NoFile(t *testing.T) {
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "store.data")
	walPath := deriveWALPath(dataPath)

	w := &WAL{path: walPath}
	entries, err := w.Replay()
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries for missing wal file, got %d", len(entries))
	}
}

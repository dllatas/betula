package lib

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type WALEntry struct {
	Timestamp time.Time
	Op        string
	ViewName  string
	Keys      []string
	Unit      string
	Payload   Event
}

type WAL struct {
	mu   sync.Mutex
	file *os.File
	enc  *gob.Encoder
	path string
}

func deriveWALPath(dataPath string) string {
	dir := filepath.Dir(dataPath)
	base := filepath.Base(dataPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.Join(dir, name+".wal")
}

func NewWAL(path string) (*WAL, error) {
	if path == "" {
		return nil, fmt.Errorf("new wal: a file path must be set")
	}

	walPath := deriveWALPath(path)

	f, err := os.OpenFile(walPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file: f,
		path: walPath,
		enc:  gob.NewEncoder(f),
	}, nil
}

func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}

func (w *WAL) Write(entry WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.enc == nil {
		w.enc = gob.NewEncoder(w.file)
	}

	if err := w.enc.Encode(entry); err != nil {
		return fmt.Errorf("wal write: failed to encode entry: %w", err)
	}

	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("wal write: failed to sync file: %w", err)
	}

	return nil
}

func (w *WAL) Replay() ([]WALEntry, error) {
	exists, err := ExistsFile(w.path)
	if err != nil {
		return nil, fmt.Errorf("wal replay: failed to verify that file exists: %w", err)
	}

	if !exists {
		slog.Warn("wal file does not exist, skipping replay", "path", w.path)
		return []WALEntry{}, nil
	}

	f, err := os.Open(w.path)
	if err != nil {
		return nil, fmt.Errorf("wal replay: failed to open file: %w", err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	var data []WALEntry
	entryCount := 0

	for {
		var entry WALEntry
		err := dec.Decode(&entry)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("wal replay: finished successfully", "entries", entryCount)
				return data, nil
			}

			slog.Warn("wal replay: decode failed", "entries decoded", entryCount, "entry", entry, "err", err)
			continue
			// return nil, fmt.Errorf("wal replay: decode failed: %w", err)
		}

		slog.Info("wal replay: decoded", "count", entryCount, "entry", entry)

		data = append(data, entry)
		entryCount++
	}
}

func (w *WAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.file.Truncate(0); err != nil {
		return fmt.Errorf("wal truncate: failed to truncate file: %w", err)
	}

	if _, err := w.file.Seek(0, 0); err != nil {
		return fmt.Errorf("wal truncate: failed to seek to beginning: %w", err)
	}

	w.enc = gob.NewEncoder(w.file)

	return nil
}

func (w *WAL) RollbackLast() error {
	// Step 1: Read all entries
	entries, err := w.Replay()
	if err != nil {
		return fmt.Errorf("rollback: failed to replay wal: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("rollback: wal is empty, nothing to rollback")
	}

	// Step 2: Close original WAL file
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("rollback: failed to close original file: %w", err)
	}

	// Step 3: Rewrite all but last entry to a new temp file
	tmpPath := w.path + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("rollback: failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	enc := gob.NewEncoder(tmpFile)
	for _, entry := range entries[:len(entries)-1] {
		if err := enc.Encode(entry); err != nil {
			return fmt.Errorf("rollback: failed to encode entry: %w", err)
		}
	}

	// Step 4: Replace original file
	if err := os.Rename(tmpPath, w.path); err != nil {
		return fmt.Errorf("rollback: failed to replace wal file: %w", err)
	}

	// Step 5: Reopen WAL file in append mode
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("rollback: failed to reopen wal file: %w", err)
	}
	w.file = f
	w.enc = gob.NewEncoder(f)

	return nil
}

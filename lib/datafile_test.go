package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExistsFile(t *testing.T) {
	t.Run("dir does not exist", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing-dir", "file.gob")
		exists, err := ExistsFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Fatalf("expected exists=false for missing dir, got true")
		}
	})

	t.Run("file does not exist but dir does", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "missing.gob")
		exists, err := ExistsFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exists {
			t.Fatalf("expected exists=false for missing file, got true")
		}
	})

	t.Run("path exists and is a file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "data.gob")
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}

		exists, err := ExistsFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !exists {
			t.Fatalf("expected exists=true for file, got false")
		}
	})

	t.Run("path exists and is a directory", func(t *testing.T) {
		dir := t.TempDir()
		exists, err := ExistsFile(dir)
		if err == nil {
			t.Fatalf("expected error for directory path, got nil (exists=%v)", exists)
		}
	})
}

func TestCreateFile_CreatesParentDirs(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "nested", "dir", "file.gob")

	if err := CreateFile(path); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("expected parent dir to exist, got %v", err)
	}
}

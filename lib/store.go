package lib

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

type Store struct {
	mu       sync.RWMutex
	Views    map[string]*ViewInstance
	DataPath string
	WAL      *WAL
}

type storeExport struct {
	Views    map[string]*ViewInstance
	DataPath string
}

func LoadStoreFromDataFileIfExists(path string) (*Store, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("load from file: close defer", "err", err.Error())
		}
	}()

	dec := gob.NewDecoder(f)
	var tmp storeExport
	err = dec.Decode(&tmp)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return NewInMemoryStore(), nil
		}

		if errors.Is(err, io.EOF) {
			return NewInMemoryStore(), nil
		}

		return nil, err
	}

	store := &Store{
		Views:    tmp.Views,
		DataPath: tmp.DataPath,
		mu:       sync.RWMutex{},
	}

	return store, nil
}

func NewInMemoryStore() *Store {
	return &Store{
		Views: make(map[string]*ViewInstance),
	}
}

func (s *Store) Register(view *ViewInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Views[view.Mapper.D.Name]; exists {
		return fmt.Errorf("view %s already exists", view.Mapper.D.Name)
	}

	s.Views[view.Mapper.D.Name] = view

	return nil
}

func (s *Store) Get(name string) (*ViewInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instance, ok := s.Views[name]
	if !ok {
		return nil, fmt.Errorf("view %s not found", name)
	}

	return instance, nil
}

func (s *Store) SaveToFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Views) == 0 {
		slog.Info("save to file: nothing to persist")
		return nil
	}

	f, err := os.Create(s.DataPath)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("save to file: close defer", "err", err.Error())
		}
	}()

	enc := gob.NewEncoder(f)

	return enc.Encode(storeExport{
		Views:    s.Views,
		DataPath: s.DataPath,
	})
}

func (s *Store) LoadWALEntries(entries []WALEntry) error {
	for _, entry := range entries {
		switch entry.Op {
		case "append":
			view, ok := s.Views[entry.ViewName]
			if !ok {
				return fmt.Errorf("apply wal: view %s not found", entry.ViewName)
			}

			if err := view.Append(entry.Payload); err != nil {
				return fmt.Errorf("apply wal: failed to append to view %s: %w", entry.ViewName, err)
			}
		case "delete":
			view, ok := s.Views[entry.ViewName]
			if !ok {
				return fmt.Errorf("apply wal: view %s not found", entry.ViewName)
			}

			if err := view.Delete(entry.Payload); err != nil {
				return fmt.Errorf("apply wal: failed to delete from view %s: %w", entry.ViewName, err)
			}
		case "create-view":
			unit, err := ParseUnit(entry.Unit)
			if err != nil {
				return fmt.Errorf("replay create-view: invalid unit %s (%s)", entry.Unit, err.Error())
			}

			def := NewViewDefinition(entry.ViewName, entry.Keys, unit)
			mapper := NewViewMapper(def)
			instance := NewViewInstance(mapper)

			if err := s.Register(instance); err != nil {
				existing, _ := s.Get(instance.Mapper.D.Name)
				if existing != nil {
					areEqual := compareViewInstances(instance, existing)
					if !areEqual {
						slog.Warn("replay create-view:", "skip", true, "existing", existing, "skipped", instance)
					}
				}
			}
		default:
			return fmt.Errorf("apply wal: unsupported op %q in view %s", entry.Op, entry.ViewName)
		}

		slog.Debug("applied wal entry", "op", entry.Op, "view", entry.ViewName)
	}

	return nil
}

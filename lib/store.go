package lib

import (
	"fmt"
	"sync"
)

type Store struct {
	mu    sync.RWMutex
	views map[string]*ViewInstance
}

func NewStore() *Store {
	return &Store{
		views: make(map[string]*ViewInstance),
	}
}

func (s *Store) Register(view *ViewInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.views[view.Mapper.d.Name]; exists {
		return fmt.Errorf("view %s already exists", view.Mapper.d.Name)
	}

	s.views[view.Mapper.d.Name] = view

	return nil
}

func (s *Store) Get(name string) (*ViewInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instance, ok := s.views[name]
	if !ok {
		return nil, fmt.Errorf("view %s not found", name)
	}

	return instance, nil
}

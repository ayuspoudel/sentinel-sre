package action

import "sync"

type Store struct {
	mu      sync.RWMutex
	actions map[string]Action
}

func NewStore() *Store {
	return &Store{
		actions: make(map[string]Action),
	}
}

func (s *Store) Set(a Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions[a.GuardName] = a
}

func (s *Store) Get(name string) (Action, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.actions[name]
	return a, ok
}

func (s *Store) List() []Action {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Action, 0, len(s.actions))
	for _, a := range s.actions {
		out = append(out, a)
	}
	return out
}

package cluster

import (
	"sync"
	"time"
)

/*
@ayuspoudel
Sometimes two concurrent goroutines might be updating this cluster name at same time
so we need to have a mutex lock to ensure atomic read and writes
Also *Cluster contains {Name, AgentID, Verison, Address, CreatedAt, LastSeen}
*/
type Store struct {
	mu       sync.RWMutex
	clusters map[string]*Cluster
}

func NewStore() *Store {
	return &Store{
		clusters: make(map[string]*Cluster),
	}
}

func (s *Store) Register(c *Cluster) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if existing, ok := s.clusters[c.Name]; ok {
		existing.LastSeen = now
		existing.Version = c.Version
		existing.Address = c.Address
		return
	}

	c.CreatedAt = now
	c.LastSeen = now
	s.clusters[c.Name] = c
}

func (s *Store) Get(name string) (*Cluster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.clusters[name]
	return c, ok
}

func (s *Store) List() []*Cluster {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*Cluster, 0, len(s.clusters))
	for _, c := range s.clusters {
		out = append(out, c)
	}
	return out
}

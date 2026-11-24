// Package state holds small in-memory conversation flags shared between
// concurrent webhook requests.
package state

import "sync"

// FlagSet is a concurrency-safe set of user IDs.
type FlagSet struct {
	mu    sync.Mutex
	flags map[int64]struct{}
}

func NewFlagSet() *FlagSet { return &FlagSet{flags: make(map[int64]struct{})} }

func (s *FlagSet) Set(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flags[id] = struct{}{}
}

func (s *FlagSet) Has(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.flags[id]
	return ok
}

func (s *FlagSet) Clear(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.flags, id)
}

package state

import (
	"sync"
)

type StateManager struct {
	mu    sync.RWMutex
	index int
	max   int
}

func New(max int) *StateManager {
	return &StateManager{
		index: 0,
		max:   max,
	}
}

func (s *StateManager) SetIndex(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 {
		s.index = 0
		return
	}
	if index >= s.max {
		s.index = s.max - 1
		return
	}
	s.index = index
}

func (s *StateManager) Index() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

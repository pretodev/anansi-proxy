package state

import (
	"sync"
)

type StateManager struct {
	mu    sync.RWMutex
	index int
}

func New() *StateManager {
	return &StateManager{
		index: 0,
	}
}

func (s *StateManager) SetIndex(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 {
		s.index = 0
	} else {
		s.index = index
	}
}

func (s *StateManager) Index() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

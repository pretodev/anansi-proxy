package state

import (
	"sync"
)

type StateManager struct {
	mu         sync.RWMutex
	index      int
	max        int
	callCounts map[string]int // key: endpoint identifier (e.g., "GET /users/{id}")
}

func New(max int) *StateManager {
	return &StateManager{
		index:      0,
		max:        max,
		callCounts: make(map[string]int),
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

// IncrementCallCount increments the call count for a specific endpoint
// and returns the new count value.
func (s *StateManager) IncrementCallCount(endpoint string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCounts[endpoint]++
	return s.callCounts[endpoint]
}

// GetCallCount returns the current call count for a specific endpoint.
func (s *StateManager) GetCallCount(endpoint string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.callCounts[endpoint]
}

// ResetCallCount resets the call count for a specific endpoint to 0.
func (s *StateManager) ResetCallCount(endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.callCounts, endpoint)
}

// ResetAllCallCounts clears all call counts.
func (s *StateManager) ResetAllCallCounts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCounts = make(map[string]int)
}

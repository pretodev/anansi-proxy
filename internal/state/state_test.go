package state

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		max  int
		want int
	}{
		{
			name: "create state manager with max 10",
			max:  10,
			want: 0,
		},
		{
			name: "create state manager with max 1",
			max:  1,
			want: 0,
		},
		{
			name: "create state manager with max 100",
			max:  100,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := New(tt.max)
			if sm == nil {
				t.Fatal("New() returned nil")
			}
			if sm.index != tt.want {
				t.Errorf("New() index = %v, want %v", sm.index, tt.want)
			}
			if sm.max != tt.max {
				t.Errorf("New() max = %v, want %v", sm.max, tt.max)
			}
			if sm.callCounts == nil {
				t.Error("New() callCounts map is nil")
			}
			if len(sm.callCounts) != 0 {
				t.Errorf("New() callCounts should be empty, got %d entries", len(sm.callCounts))
			}
		})
	}
}

func TestStateManager_Index(t *testing.T) {
	tests := []struct {
		name         string
		max          int
		initialIndex int
		want         int
	}{
		{
			name:         "get initial index",
			max:          10,
			initialIndex: 0,
			want:         0,
		},
		{
			name:         "get index after setting",
			max:          10,
			initialIndex: 5,
			want:         5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := New(tt.max)
			sm.index = tt.initialIndex
			got := sm.Index()
			if got != tt.want {
				t.Errorf("StateManager.Index() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateManager_SetIndex(t *testing.T) {
	tests := []struct {
		name     string
		max      int
		setIndex int
		want     int
	}{
		{
			name:     "set valid index",
			max:      10,
			setIndex: 5,
			want:     5,
		},
		{
			name:     "set index to 0",
			max:      10,
			setIndex: 0,
			want:     0,
		},
		{
			name:     "set index to max-1",
			max:      10,
			setIndex: 9,
			want:     9,
		},
		{
			name:     "set negative index",
			max:      10,
			setIndex: -1,
			want:     0,
		},
		{
			name:     "set very negative index",
			max:      10,
			setIndex: -100,
			want:     0,
		},
		{
			name:     "set index equal to max",
			max:      10,
			setIndex: 10,
			want:     9,
		},
		{
			name:     "set index greater than max",
			max:      10,
			setIndex: 15,
			want:     9,
		},
		{
			name:     "set index much greater than max",
			max:      10,
			setIndex: 1000,
			want:     9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := New(tt.max)
			sm.SetIndex(tt.setIndex)
			got := sm.Index()
			if got != tt.want {
				t.Errorf("after SetIndex(%v), Index() = %v, want %v", tt.setIndex, got, tt.want)
			}
		})
	}
}

func TestStateManager_SetIndexMultipleTimes(t *testing.T) {
	sm := New(10)

	// Set to 5
	sm.SetIndex(5)
	if got := sm.Index(); got != 5 {
		t.Errorf("after SetIndex(5), Index() = %v, want 5", got)
	}

	// Set to 3
	sm.SetIndex(3)
	if got := sm.Index(); got != 3 {
		t.Errorf("after SetIndex(3), Index() = %v, want 3", got)
	}

	// Set to boundary
	sm.SetIndex(9)
	if got := sm.Index(); got != 9 {
		t.Errorf("after SetIndex(9), Index() = %v, want 9", got)
	}

	// Set beyond boundary
	sm.SetIndex(20)
	if got := sm.Index(); got != 9 {
		t.Errorf("after SetIndex(20), Index() = %v, want 9", got)
	}

	// Set negative
	sm.SetIndex(-5)
	if got := sm.Index(); got != 0 {
		t.Errorf("after SetIndex(-5), Index() = %v, want 0", got)
	}
}

func TestStateManager_ConcurrentAccess(t *testing.T) {
	sm := New(100)
	var wg sync.WaitGroup
	iterations := 1000

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sm.SetIndex(val)
			}
		}(i * 10)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = sm.Index()
			}
		}()
	}

	wg.Wait()

	// Verify state is still valid
	index := sm.Index()
	if index < 0 || index >= sm.max {
		t.Errorf("after concurrent access, index %v is out of bounds [0, %v)", index, sm.max)
	}
}

func TestStateManager_ConcurrentSetAndRead(t *testing.T) {
	sm := New(50)
	var wg sync.WaitGroup
	done := make(chan bool)
	errors := make(chan error, 100)

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			select {
			case <-done:
				return
			default:
				sm.SetIndex(i % 50)
			}
		}
	}()

	// Reader goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				select {
				case <-done:
					return
				default:
					index := sm.Index()
					if index < 0 || index >= 50 {
						errors <- nil
					}
				}
			}
		}()
	}

	wg.Wait()
	close(done)
	close(errors)

	if len(errors) > 0 {
		t.Error("concurrent access produced invalid index values")
	}
}

func TestStateManager_EdgeCaseMax1(t *testing.T) {
	sm := New(1)

	// Only valid index is 0
	sm.SetIndex(0)
	if got := sm.Index(); got != 0 {
		t.Errorf("Index() = %v, want 0", got)
	}

	// Any positive value should clamp to 0
	sm.SetIndex(1)
	if got := sm.Index(); got != 0 {
		t.Errorf("after SetIndex(1) with max=1, Index() = %v, want 0", got)
	}

	// Negative should also clamp to 0
	sm.SetIndex(-1)
	if got := sm.Index(); got != 0 {
		t.Errorf("after SetIndex(-1) with max=1, Index() = %v, want 0", got)
	}
}

func BenchmarkStateManager_SetIndex(b *testing.B) {
	sm := New(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.SetIndex(i % 100)
	}
}

func BenchmarkStateManager_Index(b *testing.B) {
	sm := New(100)
	sm.SetIndex(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sm.Index()
	}
}

func BenchmarkStateManager_ConcurrentSetIndex(b *testing.B) {
	sm := New(100)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			sm.SetIndex(i % 100)
			i++
		}
	})
}

func BenchmarkStateManager_ConcurrentIndex(b *testing.B) {
	sm := New(100)
	sm.SetIndex(50)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sm.Index()
		}
	})
}

func TestStateManager_CallCount(t *testing.T) {
	t.Run("GetCallCount returns 0 for new endpoint", func(t *testing.T) {
		sm := New(10)
		count := sm.GetCallCount("GET /users")
		if count != 0 {
			t.Errorf("GetCallCount() = %v, want 0", count)
		}
	})

	t.Run("IncrementCallCount increments and returns new value", func(t *testing.T) {
		sm := New(10)
		endpoint := "GET /users/{id}"

		count := sm.IncrementCallCount(endpoint)
		if count != 1 {
			t.Errorf("IncrementCallCount() = %v, want 1", count)
		}

		count = sm.IncrementCallCount(endpoint)
		if count != 2 {
			t.Errorf("IncrementCallCount() = %v, want 2", count)
		}

		count = sm.GetCallCount(endpoint)
		if count != 2 {
			t.Errorf("GetCallCount() = %v, want 2", count)
		}
	})

	t.Run("Different endpoints have separate counts", func(t *testing.T) {
		sm := New(10)

		sm.IncrementCallCount("GET /users")
		sm.IncrementCallCount("GET /users")
		sm.IncrementCallCount("POST /users")

		if got := sm.GetCallCount("GET /users"); got != 2 {
			t.Errorf("GET /users count = %v, want 2", got)
		}
		if got := sm.GetCallCount("POST /users"); got != 1 {
			t.Errorf("POST /users count = %v, want 1", got)
		}
	})

	t.Run("ResetCallCount clears count for specific endpoint", func(t *testing.T) {
		sm := New(10)
		endpoint := "GET /products"

		sm.IncrementCallCount(endpoint)
		sm.IncrementCallCount(endpoint)
		sm.ResetCallCount(endpoint)

		count := sm.GetCallCount(endpoint)
		if count != 0 {
			t.Errorf("GetCallCount() after reset = %v, want 0", count)
		}
	})

	t.Run("ResetAllCallCounts clears all counts", func(t *testing.T) {
		sm := New(10)

		sm.IncrementCallCount("GET /users")
		sm.IncrementCallCount("POST /users")
		sm.IncrementCallCount("GET /products")

		sm.ResetAllCallCounts()

		if got := sm.GetCallCount("GET /users"); got != 0 {
			t.Errorf("GET /users count after reset = %v, want 0", got)
		}
		if got := sm.GetCallCount("POST /users"); got != 0 {
			t.Errorf("POST /users count after reset = %v, want 0", got)
		}
		if got := sm.GetCallCount("GET /products"); got != 0 {
			t.Errorf("GET /products count after reset = %v, want 0", got)
		}
	})
}

func TestStateManager_ConcurrentCallCount(t *testing.T) {
	sm := New(10)
	endpoint := "GET /concurrent"
	iterations := 1000

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				sm.IncrementCallCount(endpoint)
			}
		}()
	}

	wg.Wait()

	expected := 10 * iterations
	got := sm.GetCallCount(endpoint)
	if got != expected {
		t.Errorf("After concurrent increments: got %v, want %v", got, expected)
	}
}

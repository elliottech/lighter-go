package nonce

import (
	"errors"
	"sync"
	"testing"
)

// mockFetcher is a mock implementation of NonceFetcher for testing
type mockFetcher struct {
	mu       sync.Mutex
	nonces   map[nonceKey]int64
	fetchErr error
	calls    int
}

func newMockFetcher() *mockFetcher {
	return &mockFetcher{
		nonces: make(map[nonceKey]int64),
	}
}

func (m *mockFetcher) GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++

	if m.fetchErr != nil {
		return 0, m.fetchErr
	}

	key := nonceKey{accountIndex, apiKeyIndex}
	nonce := m.nonces[key]
	m.nonces[key] = nonce + 1
	return nonce, nil
}

func (m *mockFetcher) setNonce(accountIndex int64, apiKeyIndex uint8, nonce int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nonces[nonceKey{accountIndex, apiKeyIndex}] = nonce
}

func (m *mockFetcher) setError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetchErr = err
}

func (m *mockFetcher) getCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestOptimisticManager_GetNonce_InitialFetch(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewOptimisticManager(fetcher)

	nonce, err := manager.GetNonce(1, 0)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	if nonce != 100 {
		t.Errorf("expected nonce 100, got %d", nonce)
	}

	if fetcher.getCalls() != 1 {
		t.Errorf("expected 1 fetch call, got %d", fetcher.getCalls())
	}
}

func TestOptimisticManager_GetNonce_LocalIncrement(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewOptimisticManager(fetcher)

	// First call fetches from API
	nonce1, err := manager.GetNonce(1, 0)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	// Subsequent calls increment locally
	nonce2, err := manager.GetNonce(1, 0)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	nonce3, err := manager.GetNonce(1, 0)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	if nonce1 != 100 {
		t.Errorf("expected nonce1 100, got %d", nonce1)
	}
	if nonce2 != 101 {
		t.Errorf("expected nonce2 101, got %d", nonce2)
	}
	if nonce3 != 102 {
		t.Errorf("expected nonce3 102, got %d", nonce3)
	}

	// Only 1 API call should have been made
	if fetcher.getCalls() != 1 {
		t.Errorf("expected 1 fetch call, got %d", fetcher.getCalls())
	}
}

func TestOptimisticManager_GetNonce_MultipleAccounts(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)
	fetcher.setNonce(2, 0, 200)
	fetcher.setNonce(1, 1, 300)

	manager := NewOptimisticManager(fetcher)

	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(2, 0)
	nonce3, _ := manager.GetNonce(1, 1)

	if nonce1 != 100 {
		t.Errorf("expected nonce for account 1, key 0 to be 100, got %d", nonce1)
	}
	if nonce2 != 200 {
		t.Errorf("expected nonce for account 2, key 0 to be 200, got %d", nonce2)
	}
	if nonce3 != 300 {
		t.Errorf("expected nonce for account 1, key 1 to be 300, got %d", nonce3)
	}
}

func TestOptimisticManager_GetNonce_FetchError(t *testing.T) {
	fetcher := newMockFetcher()
	expectedErr := errors.New("network error")
	fetcher.setError(expectedErr)

	manager := NewOptimisticManager(fetcher)

	_, err := manager.GetNonce(1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestOptimisticManager_AcknowledgeSuccess(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewOptimisticManager(fetcher)

	// Get nonces
	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(1, 0)

	// Both should be pending
	if manager.PendingCount(1, 0) != 2 {
		t.Errorf("expected 2 pending, got %d", manager.PendingCount(1, 0))
	}

	// Acknowledge success
	manager.AcknowledgeSuccess(1, 0, nonce1)

	if manager.PendingCount(1, 0) != 1 {
		t.Errorf("expected 1 pending, got %d", manager.PendingCount(1, 0))
	}

	manager.AcknowledgeSuccess(1, 0, nonce2)

	if manager.PendingCount(1, 0) != 0 {
		t.Errorf("expected 0 pending, got %d", manager.PendingCount(1, 0))
	}
}

func TestOptimisticManager_AcknowledgeFailure(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewOptimisticManager(fetcher)

	// Get initial nonce
	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(1, 0)

	if nonce1 != 100 || nonce2 != 101 {
		t.Fatalf("unexpected nonces: %d, %d", nonce1, nonce2)
	}

	// Simulate failure on nonce2
	manager.AcknowledgeFailure(1, 0, nonce2)

	// State should be reset, next call should fetch from API
	// Since mock fetcher increments on each call, it should now return 101
	nonce3, _ := manager.GetNonce(1, 0)
	if nonce3 != 101 {
		t.Errorf("expected nonce 101 after failure reset, got %d", nonce3)
	}

	// Should have made 2 API calls total
	if fetcher.getCalls() != 2 {
		t.Errorf("expected 2 fetch calls, got %d", fetcher.getCalls())
	}
}

func TestOptimisticManager_Reset(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)
	fetcher.setNonce(2, 0, 200)

	manager := NewOptimisticManager(fetcher)

	// Get nonces for both accounts
	_, _ = manager.GetNonce(1, 0)
	_, _ = manager.GetNonce(2, 0)

	// Reset only account 1
	manager.Reset(1, 0)

	// Account 1 should fetch from API again, account 2 should use cache
	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(2, 0)

	// Account 1 was reset, fetched 101 (mock increments on each fetch)
	if nonce1 != 101 {
		t.Errorf("expected nonce 101 for reset account, got %d", nonce1)
	}

	// Account 2 should continue from cache
	if nonce2 != 201 {
		t.Errorf("expected nonce 201 for cached account, got %d", nonce2)
	}
}

func TestOptimisticManager_ResetAll(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)
	fetcher.setNonce(2, 0, 200)

	manager := NewOptimisticManager(fetcher)

	// Get nonces for both accounts
	_, _ = manager.GetNonce(1, 0)
	_, _ = manager.GetNonce(2, 0)

	initialCalls := fetcher.getCalls()

	// Reset all
	manager.ResetAll()

	// Both should fetch from API again
	_, _ = manager.GetNonce(1, 0)
	_, _ = manager.GetNonce(2, 0)

	if fetcher.getCalls() != initialCalls+2 {
		t.Errorf("expected %d fetch calls after reset all, got %d", initialCalls+2, fetcher.getCalls())
	}
}

func TestOptimisticManager_ConcurrentAccess(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 0)

	manager := NewOptimisticManager(fetcher)

	// Run concurrent GetNonce calls
	const goroutines = 100
	var wg sync.WaitGroup
	nonces := make(chan int64, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			nonce, err := manager.GetNonce(1, 0)
			if err != nil {
				t.Errorf("GetNonce failed: %v", err)
				return
			}
			nonces <- nonce
		}()
	}

	wg.Wait()
	close(nonces)

	// Collect all nonces
	seen := make(map[int64]bool)
	for nonce := range nonces {
		if seen[nonce] {
			t.Errorf("duplicate nonce detected: %d", nonce)
		}
		seen[nonce] = true
	}

	// Should have exactly goroutines unique nonces
	if len(seen) != goroutines {
		t.Errorf("expected %d unique nonces, got %d", goroutines, len(seen))
	}
}

func TestOptimisticManager_PendingCount_Empty(t *testing.T) {
	fetcher := newMockFetcher()
	manager := NewOptimisticManager(fetcher)

	if count := manager.PendingCount(1, 0); count != 0 {
		t.Errorf("expected 0 pending for unknown account, got %d", count)
	}
}

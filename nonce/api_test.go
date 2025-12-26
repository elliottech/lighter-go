package nonce

import (
	"errors"
	"testing"
)

func TestAPIManager_GetNonce(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	nonce, err := manager.GetNonce(1, 0)
	if err != nil {
		t.Fatalf("GetNonce failed: %v", err)
	}

	if nonce != 100 {
		t.Errorf("expected nonce 100, got %d", nonce)
	}
}

func TestAPIManager_GetNonce_AlwaysFetchesFromAPI(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	// Each call should fetch from API
	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(1, 0)
	nonce3, _ := manager.GetNonce(1, 0)

	// Mock fetcher increments on each call
	if nonce1 != 100 {
		t.Errorf("expected nonce1 100, got %d", nonce1)
	}
	if nonce2 != 101 {
		t.Errorf("expected nonce2 101, got %d", nonce2)
	}
	if nonce3 != 102 {
		t.Errorf("expected nonce3 102, got %d", nonce3)
	}

	// Should have made 3 API calls
	if fetcher.getCalls() != 3 {
		t.Errorf("expected 3 fetch calls, got %d", fetcher.getCalls())
	}
}

func TestAPIManager_GetNonce_MultipleAccounts(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)
	fetcher.setNonce(2, 0, 200)

	manager := NewAPIManager(fetcher)

	nonce1, _ := manager.GetNonce(1, 0)
	nonce2, _ := manager.GetNonce(2, 0)

	if nonce1 != 100 {
		t.Errorf("expected nonce for account 1 to be 100, got %d", nonce1)
	}
	if nonce2 != 200 {
		t.Errorf("expected nonce for account 2 to be 200, got %d", nonce2)
	}
}

func TestAPIManager_GetNonce_Error(t *testing.T) {
	fetcher := newMockFetcher()
	expectedErr := errors.New("network error")
	fetcher.setError(expectedErr)

	manager := NewAPIManager(fetcher)

	_, err := manager.GetNonce(1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestAPIManager_AcknowledgeSuccess_NoOp(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	// Get a nonce
	nonce1, _ := manager.GetNonce(1, 0)

	// Acknowledge success (should be no-op)
	manager.AcknowledgeSuccess(1, 0, nonce1)

	// Next call should still fetch from API
	nonce2, _ := manager.GetNonce(1, 0)
	if nonce2 != 101 {
		t.Errorf("expected nonce 101, got %d", nonce2)
	}

	if fetcher.getCalls() != 2 {
		t.Errorf("expected 2 fetch calls, got %d", fetcher.getCalls())
	}
}

func TestAPIManager_AcknowledgeFailure_NoOp(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	// Get a nonce
	nonce1, _ := manager.GetNonce(1, 0)

	// Acknowledge failure (should be no-op)
	manager.AcknowledgeFailure(1, 0, nonce1)

	// Next call should still fetch from API
	nonce2, _ := manager.GetNonce(1, 0)
	if nonce2 != 101 {
		t.Errorf("expected nonce 101, got %d", nonce2)
	}

	if fetcher.getCalls() != 2 {
		t.Errorf("expected 2 fetch calls, got %d", fetcher.getCalls())
	}
}

func TestAPIManager_Reset_NoOp(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	// Get a nonce
	manager.GetNonce(1, 0)

	// Reset (should be no-op)
	manager.Reset(1, 0)

	// Next call should still fetch from API as normal
	nonce, _ := manager.GetNonce(1, 0)
	if nonce != 101 {
		t.Errorf("expected nonce 101, got %d", nonce)
	}
}

func TestAPIManager_ResetAll_NoOp(t *testing.T) {
	fetcher := newMockFetcher()
	fetcher.setNonce(1, 0, 100)

	manager := NewAPIManager(fetcher)

	// Get a nonce
	manager.GetNonce(1, 0)

	// Reset all (should be no-op)
	manager.ResetAll()

	// Next call should still fetch from API as normal
	nonce, _ := manager.GetNonce(1, 0)
	if nonce != 101 {
		t.Errorf("expected nonce 101, got %d", nonce)
	}
}

func TestAPIManager_ImplementsInterface(t *testing.T) {
	var _ Manager = (*APIManager)(nil)
}

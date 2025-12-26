package nonce

import (
	"sync"
)

// OptimisticManager assumes transactions succeed and increments nonce locally.
// It only queries the API when:
//  1. No local nonce is available (first use for an account/key pair)
//  2. A failure is acknowledged via AcknowledgeFailure
//  3. Reset is called
//
// This is the fastest option but requires proper failure handling.
// If a transaction fails and AcknowledgeFailure is not called, subsequent
// transactions may fail due to nonce mismatch.
type OptimisticManager struct {
	mu      sync.Mutex
	nonces  map[nonceKey]int64
	pending map[nonceKey]map[int64]struct{} // Track pending nonces
	fetcher NonceFetcher
}

// NewOptimisticManager creates a new OptimisticManager
func NewOptimisticManager(fetcher NonceFetcher) *OptimisticManager {
	return &OptimisticManager{
		nonces:  make(map[nonceKey]int64),
		pending: make(map[nonceKey]map[int64]struct{}),
		fetcher: fetcher,
	}
}

// GetNonce returns the next nonce to use for a transaction.
// If no local nonce exists, it fetches from the API.
// Otherwise, it increments and returns the local nonce.
func (m *OptimisticManager) GetNonce(accountIndex int64, apiKeyIndex uint8) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := nonceKey{accountIndex, apiKeyIndex}

	// Check if we have a local nonce
	if nonce, ok := m.nonces[key]; ok {
		// Return current and increment for next
		m.nonces[key] = nonce + 1
		m.trackPending(key, nonce)
		return nonce, nil
	}

	// Fetch from API
	nonce, err := m.fetcher.GetNextNonce(accountIndex, apiKeyIndex)
	if err != nil {
		return 0, err
	}

	// Store next nonce and track this one as pending
	m.nonces[key] = nonce + 1
	m.trackPending(key, nonce)
	return nonce, nil
}

// AcknowledgeSuccess removes the nonce from pending tracking.
func (m *OptimisticManager) AcknowledgeSuccess(accountIndex int64, apiKeyIndex uint8, nonce int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := nonceKey{accountIndex, apiKeyIndex}
	m.removePending(key, nonce)
}

// AcknowledgeFailure removes the nonce from pending and resets local state
// to force an API fetch on the next GetNonce call.
func (m *OptimisticManager) AcknowledgeFailure(accountIndex int64, apiKeyIndex uint8, nonce int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := nonceKey{accountIndex, apiKeyIndex}

	// Remove from pending
	m.removePending(key, nonce)

	// Reset local nonce to force API fetch
	delete(m.nonces, key)
}

// Reset clears the cached nonce for a specific account/key pair.
func (m *OptimisticManager) Reset(accountIndex int64, apiKeyIndex uint8) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := nonceKey{accountIndex, apiKeyIndex}
	delete(m.nonces, key)
	delete(m.pending, key)
}

// ResetAll clears all cached nonce state.
func (m *OptimisticManager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nonces = make(map[nonceKey]int64)
	m.pending = make(map[nonceKey]map[int64]struct{})
}

// PendingCount returns the number of pending transactions for an account/key pair.
func (m *OptimisticManager) PendingCount(accountIndex int64, apiKeyIndex uint8) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := nonceKey{accountIndex, apiKeyIndex}
	if pending, ok := m.pending[key]; ok {
		return len(pending)
	}
	return 0
}

func (m *OptimisticManager) trackPending(key nonceKey, nonce int64) {
	if m.pending[key] == nil {
		m.pending[key] = make(map[int64]struct{})
	}
	m.pending[key][nonce] = struct{}{}
}

func (m *OptimisticManager) removePending(key nonceKey, nonce int64) {
	if m.pending[key] != nil {
		delete(m.pending[key], nonce)
	}
}

// Ensure OptimisticManager implements Manager
var _ Manager = (*OptimisticManager)(nil)

package nonce

// APIManager always queries the API for the next nonce.
// This is slower due to network latency but always returns the correct nonce.
//
// Use this when:
//   - Transaction rate is low
//   - Reliability is more important than speed
//   - You cannot guarantee proper failure acknowledgment
type APIManager struct {
	fetcher NonceFetcher
}

// NewAPIManager creates a new APIManager
func NewAPIManager(fetcher NonceFetcher) *APIManager {
	return &APIManager{fetcher: fetcher}
}

// GetNonce fetches the next nonce from the API.
func (m *APIManager) GetNonce(accountIndex int64, apiKeyIndex uint8) (int64, error) {
	return m.fetcher.GetNextNonce(accountIndex, apiKeyIndex)
}

// AcknowledgeSuccess is a no-op for APIManager since it always queries the API.
func (m *APIManager) AcknowledgeSuccess(accountIndex int64, apiKeyIndex uint8, nonce int64) {
	// No-op: API manager always fetches fresh nonce
}

// AcknowledgeFailure is a no-op for APIManager since it always queries the API.
func (m *APIManager) AcknowledgeFailure(accountIndex int64, apiKeyIndex uint8, nonce int64) {
	// No-op: API manager always fetches fresh nonce
}

// Reset is a no-op for APIManager since it has no cached state.
func (m *APIManager) Reset(accountIndex int64, apiKeyIndex uint8) {
	// No-op: API manager has no cached state
}

// ResetAll is a no-op for APIManager since it has no cached state.
func (m *APIManager) ResetAll() {
	// No-op: API manager has no cached state
}

// Ensure APIManager implements Manager
var _ Manager = (*APIManager)(nil)

// Package nonce provides nonce management for Lighter transactions.
//
// Two implementations are provided:
//   - OptimisticNonceManager: Assumes transactions succeed and increments locally.
//     Faster but requires failure acknowledgment for recovery.
//   - APINonceManager: Queries the API for each nonce. Slower but always accurate.
package nonce

// NonceFetcher is the interface for fetching nonce from API
type NonceFetcher interface {
	GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error)
}

// Manager defines the interface for managing transaction nonces
type Manager interface {
	// GetNonce returns the next nonce to use for a transaction.
	// This may query the API or return a cached value depending on implementation.
	GetNonce(accountIndex int64, apiKeyIndex uint8) (int64, error)

	// AcknowledgeSuccess should be called after a transaction succeeds.
	// This allows the manager to update its internal state.
	AcknowledgeSuccess(accountIndex int64, apiKeyIndex uint8, nonce int64)

	// AcknowledgeFailure should be called after a transaction fails.
	// This triggers recovery logic to resync with the server.
	AcknowledgeFailure(accountIndex int64, apiKeyIndex uint8, nonce int64)

	// Reset clears the cached nonce state, forcing a fetch from API on next GetNonce.
	Reset(accountIndex int64, apiKeyIndex uint8)

	// ResetAll clears all cached nonce state.
	ResetAll()
}

// nonceKey is used as map key for account/apikey pairs
type nonceKey struct {
	accountIndex int64
	apiKeyIndex  uint8
}

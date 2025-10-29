package client

import (
	"github.com/elliottech/lighter-go/client/http"
)

type MinimalHTTPClient interface {
	GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error)
	GetApiKey(accountIndex int64, apiKeyIndex uint8) (*http.AccountApiKeys, error)
}

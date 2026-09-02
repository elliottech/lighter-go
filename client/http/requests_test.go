package http

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestApiKeyCacheIsScopedByEndpoint(t *testing.T) {
	const (
		accountIndex = int64(987654321)
		apiKeyIndex  = uint8(0)
	)

	newServer := func(publicKey string, requests *atomic.Int32) *httptest.Server {
		return httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			requests.Add(1)
			if r.URL.Path != "/api/v1/apikeys" {
				t.Fatalf("unexpected request path: %s", r.URL.Path)
			}
			if got := r.URL.Query().Get("account_index"); got != "987654321" {
				t.Fatalf("unexpected account_index: %s", got)
			}

			if err := json.NewEncoder(w).Encode(AccountApiKeys{
				ResultCode: ResultCode{Code: CodeOK},
				ApiKeys: []*ApiKey{{
					AccountIndex: accountIndex,
					ApiKeyIndex:  apiKeyIndex,
					PublicKey:    publicKey,
				}},
			}); err != nil {
				t.Fatalf("encode response: %v", err)
			}
		}))
	}

	var requestsA, requestsB atomic.Int32
	serverA := newServer("key-a", &requestsA)
	defer serverA.Close()
	serverB := newServer("key-b", &requestsB)
	defer serverB.Close()

	clientA := NewClient(serverA.URL)
	clientB := NewClient(serverB.URL)

	keyA, err := clientA.GetApiKey(accountIndex, apiKeyIndex)
	if err != nil {
		t.Fatalf("get API key from endpoint A: %v", err)
	}
	if keyA != "key-a" {
		t.Fatalf("endpoint A returned %q, want %q", keyA, "key-a")
	}

	keyB, err := clientB.GetApiKey(accountIndex, apiKeyIndex)
	if err != nil {
		t.Fatalf("get API key from endpoint B: %v", err)
	}
	if keyB != "key-b" {
		t.Fatalf("endpoint B returned %q, want %q", keyB, "key-b")
	}

	clientA.InvalidateApiKeys(accountIndex)
	if _, err := clientB.GetApiKey(accountIndex, apiKeyIndex); err != nil {
		t.Fatalf("get cached API key from endpoint B: %v", err)
	}
	if got := requestsB.Load(); got != 1 {
		t.Fatalf("endpoint B received %d requests after endpoint A invalidation, want 1", got)
	}
}

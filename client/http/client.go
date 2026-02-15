// Package http provides an HTTP client for the Lighter API.
//
// The client provides access to all Lighter REST API endpoints through
// a fluent interface with lazy-initialized API groups.
//
// Usage:
//
//	client := http.NewFullClient("https://mainnet.zklighter.elliot.ai")
//
//	// Get account information
//	account, err := client.Account().GetAccountByL1Address(ctx, "0x...")
//
//	// Get order book
//	orderbook, err := client.Order().GetOrderBook(ctx, 0, 20)
//
//	// Get candlestick data
//	candles, err := client.Candlestick().GetCandlesticks(ctx, 0, "1h", 100)
//
// API Groups:
//   - Account(): Account and position information
//   - Order(): Order book and trade data
//   - Transaction(): Transaction submission and status
//   - Candlestick(): OHLCV market data
//   - Block(): Blockchain block data
//   - Bridge(): Cross-chain bridge operations
//   - Info(): General system information
package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	core "github.com/elliottech/lighter-go/client"
)

var (
	dialer = &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second,
	}
	transport = &http.Transport{
		DialContext:         dialer.DialContext,
		MaxConnsPerHost:     1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}

	httpClient = &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}
)

// Ensure client implements both interfaces
var _ core.MinimalHTTPClient = (*client)(nil)
var _ core.FullHTTPClient = (*client)(nil)

type client struct {
	endpoint string

	// Lazy-initialized API groups
	accountAPI     *accountAPIImpl
	orderAPI       *orderAPIImpl
	transactionAPI *transactionAPIImpl
	candlestickAPI *candlestickAPIImpl
	blockAPI       *blockAPIImpl
	bridgeAPI      *bridgeAPIImpl
	infoAPI        *infoAPIImpl

	// Mutex for lazy initialization
	mu sync.Mutex
}

// NewClient creates a new HTTP client that implements FullHTTPClient.
// For backward compatibility, it returns MinimalHTTPClient but can be
// type-asserted to FullHTTPClient for full API access.
func NewClient(baseUrl string) core.MinimalHTTPClient {
	if baseUrl == "" {
		return nil
	}

	return &client{
		endpoint: baseUrl,
	}
}

// NewFullClient creates a new HTTP client with full API access.
// This is the recommended way to create a client for new code.
func NewFullClient(baseUrl string) core.FullHTTPClient {
	if baseUrl == "" {
		return nil
	}

	return &client{
		endpoint: baseUrl,
	}
}

// Account returns the AccountAPI for account-related operations
func (c *client) Account() core.AccountAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accountAPI == nil {
		c.accountAPI = &accountAPIImpl{client: c}
	}
	return c.accountAPI
}

// Order returns the OrderAPI for order-related operations
func (c *client) Order() core.OrderAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.orderAPI == nil {
		c.orderAPI = &orderAPIImpl{client: c}
	}
	return c.orderAPI
}

// Transaction returns the TransactionAPI for transaction operations
func (c *client) Transaction() core.TransactionAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transactionAPI == nil {
		c.transactionAPI = &transactionAPIImpl{client: c}
	}
	return c.transactionAPI
}

// Candlestick returns the CandlestickAPI for market data operations
func (c *client) Candlestick() core.CandlestickAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.candlestickAPI == nil {
		c.candlestickAPI = &candlestickAPIImpl{client: c}
	}
	return c.candlestickAPI
}

// Block returns the BlockAPI for block-related operations
func (c *client) Block() core.BlockAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.blockAPI == nil {
		c.blockAPI = &blockAPIImpl{client: c}
	}
	return c.blockAPI
}

// Bridge returns the BridgeAPI for bridge operations
func (c *client) Bridge() core.BridgeAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.bridgeAPI == nil {
		c.bridgeAPI = &bridgeAPIImpl{client: c}
	}
	return c.bridgeAPI
}

// Info returns the InfoAPI for general information
func (c *client) Info() core.InfoAPI {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.infoAPI == nil {
		c.infoAPI = &infoAPIImpl{client: c}
	}
	return c.infoAPI
}

// Endpoint returns the base URL of the client
func (c *client) Endpoint() string {
	return c.endpoint
}

package ws

import (
	"time"
)

// Options configures the WebSocket client behavior
type Options struct {
	// Connection settings
	PingInterval         time.Duration // Default: 30s
	PongTimeout          time.Duration // Default: 10s
	ReconnectDelay       time.Duration // Default: 1s
	MaxReconnectDelay    time.Duration // Default: 30s
	MaxReconnectAttempts int           // Default: 10 (0 = unlimited)

	// Channel buffer sizes
	OrderBookBufferSize   int // Default: 100
	TradeBufferSize       int // Default: 100
	MarketStatsBufferSize int // Default: 100
	HeightBufferSize      int // Default: 10
	AccountBufferSize     int // Default: 100
	TxResultBufferSize    int // Default: 100
	ErrorBufferSize       int // Default: 10

	// Callbacks (optional, for Python-style usage)
	OnConnect           func()
	OnDisconnect        func(error)
	OnOrderBookUpdate   func(*OrderBookUpdate)
	OnTradeUpdate       func(*TradeUpdate)
	OnMarketStatsUpdate func(*MarketStatsUpdate)
	OnHeightUpdate      func(*HeightUpdate)
	OnAccountUpdate     func(*AccountUpdate)
	OnTxResult          func(*TxResult)
	OnError             func(error)
}

// DefaultOptions returns the default WebSocket client options
func DefaultOptions() *Options {
	return &Options{
		PingInterval:          30 * time.Second,
		PongTimeout:           10 * time.Second,
		ReconnectDelay:        1 * time.Second,
		MaxReconnectDelay:     30 * time.Second,
		MaxReconnectAttempts:  10,
		OrderBookBufferSize:   100,
		TradeBufferSize:       100,
		MarketStatsBufferSize: 100,
		HeightBufferSize:      10,
		AccountBufferSize:     100,
		TxResultBufferSize:    100,
		ErrorBufferSize:       10,
	}
}

// WithPingInterval sets the ping interval
func (o *Options) WithPingInterval(d time.Duration) *Options {
	o.PingInterval = d
	return o
}

// WithPongTimeout sets the pong timeout
func (o *Options) WithPongTimeout(d time.Duration) *Options {
	o.PongTimeout = d
	return o
}

// WithReconnectDelay sets the initial reconnect delay
func (o *Options) WithReconnectDelay(d time.Duration) *Options {
	o.ReconnectDelay = d
	return o
}

// WithMaxReconnectDelay sets the maximum reconnect delay
func (o *Options) WithMaxReconnectDelay(d time.Duration) *Options {
	o.MaxReconnectDelay = d
	return o
}

// WithMaxReconnectAttempts sets the maximum reconnect attempts (0 = unlimited)
func (o *Options) WithMaxReconnectAttempts(n int) *Options {
	o.MaxReconnectAttempts = n
	return o
}

// WithOnConnect sets the connect callback
func (o *Options) WithOnConnect(fn func()) *Options {
	o.OnConnect = fn
	return o
}

// WithOnDisconnect sets the disconnect callback
func (o *Options) WithOnDisconnect(fn func(error)) *Options {
	o.OnDisconnect = fn
	return o
}

// WithOnOrderBookUpdate sets the order book update callback
func (o *Options) WithOnOrderBookUpdate(fn func(*OrderBookUpdate)) *Options {
	o.OnOrderBookUpdate = fn
	return o
}

// WithOnTradeUpdate sets the trade update callback
func (o *Options) WithOnTradeUpdate(fn func(*TradeUpdate)) *Options {
	o.OnTradeUpdate = fn
	return o
}

// WithOnMarketStatsUpdate sets the market stats update callback
func (o *Options) WithOnMarketStatsUpdate(fn func(*MarketStatsUpdate)) *Options {
	o.OnMarketStatsUpdate = fn
	return o
}

// WithOnHeightUpdate sets the height update callback
func (o *Options) WithOnHeightUpdate(fn func(*HeightUpdate)) *Options {
	o.OnHeightUpdate = fn
	return o
}

// WithOnAccountUpdate sets the account update callback
func (o *Options) WithOnAccountUpdate(fn func(*AccountUpdate)) *Options {
	o.OnAccountUpdate = fn
	return o
}

// WithOnTxResult sets the transaction result callback
func (o *Options) WithOnTxResult(fn func(*TxResult)) *Options {
	o.OnTxResult = fn
	return o
}

// WithOnError sets the error callback
func (o *Options) WithOnError(fn func(error)) *Options {
	o.OnError = fn
	return o
}

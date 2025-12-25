// Package ws provides a WebSocket client for real-time Lighter data streaming.
//
// Features:
//   - Order book streaming with automatic state management
//   - Account updates streaming
//   - Automatic reconnection with exponential backoff
//   - Ping/pong keepalive
//   - Both channel-based and callback-based APIs
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

// Client defines the WebSocket client interface
type Client interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Run(ctx context.Context) error // Blocking

	// Subscriptions
	SubscribeOrderBook(marketIndex int16) error
	UnsubscribeOrderBook(marketIndex int16) error
	SubscribeAccount(accountIndex int64, authToken string) error
	UnsubscribeAccount(accountIndex int64) error

	// Event channels (Go-idiomatic)
	OrderBookUpdates() <-chan *OrderBookUpdate
	AccountUpdates() <-chan *AccountUpdate
	Errors() <-chan error

	// State access
	GetOrderBookState(marketIndex int16) (*OrderBookState, error)

	// Status
	IsConnected() bool
}

type wsClient struct {
	// Configuration
	endpoint string
	options  *Options

	// Connection
	conn      *websocket.Conn
	connMu    sync.RWMutex
	connected atomic.Bool

	// Subscriptions
	subscriptions *subscriptionManager

	// Order book state
	orderBooks  map[int16]*OrderBookState
	orderBookMu sync.RWMutex

	// Event channels
	orderBookCh chan *OrderBookUpdate
	accountCh   chan *AccountUpdate
	errorCh     chan error

	// Lifecycle
	done   chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	// Reconnection
	reconnectAttempts int

	// Ping/pong
	lastPingTime time.Time
	lastPongTime time.Time
	pingMu       sync.Mutex
}

// NewClient creates a new WebSocket client
func NewClient(endpoint string, options *Options) Client {
	if options == nil {
		options = DefaultOptions()
	}

	return &wsClient{
		endpoint:      endpoint,
		options:       options,
		subscriptions: newSubscriptionManager(),
		orderBooks:    make(map[int16]*OrderBookState),
		orderBookCh:   make(chan *OrderBookUpdate, options.OrderBookBufferSize),
		accountCh:     make(chan *AccountUpdate, options.AccountBufferSize),
		errorCh:       make(chan error, options.ErrorBufferSize),
	}
}

// Connect establishes the WebSocket connection
func (c *wsClient) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.connected.Load() {
		return ErrAlreadyConnected
	}

	conn, _, err := websocket.Dial(ctx, c.endpoint, nil)
	if err != nil {
		return &ConnectionError{Err: err}
	}

	c.conn = conn
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.done = make(chan struct{})
	c.connected.Store(true)

	// Notify connect callback
	if c.options.OnConnect != nil {
		c.options.OnConnect()
	}

	// Start read loop
	c.wg.Add(1)
	go c.readLoop()

	// Start ping loop
	c.wg.Add(1)
	go c.pingLoop()

	return nil
}

// Close closes the WebSocket connection
func (c *wsClient) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if !c.connected.Load() && c.conn == nil {
		return nil
	}

	c.connected.Store(false)

	if c.cancel != nil {
		c.cancel()
	}

	if c.done != nil {
		close(c.done)
	}

	var err error
	if c.conn != nil {
		err = c.conn.Close(websocket.StatusNormalClosure, "client closing")
		c.conn = nil
	}

	c.wg.Wait()

	// Clear subscriptions
	c.subscriptions.Clear()

	// Notify disconnect callback
	if c.options.OnDisconnect != nil {
		c.options.OnDisconnect(nil)
	}

	return err
}

// Run connects and blocks until context is cancelled or connection fails
func (c *wsClient) Run(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// Block until context cancelled or connection closed
	select {
	case <-ctx.Done():
		return c.Close()
	case <-c.done:
		if c.options.MaxReconnectAttempts == 0 ||
			c.reconnectAttempts < c.options.MaxReconnectAttempts {
			return c.reconnect(ctx)
		}
		return ErrMaxReconnectAttemptsExceeded
	}
}

// SubscribeOrderBook subscribes to order book updates for a market
func (c *wsClient) SubscribeOrderBook(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddOrderBook(marketIndex)
	if err != nil {
		return err
	}

	req := SubscribeRequest{
		Action:  "subscribe",
		Channel: "orderbook",
		Market:  marketIndex,
	}

	if err := c.sendJSON(req); err != nil {
		c.subscriptions.Remove(orderBookKey(marketIndex))
		return err
	}

	// Wait for confirmation
	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(orderBookKey(marketIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeOrderBook unsubscribes from order book updates
func (c *wsClient) UnsubscribeOrderBook(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := orderBookKey(marketIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	req := SubscribeRequest{
		Action:  "unsubscribe",
		Channel: "orderbook",
		Market:  marketIndex,
	}

	return c.sendJSON(req)
}

// SubscribeAccount subscribes to account updates
func (c *wsClient) SubscribeAccount(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccount(accountIndex, authToken)
	if err != nil {
		return err
	}

	req := SubscribeRequest{
		Action:    "subscribe",
		Channel:   "account",
		Account:   accountIndex,
		AuthToken: authToken,
	}

	if err := c.sendJSON(req); err != nil {
		c.subscriptions.Remove(accountKey(accountIndex))
		return err
	}

	// Wait for confirmation
	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccount unsubscribes from account updates
func (c *wsClient) UnsubscribeAccount(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	req := SubscribeRequest{
		Action:  "unsubscribe",
		Channel: "account",
		Account: accountIndex,
	}

	return c.sendJSON(req)
}

// OrderBookUpdates returns the channel for order book updates
func (c *wsClient) OrderBookUpdates() <-chan *OrderBookUpdate {
	return c.orderBookCh
}

// AccountUpdates returns the channel for account updates
func (c *wsClient) AccountUpdates() <-chan *AccountUpdate {
	return c.accountCh
}

// Errors returns the channel for errors
func (c *wsClient) Errors() <-chan error {
	return c.errorCh
}

// GetOrderBookState returns a copy of the current order book state
func (c *wsClient) GetOrderBookState(marketIndex int16) (*OrderBookState, error) {
	c.orderBookMu.RLock()
	defer c.orderBookMu.RUnlock()

	state, exists := c.orderBooks[marketIndex]
	if !exists {
		return nil, ErrOrderBookNotFound
	}

	return state.Clone(), nil
}

// IsConnected returns true if connected
func (c *wsClient) IsConnected() bool {
	return c.connected.Load()
}

// Internal methods

func (c *wsClient) readLoop() {
	defer c.wg.Done()
	defer c.handleDisconnect(nil)

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, msg, err := c.conn.Read(c.ctx)
		if err != nil {
			c.handleDisconnect(err)
			return
		}

		if err := c.handleMessage(msg); err != nil {
			c.sendError(err)
		}
	}
}

func (c *wsClient) pingLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.options.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.sendPing(); err != nil {
				c.sendError(err)
			}
		}
	}
}

func (c *wsClient) sendPing() error {
	c.pingMu.Lock()
	c.lastPingTime = time.Now()
	c.pingMu.Unlock()

	return c.sendJSON(PingMessage{Action: "ping"})
}

func (c *wsClient) sendJSON(v interface{}) error {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.conn.Write(c.ctx, websocket.MessageText, data)
}

func (c *wsClient) handleDisconnect(err error) {
	c.connected.Store(false)

	if c.options.OnDisconnect != nil {
		c.options.OnDisconnect(err)
	}
}

func (c *wsClient) reconnect(ctx context.Context) error {
	c.reconnectAttempts++
	delay := c.calculateBackoff()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}

	if err := c.Connect(ctx); err != nil {
		return c.reconnect(ctx)
	}

	// Resubscribe to all active subscriptions
	return c.resubscribeAll()
}

func (c *wsClient) calculateBackoff() time.Duration {
	delay := c.options.ReconnectDelay * time.Duration(1<<c.reconnectAttempts)
	if delay > c.options.MaxReconnectDelay {
		delay = c.options.MaxReconnectDelay
	}
	return delay
}

func (c *wsClient) resubscribeAll() error {
	subs := c.subscriptions.GetAll()
	for _, sub := range subs {
		var err error
		switch sub.sType {
		case subscriptionOrderBook:
			var marketIndex int16
			fmt.Sscanf(sub.identifier, "%d", &marketIndex)
			err = c.SubscribeOrderBook(marketIndex)
		case subscriptionAccount:
			var accountIndex int64
			fmt.Sscanf(sub.identifier, "%d", &accountIndex)
			err = c.SubscribeAccount(accountIndex, sub.authToken)
		}
		if err != nil {
			c.sendError(err)
		}
	}
	return nil
}

// Ensure wsClient implements Client
var _ Client = (*wsClient)(nil)

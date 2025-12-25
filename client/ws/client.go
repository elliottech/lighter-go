// Package ws provides a WebSocket client for real-time Lighter data streaming.
//
// Features:
//   - Order book streaming with automatic state management
//   - Trade streaming
//   - Market stats streaming
//   - Block height streaming
//   - Account updates streaming (requires auth)
//   - Transaction sending via WebSocket
//   - Automatic reconnection with exponential backoff
//   - Ping/pong keepalive
//   - Both channel-based and callback-based APIs
package ws

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"nhooyr.io/websocket"
)

// Client defines the WebSocket client interface
type Client interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Run(ctx context.Context) error // Blocking

	// Public channel subscriptions
	SubscribeOrderBook(marketIndex int16) error
	UnsubscribeOrderBook(marketIndex int16) error
	SubscribeTrades(marketIndex int16) error
	UnsubscribeTrades(marketIndex int16) error
	SubscribeMarketStats(marketIndex int16) error
	SubscribeAllMarketStats() error
	UnsubscribeMarketStats(marketIndex int16) error
	UnsubscribeAllMarketStats() error
	SubscribeHeight() error
	UnsubscribeHeight() error

	// Private channel subscriptions (require auth token)
	SubscribeAccountAll(accountIndex int64, authToken string) error
	UnsubscribeAccountAll(accountIndex int64) error
	SubscribeAccountMarket(marketIndex int16, accountIndex int64, authToken string) error
	UnsubscribeAccountMarket(marketIndex int16, accountIndex int64) error
	SubscribeAccountOrders(marketIndex int16, accountIndex int64, authToken string) error
	UnsubscribeAccountOrders(marketIndex int16, accountIndex int64) error
	SubscribeAccountAllOrders(accountIndex int64, authToken string) error
	UnsubscribeAccountAllOrders(accountIndex int64) error
	SubscribeAccountAllTrades(accountIndex int64, authToken string) error
	UnsubscribeAccountAllTrades(accountIndex int64) error
	SubscribeAccountAllPositions(accountIndex int64, authToken string) error
	UnsubscribeAccountAllPositions(accountIndex int64) error
	SubscribeAccountTx(accountIndex int64, authToken string) error
	UnsubscribeAccountTx(accountIndex int64) error
	SubscribeUserStats(accountIndex int64, authToken string) error
	UnsubscribeUserStats(accountIndex int64) error
	SubscribePoolData(accountIndex int64, authToken string) error
	UnsubscribePoolData(accountIndex int64) error
	SubscribePoolInfo(accountIndex int64, authToken string) error
	UnsubscribePoolInfo(accountIndex int64) error
	SubscribeNotification(accountIndex int64, authToken string) error
	UnsubscribeNotification(accountIndex int64) error

	// Transaction sending via WebSocket
	SendTx(tx interface{}) error
	SendTxBatch(txs []interface{}) error

	// Event channels (Go-idiomatic)
	OrderBookUpdates() <-chan *OrderBookUpdate
	TradeUpdates() <-chan *TradeUpdate
	MarketStatsUpdates() <-chan *MarketStatsUpdate
	HeightUpdates() <-chan *HeightUpdate
	AccountUpdates() <-chan *AccountUpdate
	TxResults() <-chan *TxResult
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
	orderBookCh   chan *OrderBookUpdate
	tradeCh       chan *TradeUpdate
	marketStatsCh chan *MarketStatsUpdate
	heightCh      chan *HeightUpdate
	accountCh     chan *AccountUpdate
	txResultCh    chan *TxResult
	errorCh       chan error

	// Lifecycle
	done      chan struct{}
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	readyCh   chan struct{} // Signals when "connected" message received
	readyOnce sync.Once

	// Reconnection
	reconnectAttempts int
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
		tradeCh:       make(chan *TradeUpdate, options.TradeBufferSize),
		marketStatsCh: make(chan *MarketStatsUpdate, options.MarketStatsBufferSize),
		heightCh:      make(chan *HeightUpdate, options.HeightBufferSize),
		accountCh:     make(chan *AccountUpdate, options.AccountBufferSize),
		txResultCh:    make(chan *TxResult, options.TxResultBufferSize),
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

	// Set read limit to 10MB for large order books
	conn.SetReadLimit(10 * 1024 * 1024)

	c.conn = conn
	c.ctx, c.cancel = context.WithCancel(ctx)
	c.done = make(chan struct{})
	c.readyCh = make(chan struct{})
	c.readyOnce = sync.Once{}

	// Start read loop
	c.wg.Add(1)
	go c.readLoop()

	// Wait for "connected" message from server
	select {
	case <-c.readyCh:
		// Server acknowledged connection
	case <-time.After(10 * time.Second):
		c.conn.Close(websocket.StatusGoingAway, "connection timeout")
		return ErrConnectionTimeout
	case <-ctx.Done():
		c.conn.Close(websocket.StatusGoingAway, "context cancelled")
		return ctx.Err()
	}

	c.connected.Store(true)

	// Notify connect callback
	if c.options.OnConnect != nil {
		c.options.OnConnect()
	}

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

// subscribe is a helper for subscribing to channels
func (c *wsClient) subscribe(channel, authToken string) error {
	req := SubscribeRequest{
		Type:    "subscribe",
		Channel: channel,
	}
	if authToken != "" {
		req.Auth = authToken
	}
	return c.sendJSON(req)
}

// unsubscribe is a helper for unsubscribing from channels
func (c *wsClient) unsubscribe(channel string) error {
	req := SubscribeRequest{
		Type:    "unsubscribe",
		Channel: channel,
	}
	return c.sendJSON(req)
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

	channel := fmt.Sprintf("order_book/%d", marketIndex)
	if err := c.subscribe(channel, ""); err != nil {
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

	channel := fmt.Sprintf("order_book/%d", marketIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeTrades subscribes to trade updates for a market
func (c *wsClient) SubscribeTrades(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddTrade(marketIndex)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("trade/%d", marketIndex)
	if err := c.subscribe(channel, ""); err != nil {
		c.subscriptions.Remove(tradeKey(marketIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(tradeKey(marketIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeTrades unsubscribes from trade updates
func (c *wsClient) UnsubscribeTrades(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := tradeKey(marketIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("trade/%d", marketIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeMarketStats subscribes to market stats for a specific market
func (c *wsClient) SubscribeMarketStats(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddMarketStats(marketIndex)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("market_stats/%d", marketIndex)
	if err := c.subscribe(channel, ""); err != nil {
		c.subscriptions.Remove(marketStatsKey(marketIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(marketStatsKey(marketIndex))
		return ErrSubscriptionTimeout
	}
}

// SubscribeAllMarketStats subscribes to all market stats
func (c *wsClient) SubscribeAllMarketStats() error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddMarketStatsAll()
	if err != nil {
		return err
	}

	if err := c.subscribe("market_stats/all", ""); err != nil {
		c.subscriptions.Remove(marketStatsAllKey())
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(marketStatsAllKey())
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeMarketStats unsubscribes from market stats for a specific market
func (c *wsClient) UnsubscribeMarketStats(marketIndex int16) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := marketStatsKey(marketIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("market_stats/%d", marketIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// UnsubscribeAllMarketStats unsubscribes from all market stats
func (c *wsClient) UnsubscribeAllMarketStats() error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := marketStatsAllKey()
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	if err := c.unsubscribe("market_stats/all"); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeHeight subscribes to block height updates
func (c *wsClient) SubscribeHeight() error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddHeight()
	if err != nil {
		return err
	}

	if err := c.subscribe("height", ""); err != nil {
		c.subscriptions.Remove(heightKey())
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(heightKey())
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeHeight unsubscribes from block height updates
func (c *wsClient) UnsubscribeHeight() error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := heightKey()
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	if err := c.unsubscribe("height"); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountAll subscribes to all account updates
func (c *wsClient) SubscribeAccountAll(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccount(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_all/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountAll unsubscribes from all account updates
func (c *wsClient) UnsubscribeAccountAll(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_all/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountMarket subscribes to account updates for a specific market
func (c *wsClient) SubscribeAccountMarket(marketIndex int16, accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountMarket(marketIndex, accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_market/%d/%d", marketIndex, accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountMarketKey(marketIndex, accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountMarketKey(marketIndex, accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountMarket unsubscribes from account updates for a specific market
func (c *wsClient) UnsubscribeAccountMarket(marketIndex int16, accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountMarketKey(marketIndex, accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_market/%d/%d", marketIndex, accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountOrders subscribes to account orders for a specific market
func (c *wsClient) SubscribeAccountOrders(marketIndex int16, accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountOrders(marketIndex, accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_orders/%d/%d", marketIndex, accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountOrdersKey(marketIndex, accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountOrdersKey(marketIndex, accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountOrders unsubscribes from account orders
func (c *wsClient) UnsubscribeAccountOrders(marketIndex int16, accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountOrdersKey(marketIndex, accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_orders/%d/%d", marketIndex, accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountAllOrders subscribes to all account orders
func (c *wsClient) SubscribeAccountAllOrders(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountAllOrders(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_all_orders/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountAllOrdersKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountAllOrdersKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountAllOrders unsubscribes from all account orders
func (c *wsClient) UnsubscribeAccountAllOrders(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountAllOrdersKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_all_orders/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountAllTrades subscribes to all account trades
func (c *wsClient) SubscribeAccountAllTrades(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountAllTrades(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_all_trades/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountAllTradesKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountAllTradesKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountAllTrades unsubscribes from all account trades
func (c *wsClient) UnsubscribeAccountAllTrades(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountAllTradesKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_all_trades/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountAllPositions subscribes to all account positions
func (c *wsClient) SubscribeAccountAllPositions(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountAllPositions(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_all_positions/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountAllPositionsKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountAllPositionsKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountAllPositions unsubscribes from all account positions
func (c *wsClient) UnsubscribeAccountAllPositions(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountAllPositionsKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_all_positions/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeAccountTx subscribes to account transactions
func (c *wsClient) SubscribeAccountTx(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddAccountTx(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("account_tx/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(accountTxKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(accountTxKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeAccountTx unsubscribes from account transactions
func (c *wsClient) UnsubscribeAccountTx(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := accountTxKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("account_tx/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeUserStats subscribes to user stats
func (c *wsClient) SubscribeUserStats(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddUserStats(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("user_stats/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(userStatsKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(userStatsKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeUserStats unsubscribes from user stats
func (c *wsClient) UnsubscribeUserStats(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := userStatsKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("user_stats/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribePoolData subscribes to pool data
func (c *wsClient) SubscribePoolData(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddPoolData(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("pool_data/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(poolDataKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(poolDataKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribePoolData unsubscribes from pool data
func (c *wsClient) UnsubscribePoolData(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := poolDataKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("pool_data/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribePoolInfo subscribes to pool info
func (c *wsClient) SubscribePoolInfo(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddPoolInfo(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("pool_info/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(poolInfoKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(poolInfoKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribePoolInfo unsubscribes from pool info
func (c *wsClient) UnsubscribePoolInfo(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := poolInfoKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("pool_info/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SubscribeNotification subscribes to notifications
func (c *wsClient) SubscribeNotification(accountIndex int64, authToken string) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	confirmChan, err := c.subscriptions.AddNotification(accountIndex, authToken)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("notification/%d", accountIndex)
	if err := c.subscribe(channel, authToken); err != nil {
		c.subscriptions.Remove(notificationKey(accountIndex))
		return err
	}

	select {
	case err := <-confirmChan:
		return err
	case <-time.After(10 * time.Second):
		c.subscriptions.Remove(notificationKey(accountIndex))
		return ErrSubscriptionTimeout
	}
}

// UnsubscribeNotification unsubscribes from notifications
func (c *wsClient) UnsubscribeNotification(accountIndex int64) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	key := notificationKey(accountIndex)
	if !c.subscriptions.IsSubscribed(key) {
		return ErrNotSubscribed
	}

	channel := fmt.Sprintf("notification/%d", accountIndex)
	if err := c.unsubscribe(channel); err != nil {
		return err
	}

	return c.subscriptions.Remove(key)
}

// SendTx sends a single transaction via WebSocket
func (c *wsClient) SendTx(tx interface{}) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	req := SendTxRequest{
		Type: "jsonapi/sendtx",
		Data: tx,
	}
	return c.sendJSON(req)
}

// SendTxBatch sends multiple transactions via WebSocket (max 50)
func (c *wsClient) SendTxBatch(txs []interface{}) error {
	if !c.connected.Load() {
		return ErrNotConnected
	}

	if len(txs) > 50 {
		return ErrBatchTooLarge
	}

	req := SendTxBatchRequest{
		Type: "jsonapi/sendtxbatch",
		Data: txs,
	}
	return c.sendJSON(req)
}

// OrderBookUpdates returns the channel for order book updates
func (c *wsClient) OrderBookUpdates() <-chan *OrderBookUpdate {
	return c.orderBookCh
}

// TradeUpdates returns the channel for trade updates
func (c *wsClient) TradeUpdates() <-chan *TradeUpdate {
	return c.tradeCh
}

// MarketStatsUpdates returns the channel for market stats updates
func (c *wsClient) MarketStatsUpdates() <-chan *MarketStatsUpdate {
	return c.marketStatsCh
}

// HeightUpdates returns the channel for height updates
func (c *wsClient) HeightUpdates() <-chan *HeightUpdate {
	return c.heightCh
}

// AccountUpdates returns the channel for account updates
func (c *wsClient) AccountUpdates() <-chan *AccountUpdate {
	return c.accountCh
}

// TxResults returns the channel for transaction results
func (c *wsClient) TxResults() <-chan *TxResult {
	return c.txResultCh
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

// sendPong responds to server ping
func (c *wsClient) sendPong() error {
	return c.sendJSON(PongMessage{Type: "pong"})
}

func (c *wsClient) sendJSON(v interface{}) error {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	if c.conn == nil {
		return ErrNotConnected
	}

	data, err := sonic.Marshal(v)
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
		if err := c.subscribe(sub.channel, sub.authToken); err != nil {
			c.sendError(err)
		}
	}
	return nil
}

// Ensure wsClient implements Client
var _ Client = (*wsClient)(nil)

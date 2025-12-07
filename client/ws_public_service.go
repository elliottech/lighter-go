package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
)

// LighterWebsocketPublicService implements the new Bybit-style interface
type LighterWebsocketPublicService struct {
	wsClient   *WSClient
	ctx        context.Context
	cancel     context.CancelFunc
	errHandler ErrHandler
	mu         sync.RWMutex

	// Subscription management
	subscriptions map[string]*Subscription
}

type Subscription struct {
	key        string
	unsubFunc  func() error
	cancelFunc context.CancelFunc
}

// NewLighterWebsocketPublicService creates a new public service
func NewLighterWebsocketPublicService(config *WSConfig) *LighterWebsocketPublicService {
	if config == nil {
		config = DefaultWSConfig()
	}

	return &LighterWebsocketPublicService{
		wsClient:      NewWSClient(config),
		subscriptions: make(map[string]*Subscription),
	}
}

// Start implements LighterWebsocketPublicServiceI
func (s *LighterWebsocketPublicService) Start(ctx context.Context, errHandler ErrHandler) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.errHandler = errHandler

	// Set disconnect callback to notify error handler when connection is lost
	s.wsClient.SetOnDisconnected(func() {
		log.Println("[LighterWS] Public service WebSocket disconnected")
		if s.errHandler != nil {
			s.errHandler(fmt.Errorf("WebSocket connection lost"))
		}
	})

	if err := s.wsClient.Connect(s.ctx); err != nil {
		return fmt.Errorf("failed to connect websocket: %w", err)
	}

	log.Println("[LighterWS] Public service started")
	return nil
}

// Close implements LighterWebsocketPublicServiceI
func (s *LighterWebsocketPublicService) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	// Unsubscribe from all subscriptions
	s.mu.Lock()
	for _, sub := range s.subscriptions {
		if sub.cancelFunc != nil {
			sub.cancelFunc()
		}
	}
	s.subscriptions = make(map[string]*Subscription)
	s.mu.Unlock()

	if s.wsClient != nil {
		return s.wsClient.Disconnect()
	}

	log.Println("[LighterWS] Public service closed")
	return nil
}

// SubscribeOrderBook implements LighterWebsocketPublicServiceI
func (s *LighterWebsocketPublicService) SubscribeOrderBook(
	param LighterOrderBookParamKey,
	callback func(LighterOrderBookResponse) error,
) (func() error, error) {
	key := fmt.Sprintf("orderbook_%d", param.MarketId)

	// Check if already subscribed
	s.mu.RLock()
	if _, exists := s.subscriptions[key]; exists {
		s.mu.RUnlock()
		return nil, fmt.Errorf("already subscribed to order book for market %d", param.MarketId)
	}
	s.mu.RUnlock()

	// Create subscription context
	subCtx, subCancel := context.WithCancel(s.ctx)

	// Custom handler that converts internal updates to new format
	handler := func(marketId uint16, bids, asks []PriceLevel, timestamp int64, isSnapshot bool) error {
		response := LighterOrderBookResponse{
			MarketId:   marketId,
			Bids:       bids,
			Asks:       asks,
			Timestamp:  timestamp,
			IsSnapshot: isSnapshot,
		}
		return callback(response)
	}

	// Start the order book service with our custom handler
	err := s.startOrderBookService(subCtx, param.MarketId, handler)
	if err != nil {
		subCancel()
		return nil, fmt.Errorf("failed to start order book service: %w", err)
	}

	// Create unsubscribe function
	unsubFunc := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()

		if sub, exists := s.subscriptions[key]; exists {
			if sub.cancelFunc != nil {
				sub.cancelFunc()
			}
			delete(s.subscriptions, key)
			log.Printf("[LighterWS] Unsubscribed from order book market %d", param.MarketId)
		}
		return nil
	}

	// Store subscription
	s.mu.Lock()
	s.subscriptions[key] = &Subscription{
		key:        key,
		unsubFunc:  unsubFunc,
		cancelFunc: subCancel,
	}
	s.mu.Unlock()

	log.Printf("[LighterWS] Subscribed to order book market %d", param.MarketId)
	return unsubFunc, nil
}

// SubscribeTicker is not supported by Lighter - use UpdateBookTicker in wrapper instead
// This method exists to maintain interface compatibility but always returns an error
func (s *LighterWebsocketPublicService) SubscribeTicker() (func() error, error) {
	return func() error { return nil }, fmt.Errorf("ticker subscription not supported by Lighter - use UpdateBookTicker in wrapper instead")
}

// SubscribeTrades implements LighterWebsocketPublicServiceI
func (s *LighterWebsocketPublicService) SubscribeTrades(
	param LighterTradesParamKey,
	callback func(LighterTradesResponse) error,
) (func() error, error) {
	_ = fmt.Sprintf("trades_%d", param.MarketId) // TODO: implement

	// Implementation similar to SubscribeOrderBook...
	// For now, return a placeholder
	return func() error { return nil }, fmt.Errorf("trades subscription not yet implemented")
}

// SubscribeAccount implements LighterWebsocketPublicServiceI
func (s *LighterWebsocketPublicService) SubscribeAccount(
	param LighterAccountParamKey,
	callback func(LighterAccountResponse) error,
) (func() error, error) {
	_ = fmt.Sprintf("account_%d", param.AccountId) // TODO: implement

	// Implementation similar to SubscribeOrderBook...
	// For now, return a placeholder
	return func() error { return nil }, fmt.Errorf("account subscription not yet implemented")
}

// startOrderBookService is the internal method that handles order book subscriptions
func (s *LighterWebsocketPublicService) startOrderBookService(
	ctx context.Context,
	marketId uint16,
	handler func(uint16, []PriceLevel, []PriceLevel, int64, bool) error,
) error {
	// Subscribe to order book channel
	channel := fmt.Sprintf("order_book/%d", marketId)
	if err := s.wsClient.Subscribe(channel, ""); err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	// Add handlers for both snapshot and update messages
	snapshotHandler := func(data []byte) error {
		return s.handleOrderBookSnapshot(data, marketId, handler)
	}

	updateHandler := func(data []byte) error {
		return s.handleOrderBookUpdate(data, marketId, handler)
	}

	s.wsClient.AddHandler(MessageTypeOrderBookSubscribed, snapshotHandler)
	s.wsClient.AddHandler(MessageTypeOrderBookUpdate, updateHandler)

	// Wait for context cancellation
	go func() {
		<-ctx.Done()
		// Clean up handlers
		s.wsClient.RemoveHandler(MessageTypeOrderBookSubscribed)
		s.wsClient.RemoveHandler(MessageTypeOrderBookUpdate)
		// Unsubscribe
		s.wsClient.Unsubscribe(channel, "")
	}()

	return nil
}

// handleOrderBookSnapshot processes snapshot messages
func (s *LighterWebsocketPublicService) handleOrderBookSnapshot(
	data []byte,
	marketId uint16,
	handler func(uint16, []PriceLevel, []PriceLevel, int64, bool) error,
) error {
	// Parse snapshot message (similar to existing handleOrderBookSnapshot)
	var msg struct {
		Type      string `json:"type"`
		Channel   string `json:"channel"`
		OrderBook struct {
			Code   int            `json:"code"`
			Asks   []WSPriceLevel `json:"asks"`
			Bids   []WSPriceLevel `json:"bids"`
			Offset int64          `json:"offset"`
		} `json:"order_book"`
		Timestamp int64 `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal order book snapshot: %w", err)
	}

	// Extract market ID from channel if needed

	if parts := strings.Split(msg.Channel, ":"); len(parts) == 2 {
		if id, err := strconv.ParseUint(parts[1], 10, 8); err == nil {
			marketId = uint16(id)
		}
	}

	// Convert to PriceLevel format
	bids := make([]PriceLevel, 0, len(msg.OrderBook.Bids))
	for _, bid := range msg.OrderBook.Bids {
		bids = append(bids, PriceLevel{
			Price:    bid.Price,
			Quantity: bid.Size,
		})
	}

	asks := make([]PriceLevel, 0, len(msg.OrderBook.Asks))
	for _, ask := range msg.OrderBook.Asks {
		asks = append(asks, PriceLevel{
			Price:    ask.Price,
			Quantity: ask.Size,
		})
	}

	// Call handler with isSnapshot = true
	return handler(marketId, bids, asks, msg.Timestamp, true)
}

// handleOrderBookUpdate processes incremental update messages
func (s *LighterWebsocketPublicService) handleOrderBookUpdate(
	data []byte,
	marketId uint16,
	handler func(uint16, []PriceLevel, []PriceLevel, int64, bool) error,
) error {
	// Parse update message (similar to existing handleOrderBookUpdate)
	var msg struct {
		Type      string `json:"type"`
		Channel   string `json:"channel"`
		OrderBook struct {
			Code   int            `json:"code"`
			Asks   []WSPriceLevel `json:"asks"`
			Bids   []WSPriceLevel `json:"bids"`
			Offset int64          `json:"offset"`
		} `json:"order_book"`
		Timestamp int64 `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal order book update: %w", err)
	}

	// Extract market ID from channel if needed
	if marketId == 0 {
		if parts := strings.Split(msg.Channel, ":"); len(parts) == 2 {
			if id, err := strconv.ParseUint(parts[1], 10, 8); err == nil {
				marketId = uint16(id)
			}
		}
	}

	// Convert to PriceLevel format
	bids := make([]PriceLevel, 0, len(msg.OrderBook.Bids))
	for _, bid := range msg.OrderBook.Bids {
		bids = append(bids, PriceLevel{
			Price:    bid.Price,
			Quantity: bid.Size,
		})
	}

	asks := make([]PriceLevel, 0, len(msg.OrderBook.Asks))
	for _, ask := range msg.OrderBook.Asks {
		asks = append(asks, PriceLevel{
			Price:    ask.Price,
			Quantity: ask.Size,
		})
	}

	// Call handler with isSnapshot = false
	return handler(marketId, bids, asks, msg.Timestamp, false)
}

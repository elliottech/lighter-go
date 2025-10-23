package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// LighterWebsocketPrivateService implements the new Bybit-style private interface
type LighterWebsocketPrivateService struct {
	wsClient   *WSClient
	ctx        context.Context
	cancel     context.CancelFunc
	errHandler ErrHandler
	mu         sync.RWMutex

	// Subscription management
	subscriptions map[string]*Subscription
}

// TokenGenerator is a function type for generating auth tokens
type TokenGenerator func() string

// NewLighterWebsocketPrivateService creates a new private service
func NewLighterWebsocketPrivateService(config *WSConfig, tokenGen TokenGenerator) *LighterWebsocketPrivateService {
	if config == nil {
		config = DefaultWSConfig()
	}

	wsClient := NewWSClient(config)
	// Set authentication token if generator is provided
	if tokenGen != nil {
		token := tokenGen()
		if token != "" {
			wsClient.SetAuthToken(token)
		}
	}

	return &LighterWebsocketPrivateService{
		wsClient:      wsClient,
		subscriptions: make(map[string]*Subscription),
	}
}

// Start implements LighterWebsocketPrivateServiceI
func (s *LighterWebsocketPrivateService) Start(ctx context.Context, errHandler ErrHandler) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.errHandler = errHandler

	// Set disconnect callback to notify error handler when connection is lost
	s.wsClient.SetOnDisconnected(func() {
		log.Println("[LighterWS] Private service WebSocket disconnected")
		if s.errHandler != nil {
			s.errHandler(fmt.Errorf("WebSocket connection lost"))
		}
	})

	if err := s.wsClient.Connect(s.ctx); err != nil {
		return fmt.Errorf("failed to connect websocket: %w", err)
	}

	log.Println("[LighterWS] Private service started")
	return nil
}

// Close implements LighterWebsocketPrivateServiceI
func (s *LighterWebsocketPrivateService) Close() error {
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

	log.Println("[LighterWS] Private service closed")
	return nil
}

// SubscribeAccount implements LighterWebsocketPrivateServiceI
func (s *LighterWebsocketPrivateService) SubscribeAccount(
	param LighterAccountParamKey,
	callback func(LighterAccountResponse) error,
) (func() error, error) {
	key := fmt.Sprintf("account_%d", param.AccountId)

	// Check if already subscribed
	s.mu.RLock()
	if _, exists := s.subscriptions[key]; exists {
		s.mu.RUnlock()
		return nil, fmt.Errorf("already subscribed to account %d", param.AccountId)
	}
	s.mu.RUnlock()

	// Create subscription context
	subCtx, subCancel := context.WithCancel(s.ctx)

	// First, add handlers for both snapshot and update messages
	handler := func(data []byte) error {
		var accountUpdate WSAccountUpdate

		if err := json.Unmarshal(data, &accountUpdate); err != nil {
			return fmt.Errorf("failed to unmarshal account update: %v", err)
		}

		// Check if it's an account message (either snapshot or update)
		if accountUpdate.Type != MessageTypeAccount && accountUpdate.Type != MessageTypeAccountSubscribed {
			return nil
		}

		// Pass the raw WSAccountUpdate data directly to the callback
		response := LighterAccountResponse{
			AccountId:        accountUpdate.Account,
			AvailableBalance: "",  // Raw message doesn't have separate available balance
			MarketStats:      nil, // Will be populated with raw data
			Timestamp:        0,
			IsSnapshot:       accountUpdate.Type == MessageTypeAccountSubscribed,
			RawAccountUpdate: &accountUpdate,
		}
		return callback(response)
	}

	s.wsClient.AddHandler(MessageTypeAccount, handler)
	s.wsClient.AddHandler(MessageTypeAccountSubscribed, handler)

	// Connect to WebSocket first
	err := s.wsClient.Connect(subCtx)
	if err != nil {
		subCancel()
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	// Then subscribe to the account channel
	channel := fmt.Sprintf("account_all/%d", param.AccountId)
	if err := s.wsClient.Subscribe(channel, ""); err != nil {
		subCancel()
		return nil, fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
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
			log.Printf("[LighterWS] Unsubscribed from account %d", param.AccountId)
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

	log.Printf("[LighterWS] Subscribed to account %d", param.AccountId)
	return unsubFunc, nil
}

// SubscribeOrders implements LighterWebsocketPrivateServiceI
func (s *LighterWebsocketPrivateService) SubscribeOrders(
	param LighterOrdersParamKey,
	callback func(LighterOrdersResponse) error,
) (func() error, error) {
	// For now, orders are included in account updates
	// We can extract order-specific data from account updates
	return s.SubscribeAccount(
		LighterAccountParamKey{AccountId: param.AccountId},
		func(accountResponse LighterAccountResponse) error {
			// TODO: Extract order updates from account response if needed
			// For now, we don't have separate order updates
			log.Printf("[LighterWS] Order updates are included in account updates for account %d", param.AccountId)
			return nil
		},
	)
}

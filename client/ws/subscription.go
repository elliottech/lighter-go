package ws

import (
	"fmt"
	"sync"
	"time"
)

type subscriptionType int

const (
	subscriptionOrderBook subscriptionType = iota
	subscriptionAccount
)

type subscription struct {
	sType        subscriptionType
	identifier   string // market index for orderbook, account index for account
	active       bool
	authToken    string // for account subscriptions
	subscribedAt time.Time
}

type subscriptionManager struct {
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	pending       map[string]chan error // for subscription confirmations
}

func newSubscriptionManager() *subscriptionManager {
	return &subscriptionManager{
		subscriptions: make(map[string]*subscription),
		pending:       make(map[string]chan error),
	}
}

// orderBookKey returns the subscription key for an order book
func orderBookKey(marketIndex int16) string {
	return fmt.Sprintf("orderbook:%d", marketIndex)
}

// accountKey returns the subscription key for an account
func accountKey(accountIndex int64) string {
	return fmt.Sprintf("account:%d", accountIndex)
}

// AddOrderBook adds an order book subscription
func (sm *subscriptionManager) AddOrderBook(marketIndex int16) (chan error, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := orderBookKey(marketIndex)
	if sub, exists := sm.subscriptions[key]; exists && sub.active {
		return nil, ErrAlreadySubscribed
	}

	confirmChan := make(chan error, 1)
	sm.pending[key] = confirmChan
	sm.subscriptions[key] = &subscription{
		sType:      subscriptionOrderBook,
		identifier: fmt.Sprintf("%d", marketIndex),
		active:     false, // Will be set to true on confirmation
	}

	return confirmChan, nil
}

// AddAccount adds an account subscription
func (sm *subscriptionManager) AddAccount(accountIndex int64, authToken string) (chan error, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if authToken == "" {
		return nil, ErrAuthTokenRequired
	}

	key := accountKey(accountIndex)
	if sub, exists := sm.subscriptions[key]; exists && sub.active {
		return nil, ErrAlreadySubscribed
	}

	confirmChan := make(chan error, 1)
	sm.pending[key] = confirmChan
	sm.subscriptions[key] = &subscription{
		sType:      subscriptionAccount,
		identifier: fmt.Sprintf("%d", accountIndex),
		authToken:  authToken,
		active:     false,
	}

	return confirmChan, nil
}

// Remove removes a subscription
func (sm *subscriptionManager) Remove(key string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.subscriptions[key]; !exists {
		return ErrNotSubscribed
	}

	delete(sm.subscriptions, key)
	delete(sm.pending, key)
	return nil
}

// ConfirmSubscription confirms a subscription
func (sm *subscriptionManager) ConfirmSubscription(key string, err error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if confirmChan, exists := sm.pending[key]; exists {
		confirmChan <- err
		close(confirmChan)
		delete(sm.pending, key)
	}

	if err == nil {
		if sub, exists := sm.subscriptions[key]; exists {
			sub.active = true
			sub.subscribedAt = time.Now()
		}
	} else {
		delete(sm.subscriptions, key)
	}
}

// GetAll returns all active subscriptions
func (sm *subscriptionManager) GetAll() []*subscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*subscription, 0, len(sm.subscriptions))
	for _, sub := range sm.subscriptions {
		if sub.active {
			result = append(result, sub)
		}
	}
	return result
}

// IsSubscribed checks if a subscription exists and is active
func (sm *subscriptionManager) IsSubscribed(key string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sub, exists := sm.subscriptions[key]
	return exists && sub.active
}

// Clear removes all subscriptions
func (sm *subscriptionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Close all pending channels
	for _, ch := range sm.pending {
		close(ch)
	}

	sm.subscriptions = make(map[string]*subscription)
	sm.pending = make(map[string]chan error)
}

// GetSubscription returns a subscription by key
func (sm *subscriptionManager) GetSubscription(key string) (*subscription, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sub, exists := sm.subscriptions[key]
	return sub, exists
}

package ws

import (
	"sync"
	"testing"
	"time"
)

func TestNewSubscriptionManager(t *testing.T) {
	sm := newSubscriptionManager()

	if sm.subscriptions == nil {
		t.Error("subscriptions map should be initialized")
	}

	if sm.pending == nil {
		t.Error("pending map should be initialized")
	}
}

func TestSubscriptionManager_AddOrderBook(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, err := sm.AddOrderBook(0)
	if err != nil {
		t.Fatalf("AddOrderBook failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}

	// Verify subscription was added
	key := orderBookKey(0)
	sub, exists := sm.GetSubscription(key)
	if !exists {
		t.Error("subscription should exist")
	}

	if sub.channelType != ChannelOrderBook {
		t.Errorf("expected channel type %v, got %v", ChannelOrderBook, sub.channelType)
	}

	if sub.channel != "order_book/0" {
		t.Errorf("expected channel 'order_book/0', got %s", sub.channel)
	}

	if sub.active {
		t.Error("subscription should not be active yet")
	}
}

func TestSubscriptionManager_AddTrade(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, err := sm.AddTrade(1)
	if err != nil {
		t.Fatalf("AddTrade failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}

	key := tradeKey(1)
	sub, exists := sm.GetSubscription(key)
	if !exists {
		t.Error("subscription should exist")
	}

	if sub.channelType != ChannelTrade {
		t.Errorf("expected channel type %v, got %v", ChannelTrade, sub.channelType)
	}
}

func TestSubscriptionManager_AddMarketStats(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, err := sm.AddMarketStats(2)
	if err != nil {
		t.Fatalf("AddMarketStats failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}

	key := marketStatsKey(2)
	sub, exists := sm.GetSubscription(key)
	if !exists {
		t.Error("subscription should exist")
	}

	if sub.channelType != ChannelMarketStats {
		t.Errorf("expected channel type %v, got %v", ChannelMarketStats, sub.channelType)
	}
}

func TestSubscriptionManager_AddMarketStatsAll(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, err := sm.AddMarketStatsAll()
	if err != nil {
		t.Fatalf("AddMarketStatsAll failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}

	key := marketStatsAllKey()
	sub, exists := sm.GetSubscription(key)
	if !exists {
		t.Error("subscription should exist")
	}

	if sub.channel != "market_stats/all" {
		t.Errorf("expected channel 'market_stats/all', got %s", sub.channel)
	}
}

func TestSubscriptionManager_AddHeight(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, err := sm.AddHeight()
	if err != nil {
		t.Fatalf("AddHeight failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}

	key := heightKey()
	sub, exists := sm.GetSubscription(key)
	if !exists {
		t.Error("subscription should exist")
	}

	if sub.channelType != ChannelHeight {
		t.Errorf("expected channel type %v, got %v", ChannelHeight, sub.channelType)
	}
}

func TestSubscriptionManager_AddAccount_RequiresAuth(t *testing.T) {
	sm := newSubscriptionManager()

	// Without auth token
	_, err := sm.AddAccount(123, "")
	if err != ErrAuthTokenRequired {
		t.Errorf("expected ErrAuthTokenRequired, got %v", err)
	}

	// With auth token
	confirmChan, err := sm.AddAccount(123, "valid-token")
	if err != nil {
		t.Fatalf("AddAccount with token failed: %v", err)
	}

	if confirmChan == nil {
		t.Error("confirmChan should not be nil")
	}
}

func TestSubscriptionManager_AlreadySubscribed(t *testing.T) {
	sm := newSubscriptionManager()

	// First subscription
	_, err := sm.AddOrderBook(0)
	if err != nil {
		t.Fatalf("First AddOrderBook failed: %v", err)
	}

	// Confirm it
	sm.ConfirmSubscription(orderBookKey(0), nil)

	// Try to subscribe again
	_, err = sm.AddOrderBook(0)
	if err != ErrAlreadySubscribed {
		t.Errorf("expected ErrAlreadySubscribed, got %v", err)
	}
}

func TestSubscriptionManager_ConfirmSubscription_Success(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, _ := sm.AddOrderBook(0)
	key := orderBookKey(0)

	// Confirm in goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		sm.ConfirmSubscription(key, nil)
	}()

	// Wait for confirmation
	select {
	case err := <-confirmChan:
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("confirmation timeout")
	}

	// Subscription should now be active
	sub, _ := sm.GetSubscription(key)
	if !sub.active {
		t.Error("subscription should be active after confirmation")
	}

	if sub.subscribedAt.IsZero() {
		t.Error("subscribedAt should be set")
	}
}

func TestSubscriptionManager_ConfirmSubscription_Error(t *testing.T) {
	sm := newSubscriptionManager()

	confirmChan, _ := sm.AddOrderBook(0)
	key := orderBookKey(0)

	// Confirm with error in goroutine
	go func() {
		time.Sleep(10 * time.Millisecond)
		sm.ConfirmSubscription(key, ErrSubscriptionFailed)
	}()

	// Wait for confirmation
	select {
	case err := <-confirmChan:
		if err != ErrSubscriptionFailed {
			t.Errorf("expected ErrSubscriptionFailed, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("confirmation timeout")
	}

	// Subscription should be removed on error
	_, exists := sm.GetSubscription(key)
	if exists {
		t.Error("subscription should be removed after error")
	}
}

func TestSubscriptionManager_Remove(t *testing.T) {
	sm := newSubscriptionManager()

	_, _ = sm.AddOrderBook(0)
	key := orderBookKey(0)
	sm.ConfirmSubscription(key, nil)

	// Verify it exists
	if !sm.IsSubscribed(key) {
		t.Error("subscription should exist before removal")
	}

	// Remove it
	err := sm.Remove(key)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify it's gone
	if sm.IsSubscribed(key) {
		t.Error("subscription should not exist after removal")
	}
}

func TestSubscriptionManager_Remove_NotSubscribed(t *testing.T) {
	sm := newSubscriptionManager()

	err := sm.Remove("nonexistent")
	if err != ErrNotSubscribed {
		t.Errorf("expected ErrNotSubscribed, got %v", err)
	}
}

func TestSubscriptionManager_IsSubscribed(t *testing.T) {
	sm := newSubscriptionManager()

	key := orderBookKey(0)

	// Not subscribed
	if sm.IsSubscribed(key) {
		t.Error("should not be subscribed initially")
	}

	// Add subscription but not confirmed
	_, _ = sm.AddOrderBook(0)
	if sm.IsSubscribed(key) {
		t.Error("should not be subscribed until confirmed")
	}

	// Confirm
	sm.ConfirmSubscription(key, nil)
	if !sm.IsSubscribed(key) {
		t.Error("should be subscribed after confirmation")
	}
}

func TestSubscriptionManager_GetAll(t *testing.T) {
	sm := newSubscriptionManager()

	// Add multiple subscriptions
	_, _ = sm.AddOrderBook(0)
	_, _ = sm.AddOrderBook(1)
	_, _ = sm.AddTrade(0)

	// Before confirmation
	active := sm.GetAll()
	if len(active) != 0 {
		t.Errorf("expected 0 active subscriptions, got %d", len(active))
	}

	// Confirm some
	sm.ConfirmSubscription(orderBookKey(0), nil)
	sm.ConfirmSubscription(tradeKey(0), nil)

	active = sm.GetAll()
	if len(active) != 2 {
		t.Errorf("expected 2 active subscriptions, got %d", len(active))
	}
}

func TestSubscriptionManager_Clear(t *testing.T) {
	sm := newSubscriptionManager()

	// Add subscriptions
	_, _ = sm.AddOrderBook(0)
	_, _ = sm.AddTrade(0)
	sm.ConfirmSubscription(orderBookKey(0), nil)

	// Clear
	sm.Clear()

	// Verify all cleared
	if sm.IsSubscribed(orderBookKey(0)) {
		t.Error("subscriptions should be cleared")
	}

	if len(sm.subscriptions) != 0 {
		t.Errorf("subscriptions map should be empty, got %d", len(sm.subscriptions))
	}

	if len(sm.pending) != 0 {
		t.Errorf("pending map should be empty, got %d", len(sm.pending))
	}
}

func TestSubscriptionManager_PrivateChannels(t *testing.T) {
	sm := newSubscriptionManager()

	tests := []struct {
		name string
		add  func() (chan error, error)
	}{
		{"AccountMarket", func() (chan error, error) { return sm.AddAccountMarket(0, 123, "token") }},
		{"AccountOrders", func() (chan error, error) { return sm.AddAccountOrders(0, 123, "token") }},
		{"AccountAllOrders", func() (chan error, error) { return sm.AddAccountAllOrders(123, "token") }},
		{"AccountAllTrades", func() (chan error, error) { return sm.AddAccountAllTrades(123, "token") }},
		{"AccountAllPositions", func() (chan error, error) { return sm.AddAccountAllPositions(123, "token") }},
		{"AccountTx", func() (chan error, error) { return sm.AddAccountTx(123, "token") }},
		{"UserStats", func() (chan error, error) { return sm.AddUserStats(123, "token") }},
		{"PoolData", func() (chan error, error) { return sm.AddPoolData(123, "token") }},
		{"PoolInfo", func() (chan error, error) { return sm.AddPoolInfo(123, "token") }},
		{"Notification", func() (chan error, error) { return sm.AddNotification(123, "token") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmChan, err := tt.add()
			if err != nil {
				t.Errorf("Add%s failed: %v", tt.name, err)
				return
			}
			if confirmChan == nil {
				t.Error("confirmChan should not be nil")
			}
		})
	}
}

func TestSubscriptionManager_ConcurrentAccess(t *testing.T) {
	sm := newSubscriptionManager()

	var wg sync.WaitGroup
	const goroutines = 50

	wg.Add(goroutines * 4)

	// Add subscriptions
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			sm.AddOrderBook(int16(i % 10))
		}(i)
	}

	// Confirm subscriptions
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := orderBookKey(int16(i % 10))
			sm.ConfirmSubscription(key, nil)
		}(i)
	}

	// Check subscriptions
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := orderBookKey(int16(i % 10))
			sm.IsSubscribed(key)
		}(i)
	}

	// Get all
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			sm.GetAll()
		}()
	}

	wg.Wait()
}

// Test key generation functions
func TestKeyGeneration(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"orderBookKey", func() string { return orderBookKey(5) }, "order_book:5"},
		{"tradeKey", func() string { return tradeKey(3) }, "trade:3"},
		{"marketStatsKey", func() string { return marketStatsKey(2) }, "market_stats:2"},
		{"marketStatsAllKey", marketStatsAllKey, "market_stats:all"},
		{"heightKey", heightKey, "height"},
		{"accountKey", func() string { return accountKey(123) }, "account_all:123"},
		{"accountMarketKey", func() string { return accountMarketKey(0, 123) }, "account_market:0:123"},
		{"accountOrdersKey", func() string { return accountOrdersKey(1, 456) }, "account_orders:1:456"},
		{"accountAllOrdersKey", func() string { return accountAllOrdersKey(789) }, "account_all_orders:789"},
		{"accountAllTradesKey", func() string { return accountAllTradesKey(111) }, "account_all_trades:111"},
		{"accountAllPositionsKey", func() string { return accountAllPositionsKey(222) }, "account_all_positions:222"},
		{"accountTxKey", func() string { return accountTxKey(333) }, "account_tx:333"},
		{"userStatsKey", func() string { return userStatsKey(444) }, "user_stats:444"},
		{"poolDataKey", func() string { return poolDataKey(555) }, "pool_data:555"},
		{"poolInfoKey", func() string { return poolInfoKey(666) }, "pool_info:666"},
		{"notificationKey", func() string { return notificationKey(777) }, "notification:777"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestChannelType_IsPrivate(t *testing.T) {
	publicChannels := []ChannelType{
		ChannelOrderBook,
		ChannelTrade,
		ChannelMarketStats,
		ChannelHeight,
	}

	privateChannels := []ChannelType{
		ChannelAccountAll,
		ChannelAccountMarket,
		ChannelAccountOrders,
		ChannelAccountAllOrders,
		ChannelAccountAllTrades,
		ChannelAccountAllPositions,
		ChannelAccountTx,
		ChannelUserStats,
		ChannelPoolData,
		ChannelPoolInfo,
		ChannelNotification,
	}

	for _, ct := range publicChannels {
		if ct.IsPrivate() {
			t.Errorf("%s should be public", ct)
		}
	}

	for _, ct := range privateChannels {
		if !ct.IsPrivate() {
			t.Errorf("%s should be private", ct)
		}
	}
}

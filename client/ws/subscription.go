package ws

import (
	"fmt"
	"sync"
	"time"
)

// ChannelType represents the type of WebSocket channel
type ChannelType string

const (
	// Public channels
	ChannelOrderBook   ChannelType = "order_book"
	ChannelTrade       ChannelType = "trade"
	ChannelMarketStats ChannelType = "market_stats"
	ChannelHeight      ChannelType = "height"

	// Private channels (require auth)
	ChannelAccountAll          ChannelType = "account_all"
	ChannelAccountMarket       ChannelType = "account_market"
	ChannelAccountOrders       ChannelType = "account_orders"
	ChannelAccountAllOrders    ChannelType = "account_all_orders"
	ChannelAccountAllTrades    ChannelType = "account_all_trades"
	ChannelAccountAllPositions ChannelType = "account_all_positions"
	ChannelAccountTx           ChannelType = "account_tx"
	ChannelUserStats           ChannelType = "user_stats"
	ChannelPoolData            ChannelType = "pool_data"
	ChannelPoolInfo            ChannelType = "pool_info"
	ChannelNotification        ChannelType = "notification"
)

// IsPrivate returns true if the channel requires authentication
func (ct ChannelType) IsPrivate() bool {
	switch ct {
	case ChannelOrderBook, ChannelTrade, ChannelMarketStats, ChannelHeight:
		return false
	default:
		return true
	}
}

type subscription struct {
	channelType  ChannelType
	channel      string // full channel path (e.g., "order_book/0", "account_all/123")
	identifier   string // market index, account index, etc.
	active       bool
	authToken    string // for private channels
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

// Key generation functions

func orderBookKey(marketIndex int16) string {
	return fmt.Sprintf("order_book:%d", marketIndex)
}

func tradeKey(marketIndex int16) string {
	return fmt.Sprintf("trade:%d", marketIndex)
}

func marketStatsKey(marketIndex int16) string {
	return fmt.Sprintf("market_stats:%d", marketIndex)
}

func marketStatsAllKey() string {
	return "market_stats:all"
}

func heightKey() string {
	return "height"
}

func accountKey(accountIndex int64) string {
	return fmt.Sprintf("account_all:%d", accountIndex)
}

func accountMarketKey(marketIndex int16, accountIndex int64) string {
	return fmt.Sprintf("account_market:%d:%d", marketIndex, accountIndex)
}

func accountOrdersKey(marketIndex int16, accountIndex int64) string {
	return fmt.Sprintf("account_orders:%d:%d", marketIndex, accountIndex)
}

func accountAllOrdersKey(accountIndex int64) string {
	return fmt.Sprintf("account_all_orders:%d", accountIndex)
}

func accountAllTradesKey(accountIndex int64) string {
	return fmt.Sprintf("account_all_trades:%d", accountIndex)
}

func accountAllPositionsKey(accountIndex int64) string {
	return fmt.Sprintf("account_all_positions:%d", accountIndex)
}

func accountTxKey(accountIndex int64) string {
	return fmt.Sprintf("account_tx:%d", accountIndex)
}

func userStatsKey(accountIndex int64) string {
	return fmt.Sprintf("user_stats:%d", accountIndex)
}

func poolDataKey(accountIndex int64) string {
	return fmt.Sprintf("pool_data:%d", accountIndex)
}

func poolInfoKey(accountIndex int64) string {
	return fmt.Sprintf("pool_info:%d", accountIndex)
}

func notificationKey(accountIndex int64) string {
	return fmt.Sprintf("notification:%d", accountIndex)
}

// AddSubscription adds a generic subscription
func (sm *subscriptionManager) AddSubscription(key string, channelType ChannelType, channel, identifier, authToken string) (chan error, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sub, exists := sm.subscriptions[key]; exists && sub.active {
		return nil, ErrAlreadySubscribed
	}

	// Private channels require auth token
	if channelType.IsPrivate() && authToken == "" {
		return nil, ErrAuthTokenRequired
	}

	confirmChan := make(chan error, 1)
	sm.pending[key] = confirmChan
	sm.subscriptions[key] = &subscription{
		channelType: channelType,
		channel:     channel,
		identifier:  identifier,
		authToken:   authToken,
		active:      false,
	}

	return confirmChan, nil
}

// AddOrderBook adds an order book subscription
func (sm *subscriptionManager) AddOrderBook(marketIndex int16) (chan error, error) {
	key := orderBookKey(marketIndex)
	channel := fmt.Sprintf("order_book/%d", marketIndex)
	return sm.AddSubscription(key, ChannelOrderBook, channel, fmt.Sprintf("%d", marketIndex), "")
}

// AddTrade adds a trade subscription
func (sm *subscriptionManager) AddTrade(marketIndex int16) (chan error, error) {
	key := tradeKey(marketIndex)
	channel := fmt.Sprintf("trade/%d", marketIndex)
	return sm.AddSubscription(key, ChannelTrade, channel, fmt.Sprintf("%d", marketIndex), "")
}

// AddMarketStats adds a market stats subscription
func (sm *subscriptionManager) AddMarketStats(marketIndex int16) (chan error, error) {
	key := marketStatsKey(marketIndex)
	channel := fmt.Sprintf("market_stats/%d", marketIndex)
	return sm.AddSubscription(key, ChannelMarketStats, channel, fmt.Sprintf("%d", marketIndex), "")
}

// AddMarketStatsAll adds a subscription to all market stats
func (sm *subscriptionManager) AddMarketStatsAll() (chan error, error) {
	key := marketStatsAllKey()
	channel := "market_stats/all"
	return sm.AddSubscription(key, ChannelMarketStats, channel, "all", "")
}

// AddHeight adds a height subscription
func (sm *subscriptionManager) AddHeight() (chan error, error) {
	key := heightKey()
	channel := "height"
	return sm.AddSubscription(key, ChannelHeight, channel, "", "")
}

// AddAccount adds an account_all subscription
func (sm *subscriptionManager) AddAccount(accountIndex int64, authToken string) (chan error, error) {
	key := accountKey(accountIndex)
	channel := fmt.Sprintf("account_all/%d", accountIndex)
	return sm.AddSubscription(key, ChannelAccountAll, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddAccountMarket adds an account_market subscription
func (sm *subscriptionManager) AddAccountMarket(marketIndex int16, accountIndex int64, authToken string) (chan error, error) {
	key := accountMarketKey(marketIndex, accountIndex)
	channel := fmt.Sprintf("account_market/%d/%d", marketIndex, accountIndex)
	return sm.AddSubscription(key, ChannelAccountMarket, channel, fmt.Sprintf("%d:%d", marketIndex, accountIndex), authToken)
}

// AddAccountOrders adds an account_orders subscription
func (sm *subscriptionManager) AddAccountOrders(marketIndex int16, accountIndex int64, authToken string) (chan error, error) {
	key := accountOrdersKey(marketIndex, accountIndex)
	channel := fmt.Sprintf("account_orders/%d/%d", marketIndex, accountIndex)
	return sm.AddSubscription(key, ChannelAccountOrders, channel, fmt.Sprintf("%d:%d", marketIndex, accountIndex), authToken)
}

// AddAccountAllOrders adds an account_all_orders subscription
func (sm *subscriptionManager) AddAccountAllOrders(accountIndex int64, authToken string) (chan error, error) {
	key := accountAllOrdersKey(accountIndex)
	channel := fmt.Sprintf("account_all_orders/%d", accountIndex)
	return sm.AddSubscription(key, ChannelAccountAllOrders, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddAccountAllTrades adds an account_all_trades subscription
func (sm *subscriptionManager) AddAccountAllTrades(accountIndex int64, authToken string) (chan error, error) {
	key := accountAllTradesKey(accountIndex)
	channel := fmt.Sprintf("account_all_trades/%d", accountIndex)
	return sm.AddSubscription(key, ChannelAccountAllTrades, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddAccountAllPositions adds an account_all_positions subscription
func (sm *subscriptionManager) AddAccountAllPositions(accountIndex int64, authToken string) (chan error, error) {
	key := accountAllPositionsKey(accountIndex)
	channel := fmt.Sprintf("account_all_positions/%d", accountIndex)
	return sm.AddSubscription(key, ChannelAccountAllPositions, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddAccountTx adds an account_tx subscription
func (sm *subscriptionManager) AddAccountTx(accountIndex int64, authToken string) (chan error, error) {
	key := accountTxKey(accountIndex)
	channel := fmt.Sprintf("account_tx/%d", accountIndex)
	return sm.AddSubscription(key, ChannelAccountTx, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddUserStats adds a user_stats subscription
func (sm *subscriptionManager) AddUserStats(accountIndex int64, authToken string) (chan error, error) {
	key := userStatsKey(accountIndex)
	channel := fmt.Sprintf("user_stats/%d", accountIndex)
	return sm.AddSubscription(key, ChannelUserStats, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddPoolData adds a pool_data subscription
func (sm *subscriptionManager) AddPoolData(accountIndex int64, authToken string) (chan error, error) {
	key := poolDataKey(accountIndex)
	channel := fmt.Sprintf("pool_data/%d", accountIndex)
	return sm.AddSubscription(key, ChannelPoolData, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddPoolInfo adds a pool_info subscription
func (sm *subscriptionManager) AddPoolInfo(accountIndex int64, authToken string) (chan error, error) {
	key := poolInfoKey(accountIndex)
	channel := fmt.Sprintf("pool_info/%d", accountIndex)
	return sm.AddSubscription(key, ChannelPoolInfo, channel, fmt.Sprintf("%d", accountIndex), authToken)
}

// AddNotification adds a notification subscription
func (sm *subscriptionManager) AddNotification(accountIndex int64, authToken string) (chan error, error) {
	key := notificationKey(accountIndex)
	channel := fmt.Sprintf("notification/%d", accountIndex)
	return sm.AddSubscription(key, ChannelNotification, channel, fmt.Sprintf("%d", accountIndex), authToken)
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

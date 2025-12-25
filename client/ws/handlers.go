package ws

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
)

// handleMessage routes incoming messages to appropriate handlers
func (c *wsClient) handleMessage(msg []byte) error {
	var base BaseMessage
	if err := sonic.Unmarshal(msg, &base); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// Handle messages with explicit type field
	if base.Type != "" {
		return c.handleTypedMessage(base)
	}

	// Handle messages without type field (initial snapshots from subscriptions)
	// These messages have a channel field and data
	if base.Channel != "" {
		return c.handleChannelMessage(base, msg)
	}

	// Unknown message format
	return nil
}

// handleTypedMessage handles messages that have an explicit type field
func (c *wsClient) handleTypedMessage(base BaseMessage) error {
	switch base.Type {
	// Connection messages
	case MessageTypeConnected:
		return c.handleConnected()
	case MessageTypePing:
		return c.sendPong()
	case MessageTypeError:
		return c.handleError(base.Data)

	// Order book messages
	case MessageTypeSubscribedOrderBook:
		return c.handleSubscribedOrderBook(base.Channel, base.OrderBook)
	case MessageTypeUpdateOrderBook:
		return c.handleOrderBookUpdate(base.Channel, base.OrderBook)

	// Trade messages
	case MessageTypeSubscribedTrade:
		return c.handleSubscribedTrade(base.Channel)
	case MessageTypeUpdateTrade:
		return c.handleTradeUpdate(base.Channel, base.Data)

	// Market stats messages
	case MessageTypeSubscribedMarketStats:
		return c.handleSubscribedMarketStats(base.Channel)
	case MessageTypeUpdateMarketStats:
		return c.handleMarketStatsUpdate(base.Channel, base.Data)

	// Height messages
	case MessageTypeSubscribedHeight:
		return c.handleSubscribedHeight()
	case MessageTypeUpdateHeight:
		return c.handleHeightUpdate(base.Data)

	// Account messages
	case MessageTypeSubscribedAccountAll:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAll)
	case MessageTypeUpdateAccountAll:
		return c.handleAccountUpdate(base.Channel, ChannelAccountAll, base.Data)
	case MessageTypeSubscribedAccountMarket:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountMarket)
	case MessageTypeUpdateAccountMarket:
		return c.handleAccountUpdate(base.Channel, ChannelAccountMarket, base.Data)
	case MessageTypeSubscribedAccountOrders:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountOrders)
	case MessageTypeUpdateAccountOrders:
		return c.handleAccountUpdate(base.Channel, ChannelAccountOrders, base.Data)
	case MessageTypeSubscribedAccountAllOrders:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllOrders)
	case MessageTypeUpdateAccountAllOrders:
		return c.handleAccountUpdate(base.Channel, ChannelAccountAllOrders, base.Data)
	case MessageTypeSubscribedAccountAllTrades:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllTrades)
	case MessageTypeUpdateAccountAllTrades:
		return c.handleAccountUpdate(base.Channel, ChannelAccountAllTrades, base.Data)
	case MessageTypeSubscribedAccountAllPositions:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllPositions)
	case MessageTypeUpdateAccountAllPositions:
		return c.handleAccountUpdate(base.Channel, ChannelAccountAllPositions, base.Data)
	case MessageTypeSubscribedAccountTx:
		return c.handleSubscribedAccount(base.Channel, ChannelAccountTx)
	case MessageTypeUpdateAccountTx:
		return c.handleAccountUpdate(base.Channel, ChannelAccountTx, base.Data)
	case MessageTypeSubscribedUserStats:
		return c.handleSubscribedAccount(base.Channel, ChannelUserStats)
	case MessageTypeUpdateUserStats:
		return c.handleAccountUpdate(base.Channel, ChannelUserStats, base.Data)
	case MessageTypeSubscribedPoolData:
		return c.handleSubscribedAccount(base.Channel, ChannelPoolData)
	case MessageTypeUpdatePoolData:
		return c.handleAccountUpdate(base.Channel, ChannelPoolData, base.Data)
	case MessageTypeSubscribedPoolInfo:
		return c.handleSubscribedAccount(base.Channel, ChannelPoolInfo)
	case MessageTypeUpdatePoolInfo:
		return c.handleAccountUpdate(base.Channel, ChannelPoolInfo, base.Data)
	case MessageTypeSubscribedNotification:
		return c.handleSubscribedAccount(base.Channel, ChannelNotification)
	case MessageTypeUpdateNotification:
		return c.handleAccountUpdate(base.Channel, ChannelNotification, base.Data)

	// Transaction result messages
	case MessageTypeTxResult:
		return c.handleTxResult(base.Data)
	case MessageTypeTxBatchResult:
		return c.handleTxBatchResult(base.Data)

	default:
		// Ignore unknown message types
		return nil
	}
}

// handleChannelMessage handles messages without a type field (initial subscription snapshots)
// The server sends the initial data without a type field when a subscription is confirmed
func (c *wsClient) handleChannelMessage(base BaseMessage, rawMsg []byte) error {
	parts := parseChannelParts(base.Channel)
	if len(parts) == 0 {
		return nil
	}

	channelPrefix := parts[0]

	switch channelPrefix {
	case "order_book":
		// Initial order book snapshot - treat as subscription confirmation with data
		return c.handleSubscribedOrderBook(base.Channel, base.OrderBook)

	case "trade":
		// Initial trade subscription
		if err := c.handleSubscribedTrade(base.Channel); err != nil {
			return err
		}
		// If there's data, also handle it as an update
		if len(base.Data) > 0 {
			return c.handleTradeUpdate(base.Channel, base.Data)
		}
		return nil

	case "market_stats":
		// Initial market stats subscription
		if err := c.handleSubscribedMarketStats(base.Channel); err != nil {
			return err
		}
		if len(base.Data) > 0 {
			return c.handleMarketStatsUpdate(base.Channel, base.Data)
		}
		return nil

	case "height":
		// Initial height subscription
		if err := c.handleSubscribedHeight(); err != nil {
			return err
		}
		if len(base.Data) > 0 {
			return c.handleHeightUpdate(base.Data)
		}
		return nil

	case "account_all":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAll)
	case "account_market":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountMarket)
	case "account_orders":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountOrders)
	case "account_all_orders":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllOrders)
	case "account_all_trades":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllTrades)
	case "account_all_positions":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountAllPositions)
	case "account_tx":
		return c.handleSubscribedAccount(base.Channel, ChannelAccountTx)
	case "user_stats":
		return c.handleSubscribedAccount(base.Channel, ChannelUserStats)
	case "pool_data":
		return c.handleSubscribedAccount(base.Channel, ChannelPoolData)
	case "pool_info":
		return c.handleSubscribedAccount(base.Channel, ChannelPoolInfo)
	case "notification":
		return c.handleSubscribedAccount(base.Channel, ChannelNotification)

	default:
		return nil
	}
}

func (c *wsClient) handleConnected() error {
	c.reconnectAttempts = 0

	// Signal that we received the connected message
	c.readyOnce.Do(func() {
		close(c.readyCh)
	})

	return nil
}

// parseChannelParts parses a channel string and returns its parts
// Handles both formats: "order_book:0" (response) and "order_book/0" (subscribe)
func parseChannelParts(channel string) []string {
	sep := ":"
	if !strings.Contains(channel, ":") {
		sep = "/"
	}
	return strings.Split(channel, sep)
}

// parseMarketFromChannel extracts market index from channel like "order_book:0" or "order_book/0"
func parseMarketFromChannel(channel string) (int16, error) {
	parts := parseChannelParts(channel)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid channel format: %s", channel)
	}
	idx, err := strconv.ParseInt(parts[1], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid market index: %s", parts[1])
	}
	return int16(idx), nil
}

// parseAccountFromChannel extracts account index from channel like "account_all:123" or "account_all/123"
func parseAccountFromChannel(channel string) (int64, error) {
	parts := parseChannelParts(channel)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid channel format: %s", channel)
	}
	// Account index is always the last part
	idx, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid account index: %s", parts[len(parts)-1])
	}
	return idx, nil
}

// getSubscriptionKey returns the subscription manager key for a channel
func getSubscriptionKey(channel string, channelType ChannelType) string {
	parts := parseChannelParts(channel)
	switch channelType {
	case ChannelOrderBook:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 16); err == nil {
				return orderBookKey(int16(idx))
			}
		}
	case ChannelTrade:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 16); err == nil {
				return tradeKey(int16(idx))
			}
		}
	case ChannelMarketStats:
		if len(parts) >= 2 {
			if parts[1] == "all" {
				return marketStatsAllKey()
			}
			if idx, err := strconv.ParseInt(parts[1], 10, 16); err == nil {
				return marketStatsKey(int16(idx))
			}
		}
	case ChannelHeight:
		return heightKey()
	case ChannelAccountAll:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return accountKey(idx)
			}
		}
	case ChannelAccountMarket:
		if len(parts) >= 3 {
			marketIdx, err1 := strconv.ParseInt(parts[1], 10, 16)
			accountIdx, err2 := strconv.ParseInt(parts[2], 10, 64)
			if err1 == nil && err2 == nil {
				return accountMarketKey(int16(marketIdx), accountIdx)
			}
		}
	case ChannelAccountOrders:
		if len(parts) >= 3 {
			marketIdx, err1 := strconv.ParseInt(parts[1], 10, 16)
			accountIdx, err2 := strconv.ParseInt(parts[2], 10, 64)
			if err1 == nil && err2 == nil {
				return accountOrdersKey(int16(marketIdx), accountIdx)
			}
		}
	case ChannelAccountAllOrders:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return accountAllOrdersKey(idx)
			}
		}
	case ChannelAccountAllTrades:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return accountAllTradesKey(idx)
			}
		}
	case ChannelAccountAllPositions:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return accountAllPositionsKey(idx)
			}
		}
	case ChannelAccountTx:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return accountTxKey(idx)
			}
		}
	case ChannelUserStats:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return userStatsKey(idx)
			}
		}
	case ChannelPoolData:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return poolDataKey(idx)
			}
		}
	case ChannelPoolInfo:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return poolInfoKey(idx)
			}
		}
	case ChannelNotification:
		if len(parts) >= 2 {
			if idx, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return notificationKey(idx)
			}
		}
	}
	return channel
}

// Order book handlers

func (c *wsClient) handleSubscribedOrderBook(channel string, data RawMessage) error {
	marketIndex, err := parseMarketFromChannel(channel)
	if err != nil {
		return err
	}

	key := orderBookKey(marketIndex)
	c.subscriptions.ConfirmSubscription(key, nil)

	// The initial snapshot comes with the subscribed message
	if len(data) > 0 {
		return c.handleOrderBookData(marketIndex, data, true)
	}

	return nil
}

func (c *wsClient) handleOrderBookUpdate(channel string, data RawMessage) error {
	marketIndex, err := parseMarketFromChannel(channel)
	if err != nil {
		return err
	}

	return c.handleOrderBookData(marketIndex, data, false)
}

func (c *wsClient) handleOrderBookData(marketIndex int16, data RawMessage, isInitial bool) error {
	// Parse order book data - format: {"bids": [{"price": "...", "size": "..."}, ...], "asks": [...]}
	var obData struct {
		Bids []OrderBookLevel `json:"bids"`
		Asks []OrderBookLevel `json:"asks"`
	}
	if err := sonic.Unmarshal(data, &obData); err != nil {
		return fmt.Errorf("failed to parse order book data: %w", err)
	}

	c.orderBookMu.Lock()
	state, exists := c.orderBooks[marketIndex]
	if !exists {
		state = NewOrderBookState(marketIndex)
		c.orderBooks[marketIndex] = state
	}
	c.orderBookMu.Unlock()

	bids := obData.Bids
	asks := obData.Asks

	if isInitial {
		// Apply as snapshot
		snapshot := &OrderBookSnapshot{
			MarketIndex: marketIndex,
			Bids:        bids,
			Asks:        asks,
		}
		if err := state.ApplySnapshot(snapshot); err != nil {
			return err
		}

		update := &OrderBookUpdate{
			MarketIndex: marketIndex,
			IsSnapshot:  true,
			Snapshot:    snapshot,
			State:       state.Clone(),
		}

		select {
		case c.orderBookCh <- update:
		default:
		}

		if c.options.OnOrderBookUpdate != nil {
			c.options.OnOrderBookUpdate(update)
		}
	} else {
		// Apply as delta - merge updates
		delta := &OrderBookDelta{
			MarketIndex: marketIndex,
			BidUpdates:  bids,
			AskUpdates:  asks,
		}

		state.MergeUpdates(bids, asks)

		update := &OrderBookUpdate{
			MarketIndex: marketIndex,
			IsSnapshot:  false,
			Delta:       delta,
			State:       state.Clone(),
		}

		select {
		case c.orderBookCh <- update:
		default:
		}

		if c.options.OnOrderBookUpdate != nil {
			c.options.OnOrderBookUpdate(update)
		}
	}

	return nil
}

// Trade handlers

func (c *wsClient) handleSubscribedTrade(channel string) error {
	marketIndex, err := parseMarketFromChannel(channel)
	if err != nil {
		return err
	}

	key := tradeKey(marketIndex)
	c.subscriptions.ConfirmSubscription(key, nil)
	return nil
}

func (c *wsClient) handleTradeUpdate(channel string, data RawMessage) error {
	marketIndex, err := parseMarketFromChannel(channel)
	if err != nil {
		return err
	}

	var trades []Trade
	if err := sonic.Unmarshal(data, &trades); err != nil {
		// Try single trade
		var trade Trade
		if err := sonic.Unmarshal(data, &trade); err != nil {
			return fmt.Errorf("failed to parse trade data: %w", err)
		}
		trades = []Trade{trade}
	}

	update := &TradeUpdate{
		MarketIndex: marketIndex,
		Trades:      trades,
	}

	select {
	case c.tradeCh <- update:
	default:
	}

	if c.options.OnTradeUpdate != nil {
		c.options.OnTradeUpdate(update)
	}

	return nil
}

// Market stats handlers

func (c *wsClient) handleSubscribedMarketStats(channel string) error {
	parts := parseChannelParts(channel)
	if len(parts) >= 2 && parts[1] == "all" {
		c.subscriptions.ConfirmSubscription(marketStatsAllKey(), nil)
	} else {
		marketIndex, err := parseMarketFromChannel(channel)
		if err != nil {
			return err
		}
		c.subscriptions.ConfirmSubscription(marketStatsKey(marketIndex), nil)
	}
	return nil
}

func (c *wsClient) handleMarketStatsUpdate(channel string, data RawMessage) error {
	parts := parseChannelParts(channel)
	isAll := len(parts) >= 2 && parts[1] == "all"

	var update *MarketStatsUpdate

	if isAll {
		var allStats []MarketStats
		if err := sonic.Unmarshal(data, &allStats); err != nil {
			return fmt.Errorf("failed to parse market stats data: %w", err)
		}
		update = &MarketStatsUpdate{
			MarketIndex: -1, // indicates all markets
			AllStats:    allStats,
		}
	} else {
		marketIndex, err := parseMarketFromChannel(channel)
		if err != nil {
			return err
		}

		var stats MarketStats
		if err := sonic.Unmarshal(data, &stats); err != nil {
			return fmt.Errorf("failed to parse market stats data: %w", err)
		}
		update = &MarketStatsUpdate{
			MarketIndex: marketIndex,
			Stats:       &stats,
		}
	}

	select {
	case c.marketStatsCh <- update:
	default:
	}

	if c.options.OnMarketStatsUpdate != nil {
		c.options.OnMarketStatsUpdate(update)
	}

	return nil
}

// Height handlers

func (c *wsClient) handleSubscribedHeight() error {
	c.subscriptions.ConfirmSubscription(heightKey(), nil)
	return nil
}

func (c *wsClient) handleHeightUpdate(data RawMessage) error {
	var update HeightUpdate
	if err := sonic.Unmarshal(data, &update); err != nil {
		return fmt.Errorf("failed to parse height data: %w", err)
	}

	select {
	case c.heightCh <- &update:
	default:
	}

	if c.options.OnHeightUpdate != nil {
		c.options.OnHeightUpdate(&update)
	}

	return nil
}

// Account handlers

func (c *wsClient) handleSubscribedAccount(channel string, channelType ChannelType) error {
	key := getSubscriptionKey(channel, channelType)
	c.subscriptions.ConfirmSubscription(key, nil)
	return nil
}

func (c *wsClient) handleAccountUpdate(channel string, channelType ChannelType, data RawMessage) error {
	accountIndex, err := parseAccountFromChannel(channel)
	if err != nil {
		return err
	}

	update := &AccountUpdate{
		AccountIndex: accountIndex,
		Channel:      string(channelType),
		Data:         data,
	}

	select {
	case c.accountCh <- update:
	default:
	}

	if c.options.OnAccountUpdate != nil {
		c.options.OnAccountUpdate(update)
	}

	return nil
}

// Transaction result handlers

func (c *wsClient) handleTxResult(data RawMessage) error {
	var result TxResult
	if err := sonic.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to parse tx result: %w", err)
	}

	select {
	case c.txResultCh <- &result:
	default:
	}

	if c.options.OnTxResult != nil {
		c.options.OnTxResult(&result)
	}

	return nil
}

func (c *wsClient) handleTxBatchResult(data RawMessage) error {
	var batchResult TxBatchResult
	if err := sonic.Unmarshal(data, &batchResult); err != nil {
		return fmt.Errorf("failed to parse tx batch result: %w", err)
	}

	// Send each result individually
	for _, result := range batchResult.Results {
		r := result // avoid closure issue
		select {
		case c.txResultCh <- &r:
		default:
		}

		if c.options.OnTxResult != nil {
			c.options.OnTxResult(&r)
		}
	}

	return nil
}

// Error handling

func (c *wsClient) handleError(data RawMessage) error {
	var errData ErrorData
	if err := sonic.Unmarshal(data, &errData); err != nil {
		return fmt.Errorf("failed to parse error data: %w", err)
	}

	wsErr := NewWsError(errData.Code, errData.Message)

	if errData.Channel != "" {
		c.subscriptions.ConfirmSubscription(errData.Channel, wsErr)
	}

	c.sendError(wsErr)
	return nil
}

func (c *wsClient) sendError(err error) {
	select {
	case c.errorCh <- err:
	default:
	}

	if c.options.OnError != nil {
		c.options.OnError(err)
	}
}

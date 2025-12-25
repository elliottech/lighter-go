package ws

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// handleMessage routes incoming messages to appropriate handlers
func (c *wsClient) handleMessage(msg []byte) error {
	var base BaseMessage
	if err := json.Unmarshal(msg, &base); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	switch base.Type {
	case MessageTypeConnected:
		return c.handleConnected()
	case MessageTypeSubscribedOrderBook:
		return c.handleSubscribedOrderBook(base.Channel, base.OrderBook)
	case MessageTypeUpdateOrderBook:
		return c.handleOrderBookUpdate(base.Channel, base.OrderBook)
	case MessageTypeSubscribedAccountAll:
		return c.handleSubscribedAccount(base.Channel)
	case MessageTypeUpdateAccountAll:
		return c.handleAccountUpdate(base.Channel, base.Data)
	case MessageTypePing:
		return c.sendPong()
	case MessageTypeError:
		return c.handleError(base.Data)
	default:
		// Ignore unknown message types
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

// parseMarketFromChannel extracts market index from channel like "order_book:0" or "order_book/0"
func parseMarketFromChannel(channel string) (int16, error) {
	// Try colon first (server response format), then slash
	sep := ":"
	if !strings.Contains(channel, ":") {
		sep = "/"
	}
	parts := strings.Split(channel, sep)
	if len(parts) != 2 {
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
	// Try colon first (server response format), then slash
	sep := ":"
	if !strings.Contains(channel, ":") {
		sep = "/"
	}
	parts := strings.Split(channel, sep)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid channel format: %s", channel)
	}
	idx, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid account index: %s", parts[1])
	}
	return idx, nil
}

func (c *wsClient) handleSubscribedOrderBook(channel string, data json.RawMessage) error {
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

func (c *wsClient) handleOrderBookUpdate(channel string, data json.RawMessage) error {
	marketIndex, err := parseMarketFromChannel(channel)
	if err != nil {
		return err
	}

	return c.handleOrderBookData(marketIndex, data, false)
}

func (c *wsClient) handleOrderBookData(marketIndex int16, data json.RawMessage, isInitial bool) error {
	// Parse order book data - format: {"bids": [{"price": "...", "size": "..."}, ...], "asks": [...]}
	var obData struct {
		Bids []OrderBookLevel `json:"bids"`
		Asks []OrderBookLevel `json:"asks"`
	}
	if err := json.Unmarshal(data, &obData); err != nil {
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

func (c *wsClient) handleSubscribedAccount(channel string) error {
	accountIndex, err := parseAccountFromChannel(channel)
	if err != nil {
		return err
	}

	key := accountKey(accountIndex)
	c.subscriptions.ConfirmSubscription(key, nil)
	return nil
}

func (c *wsClient) handleAccountUpdate(channel string, data json.RawMessage) error {
	accountIndex, err := parseAccountFromChannel(channel)
	if err != nil {
		return err
	}

	update := &AccountUpdate{
		AccountIndex: accountIndex,
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

func (c *wsClient) handleError(data json.RawMessage) error {
	var errData ErrorData
	if err := json.Unmarshal(data, &errData); err != nil {
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

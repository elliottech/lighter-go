package ws

import (
	"encoding/json"
	"fmt"
)

// handleMessage routes incoming messages to appropriate handlers
func (c *wsClient) handleMessage(msg []byte) error {
	var base BaseMessage
	if err := json.Unmarshal(msg, &base); err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	switch base.Type {
	case MessageTypeConnected:
		return c.handleConnected(base.Data)
	case MessageTypeSubscribed:
		return c.handleSubscribed(base.Data)
	case MessageTypeUnsubscribed:
		return c.handleUnsubscribed(base.Data)
	case MessageTypeOrderBookUpdate:
		return c.handleOrderBookUpdate(base.Data)
	case MessageTypeAccountUpdate:
		return c.handleAccountUpdate(base.Data)
	case MessageTypePong:
		return c.handlePong()
	case MessageTypeError:
		return c.handleError(base.Data)
	default:
		return fmt.Errorf("unknown message type: %s", base.Type)
	}
}

func (c *wsClient) handleConnected(data json.RawMessage) error {
	var connData ConnectedData
	if err := json.Unmarshal(data, &connData); err != nil {
		return fmt.Errorf("failed to parse connected data: %w", err)
	}

	c.connected.Store(true)
	c.reconnectAttempts = 0

	// Call callback if set
	if c.options.OnConnect != nil {
		c.options.OnConnect()
	}

	return nil
}

func (c *wsClient) handleSubscribed(data json.RawMessage) error {
	var subData SubscribedData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscribed data: %w", err)
	}

	var key string
	switch subData.Channel {
	case "orderbook":
		key = orderBookKey(subData.Market)
	case "account":
		key = accountKey(subData.Account)
	default:
		return fmt.Errorf("unknown channel: %s", subData.Channel)
	}

	c.subscriptions.ConfirmSubscription(key, nil)
	return nil
}

func (c *wsClient) handleUnsubscribed(data json.RawMessage) error {
	var subData SubscribedData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse unsubscribed data: %w", err)
	}

	var key string
	switch subData.Channel {
	case "orderbook":
		key = orderBookKey(subData.Market)
		c.orderBookMu.Lock()
		delete(c.orderBooks, subData.Market)
		c.orderBookMu.Unlock()
	case "account":
		key = accountKey(subData.Account)
	}

	c.subscriptions.Remove(key)
	return nil
}

func (c *wsClient) handleOrderBookUpdate(data json.RawMessage) error {
	// Try parsing as snapshot first
	var snapshot OrderBookSnapshot
	if err := json.Unmarshal(data, &snapshot); err == nil && len(snapshot.Bids) > 0 || len(snapshot.Asks) > 0 {
		return c.handleOrderBookSnapshot(&snapshot)
	}

	// Try parsing as delta
	var delta OrderBookDelta
	if err := json.Unmarshal(data, &delta); err != nil {
		return fmt.Errorf("failed to parse order book update: %w", err)
	}

	return c.handleOrderBookDelta(&delta)
}

func (c *wsClient) handleOrderBookSnapshot(snapshot *OrderBookSnapshot) error {
	c.orderBookMu.Lock()

	state, exists := c.orderBooks[snapshot.MarketIndex]
	if !exists {
		state = NewOrderBookState(snapshot.MarketIndex)
		c.orderBooks[snapshot.MarketIndex] = state
	}

	c.orderBookMu.Unlock()

	if err := state.ApplySnapshot(snapshot); err != nil {
		return err
	}

	update := &OrderBookUpdate{
		MarketIndex: snapshot.MarketIndex,
		IsSnapshot:  true,
		Snapshot:    snapshot,
		State:       state.Clone(),
	}

	// Send to channel
	select {
	case c.orderBookCh <- update:
	default:
		// Channel full, drop update
	}

	// Call callback if set
	if c.options.OnOrderBookUpdate != nil {
		c.options.OnOrderBookUpdate(update)
	}

	return nil
}

func (c *wsClient) handleOrderBookDelta(delta *OrderBookDelta) error {
	c.orderBookMu.RLock()
	state, exists := c.orderBooks[delta.MarketIndex]
	c.orderBookMu.RUnlock()

	if !exists {
		return ErrOrderBookNotFound
	}

	if err := state.ApplyDelta(delta); err != nil {
		if err == ErrSequenceGap {
			c.sendError(err)
			// TODO: Request snapshot recovery
		}
		return err
	}

	update := &OrderBookUpdate{
		MarketIndex: delta.MarketIndex,
		IsSnapshot:  false,
		Delta:       delta,
		State:       state.Clone(),
	}

	// Send to channel
	select {
	case c.orderBookCh <- update:
	default:
		// Channel full, drop update
	}

	// Call callback if set
	if c.options.OnOrderBookUpdate != nil {
		c.options.OnOrderBookUpdate(update)
	}

	return nil
}

func (c *wsClient) handleAccountUpdate(data json.RawMessage) error {
	var updateData AccountUpdateData
	if err := json.Unmarshal(data, &updateData); err != nil {
		return fmt.Errorf("failed to parse account update: %w", err)
	}

	update := &AccountUpdate{
		AccountIndex: updateData.AccountIndex,
		Type:         updateData.Type,
		Data:         updateData.Data,
		Timestamp:    updateData.Timestamp,
	}

	// Send to channel
	select {
	case c.accountCh <- update:
	default:
		// Channel full, drop update
	}

	// Call callback if set
	if c.options.OnAccountUpdate != nil {
		c.options.OnAccountUpdate(update)
	}

	return nil
}

func (c *wsClient) handlePong() error {
	c.pingMu.Lock()
	c.lastPongTime = c.lastPingTime
	c.pingMu.Unlock()
	return nil
}

func (c *wsClient) handleError(data json.RawMessage) error {
	var errData ErrorData
	if err := json.Unmarshal(data, &errData); err != nil {
		return fmt.Errorf("failed to parse error data: %w", err)
	}

	wsErr := NewWsError(errData.Code, errData.Message)

	// If this is a subscription error, notify the pending subscription
	if errData.Channel != "" {
		c.subscriptions.ConfirmSubscription(errData.Channel, wsErr)
	}

	c.sendError(wsErr)
	return nil
}

func (c *wsClient) sendError(err error) {
	// Send to channel
	select {
	case c.errorCh <- err:
	default:
		// Channel full, drop error
	}

	// Call callback if set
	if c.options.OnError != nil {
		c.options.OnError(err)
	}
}

package ws

import (
	"encoding/json"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	MessageTypeConnected       MessageType = "connected"
	MessageTypeSubscribed      MessageType = "subscribed"
	MessageTypeUnsubscribed    MessageType = "unsubscribed"
	MessageTypeOrderBookUpdate MessageType = "order_book_update"
	MessageTypeAccountUpdate   MessageType = "account_update"
	MessageTypePong            MessageType = "pong"
	MessageTypeError           MessageType = "error"
)

// BaseMessage is the envelope for all WebSocket messages
type BaseMessage struct {
	Type    MessageType     `json:"type"`
	Channel string          `json:"channel,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// SubscribeRequest is sent to subscribe to a channel
type SubscribeRequest struct {
	Action    string `json:"action"`               // "subscribe" or "unsubscribe"
	Channel   string `json:"channel"`              // "orderbook" or "account"
	Market    int16  `json:"market,omitempty"`     // For orderbook subscriptions
	Account   int64  `json:"account,omitempty"`    // For account subscriptions
	AuthToken string `json:"auth_token,omitempty"` // For authenticated subscriptions
}

// PingMessage is sent to keep the connection alive
type PingMessage struct {
	Action string `json:"action"`
}

// OrderBookLevel represents a price level in the order book
type OrderBookLevel struct {
	Price      string `json:"price"`
	Size       string `json:"size"`
	OrderCount int    `json:"order_count,omitempty"`
}

// OrderBookSnapshot represents a full order book snapshot
type OrderBookSnapshot struct {
	MarketIndex int16            `json:"market_index"`
	Sequence    int64            `json:"sequence"`
	Bids        []OrderBookLevel `json:"bids"`
	Asks        []OrderBookLevel `json:"asks"`
	Timestamp   int64            `json:"timestamp"`
}

// OrderBookDelta represents an incremental update
type OrderBookDelta struct {
	MarketIndex int16            `json:"market_index"`
	Sequence    int64            `json:"sequence"`
	BidUpdates  []OrderBookLevel `json:"bid_updates,omitempty"`
	AskUpdates  []OrderBookLevel `json:"ask_updates,omitempty"`
	Timestamp   int64            `json:"timestamp"`
}

// OrderBookUpdate is sent through the update channel
type OrderBookUpdate struct {
	MarketIndex int16
	IsSnapshot  bool
	Snapshot    *OrderBookSnapshot
	Delta       *OrderBookDelta
	State       *OrderBookState // Current merged state
}

// AccountUpdateData represents the data in an account update
type AccountUpdateData struct {
	AccountIndex int64           `json:"account_index"`
	Type         string          `json:"type"` // "position", "balance", "order", etc.
	Data         json.RawMessage `json:"data"`
	Timestamp    int64           `json:"timestamp"`
}

// AccountUpdate is sent through the account update channel
type AccountUpdate struct {
	AccountIndex int64
	Type         string
	Data         json.RawMessage
	Timestamp    int64
}

// PositionUpdate represents a position change
type PositionUpdate struct {
	MarketIndex      int16  `json:"market_index"`
	Size             string `json:"size"`
	Side             string `json:"side"`
	EntryPrice       string `json:"entry_price"`
	MarkPrice        string `json:"mark_price"`
	UnrealizedPnl    string `json:"unrealized_pnl"`
	LiquidationPrice string `json:"liquidation_price,omitempty"`
}

// BalanceUpdate represents a balance change
type BalanceUpdate struct {
	AssetIndex int16  `json:"asset_index"`
	Balance    string `json:"balance"`
	Available  string `json:"available"`
	Locked     string `json:"locked,omitempty"`
}

// OrderUpdate represents an order status change
type OrderUpdate struct {
	OrderIndex   int64  `json:"order_index"`
	MarketIndex  int16  `json:"market_index"`
	Status       string `json:"status"`
	FilledSize   string `json:"filled_size,omitempty"`
	RemainingSize string `json:"remaining_size,omitempty"`
	Price        string `json:"price"`
	Timestamp    int64  `json:"timestamp"`
}

// ConnectedData is returned when connection is established
type ConnectedData struct {
	SessionID string `json:"session_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// SubscribedData is returned when subscription is confirmed
type SubscribedData struct {
	Channel string `json:"channel"`
	Market  int16  `json:"market,omitempty"`
	Account int64  `json:"account,omitempty"`
}

// ErrorData represents error data from the server
type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Channel string `json:"channel,omitempty"`
}

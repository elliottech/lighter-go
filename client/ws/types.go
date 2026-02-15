package ws

import "github.com/bytedance/sonic"

// RawMessage is a raw encoded JSON value.
// It can be used to delay JSON decoding or precompute a JSON encoding.
// Uses sonic's NoCopyRawMessage for efficient JSON handling.
type RawMessage = sonic.NoCopyRawMessage

// MessageType represents WebSocket message types
type MessageType string

const (
	// Connection messages
	MessageTypeConnected MessageType = "connected"
	MessageTypePing      MessageType = "ping"
	MessageTypeError     MessageType = "error"

	// Order book messages
	MessageTypeSubscribedOrderBook MessageType = "subscribed/order_book"
	MessageTypeUpdateOrderBook     MessageType = "update/order_book"

	// Trade messages
	MessageTypeSubscribedTrade MessageType = "subscribed/trade"
	MessageTypeUpdateTrade     MessageType = "update/trade"

	// Market stats messages
	MessageTypeSubscribedMarketStats MessageType = "subscribed/market_stats"
	MessageTypeUpdateMarketStats     MessageType = "update/market_stats"

	// Height messages
	MessageTypeSubscribedHeight MessageType = "subscribed/height"
	MessageTypeUpdateHeight     MessageType = "update/height"

	// Account messages
	MessageTypeSubscribedAccountAll       MessageType = "subscribed/account_all"
	MessageTypeUpdateAccountAll           MessageType = "update/account_all"
	MessageTypeSubscribedAccountMarket    MessageType = "subscribed/account_market"
	MessageTypeUpdateAccountMarket        MessageType = "update/account_market"
	MessageTypeSubscribedAccountOrders    MessageType = "subscribed/account_orders"
	MessageTypeUpdateAccountOrders        MessageType = "update/account_orders"
	MessageTypeSubscribedAccountAllOrders MessageType = "subscribed/account_all_orders"
	MessageTypeUpdateAccountAllOrders     MessageType = "update/account_all_orders"
	MessageTypeSubscribedAccountAllTrades MessageType = "subscribed/account_all_trades"
	MessageTypeUpdateAccountAllTrades     MessageType = "update/account_all_trades"
	MessageTypeSubscribedAccountAllPositions MessageType = "subscribed/account_all_positions"
	MessageTypeUpdateAccountAllPositions     MessageType = "update/account_all_positions"
	MessageTypeSubscribedAccountTx        MessageType = "subscribed/account_tx"
	MessageTypeUpdateAccountTx            MessageType = "update/account_tx"
	MessageTypeSubscribedUserStats        MessageType = "subscribed/user_stats"
	MessageTypeUpdateUserStats            MessageType = "update/user_stats"
	MessageTypeSubscribedPoolData         MessageType = "subscribed/pool_data"
	MessageTypeUpdatePoolData             MessageType = "update/pool_data"
	MessageTypeSubscribedPoolInfo         MessageType = "subscribed/pool_info"
	MessageTypeUpdatePoolInfo             MessageType = "update/pool_info"
	MessageTypeSubscribedNotification     MessageType = "subscribed/notification"
	MessageTypeUpdateNotification         MessageType = "update/notification"

	// Transaction response messages
	MessageTypeTxResult      MessageType = "tx_result"
	MessageTypeTxBatchResult MessageType = "tx_batch_result"
)

// BaseMessage is the envelope for all WebSocket messages
type BaseMessage struct {
	Type      MessageType     `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	OrderBook RawMessage `json:"order_book,omitempty"`
	Data      RawMessage `json:"data,omitempty"`
}

// SubscribeRequest is sent to subscribe to a channel
// Format: {"type": "subscribe", "channel": "order_book/0"}
// For private channels: {"type": "subscribe", "channel": "account_all/123", "auth": "token"}
type SubscribeRequest struct {
	Type    string `json:"type"`              // "subscribe" or "unsubscribe"
	Channel string `json:"channel"`           // channel path
	Auth    string `json:"auth,omitempty"`    // auth token for private channels
}

// PongMessage is sent in response to server ping
type PongMessage struct {
	Type string `json:"type"` // "pong"
}

// SendTxRequest is sent to submit a transaction via WebSocket
type SendTxRequest struct {
	Type string      `json:"type"` // "jsonapi/sendtx"
	Data interface{} `json:"data"`
}

// SendTxBatchRequest is sent to submit multiple transactions via WebSocket
type SendTxBatchRequest struct {
	Type string        `json:"type"` // "jsonapi/sendtxbatch"
	Data []interface{} `json:"data"` // max 50 transactions
}

// TxResult is the response for a single transaction
type TxResult struct {
	Success bool            `json:"success"`
	TxHash  string          `json:"tx_hash,omitempty"`
	Error   string          `json:"error,omitempty"`
	Data    RawMessage `json:"data,omitempty"`
}

// TxBatchResult is the response for a batch transaction
type TxBatchResult struct {
	Results []TxResult `json:"results"`
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

// Trade represents a single trade
type Trade struct {
	TradeIndex  int64  `json:"trade_index"`
	MarketIndex int16  `json:"market_index"`
	Price       string `json:"price"`
	Size        string `json:"size"`
	Side        string `json:"side"` // "buy" or "sell"
	Timestamp   int64  `json:"timestamp"`
	MakerIndex  int64  `json:"maker_index,omitempty"`
	TakerIndex  int64  `json:"taker_index,omitempty"`
}

// TradeUpdate is sent through the trade update channel
type TradeUpdate struct {
	MarketIndex int16
	Trades      []Trade
}

// MarketStats represents market statistics
type MarketStats struct {
	MarketIndex     int16  `json:"market_index"`
	LastPrice       string `json:"last_price"`
	MarkPrice       string `json:"mark_price"`
	IndexPrice      string `json:"index_price"`
	High24h         string `json:"high_24h"`
	Low24h          string `json:"low_24h"`
	Volume24h       string `json:"volume_24h"`
	QuoteVolume24h  string `json:"quote_volume_24h"`
	PriceChange24h  string `json:"price_change_24h"`
	PriceChangePct  string `json:"price_change_pct"`
	OpenInterest    string `json:"open_interest"`
	FundingRate     string `json:"funding_rate"`
	NextFundingTime int64  `json:"next_funding_time"`
}

// MarketStatsUpdate is sent through the market stats update channel
type MarketStatsUpdate struct {
	MarketIndex int16
	Stats       *MarketStats
	AllStats    []MarketStats // For market_stats/all subscription
}

// HeightUpdate is sent through the height update channel
type HeightUpdate struct {
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

// AccountUpdateData represents the data in an account update
type AccountUpdateData struct {
	AccountIndex int64           `json:"account_index"`
	Type         string          `json:"type"` // "position", "balance", "order", etc.
	Data         RawMessage `json:"data"`
	Timestamp    int64           `json:"timestamp"`
}

// AccountUpdate is sent through the account update channel
type AccountUpdate struct {
	AccountIndex int64
	Channel      string // which channel this came from
	Type         string
	Data         RawMessage
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
	OrderIndex    int64  `json:"order_index"`
	MarketIndex   int16  `json:"market_index"`
	Status        string `json:"status"`
	FilledSize    string `json:"filled_size,omitempty"`
	RemainingSize string `json:"remaining_size,omitempty"`
	Price         string `json:"price"`
	Timestamp     int64  `json:"timestamp"`
}

// UserStats represents user statistics
type UserStats struct {
	AccountIndex   int64  `json:"account_index"`
	TotalVolume    string `json:"total_volume"`
	TotalTrades    int64  `json:"total_trades"`
	TotalPnl       string `json:"total_pnl"`
	MakerVolume    string `json:"maker_volume"`
	TakerVolume    string `json:"taker_volume"`
	FeesGenerated  string `json:"fees_generated"`
}

// PoolData represents pool data update
type PoolData struct {
	AccountIndex int64           `json:"account_index"`
	Data         RawMessage `json:"data"`
}

// PoolInfo represents pool info update
type PoolInfo struct {
	AccountIndex int64           `json:"account_index"`
	Data         RawMessage `json:"data"`
}

// Notification represents a notification
type Notification struct {
	AccountIndex int64           `json:"account_index"`
	Type         string          `json:"type"`
	Message      string          `json:"message"`
	Data         RawMessage `json:"data,omitempty"`
	Timestamp    int64           `json:"timestamp"`
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

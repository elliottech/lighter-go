package client

import (
	"context"
	"time"
)

// WebSocket message types
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type WSSubscribeMessage struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol,omitempty"`
}

type WSUnsubscribeMessage struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol,omitempty"`
}

// Note: SubscriptionRequest and related helper functions removed as they were unused

// Market data types
type WSOrderBookUpdate struct {
	Symbol    string     `json:"symbol"`
	Bids      [][]string `json:"bids"`
	Asks      [][]string `json:"asks"`
	Timestamp int64      `json:"timestamp"`
}

// Note: WSTickerUpdate and WSTradeUpdate types removed because
// ticker and trade streams are not supported by Lighter WebSocket API

// Account data types
type WSAccountUpdate struct {
	Account            int64                  `json:"account"`
	Channel            string                 `json:"channel"`
	Type               string                 `json:"type"`
	DailyTradesCount   int                    `json:"daily_trades_count"`
	DailyVolume        float64                `json:"daily_volume"`
	MonthlyTradesCount int                    `json:"monthly_trades_count"`
	MonthlyVolume      float64                `json:"monthly_volume"`
	TotalTradesCount   int                    `json:"total_trades_count"`
	TotalVolume        float64                `json:"total_volume"`
	WeeklyTradesCount  int                    `json:"weekly_trades_count"`
	WeeklyVolume       float64                `json:"weekly_volume"`
	Positions          map[string]*WSPosition `json:"positions"`
	Shares             []WSShare              `json:"shares"`
	Trades             map[string][]WSTrade   `json:"trades"`
	FundingHistories   map[string]interface{} `json:"funding_histories"`
}

type WSPosition struct {
	MarketId               uint16 `json:"market_id"`
	Symbol                 string `json:"symbol"`
	InitialMarginFraction  string `json:"initial_margin_fraction"`
	OpenOrderCount         int    `json:"open_order_count"`
	PendingOrderCount      int    `json:"pending_order_count"`
	PositionTiedOrderCount int    `json:"position_tied_order_count"`
	Sign                   int8   `json:"sign"`
	Position               string `json:"position"`
	AvgEntryPrice          string `json:"avg_entry_price"`
	PositionValue          string `json:"position_value"`
	UnrealizedPnl          string `json:"unrealized_pnl"`
	RealizedPnl            string `json:"realized_pnl"`
	LiquidationPrice       string `json:"liquidation_price"`
	MarginMode             int    `json:"margin_mode"`
	AllocatedMargin        string `json:"allocated_margin"`
}

type WSShare struct {
	PublicPoolIndex int64  `json:"public_pool_index"`
	SharesAmount    int64  `json:"shares_amount"`
	EntryUsdc       string `json:"entry_usdc"`
}

type WSTrade struct {
	TradeId                          int64  `json:"trade_id"`
	TxHash                           string `json:"tx_hash"`
	Type                             string `json:"type"`
	MarketId                         int    `json:"market_id"`
	Size                             string `json:"size"`
	Price                            string `json:"price"`
	UsdAmount                        string `json:"usd_amount"`
	AskId                            int64  `json:"ask_id"`
	BidId                            int64  `json:"bid_id"`
	AskClientId                      int64  `json:"ask_client_id"`
	BidClientId                      int64  `json:"bid_client_id"`
	AskAccountId                     int64  `json:"ask_account_id"`
	BidAccountId                     int64  `json:"bid_account_id"`
	IsMakerAsk                       bool   `json:"is_maker_ask"`
	BlockHeight                      int64  `json:"block_height"`
	Timestamp                        int64  `json:"timestamp"`
	TakerFee                         int    `json:"taker_fee"`
	TakerPositionSizeBefore          string `json:"taker_position_size_before"`
	TakerEntryQuoteBefore            string `json:"taker_entry_quote_before"`
	TakerInitialMarginFractionBefore int    `json:"taker_initial_margin_fraction_before"`
	MakerFee                         int    `json:"maker_fee"`
	MakerPositionSizeBefore          string `json:"maker_position_size_before"`
	MakerEntryQuoteBefore            string `json:"maker_entry_quote_before"`
	MakerInitialMarginFractionBefore int    `json:"maker_initial_margin_fraction_before"`
}

type WSOrderUpdate struct {
	AccountIndex     int64  `json:"account_index"`
	OrderId          string `json:"order_id"`
	ClientOrderIndex int64  `json:"client_order_index"`
	MarketId         uint16 `json:"market_id"`
	Status           string `json:"status"`
	BaseQuantity     string `json:"base_quantity"`
	FilledQuantity   string `json:"filled_quantity"`
	Price            string `json:"price"`
	IsAsk            uint8  `json:"is_ask"`
	Timestamp        int64  `json:"timestamp"`
}

// Channel constants - based on Python implementation
// Note: Only order_book and account_all are actually supported by Lighter WebSocket API
const (
	ChannelOrderBook = "order_book"
	ChannelAccount   = "account_all"
	ChannelOrders    = "orders"
	// The following channels are not supported by Lighter WebSocket API:
	// ChannelTicker    = "ticker"      // REMOVED - not supported
	// ChannelTrades    = "trades"      // REMOVED - not supported
	// ChannelMarkPrice = "markprice"   // REMOVED - not supported
)

// Message types - based on Python implementation
const (
	MessageTypeSubscribe    = "subscribe"
	MessageTypeUnsubscribe  = "unsubscribe"
	MessageTypePing         = "ping"
	MessageTypePong         = "pong"
	MessageTypeConnected    = "connected"
	MessageTypeSubscribed   = "subscribed"
	MessageTypeUnsubscribed = "unsubscribed"

	// Subscription confirmation messages
	MessageTypeOrderBookSubscribed = "subscribed/order_book"
	MessageTypeAccountSubscribed   = "subscribed/account_all"

	// Data update messages (the actual data streams)
	MessageTypeOrderBookUpdate = "update/order_book"
	MessageTypeAccountUpdate   = "update/account_all"

	// Deprecated: Use MessageTypeOrderBookUpdate instead
	MessageTypeOrderBook = "update/order_book"
	// Deprecated: Use MessageTypeAccountUpdate instead
	MessageTypeAccount = "update/account_all"
)

// WSPriceLevel represents a single price level in WebSocket orderbook messages
type WSPriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// WebSocket order book state for incremental updates
type WSOrderBookState struct {
	MarketId  uint16            `json:"market_id"`
	Bids      map[string]string `json:"bids"`
	Asks      map[string]string `json:"asks"`
	Timestamp int64             `json:"timestamp"`
}

// WebSocket configuration
type WSConfig struct {
	URL            string
	ReconnectDelay time.Duration
	PingInterval   time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxReconnects  int
}

func DefaultWSConfig() *WSConfig {
	return &WSConfig{
		URL:            "wss://mainnet.zklighter.elliot.ai/stream",
		ReconnectDelay: 5 * time.Second,
		PingInterval:   30 * time.Second,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxReconnects:  10,
	}
}

// New Bybit-style interface design for Lighter WebSocket
type ErrHandler func(error)

type LighterWebsocketPublicServiceI interface {
	Start(context.Context, ErrHandler) error
	Close() error

	SubscribeOrderBook(
		LighterOrderBookParamKey,
		func(LighterOrderBookResponse) error,
	) (func() error, error)

	// SubscribeTicker removed - not supported by Lighter

	SubscribeTrades(
		LighterTradesParamKey,
		func(LighterTradesResponse) error,
	) (func() error, error)
}

// Private service interface for authenticated subscriptions
type LighterWebsocketPrivateServiceI interface {
	Start(context.Context, ErrHandler) error
	Close() error

	SubscribeAccount(
		LighterAccountParamKey,
		func(LighterAccountResponse) error,
	) (func() error, error)

	SubscribeOrders(
		LighterOrdersParamKey,
		func(LighterOrdersResponse) error,
	) (func() error, error)
}

// Parameter types
type LighterOrderBookParamKey struct {
	MarketId uint16
}

// LighterTickerParamKey removed - not supported by Lighter

type LighterTradesParamKey struct {
	MarketId uint16
}

type LighterAccountParamKey struct {
	AccountId int64
}

type LighterOrdersParamKey struct {
	AccountId int64
}

// Response types
type LighterOrderBookResponse struct {
	MarketId   uint16       `json:"market_id"`
	Bids       []PriceLevel `json:"bids"`
	Asks       []PriceLevel `json:"asks"`
	Timestamp  int64        `json:"timestamp"`
	IsSnapshot bool         `json:"is_snapshot"`
}

// LighterTickerResponse removed - not supported by Lighter

type LighterTradesResponse struct {
	MarketId  uint16 `json:"market_id"`
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Side      string `json:"side"`
	Timestamp int64  `json:"timestamp"`
}

type LighterAccountResponse struct {
	AccountId        int64                `json:"account_id"`
	AvailableBalance string               `json:"available_balance"`
	MarketStats      []AccountMarketStats `json:"market_stats,omitempty"`
	Timestamp        int64                `json:"timestamp"`
	IsSnapshot       bool                 `json:"is_snapshot"`
	RawAccountUpdate *WSAccountUpdate     `json:"-"` // Raw data for ws_manager processing
}

type LighterOrdersResponse struct {
	AccountId        int64  `json:"account_id"`
	OrderId          string `json:"order_id"`
	ClientOrderIndex int64  `json:"client_order_index"`
	MarketId         uint16 `json:"market_id"`
	Status           string `json:"status"`
	BaseQuantity     string `json:"base_quantity"`
	FilledQuantity   string `json:"filled_quantity"`
	Price            string `json:"price"`
	IsAsk            uint8  `json:"is_ask"`
	Timestamp        int64  `json:"timestamp"`
	IsSnapshot       bool   `json:"is_snapshot"`
}

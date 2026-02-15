package api

// Order represents an order
type Order struct {
	Index             int64     `json:"index"`
	ClientOrderIndex  int64     `json:"client_order_index,omitempty"`
	AccountIndex      int64     `json:"account_index"`
	MarketIndex       int16     `json:"market_index"`
	MarketSymbol      string    `json:"market_symbol,omitempty"`
	Type              OrderType `json:"type"`
	Side              OrderSide `json:"side"`
	Price             string    `json:"price"`
	Size              string    `json:"size"`
	FilledSize        string    `json:"filled_size"`
	RemainingSize     string    `json:"remaining_size"`
	TriggerPrice      string    `json:"trigger_price,omitempty"`
	TimeInForce       TimeInForce `json:"time_in_force"`
	ReduceOnly        bool      `json:"reduce_only"`
	PostOnly          bool      `json:"post_only"`
	Status            string    `json:"status"` // "open", "filled", "cancelled", "expired"
	GroupIndex        int64     `json:"group_index,omitempty"`
	GroupingType      GroupingType `json:"grouping_type,omitempty"`
	ExpiredAt         int64     `json:"expired_at,omitempty"`
	CreatedAt         int64     `json:"created_at"`
	UpdatedAt         int64     `json:"updated_at,omitempty"`
	TxHash            string    `json:"tx_hash,omitempty"`
}

// SimpleOrder represents a simplified order view
type SimpleOrder struct {
	Index        int64  `json:"index"`
	AccountIndex int64  `json:"account_index"`
	MarketIndex  int16  `json:"market_index"`
	Side         string `json:"side"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	FilledSize   string `json:"filled_size"`
	Status       string `json:"status"`
}

// Orders is the response for order queries
type Orders struct {
	BaseResponse
	Orders []Order `json:"orders"`
	Cursor Cursor  `json:"cursor,omitempty"`
}

// OrderBook represents an order book snapshot
type OrderBook struct {
	MarketIndex int16           `json:"market_index"`
	MarketSymbol string         `json:"market_symbol,omitempty"`
	Timestamp   int64           `json:"timestamp"`
	Bids        []PriceLevel    `json:"bids"`
	Asks        []PriceLevel    `json:"asks"`
	Sequence    int64           `json:"sequence,omitempty"`
}

// OrderBooks is the response for order book queries
type OrderBooks struct {
	BaseResponse
	OrderBooks []OrderBook `json:"order_books"`
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price      string `json:"price"`
	Size       string `json:"size"`
	OrderCount int    `json:"order_count,omitempty"`
}

// OrderBookDepth represents order book depth data
type OrderBookDepth struct {
	MarketIndex  int16  `json:"market_index"`
	BidDepth     string `json:"bid_depth"`
	AskDepth     string `json:"ask_depth"`
	Spread       string `json:"spread"`
	SpreadBps    string `json:"spread_bps"`
	MidPrice     string `json:"mid_price"`
}

// PerpsOrderBookDetail represents detailed perps order book
type PerpsOrderBookDetail struct {
	MarketIndex       int16        `json:"market_index"`
	MarketSymbol      string       `json:"market_symbol"`
	Bids              []PriceLevel `json:"bids"`
	Asks              []PriceLevel `json:"asks"`
	LastPrice         string       `json:"last_price"`
	MarkPrice         string       `json:"mark_price"`
	IndexPrice        string       `json:"index_price"`
	FundingRate       string       `json:"funding_rate"`
	NextFundingTime   int64        `json:"next_funding_time"`
	OpenInterest      string       `json:"open_interest"`
	Volume24h         string       `json:"volume_24h"`
	Timestamp         int64        `json:"timestamp"`
}

// SpotOrderBookDetail represents detailed spot order book
type SpotOrderBookDetail struct {
	MarketIndex   int16        `json:"market_index"`
	MarketSymbol  string       `json:"market_symbol"`
	BaseAsset     string       `json:"base_asset"`
	QuoteAsset    string       `json:"quote_asset"`
	Bids          []PriceLevel `json:"bids"`
	Asks          []PriceLevel `json:"asks"`
	LastPrice     string       `json:"last_price"`
	Volume24h     string       `json:"volume_24h"`
	Timestamp     int64        `json:"timestamp"`
}

// OrderBookDetails is the response for detailed order book queries
type OrderBookDetails struct {
	BaseResponse
	PerpsOrderBooks []PerpsOrderBookDetail `json:"perps_order_books,omitempty"`
	SpotOrderBooks  []SpotOrderBookDetail  `json:"spot_order_books,omitempty"`
}

// OrderBookOrder represents an individual order in the order book
type OrderBookOrder struct {
	OrderIndex   int64  `json:"order_index"`
	AccountIndex int64  `json:"account_index"`
	Side         string `json:"side"`
	Price        string `json:"price"`
	Size         string `json:"size"`
	Timestamp    int64  `json:"timestamp"`
}

// OrderBookOrders is the response for order book orders queries
type OrderBookOrders struct {
	BaseResponse
	MarketIndex int16            `json:"market_index"`
	Bids        []OrderBookOrder `json:"bids"`
	Asks        []OrderBookOrder `json:"asks"`
}

// Ticker represents market ticker data
type Ticker struct {
	MarketIndex      int16  `json:"market_index"`
	MarketSymbol     string `json:"market_symbol"`
	LastPrice        string `json:"last_price"`
	BestBidPrice     string `json:"best_bid_price"`
	BestBidSize      string `json:"best_bid_size"`
	BestAskPrice     string `json:"best_ask_price"`
	BestAskSize      string `json:"best_ask_size"`
	MarkPrice        string `json:"mark_price,omitempty"`
	IndexPrice       string `json:"index_price,omitempty"`
	PriceChange24h   string `json:"price_change_24h"`
	PriceChangePct24h string `json:"price_change_pct_24h"`
	High24h          string `json:"high_24h"`
	Low24h           string `json:"low_24h"`
	Volume24h        string `json:"volume_24h"`
	QuoteVolume24h   string `json:"quote_volume_24h,omitempty"`
	OpenInterest     string `json:"open_interest,omitempty"`
	FundingRate      string `json:"funding_rate,omitempty"`
	NextFundingTime  int64  `json:"next_funding_time,omitempty"`
	Timestamp        int64  `json:"timestamp"`
}

// Tickers is the response for ticker queries
type Tickers struct {
	BaseResponse
	Tickers []Ticker `json:"tickers"`
}

// ExchangeStats represents exchange-wide statistics
type ExchangeStats struct {
	BaseResponse
	TotalVolume24h     string `json:"total_volume_24h"`
	TotalTrades24h     int64  `json:"total_trades_24h"`
	TotalOpenInterest  string `json:"total_open_interest"`
	TotalUsers         int64  `json:"total_users"`
	TotalMarkets       int    `json:"total_markets"`
	Timestamp          int64  `json:"timestamp"`
}

// AssetDetail represents detailed asset information
type AssetDetail struct {
	AssetIndex       int16  `json:"asset_index"`
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Decimals         int    `json:"decimals"`
	ContractAddress  string `json:"contract_address,omitempty"`
	MinWithdrawal    string `json:"min_withdrawal,omitempty"`
	MaxWithdrawal    string `json:"max_withdrawal,omitempty"`
	WithdrawalFee    string `json:"withdrawal_fee,omitempty"`
	IsActive         bool   `json:"is_active"`
}

// AssetDetails is the response for asset detail queries
type AssetDetails struct {
	BaseResponse
	Assets []AssetDetail `json:"assets"`
}

// Market represents market configuration
type Market struct {
	MarketIndex          int16  `json:"market_index"`
	Symbol               string `json:"symbol"`
	BaseAsset            string `json:"base_asset"`
	QuoteAsset           string `json:"quote_asset"`
	Type                 string `json:"type"` // "perps" or "spot"
	PricePrecision       int    `json:"price_precision"`
	SizePrecision        int    `json:"size_precision"`
	MinSize              string `json:"min_size"`
	MaxSize              string `json:"max_size"`
	MinPrice             string `json:"min_price"`
	MaxPrice             string `json:"max_price"`
	MaxLeverage          int    `json:"max_leverage,omitempty"`
	MakerFeeRate         string `json:"maker_fee_rate"`
	TakerFeeRate         string `json:"taker_fee_rate"`
	MaintenanceMarginRate string `json:"maintenance_margin_rate,omitempty"`
	InitialMarginRate    string `json:"initial_margin_rate,omitempty"`
	IsActive             bool   `json:"is_active"`
}

// Markets is the response for market queries
type Markets struct {
	BaseResponse
	Markets []Market `json:"markets"`
}

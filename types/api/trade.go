package api

// Trade represents a trade execution
type Trade struct {
	TradeIndex     int64  `json:"trade_index"`
	MarketIndex    int16  `json:"market_index"`
	MarketSymbol   string `json:"market_symbol,omitempty"`
	MakerOrderIndex int64 `json:"maker_order_index"`
	TakerOrderIndex int64 `json:"taker_order_index"`
	MakerAccountIndex int64 `json:"maker_account_index"`
	TakerAccountIndex int64 `json:"taker_account_index"`
	Price          string `json:"price"`
	Size           string `json:"size"`
	QuoteAmount    string `json:"quote_amount,omitempty"`
	Side           string `json:"side"` // taker side: "buy" or "sell"
	MakerFee       string `json:"maker_fee,omitempty"`
	TakerFee       string `json:"taker_fee,omitempty"`
	Timestamp      int64  `json:"timestamp"`
	TxHash         string `json:"tx_hash,omitempty"`
	BlockHeight    int64  `json:"block_height,omitempty"`
}

// Trades is the response for trade queries
type Trades struct {
	BaseResponse
	Trades []Trade `json:"trades"`
	Cursor Cursor  `json:"cursor,omitempty"`
}

// RecentTrades is the response for recent trades
type RecentTrades struct {
	BaseResponse
	MarketIndex int16   `json:"market_index"`
	Trades      []Trade `json:"trades"`
}

// TradeStats represents trading statistics
type TradeStats struct {
	MarketIndex     int16  `json:"market_index"`
	Volume24h       string `json:"volume_24h"`
	QuoteVolume24h  string `json:"quote_volume_24h"`
	TradeCount24h   int64  `json:"trade_count_24h"`
	High24h         string `json:"high_24h"`
	Low24h          string `json:"low_24h"`
	Open24h         string `json:"open_24h"`
	Close24h        string `json:"close_24h"`
	PriceChange24h  string `json:"price_change_24h"`
	PriceChangePct  string `json:"price_change_pct"`
}

// LiqTrade represents a liquidation trade
type LiqTrade struct {
	TradeIndex      int64  `json:"trade_index"`
	MarketIndex     int16  `json:"market_index"`
	LiquidatedAccount int64 `json:"liquidated_account"`
	LiquidatorAccount int64 `json:"liquidator_account,omitempty"`
	Price           string `json:"price"`
	Size            string `json:"size"`
	Side            string `json:"side"`
	LiquidationFee  string `json:"liquidation_fee"`
	InsuranceFund   string `json:"insurance_fund,omitempty"`
	Timestamp       int64  `json:"timestamp"`
	TxHash          string `json:"tx_hash,omitempty"`
}

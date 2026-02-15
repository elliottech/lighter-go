package api

// Candlestick represents OHLCV data for a time period
type Candlestick struct {
	MarketIndex int16  `json:"market_index,omitempty"`
	Timestamp   int64  `json:"timestamp"` // Start time of the candle
	Open        string `json:"open"`
	High        string `json:"high"`
	Low         string `json:"low"`
	Close       string `json:"close"`
	Volume      string `json:"volume"`
	QuoteVolume string `json:"quote_volume,omitempty"`
	TradeCount  int64  `json:"trade_count,omitempty"`
}

// Candlesticks is the response for candlestick queries
type Candlesticks struct {
	BaseResponse
	MarketIndex  int16         `json:"market_index"`
	Resolution   string        `json:"resolution"`
	Candlesticks []Candlestick `json:"candlesticks"`
}

// DetailedCandlestick includes additional statistics
type DetailedCandlestick struct {
	Candlestick
	VWAP           string `json:"vwap,omitempty"`    // Volume-weighted average price
	TakerBuyVolume string `json:"taker_buy_volume,omitempty"`
	TakerSellVolume string `json:"taker_sell_volume,omitempty"`
}

// Funding represents funding data for a period
type Funding struct {
	MarketIndex  int16  `json:"market_index"`
	Timestamp    int64  `json:"timestamp"`
	FundingRate  string `json:"funding_rate"`
	MarkPrice    string `json:"mark_price"`
	IndexPrice   string `json:"index_price"`
}

// Fundings is the response for funding queries
type Fundings struct {
	BaseResponse
	MarketIndex int16     `json:"market_index"`
	Resolution  string    `json:"resolution"`
	Fundings    []Funding `json:"fundings"`
}

// FundingRate represents current funding rate
type FundingRate struct {
	MarketIndex      int16  `json:"market_index"`
	MarketSymbol     string `json:"market_symbol,omitempty"`
	FundingRate      string `json:"funding_rate"`
	PredictedRate    string `json:"predicted_rate,omitempty"`
	MarkPrice        string `json:"mark_price"`
	IndexPrice       string `json:"index_price"`
	NextFundingTime  int64  `json:"next_funding_time"`
	FundingInterval  int64  `json:"funding_interval"` // In milliseconds
}

// FundingRates is the response for funding rate queries
type FundingRates struct {
	BaseResponse
	FundingRates []FundingRate `json:"funding_rates"`
}

// MarketConfig represents detailed market configuration
type MarketConfig struct {
	MarketIndex            int16  `json:"market_index"`
	Symbol                 string `json:"symbol"`
	BaseAsset              string `json:"base_asset"`
	QuoteAsset             string `json:"quote_asset"`
	Type                   string `json:"type"` // "perps" or "spot"
	Status                 string `json:"status"` // "active", "halted", "delisted"
	PricePrecision         int    `json:"price_precision"`
	SizePrecision          int    `json:"size_precision"`
	TickSize               string `json:"tick_size"`
	StepSize               string `json:"step_size"`
	MinNotional            string `json:"min_notional"`
	MaxNotional            string `json:"max_notional,omitempty"`
	MinSize                string `json:"min_size"`
	MaxSize                string `json:"max_size"`
	MinPrice               string `json:"min_price"`
	MaxPrice               string `json:"max_price"`
	MaxLeverage            int    `json:"max_leverage,omitempty"`
	MakerFeeRate           string `json:"maker_fee_rate"`
	TakerFeeRate           string `json:"taker_fee_rate"`
	MaintenanceMarginRate  string `json:"maintenance_margin_rate,omitempty"`
	InitialMarginRate      string `json:"initial_margin_rate,omitempty"`
	LiquidationFeeRate     string `json:"liquidation_fee_rate,omitempty"`
	FundingInterval        int64  `json:"funding_interval,omitempty"` // In milliseconds
	MaxFundingRate         string `json:"max_funding_rate,omitempty"`
}

// MarketInfo combines market config with current stats
type MarketInfo struct {
	MarketConfig
	LastPrice       string `json:"last_price"`
	MarkPrice       string `json:"mark_price,omitempty"`
	IndexPrice      string `json:"index_price,omitempty"`
	Volume24h       string `json:"volume_24h"`
	OpenInterest    string `json:"open_interest,omitempty"`
	FundingRate     string `json:"funding_rate,omitempty"`
	NextFundingTime int64  `json:"next_funding_time,omitempty"`
}

// ZkLighterInfo represents general exchange information
type ZkLighterInfo struct {
	BaseResponse
	ChainId            int64  `json:"chain_id"`
	ContractAddress    string `json:"contract_address"`
	L1ContractAddress  string `json:"l1_contract_address,omitempty"`
	Version            string `json:"version"`
	Environment        string `json:"environment"` // "mainnet", "testnet"
}

// Status represents service status
type Status struct {
	BaseResponse
	Status    string `json:"status"` // "ok", "degraded", "down"
	Timestamp int64  `json:"timestamp"`
}

// DailyReturn represents daily return data
type DailyReturn struct {
	Date       string `json:"date"` // YYYY-MM-DD format
	Return     string `json:"return"`
	ReturnPct  string `json:"return_pct"`
	StartValue string `json:"start_value"`
	EndValue   string `json:"end_value"`
}

// SharePrice represents pool share price at a point in time
type SharePrice struct {
	Timestamp  int64  `json:"timestamp"`
	SharePrice string `json:"share_price"`
}

package http

type OrderBook struct {
	Symbol                 string `json:"symbol"`
	MarketId               int32  `json:"market_id"`
	MarketType             string `json:"market_type"`
	BaseAssetId            int32  `json:"base_asset_id"`
	QuoteAssetId           int32  `json:"quote_asset_id"`
	Status                 string `json:"status"`
	TakerFee               string `json:"taker_fee"`
	MakerFee               string `json:"maker_fee"`
	LiquidationFee         string `json:"liquidation_fee"`
	MinBaseAmount          string `json:"min_base_amount"`
	MinQuoteAmount         string `json:"min_quote_amount"`
	OrderQuoteLimit        string `json:"order_quote_limit"`
	SupportedSizeDecimals  int32  `json:"supported_size_decimals"`
	SupportedPriceDecimals int32  `json:"supported_price_decimals"`
	SupportedQuoteDecimals int32  `json:"supported_quote_decimals"`
}

type OrderBooksResponse struct {
	ResultCode
	OrderBooks []OrderBook `json:"order_books"`
}

type PerpsOrderBookDetail struct {
	OrderBook
	SizeDecimals                 int32   `json:"size_decimals"`
	PriceDecimals                int32   `json:"price_decimals"`
	QuoteMultiplier              int32   `json:"quote_multiplier"`
	DefaultInitialMarginFraction int32   `json:"default_initial_margin_fraction"`
	MinInitialMarginFraction     int32   `json:"min_initial_margin_fraction"`
	MaintenanceMarginFraction    int32   `json:"maintenance_margin_fraction"`
	CloseoutMarginFraction       int32   `json:"closeout_margin_fraction"`
	LastTradePrice               float64 `json:"last_trade_price"`
	DailyTradesCount             int32   `json:"daily_trades_count"`
	DailyBaseTokenVolume         float64 `json:"daily_base_token_volume"`
	DailyQuoteTokenVolume        float64 `json:"daily_quote_token_volume"`
	DailyPriceLow                float64 `json:"daily_price_low"`
	DailyPriceHigh               float64 `json:"daily_price_high"`
	DailyPriceChange             float64 `json:"daily_price_change"`
	OpenInterest                 float64 `json:"open_interest"`
	StrategyIndex                int32   `json:"strategy_index"`
}

type SpotOrderBookDetail struct {
	OrderBook
	SizeDecimals          int32   `json:"size_decimals"`
	PriceDecimals         int32   `json:"price_decimals"`
	LastTradePrice        float64 `json:"last_trade_price"`
	DailyTradesCount      int32   `json:"daily_trades_count"`
	DailyBaseTokenVolume  float64 `json:"daily_base_token_volume"`
	DailyQuoteTokenVolume float64 `json:"daily_quote_token_volume"`
	DailyPriceLow         float64 `json:"daily_price_low"`
	DailyPriceHigh        float64 `json:"daily_price_high"`
	DailyPriceChange      float64 `json:"daily_price_change"`
}

type OrderBookDetailsResponse struct {
	ResultCode
	OrderBookDetails     []PerpsOrderBookDetail `json:"order_book_details"`
	SpotOrderBookDetails []SpotOrderBookDetail  `json:"spot_order_book_details"`
}

type SimpleOrder struct {
	OrderIndex          int64  `json:"order_index"`
	OrderId             string `json:"order_id"`
	OwnerAccountIndex   int64  `json:"owner_account_index"`
	InitialBaseAmount   string `json:"initial_base_amount"`
	RemainingBaseAmount string `json:"remaining_base_amount"`
	Price               string `json:"price"`
	OrderExpiry         int64  `json:"order_expiry"`
	TransactionTime     int64  `json:"transaction_time"`
}

type OrderBookOrdersResponse struct {
	ResultCode
	TotalAsks int32         `json:"total_asks"`
	Asks      []SimpleOrder `json:"asks"`
	TotalBids int32         `json:"total_bids"`
	Bids      []SimpleOrder `json:"bids"`
}

type Trade struct {
	TradeId                          int64  `json:"trade_id"`
	TxHash                           string `json:"tx_hash"`
	Type                             string `json:"type"`
	MarketId                         int32  `json:"market_id"`
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
	TakerFee                         int64  `json:"taker_fee"`
	TakerPositionSizeBefore          string `json:"taker_position_size_before"`
	TakerEntryQuoteBefore            string `json:"taker_entry_quote_before"`
	TakerInitialMarginFractionBefore int32  `json:"taker_initial_margin_fraction_before"`
	TakerPositionSignChanged         bool   `json:"taker_position_sign_changed"`
	MakerFee                         int64  `json:"maker_fee"`
	MakerPositionSizeBefore          string `json:"maker_position_size_before"`
	MakerEntryQuoteBefore            string `json:"maker_entry_quote_before"`
	MakerInitialMarginFractionBefore int32  `json:"maker_initial_margin_fraction_before"`
	MakerPositionSignChanged         bool   `json:"maker_position_sign_changed"`
	TransactionTime                  int64  `json:"transaction_time"`
	AskAccountPnl                    string `json:"ask_account_pnl"`
	BidAccountPnl                    string `json:"bid_account_pnl"`
}

type TradesResponse struct {
	ResultCode
	NextCursor *string `json:"next_cursor,omitempty"`
	Trades     []Trade `json:"trades"`
}

type Order struct {
	OrderIndex                 int64  `json:"order_index"`
	ClientOrderIndex           int64  `json:"client_order_index"`
	OrderId                    string `json:"order_id"`
	ClientOrderId              string `json:"client_order_id"`
	MarketIndex                int32  `json:"market_index"`
	OwnerAccountIndex          int64  `json:"owner_account_index"`
	InitialBaseAmount          string `json:"initial_base_amount"`
	Price                      string `json:"price"`
	Nonce                      int64  `json:"nonce"`
	RemainingBaseAmount        string `json:"remaining_base_amount"`
	IsAsk                      bool   `json:"is_ask"`
	BaseSize                   int64  `json:"base_size"`
	BasePrice                  int64  `json:"base_price"`
	FilledBaseAmount           string `json:"filled_base_amount"`
	FilledQuoteAmount          string `json:"filled_quote_amount"`
	Side                       string `json:"side"`
	Type                       string `json:"type"`
	TimeInForce                string `json:"time_in_force"`
	ReduceOnly                 bool   `json:"reduce_only"`
	TriggerPrice               string `json:"trigger_price"`
	OrderExpiry                int64  `json:"order_expiry"`
	Status                     string `json:"status"`
	TriggerStatus              string `json:"trigger_status"`
	TriggerTime                int64  `json:"trigger_time"`
	ParentOrderIndex           int64  `json:"parent_order_index"`
	ParentOrderId              string `json:"parent_order_id"`
	ToTriggerOrderId0          string `json:"to_trigger_order_id_0"`
	ToTriggerOrderId1          string `json:"to_trigger_order_id_1"`
	ToCancelOrderId0           string `json:"to_cancel_order_id_0"`
	IntegratorFeeCollectorIndex string `json:"integrator_fee_collector_index"`
	IntegratorTakerFee         string `json:"integrator_taker_fee"`
	IntegratorMakerFee         string `json:"integrator_maker_fee"`
	BlockHeight                int64  `json:"block_height"`
	Timestamp                  int64  `json:"timestamp"`
	CreatedAt                  int64  `json:"created_at"`
	UpdatedAt                  int64  `json:"updated_at"`
	TransactionTime            int64  `json:"transaction_time"`
}

type OrdersResponse struct {
	ResultCode
	NextCursor *string `json:"next_cursor,omitempty"`
	Orders     []Order `json:"orders"`
}

type TradesParams struct {
	MarketId      *int32
	AccountIndex  *int64
	OrderIndex    *int64
	SortBy        string
	SortDir       string
	Cursor        string
	From          string
	AskFilter     *int32
	Role          string
	Type          string
	Limit         int32
	Aggregate     *bool
	Auth          string
	Authorization string
}

type InactiveOrdersParams struct {
	AccountIndex      int64
	Limit             int32
	Auth              string
	Authorization     string
	MarketId          *int32
	AskFilter         *int32
	BetweenTimestamps string
	Cursor            string
}

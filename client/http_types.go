package client

const (
	CodeOK = 200
)

type ResultCode struct {
	Code    int32  `json:"code,example=200"`
	Message string `json:"message,omitempty"`
}

type NextNonce struct {
	ResultCode
	Nonce int64 `json:"nonce,example=722"`
}

type ApiKey struct {
	AccountIndex int64  `json:"account_index,example=3"`
	ApiKeyIndex  uint8  `json:"api_key_index,example=0"`
	Nonce        int64  `json:"nonce,example=722"`
	PublicKey    string `json:"public_key"`
}

type AccountApiKeys struct {
	ResultCode
	ApiKeys []*ApiKey `json:"api_keys"`
}

type TxHash struct {
	ResultCode
	TxHash string `json:"tx_hash,example=0x70997970C51812dc3A010C7d01b50e0d17dc79C8"`
}

type TxHashBatch struct {
	ResultCode
	TxHash                    []string `json:"tx_hash"`
	PredictedExecutionTimeMs  int      `json:"predicted_execution_time_ms"`
}

type TransferFeeInfo struct {
	ResultCode
	TransferFee int64 `json:"transfer_fee_usdc"`
}

type AccountInfo struct {
	ResultCode
	AccountType              uint8  `json:"account_type,omitempty"`
	Index                    int64  `json:"index,omitempty"`
	L1Address               string `json:"l1_address,omitempty"`
	CancelAllTime           int64  `json:"cancel_all_time,omitempty"`
	TotalOrderCount         int64  `json:"total_order_count,omitempty"`
	TotalIsolatedOrderCount int64  `json:"total_isolated_order_count,omitempty"`
	PendingOrderCount       int64  `json:"pending_order_count,omitempty"`
	AvailableBalance        string `json:"available_balance,omitempty"`
	Status                  uint8  `json:"status,omitempty"`
	CreatedAt               int64  `json:"created_at,omitempty"`
	LastActiveAt            int64  `json:"last_active_at,omitempty"`
	MarketStats             []AccountMarketStats `json:"market_stats,omitempty"`
}

type AccountMarketStats struct {
	MarketId              uint8  `json:"market_id,omitempty"`
	OpenOrderCount        int64  `json:"open_order_count,omitempty"`
	Sign                  int8   `json:"sign,omitempty"`
	Position              string `json:"position,omitempty"`
	AvgEntryPrice         string `json:"avg_entry_price,omitempty"`
	PositionValue         string `json:"position_value,omitempty"`
	UnrealizedPnl         string `json:"unrealized_pnl,omitempty"`
	RealizedPnl           string `json:"realized_pnl,omitempty"`
}

type DetailedAccountsResponse struct {
	ResultCode
	L1Address        string        `json:"l1_address,omitempty"`
	DetailedAccounts []AccountInfo `json:"sub_accounts,omitempty"`
}

type SubAccount struct {
	Code                     int    `json:"code"`
	AccountType              int    `json:"account_type"`
	Index                    int64  `json:"index"`
	L1Address                string `json:"l1_address"`
	CancelAllTime            int    `json:"cancel_all_time"`
	TotalOrderCount          int    `json:"total_order_count"`
	TotalIsolatedOrderCount  int    `json:"total_isolated_order_count"`
	PendingOrderCount        int    `json:"pending_order_count"`
	AvailableBalance         string `json:"available_balance"`
	Status                   int    `json:"status"`
	Collateral               string `json:"collateral"`
}

type AccountByL1AddressResponse struct {
	Code        int          `json:"code"`
	L1Address   string       `json:"l1_address"`
	SubAccounts []SubAccount `json:"sub_accounts"`
}

type Position struct {
	MarketId                 uint8  `json:"market_id"`
	Symbol                   string `json:"symbol"`
	InitialMarginFraction    string `json:"initial_margin_fraction"`
	OpenOrderCount           int    `json:"open_order_count"`
	PendingOrderCount        int    `json:"pending_order_count"`
	PositionTiedOrderCount   int    `json:"position_tied_order_count"`
	Sign                     int    `json:"sign"`
	Position                 string `json:"position"`
	AvgEntryPrice            string `json:"avg_entry_price"`
	PositionValue            string `json:"position_value"`
	UnrealizedPnl            string `json:"unrealized_pnl"`
	RealizedPnl              string `json:"realized_pnl"`
	LiquidationPrice         string `json:"liquidation_price"`
	MarginMode               int    `json:"margin_mode"`
	AllocatedMargin          string `json:"allocated_margin"`
}

type Share struct {
	PublicPoolIndex int64  `json:"public_pool_index"`
	SharesAmount    int    `json:"shares_amount"`
	EntryUsdc       string `json:"entry_usdc"`
}

type Account struct {
	Code                     int        `json:"code"`
	AccountType              int        `json:"account_type"`
	Index                    int64      `json:"index"`
	L1Address                string     `json:"l1_address"`
	CancelAllTime            int        `json:"cancel_all_time"`
	TotalOrderCount          int        `json:"total_order_count"`
	TotalIsolatedOrderCount  int        `json:"total_isolated_order_count"`
	PendingOrderCount        int        `json:"pending_order_count"`
	AvailableBalance         string     `json:"available_balance"`
	Status                   int        `json:"status"`
	Collateral               string     `json:"collateral"`
	AccountIndex             int64      `json:"account_index"`
	Name                     string     `json:"name"`
	Description              string     `json:"description"`
	CanInvite                bool       `json:"can_invite"`
	ReferralPointsPercentage string     `json:"referral_points_percentage"`
	Positions                []Position `json:"positions"`
	TotalAssetValue          string     `json:"total_asset_value"`
	CrossAssetValue          string     `json:"cross_asset_value"`
	Shares                   []Share    `json:"shares"`
}

type AccountResponse struct {
	Code     int       `json:"code"`
	Total    int       `json:"total"`
	Accounts []Account `json:"accounts"`
}

type OrderBookResponse struct {
	ResultCode
	OrderBooks []OrderBook `json:"order_books,omitempty"`
}

type OrderBook struct {
	Symbol                   string `json:"symbol"`
	MarketId                 uint8  `json:"market_id"`
	Status                   string `json:"status"`
	TakerFee                 string `json:"taker_fee"`
	MakerFee                 string `json:"maker_fee"`
	LiquidationFee           string `json:"liquidation_fee"`
	MinBaseAmount            string `json:"min_base_amount"`
	MinQuoteAmount           string `json:"min_quote_amount"`
	SupportedSizeDecimals    uint8  `json:"supported_size_decimals"`
	SupportedPriceDecimals   uint8  `json:"supported_price_decimals"`
	SupportedQuoteDecimals   uint8  `json:"supported_quote_decimals"`
}

type OrderBookData struct {
	MarketId uint8        `json:"market_id"`
	Bids     []PriceLevel `json:"bids"`
	Asks     []PriceLevel `json:"asks"`
}

type OrderBookDataResponse struct {
	ResultCode
	OrderBooks []OrderBookData `json:"order_books"`
}

type OrderBookDetail struct {
	Symbol                       string  `json:"symbol"`
	MarketId                     uint8   `json:"market_id"`
	Status                       string  `json:"status"`
	TakerFee                     string  `json:"taker_fee"`
	MakerFee                     string  `json:"maker_fee"`
	LiquidationFee               string  `json:"liquidation_fee"`
	MinBaseAmount                string  `json:"min_base_amount"`
	MinQuoteAmount               string  `json:"min_quote_amount"`
	SupportedSizeDecimals        uint8   `json:"supported_size_decimals"`
	SupportedPriceDecimals       uint8   `json:"supported_price_decimals"`
	SupportedQuoteDecimals       uint8   `json:"supported_quote_decimals"`
	SizeDecimals                 uint8   `json:"size_decimals"`
	PriceDecimals                uint8   `json:"price_decimals"`
	QuoteMultiplier              uint8   `json:"quote_multiplier"`
	DefaultInitialMarginFraction uint32  `json:"default_initial_margin_fraction"`
	MinInitialMarginFraction     uint32  `json:"min_initial_margin_fraction"`
	MaintenanceMarginFraction    uint32  `json:"maintenance_margin_fraction"`
	CloseoutMarginFraction       uint32  `json:"closeout_margin_fraction"`
	LastTradePrice               float64 `json:"last_trade_price"`
	DailyTradesCount             uint32  `json:"daily_trades_count"`
	DailyBaseTokenVolume         float64 `json:"daily_base_token_volume"`
	DailyQuoteTokenVolume        float64 `json:"daily_quote_token_volume"`
	DailyPriceLow                float64 `json:"daily_price_low"`
	DailyPriceHigh               float64 `json:"daily_price_high"`
	DailyPriceChange             float64 `json:"daily_price_change"`
	OpenInterest                 float64 `json:"open_interest"`
	DailyChart                   map[string]interface{} `json:"daily_chart"`
}

type OrderBookDetailsResponse struct {
	ResultCode
	OrderBookDetails []OrderBookDetail `json:"order_book_details"`
}

type PriceLevel struct {
	Price    string `json:"price,omitempty"`
	Quantity string `json:"quantity,omitempty"`
}

type OrdersResponse struct {
	ResultCode
	Orders []Order `json:"orders,omitempty"`
}

type FundingRate struct {
	MarketId int    `json:"market_id"`
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
	Rate     float64 `json:"rate"`
}

type FundingRatesResponse struct {
	ResultCode
	FundingRates []FundingRate `json:"funding_rates"`
}

type Order struct {
	// Core order identification
	OrderIndex       int64  `json:"order_index,omitempty"`
	ClientOrderIndex int64  `json:"client_order_index,omitempty"`
	OrderId          string `json:"order_id,omitempty"`
	ClientOrderId    string `json:"client_order_id,omitempty"`
	
	// Market and account info
	MarketIndex          uint8 `json:"market_index,omitempty"`
	OwnerAccountIndex    int64 `json:"owner_account_index,omitempty"`
	
	// Order details - using human-readable formats from API
	InitialBaseAmount    string `json:"initial_base_amount,omitempty"`    // e.g. "0.100"
	Price               string `json:"price,omitempty"`                   // e.g. "203.577"
	RemainingBaseAmount  string `json:"remaining_base_amount,omitempty"`   // e.g. "0.100"
	FilledBaseAmount     string `json:"filled_base_amount,omitempty"`      // e.g. "0.000"
	FilledQuoteAmount    string `json:"filled_quote_amount,omitempty"`     // e.g. "0.000000"
	
	// Order properties
	IsAsk        bool   `json:"is_ask,omitempty"`
	Side         string `json:"side,omitempty"`          // "buy", "sell", or ""
	Type         string `json:"type,omitempty"`          // "limit", "market", etc.
	TimeInForce  string `json:"time_in_force,omitempty"` // "post-only", "good-till-time", etc.
	ReduceOnly   bool   `json:"reduce_only,omitempty"`
	TriggerPrice string `json:"trigger_price,omitempty"`
	
	// Status and timing
	Status       string `json:"status,omitempty"`        // "open", "filled", "cancelled", etc.
	Timestamp    int64  `json:"timestamp,omitempty"`     // Unix timestamp in seconds
	OrderExpiry  int64  `json:"order_expiry,omitempty"`
	
	// Internal fields (still using scaled values)
	Nonce     int64 `json:"nonce,omitempty"`
	BaseSize  int   `json:"base_size,omitempty"`         // Scaled value: 100
	BasePrice int   `json:"base_price,omitempty"`        // Scaled value: 203577
	
	// Trigger and parent order fields
	TriggerStatus     string `json:"trigger_status,omitempty"`
	TriggerTime       int64  `json:"trigger_time,omitempty"`
	ParentOrderIndex  int64  `json:"parent_order_index,omitempty"`
	ParentOrderId     string `json:"parent_order_id,omitempty"`
	ToTriggerOrderId0 string `json:"to_trigger_order_id_0,omitempty"`
	ToTriggerOrderId1 string `json:"to_trigger_order_id_1,omitempty"`
	ToCancelOrderId0  string `json:"to_cancel_order_id_0,omitempty"`
	
	// Block info
	BlockHeight int64 `json:"block_height,omitempty"`
}

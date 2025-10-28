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

type AccountByL1Address struct {
	ResultCode
	L1Address   string                  `json:"l1_address"`
	SubAccounts []SubAccountByL1Address `json:"sub_accounts"`
}

type SubAccountByL1Address struct {
	// The code field is also returned in the SubAccounts structure, but with code = 0
	// SubAccounts结构里也会返回code字段，但是code=0
	ResultCode

	// Lighter API users can operate under a Standard or Premium accounts.
	// The Standard account is fee-less.
	// Premium accounts pay 0.2 bps maker and 2 bps taker fees. Find out more in Account Types.
	AccountType             uint8  `json:"account_type"`
	Index                   int64  `json:"index"`
	L1Address               string `json:"l1_address"`
	CancelAllTime           int64  `json:"cancel_all_time"`
	TotalOrderCount         int64  `json:"total_order_count"`
	TotalIsolatedOrderCount int64  `json:"total_isolated_order_count"`
	PendingOrderCount       int64  `json:"pending_order_count"`
	AvailableBalance        string `json:"available_balance"`
	Status                  uint8  `json:"status"`     // 1 is active, 0 is inactive.
	Collateral              string `json:"collateral"` // The amount of collateral in the account.
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

type TxInfo struct {
	ResultCode
	Hash             string `json:"hash"`
	Type             uint8  `json:"type"`
	Info             string `json:"info"`
	EventInfo        string `json:"event_info"`
	Status           int64  `json:"status"`
	TransactionIndex int64  `json:"transaction_index"`
	L1Address        string `json:"l1_address"`
	AccountIndex     int64  `json:"account_index"`
	Nonce            int64  `json:"nonce"`
	ExpireAt         int64  `json:"expire_at"`
	BlockHeight      int64  `json:"block_height"`
	QueuedAt         int64  `json:"queued_at"`
	SequenceIndex    int64  `json:"sequence_index"`
	ParentHash       string `json:"parent_hash"`
	CommittedAt      int64  `json:"committed_at"`
	VerifiedAt       int64  `json:"verified_at"`
	ExecutedAt       int64  `json:"executed_at"`
}

type TransferFeeInfo struct {
	ResultCode
	TransferFee int64 `json:"transfer_fee_usdc"`
}

type OrderBookDetails struct {
	ResultCode
	OrderBookDetails []OrderBookDetail `json:"order_book_details"`
}

type OrderBookDetail struct {
	Symbol                       string  `json:"symbol"`
	MarketID                     int     `json:"market_id"`
	Status                       string  `json:"status"`
	TakerFee                     string  `json:"taker_fee"`
	MakerFee                     string  `json:"maker_fee"`
	LiquidationFee               string  `json:"liquidation_fee"`
	MinBaseAmount                string  `json:"min_base_amount"`
	MinQuoteAmount               string  `json:"min_quote_amount"`
	OrderQuoteLimit              string  `json:"order_quote_limit"`
	SupportedSizeDecimals        int     `json:"supported_size_decimals"`
	SupportedPriceDecimals       int     `json:"supported_price_decimals"`
	SupportedQuoteDecimals       int     `json:"supported_quote_decimals"`
	SizeDecimals                 int     `json:"size_decimals"`
	PriceDecimals                int     `json:"price_decimals"`
	QuoteMultiplier              int     `json:"quote_multiplier"`
	DefaultInitialMarginFraction int     `json:"default_initial_margin_fraction"`
	MinInitialMarginFraction     int     `json:"min_initial_margin_fraction"`
	MaintenanceMarginFraction    int     `json:"maintenance_margin_fraction"`
	CloseoutMarginFraction       int     `json:"closeout_margin_fraction"`
	LastTradePrice               float64 `json:"last_trade_price"`
	DailyTradesCount             int     `json:"daily_trades_count"`
	DailyBaseTokenVolume         float64 `json:"daily_base_token_volume"`
	DailyQuoteTokenVolume        float64 `json:"daily_quote_token_volume"`
	DailyPriceLow                float64 `json:"daily_price_low"`
	DailyPriceHigh               float64 `json:"daily_price_high"`
	DailyPriceChange             float64 `json:"daily_price_change"`
	OpenInterest                 float64 `json:"open_interest"`
	DailyChart                   struct {
	} `json:"daily_chart"`
	MarketConfig MarketConfig `json:"market_config"`
}

type MarketConfig struct {
	MarketMarginMode          int   `json:"market_margin_mode"`
	InsuranceFundAccountIndex int64 `json:"insurance_fund_account_index"`
}

type OrderBookOrders struct {
	ResultCode
	TotalAsks int     `json:"total_asks"`
	Asks      []Order `json:"asks"`
	TotalBids int     `json:"total_bids"`
	Bids      []Order `json:"bids"`
}

type Order struct {
	OrderIndex          int64  `json:"order_index"`
	OrderID             string `json:"order_id"`
	OwnerAccountIndex   int64  `json:"owner_account_index"`
	InitialBaseAmount   string `json:"initial_base_amount"`
	RemainingBaseAmount string `json:"remaining_base_amount"`
	Price               string `json:"price"`
	OrderExpiry         int64  `json:"order_expiry"`
}

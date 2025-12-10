package client

const (
	CodeOK         = 200
	MarketTypeSpot = "spot"
	MarketTypePerp = "perp"
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
	TxHash                   string `json:"tx_hash,example=0x70997970C51812dc3A010C7d01b50e0d17dc79C8"`
	PredictedExecutionTimeMs int64  `json:"predicted_execution_time_ms"`
}

type TxHashBatch struct {
	ResultCode
	TxHash                   []string `json:"tx_hash"`
	PredictedExecutionTimeMs int64    `json:"predicted_execution_time_ms"`
}

type TransferFeeInfo struct {
	ResultCode
	TransferFee int64 `json:"transfer_fee_usdc"`
}

type AccountInfo struct {
	ResultCode
	AccountType             uint8                `json:"account_type,omitempty"`
	Index                   int64                `json:"index,omitempty"`
	L1Address               string               `json:"l1_address,omitempty"`
	CancelAllTime           int64                `json:"cancel_all_time,omitempty"`
	TotalOrderCount         int64                `json:"total_order_count,omitempty"`
	TotalIsolatedOrderCount int64                `json:"total_isolated_order_count,omitempty"`
	PendingOrderCount       int64                `json:"pending_order_count,omitempty"`
	AvailableBalance        string               `json:"available_balance,omitempty"`
	Status                  uint8                `json:"status,omitempty"`
	CreatedAt               int64                `json:"created_at,omitempty"`
	LastActiveAt            int64                `json:"last_active_at,omitempty"`
	MarketStats             []AccountMarketStats `json:"market_stats,omitempty"`
}

type AccountMarketStats struct {
	MarketId       uint16 `json:"market_id,omitempty"`
	OpenOrderCount int64  `json:"open_order_count,omitempty"`
	Sign           int8   `json:"sign,omitempty"`
	Position       string `json:"position,omitempty"`
	AvgEntryPrice  string `json:"avg_entry_price,omitempty"`
	PositionValue  string `json:"position_value,omitempty"`
	UnrealizedPnl  string `json:"unrealized_pnl,omitempty"`
	RealizedPnl    string `json:"realized_pnl,omitempty"`
}

type DetailedAccountsResponse struct {
	ResultCode
	L1Address        string        `json:"l1_address,omitempty"`
	DetailedAccounts []AccountInfo `json:"sub_accounts,omitempty"`
}

type SubAccount struct {
	Code                    int    `json:"code"`
	AccountType             int    `json:"account_type"`
	Index                   int64  `json:"index"`
	L1Address               string `json:"l1_address"`
	CancelAllTime           int    `json:"cancel_all_time"`
	TotalOrderCount         int    `json:"total_order_count"`
	TotalIsolatedOrderCount int    `json:"total_isolated_order_count"`
	PendingOrderCount       int    `json:"pending_order_count"`
	AvailableBalance        string `json:"available_balance"`
	Status                  int    `json:"status"`
	Collateral              string `json:"collateral"`
}

type AccountByL1AddressResponse struct {
	Code        int          `json:"code"`
	L1Address   string       `json:"l1_address"`
	SubAccounts []SubAccount `json:"sub_accounts"`
}

type Position struct {
	MarketId               uint16 `json:"market_id"`
	Symbol                 string `json:"symbol"`
	InitialMarginFraction  string `json:"initial_margin_fraction"`
	OpenOrderCount         int    `json:"open_order_count"`
	PendingOrderCount      int    `json:"pending_order_count"`
	PositionTiedOrderCount int    `json:"position_tied_order_count"`
	Sign                   int    `json:"sign"`
	Position               string `json:"position"`
	AvgEntryPrice          string `json:"avg_entry_price"`
	PositionValue          string `json:"position_value"`
	UnrealizedPnl          string `json:"unrealized_pnl"`
	RealizedPnl            string `json:"realized_pnl"`
	LiquidationPrice       string `json:"liquidation_price"`
	MarginMode             int    `json:"margin_mode"`
	AllocatedMargin        string `json:"allocated_margin"`
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
	Symbol          string `json:"symbol"`
	MarketId        uint16 `json:"market_id"`
	MarketType      string `json:"market_type"`
	Status          string `json:"status"`
	TakerFee        string `json:"taker_fee"`
	MakerFee        string `json:"maker_fee"`
	LiquidationFee  string `json:"liquidation_fee"`
	MinBaseAmount   string `json:"min_base_amount"`
	MinQuoteAmount  string `json:"min_quote_amount"`
	OrderQuoteLimit string `json:"order_quote_limit"`

	SupportedSizeDecimals  uint8 `json:"supported_size_decimals"`
	SupportedPriceDecimals uint8 `json:"supported_price_decimals"`
	SupportedQuoteDecimals uint8 `json:"supported_quote_decimals"`
}

type OrderBookData struct {
	MarketId uint16       `json:"market_id"`
	Bids     []PriceLevel `json:"bids"`
	Asks     []PriceLevel `json:"asks"`
}

type OrderBookDataResponse struct {
	ResultCode
	OrderBooks []OrderBookData `json:"order_books"`
}

type OrderBookDetail struct {
	Symbol                       string                 `json:"symbol"`
	MarketId                     uint16                 `json:"market_id"`
	MarketType                   string                 `json:"market_type"`
	Status                       string                 `json:"status"`
	TakerFee                     string                 `json:"taker_fee"`
	MakerFee                     string                 `json:"maker_fee"`
	LiquidationFee               string                 `json:"liquidation_fee"`
	MinBaseAmount                string                 `json:"min_base_amount"`
	MinQuoteAmount               string                 `json:"min_quote_amount"`
	SupportedSizeDecimals        uint8                  `json:"supported_size_decimals"`
	SupportedPriceDecimals       uint8                  `json:"supported_price_decimals"`
	SupportedQuoteDecimals       uint8                  `json:"supported_quote_decimals"`
	SizeDecimals                 uint8                  `json:"size_decimals"`
	PriceDecimals                uint8                  `json:"price_decimals"`
	QuoteMultiplier              uint8                  `json:"quote_multiplier"`
	DefaultInitialMarginFraction uint32                 `json:"default_initial_margin_fraction"`
	MinInitialMarginFraction     uint32                 `json:"min_initial_margin_fraction"`
	MaintenanceMarginFraction    uint32                 `json:"maintenance_margin_fraction"`
	CloseoutMarginFraction       uint32                 `json:"closeout_margin_fraction"`
	LastTradePrice               float64                `json:"last_trade_price"`
	DailyTradesCount             uint32                 `json:"daily_trades_count"`
	DailyBaseTokenVolume         float64                `json:"daily_base_token_volume"`
	DailyQuoteTokenVolume        float64                `json:"daily_quote_token_volume"`
	DailyPriceLow                float64                `json:"daily_price_low"`
	DailyPriceHigh               float64                `json:"daily_price_high"`
	DailyPriceChange             float64                `json:"daily_price_change"`
	OpenInterest                 float64                `json:"open_interest"`
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
	MarketId int     `json:"market_id"`
	Exchange string  `json:"exchange"`
	Symbol   string  `json:"symbol"`
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
	MarketIndex       uint16 `json:"market_index,omitempty"`
	OwnerAccountIndex int64  `json:"owner_account_index,omitempty"`

	// Order details - using human-readable formats from API
	InitialBaseAmount   string `json:"initial_base_amount,omitempty"`   // e.g. "0.100"
	Price               string `json:"price,omitempty"`                 // e.g. "203.577"
	RemainingBaseAmount string `json:"remaining_base_amount,omitempty"` // e.g. "0.100"
	FilledBaseAmount    string `json:"filled_base_amount,omitempty"`    // e.g. "0.000"
	FilledQuoteAmount   string `json:"filled_quote_amount,omitempty"`   // e.g. "0.000000"

	// Order properties
	IsAsk        bool   `json:"is_ask,omitempty"`
	Side         string `json:"side,omitempty"`          // "buy", "sell", or ""
	Type         string `json:"type,omitempty"`          // "limit", "market", etc.
	TimeInForce  string `json:"time_in_force,omitempty"` // "post-only", "good-till-time", etc.
	ReduceOnly   bool   `json:"reduce_only,omitempty"`
	TriggerPrice string `json:"trigger_price,omitempty"`

	// Status and timing
	Status      string `json:"status,omitempty"`    // "open", "filled", "cancelled", etc.
	Timestamp   int64  `json:"timestamp,omitempty"` // Unix timestamp in seconds
	OrderExpiry int64  `json:"order_expiry,omitempty"`

	// Internal fields (still using scaled values)
	Nonce     int64 `json:"nonce,omitempty"`
	BaseSize  int   `json:"base_size,omitempty"`  // Scaled value: 100
	BasePrice int   `json:"base_price,omitempty"` // Scaled value: 203577

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

// ============= Phase 1: Core Data Query Types =============

// Candlestick represents a single candlestick data point
type Candlestick struct {
	MarketId            uint16 `json:"market_id"`
	Symbol              string `json:"symbol"`
	Resolution          string `json:"resolution"`             // "1", "5", "15", "60", "240", "1D"
	Timestamp           int64  `json:"timestamp"`              // Unix timestamp
	Open                string `json:"open"`                   // Opening price
	High                string `json:"high"`                   // Highest price
	Low                 string `json:"low"`                    // Lowest price
	Close               string `json:"close"`                  // Closing price
	Volume              string `json:"volume"`                 // Trading volume
	QuoteVolume         string `json:"quote_volume"`           // Quote volume
	TradesCount         int32  `json:"trades_count"`           // Number of trades
	TakerBuyVolume      string `json:"taker_buy_volume"`       // Taker buy volume
	TakerBuyQuoteVolume string `json:"taker_buy_quote_volume"` // Taker buy quote volume
}

// CandlesticksResponse represents the response for candlesticks API
type CandlesticksResponse struct {
	ResultCode
	Candlesticks []Candlestick `json:"candlesticks"`
}

// FundingHistory represents historical funding data
type FundingHistory struct {
	MarketId        uint16 `json:"market_id"`
	Symbol          string `json:"symbol"`
	Timestamp       int64  `json:"timestamp"`         // Unix timestamp
	FundingRate     string `json:"funding_rate"`      // Funding rate
	IndexPrice      string `json:"index_price"`       // Index price at funding time
	MarkPrice       string `json:"mark_price"`        // Mark price at funding time
	PremiumRate     string `json:"premium_rate"`      // Premium rate
	NextFundingTime int64  `json:"next_funding_time"` // Next funding timestamp
}

// FundingsResponse represents the response for fundings history API
type FundingsResponse struct {
	ResultCode
	Fundings []FundingHistory `json:"fundings"`
}

// Trade represents a single trade
type Trade struct {
	TradeId       int64  `json:"trade_id"`
	TxHash        string `json:"tx_hash"`
	MarketId      uint16 `json:"market_id"`
	Symbol        string `json:"symbol"`
	Price         string `json:"price"`
	Size          string `json:"size"`
	QuoteQuantity string `json:"quote_quantity"`
	Side          string `json:"side"` // "buy" or "sell"
	IsMakerAsk    bool   `json:"is_maker_ask"`
	Timestamp     int64  `json:"timestamp"`
	BlockHeight   int64  `json:"block_height"`

	// Optional account info for user trades
	MakerAccountIndex int64  `json:"maker_account_index,omitempty"`
	TakerAccountIndex int64  `json:"taker_account_index,omitempty"`
	MakerOrderId      string `json:"maker_order_id,omitempty"`
	TakerOrderId      string `json:"taker_order_id,omitempty"`
}

// TradesResponse represents the response for trades API
type TradesResponse struct {
	ResultCode
	Trades []Trade `json:"trades"`
}

// RecentTradesResponse represents the response for recent trades API
type RecentTradesResponse struct {
	ResultCode
	Trades []Trade `json:"recent_trades"`
}

// ExchangeStats represents exchange-wide statistics
type ExchangeStats struct {
	TotalVolume24h      string `json:"total_volume_24h"`      // 24h total volume
	TotalTrades24h      int64  `json:"total_trades_24h"`      // 24h total trades count
	TotalUsers          int64  `json:"total_users"`           // Total registered users
	ActiveUsers24h      int64  `json:"active_users_24h"`      // Active users in 24h
	TotalMarkets        int32  `json:"total_markets"`         // Total number of markets
	ActiveMarkets       int32  `json:"active_markets"`        // Currently active markets
	SystemStatus        string `json:"system_status"`         // "normal", "maintenance", etc
	LastUpdateTimestamp int64  `json:"last_update_timestamp"` // Last update time
}

// ExchangeStatsResponse represents the response for exchange stats API
type ExchangeStatsResponse struct {
	ResultCode
	Stats ExchangeStats `json:"exchange_stats"`
}

// ExportData represents data export information
type ExportData struct {
	RequestId     string `json:"request_id"`
	Status        string `json:"status"`                    // "pending", "processing", "completed", "failed"
	DataType      string `json:"data_type"`                 // "trades", "orders", "positions", etc
	StartDate     string `json:"start_date"`                // ISO date string
	EndDate       string `json:"end_date"`                  // ISO date string
	DownloadUrl   string `json:"download_url,omitempty"`    // Available when completed
	CreatedAt     int64  `json:"created_at"`                // Request timestamp
	CompletedAt   int64  `json:"completed_at,omitempty"`    // Completion timestamp
	ExpiresAt     int64  `json:"expires_at,omitempty"`      // URL expiration timestamp
	RecordCount   int64  `json:"record_count,omitempty"`    // Number of records
	FileSizeBytes int64  `json:"file_size_bytes,omitempty"` // File size
}

// ExportResponse represents the response for data export API
type ExportResponse struct {
	ResultCode
	Export ExportData `json:"export"`
}

// ============= Phase 2: Account Management Types =============

// AccountLimits represents account trading limits and restrictions
type AccountLimits struct {
	AccountIndex       int64    `json:"account_index"`
	MaxDailyTrades     int32    `json:"max_daily_trades"`
	MaxOrderCount      int32    `json:"max_order_count"`
	MaxPositionSize    string   `json:"max_position_size"`   // In base currency
	MaxOrderSize       string   `json:"max_order_size"`      // In base currency
	MaxNotionalValue   string   `json:"max_notional_value"`  // In quote currency
	WithdrawalLimit    string   `json:"withdrawal_limit"`    // Daily withdrawal limit
	DepositLimit       string   `json:"deposit_limit"`       // Daily deposit limit
	TierLevel          int32    `json:"tier_level"`          // Account tier (0=basic, 1=verified, etc)
	RequiresKyc        bool     `json:"requires_kyc"`        // Whether KYC is required
	RestrictedMarkets  []uint8  `json:"restricted_markets"`  // Markets user cannot trade
	RestrictedFeatures []string `json:"restricted_features"` // Features not available
	LastUpdated        int64    `json:"last_updated"`        // When limits were last updated
}

// AccountLimitsResponse represents the response for account limits API
type AccountLimitsResponse struct {
	ResultCode
	Limits AccountLimits `json:"limits"`
}

// AccountMetadata represents extended account information
type AccountMetadata struct {
	AccountIndex      int64                  `json:"account_index"`
	L1Address         string                 `json:"l1_address"`
	CreatedAt         int64                  `json:"created_at"`
	LastActiveAt      int64                  `json:"last_active_at"`
	TotalTradeCount   int64                  `json:"total_trade_count"`
	TotalVolume       string                 `json:"total_volume"`
	AverageTradeSize  string                 `json:"average_trade_size"`
	PreferredMarkets  []uint8                `json:"preferred_markets"`
	ReferralCode      string                 `json:"referral_code,omitempty"`
	ReferredBy        string                 `json:"referred_by,omitempty"`
	AccountTier       int32                  `json:"account_tier"`
	VerificationLevel string                 `json:"verification_level"` // "none", "basic", "advanced"
	Tags              []string               `json:"tags,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"` // Additional flexible data
}

// AccountMetadataResponse represents the response for account metadata API
type AccountMetadataResponse struct {
	ResultCode
	AccountMetadata AccountMetadata `json:"account_metadata"`
}

// Liquidation represents a single liquidation event
type Liquidation struct {
	LiquidationId    int64  `json:"liquidation_id"`
	AccountIndex     int64  `json:"account_index"`
	MarketId         uint16 `json:"market_id"`
	Symbol           string `json:"symbol"`
	LiquidatedSize   string `json:"liquidated_size"`   // Size liquidated
	LiquidationPrice string `json:"liquidation_price"` // Price at liquidation
	MarkPrice        string `json:"mark_price"`        // Mark price at liquidation
	UnrealizedPnL    string `json:"unrealized_pnl"`    // PnL realized from liquidation
	Fee              string `json:"fee"`               // Liquidation fee charged
	LiquidationType  string `json:"liquidation_type"`  // "auto", "forced", "insurance"
	TriggerReason    string `json:"trigger_reason"`    // "margin_call", "adl", etc
	Timestamp        int64  `json:"timestamp"`         // When liquidation occurred
	BlockHeight      int64  `json:"block_height"`
	TxHash           string `json:"tx_hash"`
}

// LiquidationsResponse represents the response for liquidations API
type LiquidationsResponse struct {
	ResultCode
	Liquidations []Liquidation `json:"liquidations"`
}

// PnLEntry represents a single profit/loss record
type PnLEntry struct {
	AccountIndex   int64  `json:"account_index"`
	MarketId       uint16 `json:"market_id"`
	Symbol         string `json:"symbol"`
	Date           string `json:"date"`            // YYYY-MM-DD format
	RealizedPnL    string `json:"realized_pnl"`    // Daily realized PnL
	UnrealizedPnL  string `json:"unrealized_pnl"`  // End of day unrealized PnL
	TradingFees    string `json:"trading_fees"`    // Total fees paid
	FundingFees    string `json:"funding_fees"`    // Total funding fees
	NetPnL         string `json:"net_pnl"`         // Net PnL (realized - fees)
	OpeningBalance string `json:"opening_balance"` // Balance at start of day
	ClosingBalance string `json:"closing_balance"` // Balance at end of day
	TradesCount    int32  `json:"trades_count"`    // Number of trades
	Volume         string `json:"volume"`          // Total trading volume
}

// PnLResponse represents the response for PnL history API
type PnLResponse struct {
	ResultCode
	PnLEntries []PnLEntry `json:"pnl_entries"`
}

// PositionFunding represents funding fee for a position
type PositionFunding struct {
	AccountIndex     int64  `json:"account_index"`
	MarketId         uint16 `json:"market_id"`
	Symbol           string `json:"symbol"`
	PositionSize     string `json:"position_size"`     // Size of position when funding was applied
	FundingRate      string `json:"funding_rate"`      // Funding rate applied
	FundingFee       string `json:"funding_fee"`       // Fee paid (negative) or received (positive)
	IndexPrice       string `json:"index_price"`       // Index price at funding time
	MarkPrice        string `json:"mark_price"`        // Mark price at funding time
	FundingTimestamp int64  `json:"funding_timestamp"` // When funding was applied
	BlockHeight      int64  `json:"block_height"`
}

// PositionFundingResponse represents the response for position funding API
type PositionFundingResponse struct {
	ResultCode
	PositionFundings []PositionFunding `json:"position_fundings"`
}

// PublicPool represents information about a public liquidity pool
type PublicPool struct {
	PoolIndex     int64  `json:"pool_index"`
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	TotalShares   string `json:"total_shares"`   // Total shares outstanding
	TotalValue    string `json:"total_value"`    // Total value in USDC
	SharePrice    string `json:"share_price"`    // Current price per share
	DailyReturn   string `json:"daily_return"`   // 24h return percentage
	WeeklyReturn  string `json:"weekly_return"`  // 7d return percentage
	MonthlyReturn string `json:"monthly_return"` // 30d return percentage
	CreatedAt     int64  `json:"created_at"`
	Manager       string `json:"manager"`        // Pool manager address
	Fee           string `json:"fee"`            // Management fee percentage
	MinInvestment string `json:"min_investment"` // Minimum investment
	Status        string `json:"status"`         // "active", "closed", "paused"
}

// PublicPoolsResponse represents the response for public pools API
type PublicPoolsResponse struct {
	ResultCode
	PublicPools []PublicPool `json:"public_pools"`
}

// PublicPoolMetadata represents metadata for public pools
type PublicPoolMetadata struct {
	PoolIndex       int64                  `json:"pool_index"`
	CreatedAt       int64                  `json:"created_at"`
	UpdatedAt       int64                  `json:"updated_at"`
	Version         string                 `json:"version"`
	StrategyType    string                 `json:"strategy_type"` // "market_making", "arbitrage", etc
	RiskLevel       string                 `json:"risk_level"`    // "low", "medium", "high"
	SupportedAssets []string               `json:"supported_assets"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PublicPoolsMetadataResponse represents the response for public pools metadata API
type PublicPoolsMetadataResponse struct {
	ResultCode
	PoolsMetadata []PublicPoolMetadata `json:"pools_metadata"`
}

// ChangeAccountTierRequest represents request to change account tier
type ChangeAccountTierRequest struct {
	AccountIndex int64  `json:"account_index"`
	NewTier      int32  `json:"new_tier"`
	Auth         string `json:"auth"`
}

// ChangeAccountTierResponse represents the response for account tier change
type ChangeAccountTierResponse struct {
	ResultCode
	Success     bool   `json:"success"`
	OldTier     int32  `json:"old_tier"`
	NewTier     int32  `json:"new_tier"`
	Message     string `json:"message,omitempty"`
	EffectiveAt int64  `json:"effective_at"` // When change takes effect
}

// ============= Phase 3: Transaction History Types =============

// Transaction represents a blockchain transaction
type Transaction struct {
	TxHash       string                 `json:"tx_hash"`
	TxType       int32                  `json:"tx_type"`      // Transaction type ID
	TxTypeName   string                 `json:"tx_type_name"` // Human readable type
	AccountIndex int64                  `json:"account_index"`
	MarketId     uint16                 `json:"market_id,omitempty"`
	Nonce        int64                  `json:"nonce"`
	Status       string                 `json:"status"` // "pending", "confirmed", "failed"
	BlockHeight  int64                  `json:"block_height,omitempty"`
	Timestamp    int64                  `json:"timestamp"`
	GasFee       string                 `json:"gas_fee,omitempty"`
	TxData       map[string]interface{} `json:"tx_data,omitempty"` // Transaction specific data
}

// AccountTxsResponse represents the response for account transactions API
type AccountTxsResponse struct {
	ResultCode
	Transactions []Transaction `json:"transactions"`
}

// BlockTx represents a transaction within a specific block
type BlockTx struct {
	TxHash      string                 `json:"tx_hash"`
	TxType      int32                  `json:"tx_type"`
	TxTypeName  string                 `json:"tx_type_name"`
	TxIndex     int32                  `json:"tx_index"` // Position in block
	Status      string                 `json:"status"`
	GasFee      string                 `json:"gas_fee"`
	FromAccount int64                  `json:"from_account"`
	ToAccount   int64                  `json:"to_account,omitempty"`
	TxData      map[string]interface{} `json:"tx_data"`
}

// Block represents blockchain block information
type Block struct {
	BlockHeight  int64     `json:"block_height"`
	BlockHash    string    `json:"block_hash"`
	Timestamp    int64     `json:"timestamp"`
	TxCount      int32     `json:"tx_count"`
	Transactions []BlockTx `json:"transactions"`
}

// BlockTxsResponse represents the response for block transactions API
type BlockTxsResponse struct {
	ResultCode
	Block Block `json:"block"`
}

// DepositHistoryItem represents a single deposit record
type DepositHistoryItem struct {
	DepositId     int64  `json:"deposit_id"`
	AccountIndex  int64  `json:"account_index"`
	L1TxHash      string `json:"l1_tx_hash"` // L1 transaction hash
	L2TxHash      string `json:"l2_tx_hash"` // L2 transaction hash
	Amount        string `json:"amount"`     // Deposit amount in USDC
	Status        string `json:"status"`     // "pending", "confirmed", "failed"
	L1BlockHeight int64  `json:"l1_block_height"`
	L2BlockHeight int64  `json:"l2_block_height"`
	CreatedAt     int64  `json:"created_at"`             // When deposit was initiated
	ConfirmedAt   int64  `json:"confirmed_at,omitempty"` // When deposit was confirmed
	Fee           string `json:"fee,omitempty"`          // Bridge fee
	FromAddress   string `json:"from_address"`           // L1 source address
	ToAddress     string `json:"to_address"`             // L2 destination address
}

// DepositHistoryResponse represents the response for deposit history API
type DepositHistoryResponse struct {
	ResultCode
	Deposits []DepositHistoryItem `json:"deposits"`
}

// TransferHistoryItem represents a single transfer record
type TransferHistoryItem struct {
	TransferId   int64  `json:"transfer_id"`
	FromAccount  int64  `json:"from_account"`
	ToAccount    int64  `json:"to_account"`
	Amount       string `json:"amount"` // Transfer amount
	Fee          string `json:"fee"`    // Transfer fee
	TxHash       string `json:"tx_hash"`
	Status       string `json:"status"` // "pending", "confirmed", "failed"
	BlockHeight  int64  `json:"block_height"`
	Timestamp    int64  `json:"timestamp"`
	TransferType string `json:"transfer_type"`   // "internal", "external"
	Notes        string `json:"notes,omitempty"` // Optional transfer notes
}

// TransferHistoryResponse represents the response for transfer history API
type TransferHistoryResponse struct {
	ResultCode
	Transfers []TransferHistoryItem `json:"transfers"`
}

// WithdrawHistoryItem represents a single withdrawal record
type WithdrawHistoryItem struct {
	WithdrawId    int64  `json:"withdraw_id"`
	AccountIndex  int64  `json:"account_index"`
	L2TxHash      string `json:"l2_tx_hash"` // L2 transaction hash
	L1TxHash      string `json:"l1_tx_hash"` // L1 transaction hash (when completed)
	Amount        string `json:"amount"`     // Withdrawal amount
	Fee           string `json:"fee"`        // Withdrawal fee
	Status        string `json:"status"`     // "pending", "processing", "completed", "failed"
	L2BlockHeight int64  `json:"l2_block_height"`
	L1BlockHeight int64  `json:"l1_block_height,omitempty"`
	RequestedAt   int64  `json:"requested_at"`           // When withdrawal was requested
	ProcessedAt   int64  `json:"processed_at,omitempty"` // When withdrawal was processed
	CompletedAt   int64  `json:"completed_at,omitempty"` // When withdrawal completed on L1
	ToAddress     string `json:"to_address"`             // L1 destination address
	WithdrawDelay int32  `json:"withdraw_delay"`         // Delay in seconds before processing
}

// WithdrawHistoryResponse represents the response for withdraw history API
type WithdrawHistoryResponse struct {
	ResultCode
	Withdrawals []WithdrawHistoryItem `json:"withdrawals"`
}

// TxInfo represents the response payload for a single transaction lookup
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

// TxsResponse represents the response for transactions lookup
type TxsResponse struct {
	ResultCode
	Transactions []Transaction `json:"transactions"`
}

// ============= Phase 4: System Information Types =============

// SystemInfo represents general system information
type SystemInfo struct {
	Version           string                 `json:"version"`
	ChainId           int64                  `json:"chain_id"`
	NetworkName       string                 `json:"network_name"`
	BlockHeight       int64                  `json:"block_height"`
	LastBlockTime     int64                  `json:"last_block_time"`
	TotalAccounts     int64                  `json:"total_accounts"`
	TotalMarkets      int32                  `json:"total_markets"`
	SystemStatus      string                 `json:"system_status"`
	MaintenanceMode   bool                   `json:"maintenance_mode"`
	SupportedFeatures []string               `json:"supported_features"`
	Configuration     map[string]interface{} `json:"configuration,omitempty"`
}

// SystemInfoResponse represents the response for system info API
type SystemInfoResponse struct {
	ResultCode
	Info SystemInfo `json:"info"`
}

// SystemStatus represents current system status
type SystemStatus struct {
	Status               string                `json:"status"` // "operational", "maintenance", "degraded"
	LastUpdate           int64                 `json:"last_update"`
	Message              string                `json:"message,omitempty"`
	ScheduledMaintenance *ScheduledMaintenance `json:"scheduled_maintenance,omitempty"`
}

// ScheduledMaintenance represents scheduled maintenance information
type ScheduledMaintenance struct {
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Description string `json:"description"`
	Impact      string `json:"impact"` // "low", "medium", "high"
}

// SystemStatusResponse represents the response for system status API
type SystemStatusResponse struct {
	ResultCode
	Status SystemStatus `json:"status"`
}

// WithdrawalDelayInfo represents withdrawal delay information
type WithdrawalDelayInfo struct {
	AccountIndex   int64 `json:"account_index"`
	CurrentDelay   int32 `json:"current_delay"`   // Current delay in seconds
	MinDelay       int32 `json:"min_delay"`       // Minimum delay in seconds
	MaxDelay       int32 `json:"max_delay"`       // Maximum delay in seconds
	LastWithdrawal int64 `json:"last_withdrawal"` // Timestamp of last withdrawal
	NextAvailable  int64 `json:"next_available"`  // When next withdrawal is available
}

// WithdrawalDelayResponse represents the response for withdrawal delay API
type WithdrawalDelayResponse struct {
	ResultCode
	DelayInfo WithdrawalDelayInfo `json:"delay_info"`
}

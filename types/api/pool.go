package api

// PublicPoolInfo represents public pool information
type PublicPoolInfo struct {
	PoolIndex        int64  `json:"pool_index"`
	OperatorAccount  int64  `json:"operator_account"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	TotalShares      string `json:"total_shares"`
	SharePrice       string `json:"share_price"`
	TotalValue       string `json:"total_value"`
	AvailableValue   string `json:"available_value"`
	LockedValue      string `json:"locked_value,omitempty"`
	OperatorFeeRate  string `json:"operator_fee_rate"`
	MinShareRate     string `json:"min_share_rate,omitempty"`
	MaxShareRate     string `json:"max_share_rate,omitempty"`
	TotalInvestors   int64  `json:"total_investors"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at,omitempty"`
}

// PublicPoolMetadata represents pool metadata
type PublicPoolMetadata struct {
	PoolIndex       int64    `json:"pool_index"`
	Name            string   `json:"name,omitempty"`
	Description     string   `json:"description,omitempty"`
	Website         string   `json:"website,omitempty"`
	Twitter         string   `json:"twitter,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	PerformanceData *PoolPerformance `json:"performance_data,omitempty"`
}

// PoolPerformance represents pool performance metrics
type PoolPerformance struct {
	Return24h   string `json:"return_24h"`
	Return7d    string `json:"return_7d"`
	Return30d   string `json:"return_30d"`
	ReturnAll   string `json:"return_all"`
	MaxDrawdown string `json:"max_drawdown,omitempty"`
	SharpeRatio string `json:"sharpe_ratio,omitempty"`
}

// RespPublicPoolsMetadata is the response for pool metadata queries
type RespPublicPoolsMetadata struct {
	BaseResponse
	Pools  []PublicPoolMetadata `json:"pools"`
	Cursor Cursor               `json:"cursor,omitempty"`
}

// PublicPoolShare represents a user's share in a pool
type PublicPoolShare struct {
	PoolIndex      int64  `json:"pool_index"`
	AccountIndex   int64  `json:"account_index"`
	Shares         string `json:"shares"`
	ShareValue     string `json:"share_value"`
	EntryValue     string `json:"entry_value"`
	UnrealizedPnl  string `json:"unrealized_pnl"`
	InvestedAt     int64  `json:"invested_at"`
}

// PublicPoolShares is the response for user pool shares query
type PublicPoolShares struct {
	BaseResponse
	Shares []PublicPoolShare `json:"shares"`
}

// PoolPosition represents a pool's position in a market
type PoolPosition struct {
	PoolIndex       int64  `json:"pool_index"`
	MarketIndex     int16  `json:"market_index"`
	Size            string `json:"size"`
	Side            string `json:"side"`
	EntryPrice      string `json:"entry_price"`
	MarkPrice       string `json:"mark_price"`
	UnrealizedPnl   string `json:"unrealized_pnl"`
	Leverage        string `json:"leverage"`
}

// PoolHistory represents pool value history
type PoolHistory struct {
	PoolIndex   int64        `json:"pool_index"`
	History     []PoolPoint  `json:"history"`
}

// PoolPoint represents a point in pool history
type PoolPoint struct {
	Timestamp  int64  `json:"timestamp"`
	TotalValue string `json:"total_value"`
	SharePrice string `json:"share_price"`
	TotalShares string `json:"total_shares"`
}

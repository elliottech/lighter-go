package api

// AccountPnL represents account PnL data
type AccountPnL struct {
	BaseResponse
	AccountIndex int64      `json:"account_index"`
	Resolution   string     `json:"resolution"`
	Entries      []PnLEntry `json:"entries"`
}

// PnLEntry represents a single PnL data point
type PnLEntry struct {
	Timestamp       int64  `json:"timestamp"`
	PortfolioValue  string `json:"portfolio_value"`
	CollateralValue string `json:"collateral_value"`
	PositionValue   string `json:"position_value"`
	UnrealizedPnl   string `json:"unrealized_pnl"`
	RealizedPnl     string `json:"realized_pnl"`
	TotalPnl        string `json:"total_pnl"`
	PnlChange       string `json:"pnl_change,omitempty"`
	PnlChangePct    string `json:"pnl_change_pct,omitempty"`
}

// DailyPnL represents daily PnL summary
type DailyPnL struct {
	Date            string `json:"date"` // YYYY-MM-DD format
	StartValue      string `json:"start_value"`
	EndValue        string `json:"end_value"`
	HighValue       string `json:"high_value"`
	LowValue        string `json:"low_value"`
	RealizedPnl     string `json:"realized_pnl"`
	UnrealizedPnl   string `json:"unrealized_pnl"`
	TotalPnl        string `json:"total_pnl"`
	TotalReturn     string `json:"total_return"`
	TotalReturnPct  string `json:"total_return_pct"`
	Volume          string `json:"volume"`
	TradeCount      int64  `json:"trade_count"`
	Fees            string `json:"fees"`
	FundingPayments string `json:"funding_payments"`
}

// MarketPnL represents PnL for a specific market
type MarketPnL struct {
	MarketIndex    int16  `json:"market_index"`
	MarketSymbol   string `json:"market_symbol,omitempty"`
	RealizedPnl    string `json:"realized_pnl"`
	UnrealizedPnl  string `json:"unrealized_pnl"`
	TotalPnl       string `json:"total_pnl"`
	Volume         string `json:"volume"`
	TradeCount     int64  `json:"trade_count"`
	Fees           string `json:"fees"`
	FundingPayments string `json:"funding_payments"`
}

// PnLSummary represents a summary of PnL
type PnLSummary struct {
	AccountIndex   int64       `json:"account_index"`
	TotalPnl       string      `json:"total_pnl"`
	RealizedPnl    string      `json:"realized_pnl"`
	UnrealizedPnl  string      `json:"unrealized_pnl"`
	TotalVolume    string      `json:"total_volume"`
	TotalTrades    int64       `json:"total_trades"`
	TotalFees      string      `json:"total_fees"`
	WinRate        string      `json:"win_rate,omitempty"`
	ProfitFactor   string      `json:"profit_factor,omitempty"`
	ByMarket       []MarketPnL `json:"by_market,omitempty"`
}

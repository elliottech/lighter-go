package client

import (
	"github.com/elliottech/lighter-go/types/api"
)

// MinimalHTTPClient is the minimal interface for HTTP operations.
// This is maintained for backward compatibility.
type MinimalHTTPClient interface {
	GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error)
	GetApiKey(accountIndex int64, apiKeyIndex uint8) (string, error)
}

// FullHTTPClient extends MinimalHTTPClient with complete API coverage.
// Use this interface for full access to all Lighter API endpoints.
type FullHTTPClient interface {
	MinimalHTTPClient

	// API group accessors - each returns a specialized API client
	Account() AccountAPI
	Order() OrderAPI
	Transaction() TransactionAPI
	Candlestick() CandlestickAPI
	Block() BlockAPI
	Bridge() BridgeAPI
	Info() InfoAPI
}

// AccountAPI provides access to account-related endpoints
type AccountAPI interface {
	// GetAccount retrieves account details
	GetAccount(by api.QueryBy, value string) (*api.DetailedAccounts, error)

	// GetAccountsByL1Address retrieves accounts linked to an L1 address
	GetAccountsByL1Address(l1Address string) (*api.SubAccounts, error)

	// GetAccountMetadata retrieves account metadata
	GetAccountMetadata(by api.QueryBy, value string, auth string) (*api.AccountMetadatas, error)

	// GetAccountLimits retrieves account trading limits
	GetAccountLimits(accountIndex int64, auth string) (*api.AccountLimits, error)

	// GetLiquidations retrieves account liquidation history
	GetLiquidations(accountIndex int64, limit int, auth string, opts *LiquidationOpts) (*api.LiquidationInfos, error)

	// GetPositionFunding retrieves position funding history
	GetPositionFunding(accountIndex int64, limit int, auth string, opts *PositionFundingOpts) (*api.PositionFundings, error)

	// GetPnL retrieves account PnL history
	GetPnL(accountIndex int64, resolution string, timestamps api.TimestampRange, countBack int, auth string, ignoreTransfers bool) (*api.AccountPnL, error)

	// GetPublicPoolsMetadata retrieves public pool metadata
	GetPublicPoolsMetadata(filter string, index int, limit int, auth string, accountIndex *int64) (*api.RespPublicPoolsMetadata, error)

	// ChangeAccountTier changes the account tier (requires auth)
	ChangeAccountTier(accountIndex int64, newTier string, auth string) (*api.RespChangeAccountTier, error)

	// GetL1Metadata retrieves L1 address metadata
	GetL1Metadata(l1Address string, auth string) (*api.L1Metadata, error)

	// GetApiKeys retrieves API keys for an account
	GetApiKeys(accountIndex int64, apiKeyIndex *uint8) (*api.AccountApiKeys, error)
}

// LiquidationOpts contains options for liquidation queries
type LiquidationOpts struct {
	MarketID *int16
	Cursor   string
}

// PositionFundingOpts contains options for position funding queries
type PositionFundingOpts struct {
	MarketID *int16
	Cursor   string
	Side     string
}

// OrderAPI provides access to order-related endpoints
type OrderAPI interface {
	// GetActiveOrders retrieves active orders for an account
	GetActiveOrders(accountIndex int64, marketID *int16, auth string) (*api.Orders, error)

	// GetInactiveOrders retrieves order history
	GetInactiveOrders(accountIndex int64, marketID *int16, opts *InactiveOrdersOpts) (*api.Orders, error)

	// GetOrderBooks retrieves order book snapshots
	GetOrderBooks(marketID *int16, filter api.MarketFilter) (*api.OrderBooks, error)

	// GetOrderBookDetails retrieves detailed order book data
	GetOrderBookDetails(marketID int16, filter api.MarketFilter) (*api.OrderBookDetails, error)

	// GetOrderBookOrders retrieves individual orders in the order book
	GetOrderBookOrders(marketID int16, limit int) (*api.OrderBookOrders, error)

	// GetRecentTrades retrieves recent trades for a market
	GetRecentTrades(marketID int16, limit int) (*api.Trades, error)

	// GetTrades retrieves trade history
	GetTrades(marketID int16, accountIndex *int64, opts *TradesOpts) (*api.Trades, error)

	// GetAssetDetails retrieves asset information
	GetAssetDetails(assetID *int16) (*api.AssetDetails, error)

	// GetExchangeStats retrieves exchange-wide statistics
	GetExchangeStats() (*api.ExchangeStats, error)
}

// InactiveOrdersOpts contains options for inactive order queries
type InactiveOrdersOpts struct {
	Status   api.OrderStatusFilter
	Limit    int
	Cursor   string
	SortBy   string
	SortOrder string
}

// TradesOpts contains options for trade queries
type TradesOpts struct {
	Limit     int
	Cursor    string
	SortBy    string
	SortOrder string
}

// TransactionAPI provides access to transaction-related endpoints
type TransactionAPI interface {
	// SendTx submits a signed transaction
	SendTx(txType uint8, txInfo string, priceProtection *api.PriceProtection) (*api.RespSendTx, error)

	// SendTxBatch submits multiple transactions
	SendTxBatch(txTypes []uint8, txInfos []string) (*api.RespSendTxBatch, error)

	// GetTx retrieves a transaction by hash or sequence index
	GetTx(by api.QueryBy, value string) (*api.EnrichedTx, error)

	// GetTxs retrieves transactions with pagination
	GetTxs(index *int64, limit int) (*api.Txs, error)

	// GetAccountTxs retrieves transactions for an account
	GetAccountTxs(by api.QueryBy, value string, limit int, types []api.TxType) (*api.Txs, error)

	// GetTxFromL1TxHash retrieves a transaction by its L1 hash
	GetTxFromL1TxHash(hash string) (*api.EnrichedTx, error)

	// GetDepositHistory retrieves deposit history
	GetDepositHistory(accountIndex int64, l1Address string, filter string, cursor string) (*api.DepositHistory, error)

	// GetWithdrawHistory retrieves withdrawal history
	GetWithdrawHistory(accountIndex int64, filter string, cursor string) (*api.WithdrawHistory, error)

	// GetTransferHistory retrieves transfer history
	GetTransferHistory(accountIndex int64, cursor string) (*api.TransferHistory, error)

	// GetTransferFeeInfo retrieves transfer fee information
	GetTransferFeeInfo(accountIndex int64, toAccountIndex *int64) (*api.TransferFeeInfo, error)

	// GetWithdrawalDelay retrieves current withdrawal delay
	GetWithdrawalDelay() (*api.RespWithdrawalDelay, error)
}

// CandlestickAPI provides access to market data endpoints
type CandlestickAPI interface {
	// GetCandlesticks retrieves OHLCV data
	GetCandlesticks(marketID int16, resolution api.CandlestickResolution, timestamps api.TimestampRange, countBack int) (*api.Candlesticks, error)

	// GetFundings retrieves funding data
	GetFundings(marketID int16, resolution api.FundingResolution, timestamps api.TimestampRange, countBack int) (*api.Fundings, error)

	// GetFundingRates retrieves current funding rates
	GetFundingRates() (*api.FundingRates, error)
}

// BlockAPI provides access to block-related endpoints
type BlockAPI interface {
	// GetBlock retrieves a block by height or commitment
	GetBlock(by api.QueryBy, value string) (*api.Blocks, error)

	// GetBlocks retrieves blocks with pagination
	GetBlocks(index *int64, limit int, sort string) (*api.Blocks, error)

	// GetBlockTxs retrieves transactions for a block
	GetBlockTxs(by api.QueryBy, value string) (*api.Txs, error)

	// GetCurrentHeight retrieves the current block height
	GetCurrentHeight() (*api.CurrentHeight, error)
}

// BridgeAPI provides access to bridge-related endpoints
type BridgeAPI interface {
	// GetBridges retrieves bridge transactions for an L1 address
	GetBridges(l1Address string) (*api.RespGetBridgesByL1Addr, error)

	// GetIsNextBridgeFast checks if the next bridge will be fast
	GetIsNextBridgeFast(l1Address string) (*api.RespGetIsNextBridgeFast, error)

	// GetFastBridgeInfo retrieves fast bridge information
	GetFastBridgeInfo() (*api.RespGetFastBridgeInfo, error)

	// CreateIntentAddress creates an intent address for external deposits
	CreateIntentAddress(chainID int64, fromAddr string, amount string, isExternalDeposit bool) (*api.RespCreateIntentAddress, error)
}

// InfoAPI provides access to general information endpoints
type InfoAPI interface {
	// GetStatus retrieves service status
	GetStatus() (*api.Status, error)

	// GetInfo retrieves exchange information
	GetInfo() (*api.ZkLighterInfo, error)

	// GetAnnouncements retrieves exchange announcements
	GetAnnouncements() (*api.Announcements, error)

	// Export exports account data
	Export(accountIndex int64, marketID int16, exportType api.ExportType) (*api.ExportData, error)
}

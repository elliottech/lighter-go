package client

import (
	"fmt"
	"time"

	"github.com/elliottech/lighter-go/nonce"
	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/api"
	"github.com/elliottech/lighter-go/types/txtypes"
)

// SignerClient extends TxClient with convenience methods for common trading patterns.
// It provides higher-level APIs similar to the Python SDK's SignerClient.
type SignerClient struct {
	*TxClient
	fullHTTP     FullHTTPClient
	nonceManager nonce.Manager
}

// NewSignerClient creates a SignerClient with full HTTP capabilities.
// If nonceManager is nil, a new OptimisticNonceManager will be created.
func NewSignerClient(httpClient FullHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64, nonceManager nonce.Manager) (*SignerClient, error) {
	txClient, err := createTxClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
	if err != nil {
		return nil, err
	}

	if nonceManager == nil {
		nonceManager = nonce.NewOptimisticManager(httpClient)
	}

	return &SignerClient{
		TxClient:     txClient,
		fullHTTP:     httpClient,
		nonceManager: nonceManager,
	}, nil
}

// createTxClient is a helper that creates a TxClient without registering it globally
func createTxClient(httpClient MinimalHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64) (*TxClient, error) {
	// Use the existing CreateClient function but we need to get the client back
	return CreateClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
}

// FullHTTP returns the full HTTP client for direct API access
func (c *SignerClient) FullHTTP() FullHTTPClient {
	return c.fullHTTP
}

// NonceManager returns the nonce manager
func (c *SignerClient) NonceManager() nonce.Manager {
	return c.nonceManager
}

// CreateMarketOrder creates a market order with minimal parameters
func (c *SignerClient) CreateMarketOrder(marketIndex int16, size int64, isBuy bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For market orders, use max/min price depending on side
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0, // Auto-generate
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.MarketOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateMarketOrderWithSlippage creates a market order with slippage protection
func (c *SignerClient) CreateMarketOrderWithSlippage(marketIndex int16, size int64, isBuy bool, slippageBps int, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	// Fetch current market price
	orderBooks, err := c.fullHTTP.Order().GetOrderBooks(&marketIndex, api.MarketFilterAll)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	if len(orderBooks.OrderBooks) == 0 {
		return nil, fmt.Errorf("no order book data for market %d", marketIndex)
	}

	ob := orderBooks.OrderBooks[0]
	var referencePrice string
	if isBuy && len(ob.Asks) > 0 {
		referencePrice = ob.Asks[0].Price
	} else if !isBuy && len(ob.Bids) > 0 {
		referencePrice = ob.Bids[0].Price
	} else {
		return nil, fmt.Errorf("no liquidity in order book")
	}

	price, err := calculateSlippagePrice(referencePrice, slippageBps, isBuy)
	if err != nil {
		return nil, err
	}

	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder, // Use limit with IOC for slippage protection
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateLimitOrder creates a limit order
func (c *SignerClient) CreateLimitOrder(marketIndex int16, size int64, price uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateTakeProfitOrder creates a take-profit market order
func (c *SignerClient) CreateTakeProfitOrder(marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For TP market orders, use limit price that will fill immediately once triggered
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateTakeProfitLimitOrder creates a take-profit limit order
func (c *SignerClient) CreateTakeProfitLimitOrder(marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateStopLossOrder creates a stop-loss market order
func (c *SignerClient) CreateStopLossOrder(marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateStopLossLimitOrder creates a stop-loss limit order
func (c *SignerClient) CreateStopLossLimitOrder(marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// SendAndSubmit signs a transaction and submits it to the API
func (c *SignerClient) SendAndSubmit(txInfo txtypes.TxInfo) (*api.RespSendTx, error) {
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}

	resp, err := c.fullHTTP.Transaction().SendTx(txInfo.GetTxType(), txInfoJSON, nil)
	if err != nil {
		// Acknowledge failure for nonce recovery
		if optManager, ok := c.nonceManager.(*nonce.OptimisticManager); ok {
			// Extract nonce from tx info if possible
			optManager.AcknowledgeFailure(c.GetAccountIndex(), c.GetApiKeyIndex(), -1)
		}
		return nil, err
	}

	return resp, nil
}

// SendTxBatch submits multiple transactions
func (c *SignerClient) SendTxBatch(txInfos []txtypes.TxInfo) (*api.RespSendTxBatch, error) {
	txTypes := make([]uint8, len(txInfos))
	txInfoJSONs := make([]string, len(txInfos))

	for i, tx := range txInfos {
		txTypes[i] = tx.GetTxType()
		jsonStr, err := tx.GetTxInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize tx %d: %w", i, err)
		}
		txInfoJSONs[i] = jsonStr
	}

	return c.fullHTTP.Transaction().SendTxBatch(txTypes, txInfoJSONs)
}

// GetOpenOrders retrieves open orders for the account
func (c *SignerClient) GetOpenOrders(marketID *int16) (*api.Orders, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, err
	}
	return c.fullHTTP.Order().GetActiveOrders(c.GetAccountIndex(), marketID, authToken)
}

// CancelAllOrders cancels all open orders
func (c *SignerClient) CancelAllOrders(opts *types.TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	req := &types.CancelAllOrdersTxReq{
		TimeInForce: txtypes.ImmediateCancelAll,
		Time:        0,
	}
	return c.GetCancelAllOrdersTransaction(req, opts)
}

// GetPositions retrieves current positions
func (c *SignerClient) GetPositions() (*api.DetailedAccounts, error) {
	return c.fullHTTP.Account().GetAccount(api.QueryByIndex, fmt.Sprintf("%d", c.GetAccountIndex()))
}

// Helper to get auth token
func (c *SignerClient) getAuthToken() (string, error) {
	deadline := time.Now().Add(8 * time.Hour)
	authInfo, err := c.GetAuthToken(deadline)
	if err != nil {
		return "", err
	}

	// The auth token is returned as a string directly
	return authInfo, nil
}

// calculateSlippagePrice calculates price with slippage
func calculateSlippagePrice(priceStr string, slippageBps int, isBuy bool) (uint32, error) {
	// Parse price as integer (prices are typically scaled integers)
	var price int64
	if _, err := fmt.Sscanf(priceStr, "%d", &price); err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	// Calculate slippage adjustment
	adjustment := (price * int64(slippageBps)) / 10000

	if isBuy {
		price += adjustment
		if price > int64(txtypes.MaxOrderPrice) {
			price = int64(txtypes.MaxOrderPrice)
		}
	} else {
		price -= adjustment
		if price < int64(txtypes.MinOrderPrice) {
			price = int64(txtypes.MinOrderPrice)
		}
	}

	return uint32(price), nil
}

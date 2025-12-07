package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/elliottech/lighter-go/types/txtypes"
)

func (c *HTTPClient) parseResultStatus(respBody []byte) error {
	resultStatus := &ResultCode{}
	if err := json.Unmarshal(respBody, resultStatus); err != nil {
		return err
	}
	if resultStatus.Code != CodeOK {
		return errors.New(resultStatus.Message)
	}
	return nil
}

func (c *HTTPClient) getAndParseL2HTTPResponse(path string, params map[string]any, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = path

	q := u.Query()
	for k, v := range params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()
	resp, err := httpClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	// log.Println("Response: of ", u.String(), " is ", string(body))
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}
	if err = c.parseResultStatus(body); err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}
	return nil
}

func (c *HTTPClient) GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error) {
	result := &NextNonce{}
	err := c.getAndParseL2HTTPResponse("api/v1/nextNonce", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return -1, err
	}
	return result.Nonce, nil
}

func (c *HTTPClient) GetApiKey(accountIndex int64, apiKeyIndex uint8) (*AccountApiKeys, error) {
	result := &AccountApiKeys{}
	err := c.getAndParseL2HTTPResponse("api/v1/apikeys", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) SendRawTx(tx txtypes.TxInfo) (*TxHash, error) {
	txType := tx.GetTxType()
	txInfo, err := tx.GetTxInfo()
	if err != nil {
		return nil, err
	}

	data := url.Values{"tx_type": {strconv.Itoa(int(txType))}, "tx_info": {txInfo}}

	if c.fatFingerProtection == false {
		data.Add("price_protection", "false")
	}

	req, _ := http.NewRequest("POST", c.endpoint+"/api/v1/sendTx", strings.NewReader(data.Encode()))
	req.Header.Set("Channel-Name", c.channelName)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	if err = c.parseResultStatus(body); err != nil {
		return nil, err
	}
	res := &TxHash{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

// SendTxBatch sends multiple transactions in a batch using /api/v1/sendTxBatch endpoint
func (c *HTTPClient) SendTxBatch(txTypes []int, txInfos []string) (*TxHashBatch, error) {
	// Convert slices to JSON strings as required by the API
	txTypesJson, err := json.Marshal(txTypes)
	if err != nil {
		return nil, err
	}

	txInfosJson, err := json.Marshal(txInfos)
	if err != nil {
		return nil, err
	}

	data := url.Values{
		"tx_types": {string(txTypesJson)},
		"tx_infos": {string(txInfosJson)},
	}

	req, _ := http.NewRequest("POST", c.endpoint+"/api/v1/sendTxBatch", strings.NewReader(data.Encode()))
	req.Header.Set("Channel-Name", c.channelName)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	if err = c.parseResultStatus(body); err != nil {
		return nil, err
	}

	res := &TxHashBatch{}
	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *HTTPClient) GetTransferFeeInfo(accountIndex, toAccountIndex int64, auth string) (*TransferFeeInfo, error) {
	result := &TransferFeeInfo{}
	err := c.getAndParseL2HTTPResponse("api/v1/transferFeeInfo", map[string]any{
		"account_index":    accountIndex,
		"to_account_index": toAccountIndex,
		"auth":             auth,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetAccount(accountIndex int64) (*AccountResponse, error) {
	result := &AccountResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/account", map[string]any{
		"by":    "index",
		"value": fmt.Sprintf("%d", accountIndex),
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetAccountByL1Address(l1Address string) (*AccountByL1AddressResponse, error) {
	result := &AccountByL1AddressResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/accountsByL1Address", map[string]any{
		"l1_address": l1Address,
	}, result)
	if err != nil {
		return nil, err
	} else {
		log.Println("Detailed Account: ", result)
	}
	return result, nil
}

func (c *HTTPClient) GetOrderBooks() (*OrderBookResponse, error) {
	result := &OrderBookResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/orderBooks", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetOrderBookDetails(marketId uint16) (*OrderBookDetailsResponse, error) {
	result := &OrderBookDetailsResponse{}
	params := map[string]any{}
	if marketId > 0 {
		params["market_id"] = marketId
	}
	err := c.getAndParseL2HTTPResponse("api/v1/orderBookDetails", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetActiveOrders(accountIndex int64, marketId uint16, auth string) (*OrdersResponse, error) {
	result := &OrdersResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/accountActiveOrders", map[string]any{
		"account_index": accountIndex,
		"market_id":     marketId,
		"auth":          auth,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetInactiveOrders(accountIndex int64, marketId uint16, auth string) (*OrdersResponse, error) {
	return c.GetInactiveOrdersWithLimit(accountIndex, marketId, auth, 50) // Default limit of 50
}

func (c *HTTPClient) GetInactiveOrdersWithLimit(accountIndex int64, marketId uint16, auth string, limit int32) (*OrdersResponse, error) {
	result := &OrdersResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
		"limit":         limit,
	}
	if marketId != 255 { // 255 means all markets
		params["market_id"] = marketId
	}
	err := c.getAndParseL2HTTPResponse("api/v1/accountInactiveOrders", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) GetFundingRates() (*FundingRatesResponse, error) {
	result := &FundingRatesResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/funding-rates", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ============= Phase 1: Core Data Query Methods =============

// GetCandlesticks retrieves candlestick data for a market
func (c *HTTPClient) GetCandlesticks(marketId uint16, resolution string, startTimestamp, endTimestamp int64, countBack int32, setTimestampToEnd *bool) (*CandlesticksResponse, error) {
	result := &CandlesticksResponse{}
	params := map[string]any{
		"market_id":       marketId,
		"resolution":      resolution,
		"start_timestamp": startTimestamp,
		"end_timestamp":   endTimestamp,
		"count_back":      countBack,
	}
	if setTimestampToEnd != nil {
		params["set_timestamp_to_end"] = *setTimestampToEnd
	}
	err := c.getAndParseL2HTTPResponse("api/v1/candlesticks", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetFundings retrieves funding history data for a market
func (c *HTTPClient) GetFundings(marketId uint16, startTimestamp, endTimestamp *int64, limit *int32) (*FundingsResponse, error) {
	result := &FundingsResponse{}
	params := map[string]any{
		"market_id": marketId,
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/fundings", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetRecentTrades retrieves recent trades for a market
func (c *HTTPClient) GetRecentTrades(marketId uint16, limit *int32) (*RecentTradesResponse, error) {
	result := &RecentTradesResponse{}
	params := map[string]any{
		"market_id": marketId,
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/recentTrades", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTrades retrieves trade history for a market or account
func (c *HTTPClient) GetTrades(marketId *uint16, accountIndex *int64, startTimestamp, endTimestamp *int64, limit *int32, auth *string) (*TradesResponse, error) {
	result := &TradesResponse{}
	params := map[string]any{}

	if marketId != nil {
		params["market_id"] = *marketId
	}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	if auth != nil {
		params["auth"] = *auth
	}

	err := c.getAndParseL2HTTPResponse("api/v1/trades", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetExchangeStats retrieves exchange-wide statistics
func (c *HTTPClient) GetExchangeStats() (*ExchangeStatsResponse, error) {
	result := &ExchangeStatsResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/exchangeStats", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetExport retrieves data export information
func (c *HTTPClient) GetExport(accountIndex int64, dataType string, startDate, endDate string, auth string) (*ExportResponse, error) {
	result := &ExportResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"data_type":     dataType,
		"start_date":    startDate,
		"end_date":      endDate,
		"auth":          auth,
	}
	err := c.getAndParseL2HTTPResponse("api/v1/export", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ============= Phase 2: Account Management Methods =============

// GetAccountLimits retrieves account trading limits and restrictions
func (c *HTTPClient) GetAccountLimits(accountIndex int64, auth string) (*AccountLimitsResponse, error) {
	result := &AccountLimitsResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	err := c.getAndParseL2HTTPResponse("api/v1/accountLimits", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAccountMetadata retrieves extended account information
func (c *HTTPClient) GetAccountMetadata(accountIndex int64, auth string) (*AccountMetadataResponse, error) {
	result := &AccountMetadataResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	err := c.getAndParseL2HTTPResponse("api/v1/accountMetadata", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetLiquidations retrieves liquidation history for an account
func (c *HTTPClient) GetLiquidations(accountIndex int64, marketId *uint16, startTimestamp, endTimestamp *int64, limit *int32, auth string) (*LiquidationsResponse, error) {
	result := &LiquidationsResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if marketId != nil {
		params["market_id"] = *marketId
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/liquidations", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPnL retrieves profit/loss history for an account
func (c *HTTPClient) GetPnL(accountIndex int64, marketId *uint16, startDate, endDate *string, auth string) (*PnLResponse, error) {
	result := &PnLResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if marketId != nil {
		params["market_id"] = *marketId
	}
	if startDate != nil {
		params["start_date"] = *startDate
	}
	if endDate != nil {
		params["end_date"] = *endDate
	}
	err := c.getAndParseL2HTTPResponse("api/v1/pnl", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPositionFunding retrieves funding fee history for positions
func (c *HTTPClient) GetPositionFunding(accountIndex int64, marketId *uint16, startTimestamp, endTimestamp *int64, limit *int32, auth string) (*PositionFundingResponse, error) {
	result := &PositionFundingResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if marketId != nil {
		params["market_id"] = *marketId
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/positionFunding", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPublicPools retrieves information about public liquidity pools
func (c *HTTPClient) GetPublicPools(accountIndex *int64, auth *string) (*PublicPoolsResponse, error) {
	result := &PublicPoolsResponse{}
	params := map[string]any{}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}
	if auth != nil {
		params["auth"] = *auth
	}
	err := c.getAndParseL2HTTPResponse("api/v1/publicPools", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPublicPoolsMetadata retrieves metadata for public liquidity pools
func (c *HTTPClient) GetPublicPoolsMetadata(poolIndex *int64) (*PublicPoolsMetadataResponse, error) {
	result := &PublicPoolsMetadataResponse{}
	params := map[string]any{}
	if poolIndex != nil {
		params["pool_index"] = *poolIndex
	}
	err := c.getAndParseL2HTTPResponse("api/v1/publicPoolsMetadata", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ChangeAccountTier changes the tier level of an account
func (c *HTTPClient) ChangeAccountTier(accountIndex int64, newTier int32, auth string) (*ChangeAccountTierResponse, error) {
	result := &ChangeAccountTierResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"new_tier":      newTier,
		"auth":          auth,
	}
	err := c.getAndParseL2HTTPResponse("api/v1/changeAccountTier", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ============= Phase 3: Transaction History Methods =============

// GetAccountTxs retrieves transaction history for an account
func (c *HTTPClient) GetAccountTxs(accountIndex int64, startTimestamp, endTimestamp *int64, limit *int32, txType *int32, auth string) (*AccountTxsResponse, error) {
	result := &AccountTxsResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	if txType != nil {
		params["tx_type"] = *txType
	}
	err := c.getAndParseL2HTTPResponse("api/v1/accountTxs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetBlockTxs retrieves all transactions in a specific block
func (c *HTTPClient) GetBlockTxs(blockHeight int64, limit *int32) (*BlockTxsResponse, error) {
	result := &BlockTxsResponse{}
	params := map[string]any{
		"block_height": blockHeight,
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/blockTxs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetDepositHistory retrieves deposit history for an account
func (c *HTTPClient) GetDepositHistory(accountIndex int64, startTimestamp, endTimestamp *int64, limit *int32, auth string) (*DepositHistoryResponse, error) {
	result := &DepositHistoryResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/deposit/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTransferHistory retrieves transfer history for an account
func (c *HTTPClient) GetTransferHistory(accountIndex int64, startTimestamp, endTimestamp *int64, limit *int32, auth string) (*TransferHistoryResponse, error) {
	result := &TransferHistoryResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/transfer/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetWithdrawHistory retrieves withdrawal history for an account
func (c *HTTPClient) GetWithdrawHistory(accountIndex int64, startTimestamp, endTimestamp *int64, limit *int32, auth string) (*WithdrawHistoryResponse, error) {
	result := &WithdrawHistoryResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	err := c.getAndParseL2HTTPResponse("api/v1/withdraw/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTx retrieves information about a specific transaction by hash
func (c *HTTPClient) GetTx(txHash string) (*TxInfo, error) {
	result := &TxInfo{}
	params := map[string]any{
		"by":    "hash", // Required field: specify lookup method
		"value": txHash, // Transaction hash value
	}
	err := c.getAndParseL2HTTPResponse("api/v1/tx", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTxBySequenceIndex retrieves information about a specific transaction by sequence index
func (c *HTTPClient) GetTxBySequenceIndex(sequenceIndex int64) (*TxInfo, error) {
	result := &TxInfo{}
	params := map[string]any{
		"by":    "sequence_index",                     // Required field: specify lookup method
		"value": strconv.FormatInt(sequenceIndex, 10), // Sequence index as string
	}
	err := c.getAndParseL2HTTPResponse("api/v1/tx", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTxs retrieves multiple transactions based on criteria
func (c *HTTPClient) GetTxs(startTimestamp, endTimestamp *int64, limit *int32, txType *int32, accountIndex *int64) (*TxsResponse, error) {
	result := &TxsResponse{}
	params := map[string]any{}

	if startTimestamp != nil {
		params["start_timestamp"] = *startTimestamp
	}
	if endTimestamp != nil {
		params["end_timestamp"] = *endTimestamp
	}
	if limit != nil {
		params["limit"] = *limit
	}
	if txType != nil {
		params["tx_type"] = *txType
	}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}

	err := c.getAndParseL2HTTPResponse("api/v1/txs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ============= Phase 4: System Information Methods =============

// GetSystemInfo retrieves general system information
func (c *HTTPClient) GetSystemInfo() (*SystemInfoResponse, error) {
	result := &SystemInfoResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/info", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetSystemStatus retrieves current system status
func (c *HTTPClient) GetSystemStatus() (*SystemStatusResponse, error) {
	result := &SystemStatusResponse{}
	err := c.getAndParseL2HTTPResponse("api/v1/status", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetWithdrawalDelay retrieves withdrawal delay information for an account
func (c *HTTPClient) GetWithdrawalDelay(accountIndex int64, auth string) (*WithdrawalDelayResponse, error) {
	result := &WithdrawalDelayResponse{}
	params := map[string]any{
		"account_index": accountIndex,
		"auth":          auth,
	}
	err := c.getAndParseL2HTTPResponse("api/v1/withdrawalDelay", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

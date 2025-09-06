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

func (c *HTTPClient) SendRawTx(tx txtypes.TxInfo) (string, error) {
	txType := tx.GetTxType()
	txInfo, err := tx.GetTxInfo()
	if err != nil {
		return "", err
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
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}
	if err = c.parseResultStatus(body); err != nil {
		return "", err
	}
	res := &TxHash{}
	if err := json.Unmarshal(body, res); err != nil {
		return "", err
	}

	return res.TxHash, nil
}

// SendTxBatch sends multiple transactions in a batch using /api/v1/sendTxBatch endpoint
func (c *HTTPClient) SendTxBatch(txTypes []int, txInfos []string) ([]string, error) {
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

	return res.TxHash, nil
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

func (c *HTTPClient) GetOrderBookDetails(marketId uint8) (*OrderBookDetailsResponse, error) {
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

func (c *HTTPClient) GetActiveOrders(accountIndex int64, marketId uint8, auth string) (*OrdersResponse, error) {
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

func (c *HTTPClient) GetInactiveOrders(accountIndex int64, marketId uint8, auth string) (*OrdersResponse, error) {
	return c.GetInactiveOrdersWithLimit(accountIndex, marketId, auth, 50) // Default limit of 50
}

func (c *HTTPClient) GetInactiveOrdersWithLimit(accountIndex int64, marketId uint8, auth string, limit int32) (*OrdersResponse, error) {
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

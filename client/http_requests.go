package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// AccountsByL1Address queries account information by L1 address
// Docs: https://apidocs.lighter.xyz/reference/accountsbyl1address
// GET https://mainnet.zklighter.elliot.ai/api/v1/accountsByL1Address, query params: l1_address string required
func (c *HTTPClient) AccountsByL1Address(l1Address string) (*AccountByL1Address, error) {
	result := &AccountByL1Address{}
	err := c.getAndParseL2HTTPResponse("api/v1/accountsByL1Address", map[string]any{"l1_address": l1Address}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetApiKey Get account api key. Set api_key_index to 255 to retrieve all api keys associated with the account.
// Docs: https://apidocs.lighter.xyz/reference/apikeys
// GET https://mainnet.zklighter.elliot.ai/api/v1/apikeys
func (c *HTTPClient) GetApiKey(accountIndex int64, apiKeyIndex uint8) (*AccountApiKeys, error) {
	result := &AccountApiKeys{}
	err := c.getAndParseL2HTTPResponse("api/v1/apikeys", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SendRawTx sends a raw transaction to the network
// Docs: https://apidocs.lighter.xyz/reference/sendtx
// POST https://mainnet.zklighter.elliot.ai/api/v1/sendTx
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

// GetTx Get transaction by hash or sequence index. Only one of the parameters `txHash` or `sequenceIndex` will be used. If both are provided, `txHash` takes precedence.
// Docs: https://apidocs.lighter.xyz/reference/tx
// GET https://mainnet.zklighter.elliot.ai/api/v1/tx
func (c *HTTPClient) GetTx(txHash, sequenceIndex string) (*TxInfo, error) {
	result := &TxInfo{}
	params := map[string]any{}
	if txHash != "" {
		params["by"] = "hash"
		params["value"] = txHash
	} else if sequenceIndex != "" {
		params["by"] = "sequence_index"
		params["value"] = sequenceIndex
	} else {
		return nil, fmt.Errorf("either txHash or sequenceIndex must be provided")
	}
	err := c.getAndParseL2HTTPResponse("api/v1/tx", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
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

// OrderBookDetails Get data about a specific market’s orderbook
// Docs: https://apidocs.lighter.xyz/reference/orderbookdetails
// GET https://mainnet.zklighter.elliot.ai/api/v1/orderBookDetails
func (c *HTTPClient) OrderBookDetails(marketId uint8) (*OrderBookDetails, error) {
	result := &OrderBookDetails{}
	err := c.getAndParseL2HTTPResponse("api/v1/orderBookDetails", map[string]any{
		"market_id": marketId,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// OrderBookOrders Get order book orders
// Docs: https://apidocs.lighter.xyz/reference/orderbookorders
// GET https://mainnet.zklighter.elliot.ai/api/v1/orderBookOrders
func (c *HTTPClient) OrderBookOrders(marketId uint8, limit int64) (*OrderBookOrders, error) {
	result := &OrderBookOrders{}
	err := c.getAndParseL2HTTPResponse("api/v1/orderBookOrders", map[string]any{
		"market_id": marketId,
		"limit":     limit,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

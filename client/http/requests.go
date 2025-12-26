package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bytedance/sonic"
)

func (c *client) parseResultStatus(respBody []byte) error {
	resultStatus := &ResultCode{}
	if err := sonic.Unmarshal(respBody, resultStatus); err != nil {
		return err
	}
	if resultStatus.Code != CodeOK {
		return errors.New(resultStatus.Message)
	}
	return nil
}

func (c *client) getAndParseL2HTTPResponse(path string, params map[string]any, result interface{}) error {
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
	defer resp.Body.Close() //nolint:errcheck // Response body close errors are non-actionable
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
	if err := sonic.Unmarshal(body, result); err != nil {
		return err
	}
	return nil
}

func (c *client) GetNextNonce(accountIndex int64, apiKeyIndex uint8) (int64, error) {
	result := &NextNonce{}
	err := c.getAndParseL2HTTPResponse("api/v1/nextNonce", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return -1, err
	}
	return result.Nonce, nil
}

func (c *client) GetApiKey(accountIndex int64, apiKeyIndex uint8) (string, error) {
	result := &AccountApiKeys{}
	err := c.getAndParseL2HTTPResponse("api/v1/apikeys", map[string]any{"account_index": accountIndex, "api_key_index": apiKeyIndex}, result)
	if err != nil {
		return "", err
	}
	if len(result.ApiKeys) == 0 {
		return "", fmt.Errorf("no api keys returned")
	}
	return result.ApiKeys[0].PublicKey, nil
}

// postAndParseL2HTTPResponse sends a POST request with JSON body and parses the response
func (c *client) postAndParseL2HTTPResponse(path string, body interface{}, result interface{}) error {
	u, err := url.Parse(c.endpoint)
	if err != nil {
		return err
	}
	u.Path = path

	jsonBody, err := sonic.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return &ConnectionError{Err: err}
	}
	defer resp.Body.Close() //nolint:errcheck // Response body close errors are non-actionable

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return NewAPIErrorWithStatus(int32(resp.StatusCode), string(respBody), resp.StatusCode)
	}

	if err = c.parseResultStatus(respBody); err != nil {
		return err
	}

	if err := sonic.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

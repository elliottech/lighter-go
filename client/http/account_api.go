package http

import (
	"fmt"

	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type accountAPIImpl struct {
	client *client
}

// Ensure accountAPIImpl implements AccountAPI
var _ core.AccountAPI = (*accountAPIImpl)(nil)

func (a *accountAPIImpl) GetAccount(by api.QueryBy, value string) (*api.DetailedAccounts, error) {
	result := &api.DetailedAccounts{}
	err := a.client.getAndParseL2HTTPResponse("api/v1/account", map[string]any{
		"by":    string(by),
		"value": value,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetAccountsByL1Address(l1Address string) (*api.SubAccounts, error) {
	result := &api.SubAccounts{}
	err := a.client.getAndParseL2HTTPResponse("api/v1/accountsByL1Address", map[string]any{
		"l1_address": l1Address,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetAccountMetadata(by api.QueryBy, value string, auth string) (*api.AccountMetadatas, error) {
	result := &api.AccountMetadatas{}
	params := map[string]any{
		"by":    string(by),
		"value": value,
	}
	if auth != "" {
		params["auth"] = auth
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/accountMetadata", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetAccountLimits(accountIndex int64, auth string) (*api.AccountLimits, error) {
	result := &api.AccountLimits{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if auth != "" {
		params["auth"] = auth
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/accountLimits", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetLiquidations(accountIndex int64, limit int, auth string, opts *core.LiquidationOpts) (*api.LiquidationInfos, error) {
	result := &api.LiquidationInfos{}
	params := map[string]any{
		"account_index": accountIndex,
		"limit":         limit,
	}
	if auth != "" {
		params["auth"] = auth
	}
	if opts != nil {
		if opts.MarketID != nil {
			params["market_id"] = *opts.MarketID
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/liquidations", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetPositionFunding(accountIndex int64, limit int, auth string, opts *core.PositionFundingOpts) (*api.PositionFundings, error) {
	result := &api.PositionFundings{}
	params := map[string]any{
		"account_index": accountIndex,
		"limit":         limit,
	}
	if auth != "" {
		params["auth"] = auth
	}
	if opts != nil {
		if opts.MarketID != nil {
			params["market_id"] = *opts.MarketID
		}
		if opts.Cursor != "" {
			params["cursor"] = opts.Cursor
		}
		if opts.Side != "" {
			params["side"] = opts.Side
		}
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/positionFunding", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetPnL(accountIndex int64, resolution string, timestamps api.TimestampRange, countBack int, auth string, ignoreTransfers bool) (*api.AccountPnL, error) {
	result := &api.AccountPnL{}
	params := map[string]any{
		"by":              "index",
		"value":           fmt.Sprintf("%d", accountIndex),
		"resolution":      resolution,
		"start_timestamp": timestamps.StartTimestamp,
		"end_timestamp":   timestamps.EndTimestamp,
		"count_back":      countBack,
	}
	if auth != "" {
		params["auth"] = auth
	}
	if ignoreTransfers {
		params["ignore_transfers"] = true
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/pnl", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetPublicPoolsMetadata(filter string, index int, limit int, auth string, accountIndex *int64) (*api.RespPublicPoolsMetadata, error) {
	result := &api.RespPublicPoolsMetadata{}
	params := map[string]any{
		"filter": filter,
		"index":  index,
		"limit":  limit,
	}
	if auth != "" {
		params["auth"] = auth
	}
	if accountIndex != nil {
		params["account_index"] = *accountIndex
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/publicPoolsMetadata", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) ChangeAccountTier(accountIndex int64, newTier string, auth string) (*api.RespChangeAccountTier, error) {
	result := &api.RespChangeAccountTier{}
	body := map[string]any{
		"account_index": accountIndex,
		"new_tier":      newTier,
		"auth":          auth,
	}
	err := a.client.postAndParseL2HTTPResponse("api/v1/changeAccountTier", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetL1Metadata(l1Address string, auth string) (*api.L1Metadata, error) {
	result := &api.L1Metadata{}
	params := map[string]any{
		"l1_address": l1Address,
	}
	if auth != "" {
		params["auth"] = auth
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/l1Metadata", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *accountAPIImpl) GetApiKeys(accountIndex int64, apiKeyIndex *uint8) (*api.AccountApiKeys, error) {
	result := &api.AccountApiKeys{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if apiKeyIndex != nil {
		params["api_key_index"] = *apiKeyIndex
	}
	err := a.client.getAndParseL2HTTPResponse("api/v1/apikeys", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

package http

import (
	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type candlestickAPIImpl struct {
	client *client
}

// Ensure candlestickAPIImpl implements CandlestickAPI
var _ core.CandlestickAPI = (*candlestickAPIImpl)(nil)

func (c *candlestickAPIImpl) GetCandlesticks(marketID int16, resolution api.CandlestickResolution, timestamps api.TimestampRange, countBack int) (*api.Candlesticks, error) {
	result := &api.Candlesticks{}
	params := map[string]any{
		"market_id":       marketID,
		"resolution":      string(resolution),
		"start_timestamp": timestamps.StartTimestamp,
		"end_timestamp":   timestamps.EndTimestamp,
		"count_back":      countBack,
	}
	err := c.client.getAndParseL2HTTPResponse("api/v1/candlesticks", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *candlestickAPIImpl) GetFundings(marketID int16, resolution api.FundingResolution, timestamps api.TimestampRange, countBack int) (*api.Fundings, error) {
	result := &api.Fundings{}
	params := map[string]any{
		"market_id":       marketID,
		"resolution":      string(resolution),
		"start_timestamp": timestamps.StartTimestamp,
		"end_timestamp":   timestamps.EndTimestamp,
		"count_back":      countBack,
	}
	err := c.client.getAndParseL2HTTPResponse("api/v1/fundings", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *candlestickAPIImpl) GetFundingRates() (*api.FundingRates, error) {
	result := &api.FundingRates{}
	err := c.client.getAndParseL2HTTPResponse("api/v1/funding-rates", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

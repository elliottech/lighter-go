package http

func GetOrderBooks(baseUrl string, marketId *int32, filter string) (*OrderBooksResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{}
	if marketId != nil {
		params["market_id"] = *marketId
	}
	if filter != "" {
		params["filter"] = filter
	}
	result := &OrderBooksResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/orderBooks", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetOrderBookDetails(baseUrl string, marketId *int32, filter string) (*OrderBookDetailsResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{}
	if marketId != nil {
		params["market_id"] = *marketId
	}
	if filter != "" {
		params["filter"] = filter
	}
	result := &OrderBookDetailsResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/orderBookDetails", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetOrderBookOrders(baseUrl string, marketId int32, limit int32) (*OrderBookOrdersResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{
		"market_id": marketId,
		"limit":     limit,
	}
	result := &OrderBookOrdersResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/orderBookOrders", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetRecentTrades(baseUrl string, marketId int32, limit int32) (*TradesResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{
		"market_id": marketId,
		"limit":     limit,
	}
	result := &TradesResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/recentTrades", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetTrades(baseUrl string, p *TradesParams) (*TradesResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{
		"limit": p.Limit,
	}
	if p.MarketId != nil {
		params["market_id"] = *p.MarketId
	}
	if p.AccountIndex != nil {
		params["account_index"] = *p.AccountIndex
	}
	if p.OrderIndex != nil {
		params["order_index"] = *p.OrderIndex
	}
	if p.SortBy != "" {
		params["sort_by"] = p.SortBy
	}
	if p.SortDir != "" {
		params["sort_dir"] = p.SortDir
	}
	if p.Cursor != "" {
		params["cursor"] = p.Cursor
	}
	if p.From != "" {
		params["from"] = p.From
	}
	if p.AskFilter != nil {
		params["ask_filter"] = *p.AskFilter
	}
	if p.Role != "" {
		params["role"] = p.Role
	}
	if p.Type != "" {
		params["type"] = p.Type
	}
	if p.Aggregate != nil {
		params["aggregate"] = *p.Aggregate
	}
	if p.Auth != "" {
		params["auth"] = p.Auth
	}
	if p.Authorization != "" {
		params["authorization"] = p.Authorization
	}
	result := &TradesResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/trades", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetAccountActiveOrders(baseUrl string, accountIndex int64, marketId int32, auth string) (*OrdersResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{
		"account_index": accountIndex,
		"market_id":     marketId,
		"auth":          auth,
	}
	result := &OrdersResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/accountActiveOrders", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

func GetAccountInactiveOrders(baseUrl string, p *InactiveOrdersParams) (*OrdersResponse, error) {
	c := &client{endpoint: baseUrl}
	params := map[string]any{
		"account_index": p.AccountIndex,
		"limit":         p.Limit,
		"auth":          p.Auth,
	}
	if p.Authorization != "" {
		params["authorization"] = p.Authorization
	}
	if p.MarketId != nil {
		params["market_id"] = *p.MarketId
	}
	if p.AskFilter != nil {
		params["ask_filter"] = *p.AskFilter
	}
	if p.BetweenTimestamps != "" {
		params["between_timestamps"] = p.BetweenTimestamps
	}
	if p.Cursor != "" {
		params["cursor"] = p.Cursor
	}
	result := &OrdersResponse{}
	if err := c.getAndParseL2HTTPResponse("api/v1/accountInactiveOrders", params, result); err != nil {
		return nil, err
	}
	return result, nil
}

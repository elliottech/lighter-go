package http

import (
	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type bridgeAPIImpl struct {
	client *client
}

// Ensure bridgeAPIImpl implements BridgeAPI
var _ core.BridgeAPI = (*bridgeAPIImpl)(nil)

func (b *bridgeAPIImpl) GetBridges(l1Address string) (*api.RespGetBridgesByL1Addr, error) {
	result := &api.RespGetBridgesByL1Addr{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/bridges", map[string]any{
		"l1_address": l1Address,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) GetIsNextBridgeFast(l1Address string) (*api.RespGetIsNextBridgeFast, error) {
	result := &api.RespGetIsNextBridgeFast{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/bridges/isNextBridgeFast", map[string]any{
		"l1_address": l1Address,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) GetFastBridgeInfo() (*api.RespGetFastBridgeInfo, error) {
	result := &api.RespGetFastBridgeInfo{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/fastbridge/info", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *bridgeAPIImpl) CreateIntentAddress(chainID int64, fromAddr string, amount string, isExternalDeposit bool) (*api.RespCreateIntentAddress, error) {
	result := &api.RespCreateIntentAddress{}
	err := b.client.postFormL2HTTPResponse("api/v1/createIntentAddress", map[string]any{
		"chain_id":            chainID,
		"from_addr":           fromAddr,
		"amount":              amount,
		"is_external_deposit": isExternalDeposit,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

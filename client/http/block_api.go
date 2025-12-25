package http

import (
	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type blockAPIImpl struct {
	client *client
}

// Ensure blockAPIImpl implements BlockAPI
var _ core.BlockAPI = (*blockAPIImpl)(nil)

func (b *blockAPIImpl) GetBlock(by api.QueryBy, value string) (*api.Blocks, error) {
	result := &api.Blocks{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/block", map[string]any{
		"by":    string(by),
		"value": value,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *blockAPIImpl) GetBlocks(index *int64, limit int, sort string) (*api.Blocks, error) {
	result := &api.Blocks{}
	params := map[string]any{
		"limit": limit,
	}
	if index != nil {
		params["index"] = *index
	}
	if sort != "" {
		params["sort"] = sort
	}
	err := b.client.getAndParseL2HTTPResponse("api/v1/blocks", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *blockAPIImpl) GetBlockTxs(by api.QueryBy, value string) (*api.Txs, error) {
	result := &api.Txs{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/blockTxs", map[string]any{
		"by":    string(by),
		"value": value,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (b *blockAPIImpl) GetCurrentHeight() (*api.CurrentHeight, error) {
	result := &api.CurrentHeight{}
	err := b.client.getAndParseL2HTTPResponse("api/v1/currentHeight", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

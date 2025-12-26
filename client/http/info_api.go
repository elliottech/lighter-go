package http

import (
	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type infoAPIImpl struct {
	client *client
}

// Ensure infoAPIImpl implements InfoAPI
var _ core.InfoAPI = (*infoAPIImpl)(nil)

func (i *infoAPIImpl) GetStatus() (*api.Status, error) {
	result := &api.Status{}
	err := i.client.getAndParseL2HTTPResponse("", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *infoAPIImpl) GetInfo() (*api.ZkLighterInfo, error) {
	result := &api.ZkLighterInfo{}
	err := i.client.getAndParseL2HTTPResponse("info", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *infoAPIImpl) GetAnnouncements() (*api.Announcements, error) {
	result := &api.Announcements{}
	err := i.client.getAndParseL2HTTPResponse("api/v1/announcement", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *infoAPIImpl) Export(accountIndex int64, marketID int16, exportType api.ExportType) (*api.ExportData, error) {
	result := &api.ExportData{}
	err := i.client.getAndParseL2HTTPResponse("api/v1/export", map[string]any{
		"account_index": accountIndex,
		"market_id":     marketID,
		"type":          string(exportType),
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

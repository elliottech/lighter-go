package http

import (
	"fmt"
	"strings"

	core "github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types/api"
)

type transactionAPIImpl struct {
	client *client
}

// Ensure transactionAPIImpl implements TransactionAPI
var _ core.TransactionAPI = (*transactionAPIImpl)(nil)

func (t *transactionAPIImpl) SendTx(txType uint8, txInfo string, priceProtection *api.PriceProtection) (*api.RespSendTx, error) {
	result := &api.RespSendTx{}
	body := map[string]any{
		"tx_type": txType,
		"tx_info": txInfo,
	}
	if priceProtection != nil {
		body["price_protection"] = priceProtection
	}
	err := t.client.postAndParseL2HTTPResponse("api/v1/sendTx", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) SendTxBatch(txTypes []uint8, txInfos []string) (*api.RespSendTxBatch, error) {
	result := &api.RespSendTxBatch{}
	body := map[string]any{
		"tx_types": txTypes,
		"tx_infos": txInfos,
	}
	err := t.client.postAndParseL2HTTPResponse("api/v1/sendTxBatch", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTx(by api.QueryBy, value string) (*api.EnrichedTx, error) {
	result := &api.EnrichedTx{}
	err := t.client.getAndParseL2HTTPResponse("api/v1/tx", map[string]any{
		"by":    string(by),
		"value": value,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTxs(index *int64, limit int) (*api.Txs, error) {
	result := &api.Txs{}
	params := map[string]any{
		"limit": limit,
	}
	if index != nil {
		params["index"] = *index
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/txs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetAccountTxs(by api.QueryBy, value string, limit int, types []api.TxType) (*api.Txs, error) {
	result := &api.Txs{}
	params := map[string]any{
		"by":    string(by),
		"value": value,
		"limit": limit,
	}
	if len(types) > 0 {
		typeStrs := make([]string, len(types))
		for i, t := range types {
			typeStrs[i] = fmt.Sprintf("%d", t)
		}
		params["types"] = strings.Join(typeStrs, ",")
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/accountTxs", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTxFromL1TxHash(hash string) (*api.EnrichedTx, error) {
	result := &api.EnrichedTx{}
	err := t.client.getAndParseL2HTTPResponse("api/v1/txFromL1TxHash", map[string]any{
		"hash": hash,
	}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetDepositHistory(accountIndex int64, l1Address string, filter string, cursor string) (*api.DepositHistory, error) {
	result := &api.DepositHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if l1Address != "" {
		params["l1_address"] = l1Address
	}
	if filter != "" {
		params["filter"] = filter
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/deposit/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetWithdrawHistory(accountIndex int64, filter string, cursor string) (*api.WithdrawHistory, error) {
	result := &api.WithdrawHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if filter != "" {
		params["filter"] = filter
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/withdraw/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTransferHistory(accountIndex int64, cursor string) (*api.TransferHistory, error) {
	result := &api.TransferHistory{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if cursor != "" {
		params["cursor"] = cursor
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/transfer/history", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetTransferFeeInfo(accountIndex int64, toAccountIndex *int64) (*api.TransferFeeInfo, error) {
	result := &api.TransferFeeInfo{}
	params := map[string]any{
		"account_index": accountIndex,
	}
	if toAccountIndex != nil {
		params["to_account_index"] = *toAccountIndex
	}
	err := t.client.getAndParseL2HTTPResponse("api/v1/transferFeeInfo", params, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *transactionAPIImpl) GetWithdrawalDelay() (*api.RespWithdrawalDelay, error) {
	result := &api.RespWithdrawalDelay{}
	err := t.client.getAndParseL2HTTPResponse("api/v1/withdrawalDelay", map[string]any{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/types"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func ensureClient() error {
	if txClient == nil {
		return fmt.Errorf("client is not created, call CreateClient() first")
	}
	return nil
}

func newTransactOpts(nonce int64) *types.TransactOpts {
	ops := new(types.TransactOpts)
	if nonce != -1 {
		n := nonce
		ops.Nonce = &n
	}
	return ops
}

func generateAPIKey(seed string) (privateKey string, publicKey string, err error) {
	var seedPtr *string
	if seed != "" {
		seedPtr = &seed
	}

	key := curve.SampleScalar(seedPtr)

	publicKey = hexutil.Encode(schnorr.SchnorrPkFromSk(key).ToLittleEndianBytes())
	privateKey = hexutil.Encode(key.ToLittleEndianBytes())

	return
}

func createClient(url, privateKey string, chainID uint32, apiKeyIndex uint8, accountIndex int64) error {
	if accountIndex <= 0 {
		return fmt.Errorf("invalid account index")
	}

	newClient, err := client.NewTxClient(privateKey, accountIndex, apiKeyIndex, chainID)
	if err != nil {
		return fmt.Errorf("error occurred when creating TxClient. err: %v", err)
	}

	txClient = newClient
	if backupTxClients == nil {
		backupTxClients = make(map[uint8]*client.TxClient)
	}
	backupTxClients[apiKeyIndex] = newClient

	return nil
}

func signCreateOrder(marketIndex uint8, clientOrderIndex, baseAmount int64, price uint32, isAsk, orderType, timeInForce, reduceOnly uint8, triggerPrice uint32, orderExpiry int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli()
	}

	txInfo := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: clientOrderIndex,
		BaseAmount:       baseAmount,
		Price:            price,
		IsAsk:            isAsk,
		Type:             orderType,
		TimeInForce:      timeInForce,
		ReduceOnly:       reduceOnly,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      orderExpiry,
	}

	tx, err := txClient.GetCreateOrderTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signCancelOrder(marketIndex uint8, orderIndex int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}

	tx, err := txClient.GetCancelOrderTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signModifyOrder(marketIndex uint8, index, baseAmount int64, price, triggerPrice uint32, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.ModifyOrderTxReq{
		MarketIndex:  marketIndex,
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        price,
		TriggerPrice: triggerPrice,
	}

	tx, err := txClient.GetModifyOrderTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signUpdateLeverage(marketIndex uint8, initialMarginFraction uint16, marginMode uint8, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}

	tx, err := txClient.GetUpdateLeverageTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func createAuthToken(deadline int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	token, err := txClient.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return "", err
	}

	return token, nil
}

func signUpdateMargin(marketIndex uint8, usdcAmount int64, direction uint8, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}

	tx, err := txClient.GetUpdateMarginTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

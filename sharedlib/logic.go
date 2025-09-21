package main

import (
	"encoding/json"
	"fmt"
	"strings"
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

	httpClient := client.NewHTTPClient(url)
	newClient, err := client.NewTxClient(httpClient, privateKey, accountIndex, apiKeyIndex, chainID)
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

func checkClient(apiKeyIndex uint8, accountIndex int64) error {
	if backupTxClients == nil {
		return fmt.Errorf("api key not registered")
	}

	client, ok := backupTxClients[apiKeyIndex]
	if !ok {
		return fmt.Errorf("api key not registered")
	}

	if client.GetApiKeyIndex() != apiKeyIndex {
		return fmt.Errorf("apiKeyIndex does not match. expected %v but got %v", client.GetApiKeyIndex(), apiKeyIndex)
	}
	if client.GetAccountIndex() != accountIndex {
		return fmt.Errorf("accountIndex does not match. expected %v but got %v", client.GetAccountIndex(), accountIndex)
	}

	key, err := client.HTTP().GetApiKey(accountIndex, apiKeyIndex)
	if err != nil {
		return fmt.Errorf("failed to get Api Keys. err: %v", err)
	}

	pubKeyBytes := client.GetKeyManager().PubKeyBytes()
	pubKeyStr := hexutil.Encode(pubKeyBytes[:])
	pubKeyStr = strings.Replace(pubKeyStr, "0x", "", 1)

	ak := key.ApiKeys[0]
	if ak.PublicKey != pubKeyStr {
		return fmt.Errorf("private key does not match the one on Lighter. ownPubKey: %s response: %+v", pubKeyStr, ak)
	}

	return nil
}

func signChangePubKey(pubKeyStr string, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		return "", err
	}
	if len(pubKeyBytes) != 40 {
		return "", fmt.Errorf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes))
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	txInfo := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}
	tx, err := txClient.GetChangePubKeyTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	obj := make(map[string]interface{})
	if err := json.Unmarshal(txInfoBytes, &obj); err != nil {
		return "", err
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
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

func signWithdraw(usdcAmount uint64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := types.WithdrawTxReq{
		USDCAmount: usdcAmount,
	}

	tx, err := txClient.GetWithdrawTransaction(&txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signCreateSubAccount(nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	tx, err := txClient.GetCreateSubAccountTransaction(newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signCancelAllOrders(timeInForce uint8, t, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        t,
	}

	tx, err := txClient.GetCancelAllOrdersTransaction(txInfo, newTransactOpts(nonce))
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

func signTransfer(toAccountIndex, usdcAmount, fee int64, memoStr string, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	if len(memoStr) != 32 {
		return "", fmt.Errorf("memo expected to be 32 bytes long")
	}

	var memo [32]byte
	for i := 0; i < 32; i++ {
		memo[i] = memoStr[i]
	}

	txInfo := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		USDCAmount:     usdcAmount,
		Fee:            fee,
		Memo:           memo,
	}

	tx, err := txClient.GetTransferTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	obj := make(map[string]interface{})
	if err := json.Unmarshal(txInfoBytes, &obj); err != nil {
		return "", err
	}
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signCreatePublicPool(operatorFee, initialTotalShares, minOperatorShareRate int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}

	tx, err := txClient.GetCreatePublicPoolTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signUpdatePublicPool(publicPoolIndex int64, status uint8, operatorFee, minOperatorShareRate int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      publicPoolIndex,
		Status:               status,
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}

	tx, err := txClient.GetUpdatePublicPoolTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signMintShares(publicPoolIndex, shareAmount int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}

	tx, err := txClient.GetMintSharesTransaction(txInfo, newTransactOpts(nonce))
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

func signBurnShares(publicPoolIndex, shareAmount int64, nonce int64) (string, error) {
	if err := ensureClient(); err != nil {
		return "", err
	}

	txInfo := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}

	tx, err := txClient.GetBurnSharesTransaction(txInfo, newTransactOpts(nonce))
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

func switchAPIKey(apiKeyIndex uint8) error {
	if backupTxClients == nil {
		return fmt.Errorf("no client initialized for api key")
	}

	client := backupTxClients[apiKeyIndex]
	if client == nil {
		return fmt.Errorf("no client initialized for api key")
	}

	txClient = client
	return nil
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

package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
	"github.com/elliottech/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// SharedClientManager holds the global txClient and backupTxClients
// This will be managed by both sharedlib and wasm builds
// Supports multiple accounts and API keys with thread safety
var (
	txClientMu      sync.Mutex
	defaultTxClient *TxClient
	allTxClients    map[int64]map[uint8]*TxClient // accountIndex -> apiKeyIndex -> client

	// Legacy support for old API - kept for compatibility
	txClient        *TxClient // Deprecated: Use GetClientForAPIKey instead
	backupTxClients map[uint8]*TxClient // Deprecated: Use allTxClients instead
)

// GetClientForAPIKey retrieves a client for specific account and API key
// If apiKeyIndex==255 && accountIndex==-1, returns default client
func GetClientForAPIKey(apiKeyIndex uint8, accountIndex int64) (*TxClient, error) {
	txClientMu.Lock()
	defer txClientMu.Unlock()

	// Special case: return default client
	if apiKeyIndex == 255 && accountIndex == -1 {
		if defaultTxClient == nil {
			return nil, fmt.Errorf("client is not created, call CreateClient() first")
		}
		return defaultTxClient, nil
	}

	// Look up client in double map
	var c *TxClient
	if allTxClients[accountIndex] != nil {
		c = allTxClients[accountIndex][apiKeyIndex]
	}

	if c == nil {
		return nil, fmt.Errorf("client is not created for apiKeyIndex: %v accountIndex: %v", apiKeyIndex, accountIndex)
	}
	return c, nil
}

// getDefaultClient is a helper for legacy functions that use the default client
func getDefaultClient() (*TxClient, error) {
	return GetClientForAPIKey(255, -1)
}

// GenerateAPIKey generates a new API key pair from a seed
func GenerateAPIKey(seed string) (string, string, error) {
	var seedP *string
	if seed != "" {
		seedP = &seed
	}

	key := curve.SampleScalar(seedP)
	publicKeyStr := hexutil.Encode(schnorr.SchnorrPkFromSk(key).ToLittleEndianBytes())
	privateKeyStr := hexutil.Encode(key.ToLittleEndianBytes())

	return privateKeyStr, publicKeyStr, nil
}

// CreateClient creates a new TxClient and stores it
// httpClientFactory is a function that creates an HTTP client from a URL string
func CreateClient(httpClientFactory func(string) MinimalHTTPClient, url, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64) error {
	if accountIndex <= 0 {
		return fmt.Errorf("invalid account index")
	}

	httpClient := httpClientFactory(url)
	txClientInstance, err := NewTxClient(httpClient, privateKey, accountIndex, apiKeyIndex, chainId)
	if err != nil {
		return fmt.Errorf("error occurred when creating TxClient. err: %v", err)
	}

	txClientMu.Lock()
	if allTxClients == nil {
		allTxClients = make(map[int64]map[uint8]*TxClient)
	}
	if allTxClients[accountIndex] == nil {
		allTxClients[accountIndex] = make(map[uint8]*TxClient)
	}
	allTxClients[accountIndex][apiKeyIndex] = txClientInstance
	
	// Update default client (most recently created becomes default)
	defaultTxClient = txClientInstance
	
	// Update legacy variables for backward compatibility
	txClient = txClientInstance
	if backupTxClients == nil {
		backupTxClients = make(map[uint8]*TxClient)
	}
	backupTxClients[apiKeyIndex] = txClientInstance
	txClientMu.Unlock()

	return nil
}

// GetClient returns the global txClient
func GetClient() *TxClient {
	return txClient
}

// GetBackupClients returns the backup clients map
func GetBackupClients() map[uint8]*TxClient {
	return backupTxClients
}

// CheckClient validates that the client exists and the API key matches the one on the server
func CheckClient(apiKeyIndex uint8, accountIndex int64) error {
	clientInstance, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return err
	}

	// check that the API key registered on Lighter matches this one
	publicKey, err := clientInstance.HTTP().GetApiKey(accountIndex, apiKeyIndex)
	if err != nil {
		return fmt.Errorf("failed to get Api Keys. err: %v", err)
	}

	pubKeyBytes := clientInstance.GetKeyManager().PubKeyBytes()
	pubKeyStr := hexutil.Encode(pubKeyBytes[:])
	pubKeyStr = strings.Replace(pubKeyStr, "0x", "", 1)

	if publicKey != pubKeyStr {
		return fmt.Errorf("private key does not match the one on Lighter. ownPubKey: %s response: %+v", pubKeyStr, publicKey)
	}

	return nil
}

// GetChangePubKeyTransaction generates a ChangePubKey transaction
// Uses default client (legacy API for backward compatibility)
func GetChangePubKeyTransaction(pubKey [40]byte, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", "", err
	}

	txInfo := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}
	ops := &types.TransactOpts{}
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetChangePubKeyTransaction(txInfo, ops)
	if err != nil {
		return "", "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", "", err
	}

	// Add MessageToSign to the response
	txInfoMap := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &txInfoMap)
	if err != nil {
		return "", "", err
	}
	txInfoMap["MessageToSign"] = tx.GetL1SignatureBody()
	
	txInfoBytesFinal, err := json.Marshal(txInfoMap)
	if err != nil {
		return "", "", err
	}

	return string(txInfoBytesFinal), "", nil
}

// GetCreateOrderTransaction generates a CreateOrder transaction
func GetCreateOrderTransaction(
	marketIndex uint8,
	clientOrderIndex int64,
	baseAmount int64,
	price uint32,
	isAsk uint8,
	orderType uint8,
	timeInForce uint8,
	reduceOnly uint8,
	triggerPrice uint32,
	orderExpiry int64,
	nonce int64,
	apiKeyIndex uint8,
	accountIndex int64,
) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
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
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateOrderTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// CreateAuthToken generates an auth token
func CreateAuthToken(deadline int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := txClient.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return "", err
	}

	return authToken, nil
}

// SwitchAPIKey switches the active txClient to a different API key
// Note: This is a legacy function. Consider using GetClientForAPIKey instead for multi-account support
func SwitchAPIKey(apiKeyIndex uint8) error {
	txClientMu.Lock()
	defer txClientMu.Unlock()
	
	if backupTxClients[apiKeyIndex] != nil {
		txClient = backupTxClients[apiKeyIndex]
		// Update default client as well
		if defaultTxClient == nil || defaultTxClient.GetApiKeyIndex() != apiKeyIndex {
			// Find the client in allTxClients
			for _, accountMap := range allTxClients {
				if client := accountMap[apiKeyIndex]; client != nil {
					defaultTxClient = client
					return nil
				}
			}
		}
		return nil
	}
	return fmt.Errorf("no client initialized for api key")
}

// GetCreateSubAccountTransaction generates a CreateSubAccount transaction
func GetCreateSubAccountTransaction(nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreateSubAccountTransaction(ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetCreatePublicPoolTransaction generates a CreatePublicPool transaction
func GetCreatePublicPoolTransaction(operatorFee, initialTotalShares, minOperatorShareRate, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCreatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetUpdatePublicPoolTransaction generates an UpdatePublicPool transaction
func GetUpdatePublicPoolTransaction(publicPoolIndex, status uint8, operatorFee, minOperatorShareRate, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      int64(publicPoolIndex),
		Status:               status,
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetTransferTransaction generates a Transfer transaction
func GetTransferTransaction(toAccountIndex, usdcAmount, fee, nonce int64, memo [32]byte, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		USDCAmount:     usdcAmount,
		Fee:            fee,
		Memo:           memo,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetTransferTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	// Add MessageToSign to the response
	txInfoMap := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &txInfoMap)
	if err != nil {
		return "", err
	}
	txInfoMap["MessageToSign"] = tx.GetL1SignatureBody()
	
	txInfoBytesFinal, err := json.Marshal(txInfoMap)
	if err != nil {
		return "", err
	}

	return string(txInfoBytesFinal), nil
}

// GetWithdrawTransaction generates a Withdraw transaction
func GetWithdrawTransaction(usdcAmount uint64, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.WithdrawTxReq{
		USDCAmount: usdcAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetWithdrawTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetCreateGroupedOrdersTransaction generates a CreateGroupedOrders transaction
func GetCreateGroupedOrdersTransaction(groupingType uint8, orders []*types.CreateOrderTxReq, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: groupingType,
		Orders:       orders,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	txInfo, err := txClient.GetCreateGroupedOrdersTransaction(req, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetCancelOrderTransaction generates a CancelOrder transaction
func GetCancelOrderTransaction(marketIndex uint8, orderIndex, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelOrderTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetModifyOrderTransaction generates a ModifyOrder transaction
func GetModifyOrderTransaction(marketIndex uint8, index, baseAmount int64, price, triggerPrice uint32, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.ModifyOrderTxReq{
		MarketIndex:  marketIndex,
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        price,
		TriggerPrice: triggerPrice,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetModifyOrderTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetCancelAllOrdersTransaction generates a CancelAllOrders transaction
func GetCancelAllOrdersTransaction(timeInForce uint8, timeVal, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        timeVal,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetCancelAllOrdersTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetMintSharesTransaction generates a MintShares transaction
func GetMintSharesTransaction(publicPoolIndex, shareAmount, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetMintSharesTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetBurnSharesTransaction generates a BurnShares transaction
func GetBurnSharesTransaction(publicPoolIndex, shareAmount, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetBurnSharesTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetUpdateLeverageTransaction generates an UpdateLeverage transaction
func GetUpdateLeverageTransaction(marketIndex, marginMode uint8, initialMarginFraction uint16, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateLeverageTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}

// GetUpdateMarginTransaction generates an UpdateMargin transaction
func GetUpdateMarginTransaction(marketIndex, direction uint8, usdcAmount, nonce int64, apiKeyIndex uint8, accountIndex int64) (string, error) {
	txClient, err := GetClientForAPIKey(apiKeyIndex, accountIndex)
	if err != nil {
		return "", err
	}

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := txClient.GetUpdateMarginTransaction(txInfo, ops)
	if err != nil {
		return "", err
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}

	return string(txInfoBytes), nil
}


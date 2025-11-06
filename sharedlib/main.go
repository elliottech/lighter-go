package main

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

/*
#include <stdlib.h>
#include <stdint.h>
typedef struct {
	char* str;
	char* err;
} StrOrErr;

typedef struct {
	uint8_t txType;
	char* txInfo;
	char* txHash;
	char* messageToSign;
	char* err;
} SignedTxResponse;

typedef struct {
	char* privateKey;
	char* publicKey;
	char* err;
} ApiKeyResponse;

typedef struct {
    uint8_t MarketIndex;
    int64_t ClientOrderIndex;
    int64_t BaseAmount;
    uint32_t Price;
    uint8_t IsAsk;
    uint8_t Type;
    uint8_t TimeInForce;
    uint8_t ReduceOnly;
    uint32_t TriggerPrice;
    int64_t OrderExpiry;
} CreateOrderTxReq;
*/
import "C"

func wrapErr(err error) (ret *C.char) {
	return C.CString(fmt.Sprintf("%v", err))
}

// convertSignedTxResponse takes the return type of operations which signs a message, and returns the C response
// if MessageToSign is not present, it returns nul
func convertSignedTxResponse(response *client.SignedTx, err error) C.SignedTxResponse {
	if err != nil {
		return C.SignedTxResponse{
			err: wrapErr(err),
		}
	} else {
		resp := C.SignedTxResponse{
			txType: C.uint8_t(response.TxType),
			txInfo: C.CString(response.TxInfo),
			txHash: C.CString(response.TxHash),
		}
		if len(response.MessageToSign) > 0 {
			resp.messageToSign = C.CString(response.MessageToSign)
		}

		return resp
	}
}

// getClient returns the go TxClient from the specified cApiKeyIndex and cAccountIndex
func getClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (*client.TxClient, error) {
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	return client.GetClient(apiKeyIndex, accountIndex)
}

//export GenerateAPIKey
func GenerateAPIKey(cSeed *C.char) (ret C.ApiKeyResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.ApiKeyResponse{
				err: wrapErr(fmt.Errorf("panic: %v", r)),
			}
		}
	}()

	seed := C.GoString(cSeed)
	privateKeyStr, publicKeyStr, err := client.GenerateAPIKey(seed)
	if err != nil {
		ret = C.ApiKeyResponse{
			err: wrapErr(err),
		}
	} else {
		ret = C.ApiKeyResponse{
			privateKey: C.CString(privateKeyStr),
			publicKey:  C.CString(publicKeyStr),
		}
	}
	return
}

//export CreateClient
func CreateClient(cUrl *C.char, cPrivateKey *C.char, cChainId C.int, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	defer func() {
		if r := recover(); r != nil {
			ret = wrapErr(fmt.Errorf("panic: %v", r))
		}
	}()

	url := C.GoString(cUrl)
	privateKey := C.GoString(cPrivateKey)
	chainId := uint32(cChainId)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	httpClient := http.NewClient(url)

	_, err := client.CreateClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
	return wrapErr(err)
}

//export CheckClient
func CheckClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	defer func() {
		if r := recover(); r != nil {
			ret = wrapErr(fmt.Errorf("panic: %v", r))
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return wrapErr(err)
	}

	return wrapErr(c.Check())
}

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.SignedTxResponse{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return C.SignedTxResponse{err: wrapErr(err)}
	}

	nonce := int64(cNonce)
	pubKeyStr := C.GoString(cPubKey)
	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		return C.SignedTxResponse{err: wrapErr(err)}
	}
	if len(pubKeyBytes) != 40 {
		return C.SignedTxResponse{err: wrapErr(fmt.Errorf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes)))}
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	resp, err := c.GetChangePubKeyTx(pubKey, nonce)
	return convertSignedTxResponse(resp, err)
}

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	marketIndex := uint8(cMarketIndex)
	clientOrderIndex := int64(cClientOrderIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	isAsk := uint8(cIsAsk)
	orderType := uint8(cOrderType)
	timeInForce := uint8(cTimeInForce)
	reduceOnly := uint8(cReduceOnly)
	triggerPrice := uint32(cTriggerPrice)
	orderExpiry := int64(cOrderExpiry)
	nonce := int64(cNonce)

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

	tx, err := c.GetCreateOrderTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignCreateGroupedOrders
func SignCreateGroupedOrders(cGroupingType C.uint8_t, cOrders *C.CreateOrderTxReq, cLen C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	length := int(cLen)
	orders := make([]*types.CreateOrderTxReq, length)
	size := unsafe.Sizeof(*cOrders)
	nonce := int64(cNonce)

	for i := 0; i < length; i++ {
		order := (*C.CreateOrderTxReq)(unsafe.Pointer(uintptr(unsafe.Pointer(cOrders)) + uintptr(i)*uintptr(size)))

		orderExpiry := int64(order.OrderExpiry)
		if orderExpiry == -1 {
			orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli()
		}

		orders[i] = &types.CreateOrderTxReq{
			MarketIndex:      uint8(order.MarketIndex),
			ClientOrderIndex: int64(order.ClientOrderIndex),
			BaseAmount:       int64(order.BaseAmount),
			Price:            uint32(order.Price),
			IsAsk:            uint8(order.IsAsk),
			Type:             uint8(order.Type),
			TimeInForce:      uint8(order.TimeInForce),
			ReduceOnly:       uint8(order.ReduceOnly),
			TriggerPrice:     uint32(order.TriggerPrice),
			OrderExpiry:      orderExpiry,
		}
	}

	req := &types.CreateGroupedOrdersTxReq{
		GroupingType: uint8(cGroupingType),
		Orders:       orders,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	txInfo, err := c.GetCreateGroupedOrdersTransaction(req, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	marketIndex := uint8(cMarketIndex)
	orderIndex := int64(cOrderIndex)
	nonce := int64(cNonce)

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetCancelOrderTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignWithdraw
func SignWithdraw(cUSDCAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	usdcAmount := uint64(cUSDCAmount)
	nonce := int64(cNonce)

	txInfo := &types.WithdrawTxReq{
		USDCAmount: usdcAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetWithdrawTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignCreateSubAccount
func SignCreateSubAccount(cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	nonce := int64(cNonce)

	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetCreateSubAccountTransaction(ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	timeInForce := uint8(cTimeInForce)
	t := int64(cTime)
	nonce := int64(cNonce)

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        t,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetCancelAllOrdersTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	marketIndex := uint8(cMarketIndex)
	index := int64(cIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	triggerPrice := uint32(cTriggerPrice)
	nonce := int64(cNonce)

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

	tx, err := c.GetModifyOrderTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cUSDCAmount C.longlong, cFee C.longlong, cMemo *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	toAccountIndex := int64(cToAccountIndex)
	usdcAmount := int64(cUSDCAmount)
	nonce := int64(cNonce)
	fee := int64(cFee)
	memo := [32]byte{}
	memoStr := C.GoString(cMemo)
	if len(memoStr) != 32 {
		ret = C.StrOrErr{err: wrapErr(fmt.Errorf("memo expected to be 32 bytes long"))}
		return
	}
	for i := 0; i < 32; i++ {
		memo[i] = byte(memoStr[i])
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

	tx, err := c.GetTransferTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	// Add MessageToSign to the response
	txInfoMap := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &txInfoMap)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}
	txInfoMap["MessageToSign"] = tx.GetL1SignatureBody()

	txInfoBytesFinal, err := json.Marshal(txInfoMap)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytesFinal))}
	return
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	operatorFee := int64(cOperatorFee)
	initialTotalShares := int64(cInitialTotalShares)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)

	txInfo := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetCreatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	publicPoolIndex := uint8(cPublicPoolIndex)
	status := uint8(cStatus)
	operatorFee := int64(cOperatorFee)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)

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

	tx, err := c.GetUpdatePublicPoolTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)

	txInfo := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetMintSharesTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)

	txInfo := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetBurnSharesTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	marketIndex := uint8(cMarketIndex)
	initialMarginFraction := uint16(cInitialMarginFraction)
	nonce := int64(cNonce)
	marginMode := uint8(cMarginMode)

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetUpdateLeverageTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

//export CreateAuthToken
func CreateAuthToken(cDeadline C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	deadline := int64(cDeadline)
	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := c.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(authToken)}
	return
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.StrOrErr{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	marketIndex := uint8(cMarketIndex)
	usdcAmount := int64(cUSDCAmount)
	direction := uint8(cDirection)
	nonce := int64(cNonce)

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}
	ops := new(types.TransactOpts)
	if nonce != -1 {
		ops.Nonce = &nonce
	}

	tx, err := c.GetUpdateMarginTransaction(txInfo, ops)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}

	ret = C.StrOrErr{str: C.CString(string(txInfoBytes))}
	return
}

func main() {}

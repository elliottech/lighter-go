package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/elliottech/lighter-go/executables"
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

//export GenerateAPIKey
func GenerateAPIKey(cSeed *C.char) (ret C.ApiKeyResponse) {
	seed := C.GoString(cSeed)
	privateKeyStr, publicKeyStr, err := executables.GenerateAPIKey(seed)
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
	url := C.GoString(cUrl)
	privateKey := C.GoString(cPrivateKey)
	chainId := uint32(cChainId)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	return wrapErr(executables.CreateClient(url, privateKey, chainId, apiKeyIndex, accountIndex))
}

//export CheckClient
func CheckClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	return wrapErr(executables.CheckClient(apiKeyIndex, accountIndex))
}

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	pubKeyStr := C.GoString(cPubKey)
	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
		return
	}
	if len(pubKeyBytes) != 40 {
		ret = C.StrOrErr{err: wrapErr(fmt.Errorf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes)))}
		return
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	txInfoStr, _, err := executables.GetChangePubKeyTransaction(pubKey, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetCreateOrderTransaction(
		marketIndex,
		clientOrderIndex,
		baseAmount,
		price,
		isAsk,
		orderType,
		timeInForce,
		reduceOnly,
		triggerPrice,
		orderExpiry,
		nonce,
		apiKeyIndex,
		accountIndex,
	)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCreateGroupedOrders
func SignCreateGroupedOrders(cGroupingType C.uint8_t, cOrders *C.CreateOrderTxReq, cLen C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	length := int(cLen)
	orders := make([]*types.CreateOrderTxReq, length)
	size := unsafe.Sizeof(*cOrders)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

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

	txInfoStr, err := executables.GetCreateGroupedOrdersTransaction(uint8(cGroupingType), orders, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	marketIndex := uint8(cMarketIndex)
	orderIndex := int64(cOrderIndex)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetCancelOrderTransaction(marketIndex, orderIndex, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignWithdraw
func SignWithdraw(cUSDCAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	usdcAmount := uint64(cUSDCAmount)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetWithdrawTransaction(usdcAmount, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCreateSubAccount
func SignCreateSubAccount(cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	txInfoStr, err := executables.GetCreateSubAccountTransaction(nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	timeInForce := uint8(cTimeInForce)
	t := int64(cTime)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetCancelAllOrdersTransaction(timeInForce, t, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	marketIndex := uint8(cMarketIndex)
	index := int64(cIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	triggerPrice := uint32(cTriggerPrice)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetModifyOrderTransaction(marketIndex, index, baseAmount, price, triggerPrice, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cUSDCAmount C.longlong, cFee C.longlong, cMemo *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	toAccountIndex := int64(cToAccountIndex)
	usdcAmount := int64(cUSDCAmount)
	nonce := int64(cNonce)
	fee := int64(cFee)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	memo := [32]byte{}
	memoStr := C.GoString(cMemo)
	if len(memoStr) != 32 {
		ret = C.StrOrErr{err: wrapErr(fmt.Errorf("memo expected to be 32 bytes long"))}
		return
	}
	for i := 0; i < 32; i++ {
		memo[i] = byte(memoStr[i])
	}

	txInfoStr, err := executables.GetTransferTransaction(toAccountIndex, usdcAmount, fee, nonce, memo, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	operatorFee := int64(cOperatorFee)
	initialTotalShares := int64(cInitialTotalShares)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetCreatePublicPoolTransaction(operatorFee, initialTotalShares, minOperatorShareRate, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	publicPoolIndex := uint8(cPublicPoolIndex)
	status := uint8(cStatus)
	operatorFee := int64(cOperatorFee)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetUpdatePublicPoolTransaction(publicPoolIndex, status, operatorFee, minOperatorShareRate, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetMintSharesTransaction(publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetBurnSharesTransaction(publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	marketIndex := uint8(cMarketIndex)
	initialMarginFraction := uint16(cInitialMarginFraction)
	nonce := int64(cNonce)
	marginMode := uint8(cMarginMode)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetUpdateLeverageTransaction(marketIndex, marginMode, initialMarginFraction, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

//export CreateAuthToken
func CreateAuthToken(cDeadline C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	deadline := int64(cDeadline)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	authToken, err := executables.CreateAuthToken(deadline, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(authToken)}
	}
	return
}

//export SwitchAPIKey
func SwitchAPIKey(c C.int) (ret *C.char) {
	apiKeyIndex := uint8(c)
	return wrapErr(executables.SwitchAPIKey(apiKeyIndex))
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	marketIndex := uint8(cMarketIndex)
	usdcAmount := int64(cUSDCAmount)
	direction := uint8(cDirection)
	nonce := int64(cNonce)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	txInfoStr, err := executables.GetUpdateMarginTransaction(marketIndex, direction, usdcAmount, nonce, apiKeyIndex, accountIndex)
	if err != nil {
		ret = C.StrOrErr{err: wrapErr(err)}
	} else {
		ret = C.StrOrErr{str: C.CString(txInfoStr)}
	}
	return
}

func main() {}

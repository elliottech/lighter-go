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
	var err error
	var privateKeyStr string
	var publicKeyStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
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
	}()

	seed := C.GoString(cSeed)
	privateKeyStr, publicKeyStr, err = executables.GenerateAPIKey(seed)
	return
}

//export CreateClient
func CreateClient(cUrl *C.char, cPrivateKey *C.char, cChainId C.int, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	url := C.GoString(cUrl)
	privateKey := C.GoString(cPrivateKey)
	chainId := uint32(cChainId)
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	err = executables.CreateClient(url, privateKey, chainId, apiKeyIndex, accountIndex)
	return
}

//export CheckClient
func CheckClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)

	err = executables.CheckClient(apiKeyIndex, accountIndex)
	return
}

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	nonce := int64(cNonce)

	// handle PubKey
	pubKeyStr := C.GoString(cPubKey)
	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		return
	}
	if len(pubKeyBytes) != 40 {
		err = fmt.Errorf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes))
		return
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	txInfoStr, _, err = executables.GetChangePubKeyTransaction(pubKey, nonce)
	return
}

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

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

	txInfoStr, err = executables.GetCreateOrderTransaction(
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
	)
	return
}

//export SignCreateGroupedOrders
func SignCreateGroupedOrders(cGroupingType C.uint8_t, cOrders *C.CreateOrderTxReq, cLen C.int, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

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

	txInfoStr, err = executables.GetCreateGroupedOrdersTransaction(uint8(cGroupingType), orders, nonce)
	return
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	marketIndex := uint8(cMarketIndex)
	orderIndex := int64(cOrderIndex)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetCancelOrderTransaction(marketIndex, orderIndex, nonce)
	return
}

//export SignWithdraw
func SignWithdraw(cUSDCAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	usdcAmount := uint64(cUSDCAmount)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetWithdrawTransaction(usdcAmount, nonce)
	return
}

//export SignCreateSubAccount
func SignCreateSubAccount(cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	nonce := int64(cNonce)
	txInfoStr, err = executables.GetCreateSubAccountTransaction(nonce)
	return
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	timeInForce := uint8(cTimeInForce)
	t := int64(cTime)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetCancelAllOrdersTransaction(timeInForce, t, nonce)
	return
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	marketIndex := uint8(cMarketIndex)
	index := int64(cIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	triggerPrice := uint32(cTriggerPrice)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetModifyOrderTransaction(marketIndex, index, baseAmount, price, triggerPrice, nonce)
	return
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cUSDCAmount C.longlong, cFee C.longlong, cMemo *C.char, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	toAccountIndex := int64(cToAccountIndex)
	usdcAmount := int64(cUSDCAmount)
	nonce := int64(cNonce)
	fee := int64(cFee)
	memo := [32]byte{}
	memoStr := C.GoString(cMemo)
	if len(memoStr) != 32 {
		err = fmt.Errorf("memo expected to be 32 bytes long")
		return
	}
	for i := 0; i < 32; i++ {
		memo[i] = byte(memoStr[i])
	}

	txInfoStr, err = executables.GetTransferTransaction(toAccountIndex, usdcAmount, fee, nonce, memo)
	return
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	operatorFee := int64(cOperatorFee)
	initialTotalShares := int64(cInitialTotalShares)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetCreatePublicPoolTransaction(operatorFee, initialTotalShares, minOperatorShareRate, nonce)
	return
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	publicPoolIndex := uint8(cPublicPoolIndex)
	status := uint8(cStatus)
	operatorFee := int64(cOperatorFee)
	minOperatorShareRate := int64(cMinOperatorShareRate)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetUpdatePublicPoolTransaction(publicPoolIndex, status, operatorFee, minOperatorShareRate, nonce)
	return
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetMintSharesTransaction(publicPoolIndex, shareAmount, nonce)
	return
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetBurnSharesTransaction(publicPoolIndex, shareAmount, nonce)
	return
}

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	marketIndex := uint8(cMarketIndex)
	initialMarginFraction := uint16(cInitialMarginFraction)
	nonce := int64(cNonce)
	marginMode := uint8(cMarginMode)

	txInfoStr, err = executables.GetUpdateLeverageTransaction(marketIndex, marginMode, initialMarginFraction, nonce)
	return
}

//export CreateAuthToken
func CreateAuthToken(cDeadline C.longlong) (ret C.StrOrErr) {
	var err error
	var authToken string

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(authToken),
			}
		}
	}()

	deadline := int64(cDeadline)
	authToken, err = executables.CreateAuthToken(deadline)
	return
}

//export SwitchAPIKey
func SwitchAPIKey(c C.int) (ret *C.char) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	apiKeyIndex := uint8(c)
	err = executables.SwitchAPIKey(apiKeyIndex)
	return
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
		if err != nil {
			ret = C.StrOrErr{
				err: wrapErr(err),
			}
		} else {
			ret = C.StrOrErr{
				str: C.CString(txInfoStr),
			}
		}
	}()

	marketIndex := uint8(cMarketIndex)
	usdcAmount := int64(cUSDCAmount)
	direction := uint8(cDirection)
	nonce := int64(cNonce)

	txInfoStr, err = executables.GetUpdateMarginTransaction(marketIndex, direction, usdcAmount, nonce)
	return
}

func main() {}


//go:build !js && cgo

package main

/*
#include <stdlib.h>
typedef struct {
    char* str;
    char* err;
} StrOrErr;

typedef struct {
    char* privateKey;
    char* publicKey;
    char* err;
} ApiKeyResponse;
*/
import "C"

import "fmt"

func wrapErr(err error) *C.char {
	return C.CString(fmt.Sprintf("%v", err))
}

func recoverErr(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("%v", r)
	}
}

func finalizeStrOrErr(ret *C.StrOrErr, value string, err error) {
	if err != nil {
		ret.err = wrapErr(err)
	} else {
		ret.str = C.CString(value)
	}
}

//export GenerateAPIKey
func GenerateAPIKey(cSeed *C.char) (ret C.ApiKeyResponse) {
	var err error
	var privateKeyStr string
	var publicKeyStr string

	defer func() {
		recoverErr(&err)
		if err != nil {
			ret = C.ApiKeyResponse{err: wrapErr(err)}
		} else {
			ret = C.ApiKeyResponse{
				privateKey: C.CString(privateKeyStr),
				publicKey:  C.CString(publicKeyStr),
			}
		}
	}()

	privateKeyStr, publicKeyStr, err = generateAPIKey(C.GoString(cSeed))
	return
}

//export CreateClient
func CreateClient(cUrl *C.char, cPrivateKey *C.char, cChainId C.int, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	var err error
	defer func() {
		recoverErr(&err)
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	err = createClient(
		C.GoString(cUrl),
		C.GoString(cPrivateKey),
		uint32(cChainId),
		uint8(cApiKeyIndex),
		int64(cAccountIndex),
	)
	return nil
}

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signChangePubKey(C.GoString(cPubKey), int64(cNonce))
	return
}

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signCreateOrder(
		uint8(cMarketIndex),
		int64(cClientOrderIndex),
		int64(cBaseAmount),
		uint32(cPrice),
		uint8(cIsAsk),
		uint8(cOrderType),
		uint8(cTimeInForce),
		uint8(cReduceOnly),
		uint32(cTriggerPrice),
		int64(cOrderExpiry),
		int64(cNonce),
	)
	return
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signCancelOrder(uint8(cMarketIndex), int64(cOrderIndex), int64(cNonce))
	return
}

//export SignWithdraw
func SignWithdraw(cUSDCAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signWithdraw(uint64(cUSDCAmount), int64(cNonce))
	return
}

//export SignCreateSubAccount
func SignCreateSubAccount(cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signCreateSubAccount(int64(cNonce))
	return
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signCancelAllOrders(uint8(cTimeInForce), int64(cTime), int64(cNonce))
	return
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signModifyOrder(
		uint8(cMarketIndex),
		int64(cIndex),
		int64(cBaseAmount),
		uint32(cPrice),
		uint32(cTriggerPrice),
		int64(cNonce),
	)
	return
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cUSDCAmount C.longlong, cFee C.longlong, cMemo *C.char, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signTransfer(
		int64(cToAccountIndex),
		int64(cUSDCAmount),
		int64(cFee),
		C.GoString(cMemo),
		int64(cNonce),
	)
	return
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signCreatePublicPool(
		int64(cOperatorFee),
		int64(cInitialTotalShares),
		int64(cMinOperatorShareRate),
		int64(cNonce),
	)
	return
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signUpdatePublicPool(
		int64(cPublicPoolIndex),
		uint8(cStatus),
		int64(cOperatorFee),
		int64(cMinOperatorShareRate),
		int64(cNonce),
	)
	return
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signMintShares(int64(cPublicPoolIndex), int64(cShareAmount), int64(cNonce))
	return
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signBurnShares(int64(cPublicPoolIndex), int64(cShareAmount), int64(cNonce))
	return
}

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signUpdateLeverage(
		uint8(cMarketIndex),
		uint16(cInitialMarginFraction),
		uint8(cMarginMode),
		int64(cNonce),
	)
	return
}

//export CreateAuthToken
func CreateAuthToken(cDeadline C.longlong) (ret C.StrOrErr) {
	var err error
	var authToken string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, authToken, err)
	}()

	authToken, err = createAuthToken(int64(cDeadline))
	return
}

//export SwitchAPIKey
func SwitchAPIKey(c C.int) (ret *C.char) {
	var err error
	defer func() {
		recoverErr(&err)
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	err = switchAPIKey(uint8(c))
	return nil
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cNonce C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string

	defer func() {
		recoverErr(&err)
		finalizeStrOrErr(&ret, txInfoStr, err)
	}()

	txInfoStr, err = signUpdateMargin(
		uint8(cMarketIndex),
		int64(cUSDCAmount),
		uint8(cDirection),
		int64(cNonce),
	)
	return
}

func main() {}

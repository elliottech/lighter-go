package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/types"
	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
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

var (
	txClientMu      sync.Mutex
	defaultTxClient *client.TxClient
	allTxClients    map[int64]map[uint8]*client.TxClient
)

func wrapErr(err error) (ret *C.char) {
	return C.CString(fmt.Sprintf("%v", err))
}

func getTxClient(apiKeyIndex uint8, accountIndex int64) (*client.TxClient, error) {
	txClientMu.Lock()
	defer txClientMu.Unlock()

	if apiKeyIndex == 255 && accountIndex == -1 {
		if defaultTxClient == nil {
			return nil, fmt.Errorf("client is not created, call CreateClient() first")
		} else {
			return defaultTxClient, nil
		}
	}

	var c *client.TxClient
	if allTxClients[accountIndex] != nil {
		c = allTxClients[accountIndex][apiKeyIndex]
	}

	if c == nil {
		return nil, fmt.Errorf("client is not created for apiKeyIndex: %v accountIndex: %v", apiKeyIndex, accountIndex)
	}
	return c, nil
}

func getOps(cNonce C.longlong) *types.TransactOpts {
	nonce := int64(cNonce)
	return &types.TransactOpts{
		Nonce: &nonce,
	}
}

/// === Client related ops ===

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

	if accountIndex <= 0 {
		err = fmt.Errorf("invalid account index")
		return
	}

	httpClient := http.NewClient(url)
	txClient, err := client.NewTxClient(httpClient, privateKey, accountIndex, apiKeyIndex, chainId)
	if err != nil {
		err = fmt.Errorf("error occurred when creating TxClient. err: %v", err)
		return
	}

	txClientMu.Lock()
	if allTxClients == nil {
		allTxClients = make(map[int64]map[uint8]*client.TxClient)
	}
	if allTxClients[accountIndex] == nil {
		allTxClients[accountIndex] = make(map[uint8]*client.TxClient)
	}
	allTxClients[accountIndex][apiKeyIndex] = txClient
	txClientMu.Unlock()

	return nil
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	// check that the API key registered on Lighter matches this one
	publicKey, err := txClient.HTTP().GetApiKey(txClient.GetAccountIndex(), txClient.GetApiKeyIndex())
	if err != nil {
		err = fmt.Errorf("failed to get Api Keys. err: %v", err)
		return
	}

	pubKeyBytes := txClient.GetKeyManager().PubKeyBytes()
	pubKeyStr := hexutil.Encode(pubKeyBytes[:])
	pubKeyStr = strings.Replace(pubKeyStr, "0x", "", 1)

	if publicKey != pubKeyStr {
		err = fmt.Errorf("private key does not match the one on Lighter. ownPubKey: %s response: %+v", pubKeyStr, publicKey)
		return
	}

	return
}

//export SwitchDefaultClient
func SwitchDefaultClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (ret *C.char) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			ret = wrapErr(err)
		}
	}()

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	txClientMu.Lock()
	defaultTxClient = txClient
	txClientMu.Unlock()

	return
}

// CreateAuthToken Note: in order for the deadline to be valid, it needs to be at most 8 hours from now.
// It's recommended that it'd be at most 7:55, as differences in clock times could make this
// invalid. Still, this endpoint does not enforce that so users can generate the auth tokens in advance.

//export CreateAuthToken
func CreateAuthToken(cDeadline C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	deadline := int64(cDeadline)
	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err = txClient.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return
	}

	return
}

/// === API Key related ops ===

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	// Note: The ChangePubKey TX needs to be signed by the API key that's being changed to as well.
	//       Because of that, there's no reason to add the params for apiKeyIndex & accountIndex, because this
	//       version of the SDK doesn't have support for multiple signers.
	//       Even if it'd had, the flow would look something like this:
	//       - first you select which client you're sending the TX from
	//       - then we use the ApiKeyIndex & AccountIndex from that client
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

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

	txInfo := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}
	tx, err := txClient.GetChangePubKeyTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	// === manually add MessageToSign to the response:
	// - marshal the tx
	// - unmarshal it into a generic map
	// - add the new field
	// - marshal it again
	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
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
	seedP := &seed
	if seed == "" {
		seedP = nil
	}

	key := curve.SampleScalar(seedP)

	publicKeyStr = hexutil.Encode(schnorr.SchnorrPkFromSk(key).ToLittleEndianBytes())
	privateKeyStr = hexutil.Encode(key.ToLittleEndianBytes())

	return
}

/// === Order related ops ===

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
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
	tx, err := txClient.GetCreateOrderTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignCreateGroupedOrders
func SignCreateGroupedOrders(cGroupingType C.uint8_t, cOrders *C.CreateOrderTxReq, cLen C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	length := int(cLen)
	orders := make([]*types.CreateOrderTxReq, length)
	size := unsafe.Sizeof(*cOrders)

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

	txInfo, err := txClient.GetCreateGroupedOrdersTransaction(req, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	marketIndex := uint8(cMarketIndex)
	orderIndex := int64(cOrderIndex)

	txInfo := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}
	tx, err := txClient.GetCancelOrderTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	timeInForce := uint8(cTimeInForce)
	t := int64(cTime)

	txInfo := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        t,
	}
	tx, err := txClient.GetCancelAllOrdersTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	marketIndex := uint8(cMarketIndex)
	index := int64(cIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	triggerPrice := uint32(cTriggerPrice)

	txInfo := &types.ModifyOrderTxReq{
		MarketIndex:  marketIndex,
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        price,
		TriggerPrice: triggerPrice,
	}
	tx, err := txClient.GetModifyOrderTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

/// === Leverage related ops ===

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	marketIndex := uint8(cMarketIndex)
	initialMarginFraction := uint16(cInitialMarginFraction)
	marginMode := uint8(cMarginMode)

	txInfo := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}
	tx, err := txClient.GetUpdateLeverageTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
	var err error
	var txInfoStr string
	defer func() {
		if r := recover(); r != nil {
			wrapErr(fmt.Errorf("panic: %v", r))
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	marketIndex := uint8(cMarketIndex)
	usdcAmount := int64(cUSDCAmount)
	direction := uint8(cDirection)

	txInfo := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}
	tx, err := txClient.GetUpdateMarginTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}
	txInfoStr = string(txInfoBytes)

	return ret
}

/// === Transfer related ops ===

//export SignWithdraw
func SignWithdraw(cUSDCAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	usdcAmount := uint64(cUSDCAmount)

	txInfo := types.WithdrawTxReq{
		USDCAmount: usdcAmount,
	}
	tx, err := txClient.GetWithdrawTransaction(&txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cUSDCAmount C.longlong, cFee C.longlong, cMemo *C.char, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	toAccountIndex := int64(cToAccountIndex)
	usdcAmount := int64(cUSDCAmount)
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

	txInfo := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		USDCAmount:     usdcAmount,
		Fee:            fee,
		Memo:           memo,
	}
	tx, err := txClient.GetTransferTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	obj := make(map[string]interface{})
	err = json.Unmarshal(txInfoBytes, &obj)
	obj["MessageToSign"] = tx.GetL1SignatureBody()
	txInfoBytes, err = json.Marshal(obj)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

/// === Sub accounts & pools related ops ===

//export SignCreateSubAccount
func SignCreateSubAccount(cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	tx, err := txClient.GetCreateSubAccountTransaction(getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	operatorFee := int64(cOperatorFee)
	initialTotalShares := int64(cInitialTotalShares)
	minOperatorShareRate := int64(cMinOperatorShareRate)

	txInfo := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}
	tx, err := txClient.GetCreatePublicPoolTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	status := uint8(cStatus)
	operatorFee := int64(cOperatorFee)
	minOperatorShareRate := int64(cMinOperatorShareRate)

	txInfo := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      publicPoolIndex,
		Status:               status,
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}
	tx, err := txClient.GetUpdatePublicPoolTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)

	txInfo := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	tx, err := txClient.GetMintSharesTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.StrOrErr) {
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

	txClient, err := getTxClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)

	txInfo := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	tx, err := txClient.GetBurnSharesTransaction(txInfo, getOps(cNonce))
	if err != nil {
		return
	}

	txInfoBytes, err := json.Marshal(tx)
	if err != nil {
		return
	}

	txInfoStr = string(txInfoBytes)
	return
}

func main() {}

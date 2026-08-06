package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/elliottech/lighter-go/client"
	"github.com/elliottech/lighter-go/client/http"
	"github.com/elliottech/lighter-go/signer"
	"github.com/elliottech/lighter-go/types"
	"github.com/elliottech/lighter-go/types/txtypes"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

/*
#cgo CFLAGS: -std=c11
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
	void* data;
	int64_t len;
	char* err;
} SignedTxBatchResponse;

typedef struct {
	char* privateKey;
	char* publicKey;
	char* err;
} ApiKeyResponse;

typedef struct {
    int16_t MarketIndex;
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

var chainId uint32

func wrapErr(err any) *C.char {
	if err == nil {
		return nil
	}
	return C.CString(fmt.Sprintf("%v", err))
}

func messageToSign(txInfo txtypes.TxInfo) string {
	switch typed := txInfo.(type) {
	case *txtypes.L2ChangePubKeyTxInfo:
		return typed.GetL1SignatureBody()
	case *txtypes.L2TransferTxInfo:
		return typed.GetL1SignatureBody(chainId)
	case *txtypes.L2ApproveIntegratorTxInfo:
		return typed.GetL1SignatureBody(chainId)
	default:
		return ""
	}
}

func signedTxResponseErr(err any) C.SignedTxResponse {
	return C.SignedTxResponse{err: wrapErr(err)}
}

func signedTxResponsePanic(err any) C.SignedTxResponse {
	return signedTxResponseErr(fmt.Errorf("panic: %v", err))
}

func signedTxBatchResponseErr(err any) C.SignedTxBatchResponse {
	return C.SignedTxBatchResponse{err: wrapErr(err)}
}

type packedSignedTxResponse struct {
	txType uint8
	txInfo string
	txHash string
	err    string
}

func convertTxInfoToPackedResponse(txInfo txtypes.TxInfo, err error) packedSignedTxResponse {
	if err != nil {
		return packedSignedTxResponse{err: err.Error()}
	}
	if txInfo == nil {
		return packedSignedTxResponse{err: "nil transaction info"}
	}

	txInfoStr, err := txInfo.GetTxInfo()
	if err != nil {
		return packedSignedTxResponse{err: err.Error()}
	}
	return packedSignedTxResponse{
		txType: uint8(txInfo.GetTxType()),
		txInfo: txInfoStr,
		txHash: txInfo.GetTxHash(),
	}
}

const packedSignedTxHeaderSize = 16

func marshalSignedTxResponses(results []packedSignedTxResponse) []byte {
	totalSize := 4
	for _, result := range results {
		totalSize += packedSignedTxHeaderSize + len(result.txInfo) + len(result.txHash) + len(result.err)
	}

	packed := make([]byte, totalSize)
	binary.LittleEndian.PutUint32(packed, uint32(len(results)))
	offset := 4
	for _, result := range results {
		packed[offset] = result.txType
		binary.LittleEndian.PutUint32(packed[offset+4:], uint32(len(result.txInfo)))
		binary.LittleEndian.PutUint32(packed[offset+8:], uint32(len(result.txHash)))
		binary.LittleEndian.PutUint32(packed[offset+12:], uint32(len(result.err)))
		offset += packedSignedTxHeaderSize
		offset += copy(packed[offset:], result.txInfo)
		offset += copy(packed[offset:], result.txHash)
		offset += copy(packed[offset:], result.err)
	}

	return packed
}

func packSignedTxResponses(results []packedSignedTxResponse) C.SignedTxBatchResponse {
	packed := marshalSignedTxResponses(results)
	return C.SignedTxBatchResponse{data: C.CBytes(packed), len: C.int64_t(len(packed))}
}

func convertTxInfoToResponse(txInfo txtypes.TxInfo, err error) C.SignedTxResponse {
	if err != nil {
		return signedTxResponseErr(err)
	}
	if txInfo == nil {
		return signedTxResponseErr("nil transaction info")
	}

	txInfoStr, err := txInfo.GetTxInfo()
	if err != nil {
		return signedTxResponseErr(err)
	}

	resp := C.SignedTxResponse{
		txType: C.uint8_t(txInfo.GetTxType()),
		txInfo: C.CString(txInfoStr),
		txHash: C.CString(txInfo.GetTxHash()),
	}

	if msg := messageToSign(txInfo); msg != "" {
		resp.messageToSign = C.CString(msg)
	}

	return resp
}

// getClient returns the go TxClient from the specified cApiKeyIndex and cAccountIndex
func getClient(cApiKeyIndex C.int, cAccountIndex C.longlong) (*client.TxClient, error) {
	apiKeyIndex := uint8(cApiKeyIndex)
	accountIndex := int64(cAccountIndex)
	return client.GetClient(apiKeyIndex, accountIndex)
}

func CreateTxAttributesFromIsSkipNonce(skipNonce uint8) *types.L2TxAttributes {
	attr := types.L2TxAttributes{}
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func CreateIntegratorTxAttributes(integratorAccountIndex int64, integratorTakerFee uint32, integratorMakerFee uint32, skipNonce uint8, selfTradeBehaviorMode uint8, selfTradeEqualityMode uint8) *types.L2TxAttributes {
	attr := types.L2TxAttributes{}
	if integratorAccountIndex != txtypes.NilIntegratorIndex {
		attr.IntegratorAccountIndex = &integratorAccountIndex
	}
	if integratorTakerFee != txtypes.NilIntegratorTakerFee {
		attr.IntegratorTakerFee = &integratorTakerFee
	}
	if integratorMakerFee != txtypes.NilIntegratorMakerFee {
		attr.IntegratorMakerFee = &integratorMakerFee
	}
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	if selfTradeBehaviorMode != txtypes.SelfTradeBehaviorExpireMaker {
		attr.SelfTradeBehaviorMode = &selfTradeBehaviorMode
	}
	if selfTradeEqualityMode != txtypes.SelfTradeEqualityAccountIndex {
		attr.SelfTradeEqualityMode = &selfTradeEqualityMode
	}
	return &attr
}

func CreateCancelAllTxAttributes(cancelAllMarketIndex int16, skipNonce uint8) *types.L2TxAttributes {
	attr := types.L2TxAttributes{}
	if cancelAllMarketIndex != txtypes.NilMarketIndex {
		attr.CancelAllMarketIndex = &cancelAllMarketIndex
	}
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func getTransactOpts(cSkipNonce C.uint8_t, cNonce C.longlong) *types.TransactOpts {
	nonce := int64(cNonce)
	txAttributes := CreateTxAttributesFromIsSkipNonce(uint8(cSkipNonce))
	return &types.TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func getIntegratorTransactOptsAll(cIntegratorAccountIndex C.longlong, cIntegratorTakerFee C.int, cIntegratorMakerFee C.int, cSkipNonce C.uint8_t, cNonce C.longlong, cSelfTradeBehaviorMode C.uint8_t, cSelfTradeEqualityMode C.uint8_t) *types.TransactOpts {
	nonce := int64(cNonce)
	integratorAccountIndex := int64(cIntegratorAccountIndex)
	integratorTakerFee := uint32(cIntegratorTakerFee)
	integratorMakerFee := uint32(cIntegratorMakerFee)
	skipNonce := uint8(cSkipNonce)
	selfTradeBehaviorMode := uint8(cSelfTradeBehaviorMode)
	selfTradeEqualityMode := uint8(cSelfTradeEqualityMode)
	txAttributes := CreateIntegratorTxAttributes(integratorAccountIndex, integratorTakerFee, integratorMakerFee, skipNonce, selfTradeBehaviorMode, selfTradeEqualityMode)
	return &types.TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func getCancelAllTransactOpts(cCancelAllMarketIndex C.int, cSkipNonce C.uint8_t, cNonce C.longlong) *types.TransactOpts {
	nonce := int64(cNonce)
	cancelAllMarketIndex := int16(cCancelAllMarketIndex)
	skipNonce := uint8(cSkipNonce)
	txAttributes := CreateCancelAllTxAttributes(cancelAllMarketIndex, skipNonce)
	return &types.TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

//export GenerateAPIKey
func GenerateAPIKey() (ret C.ApiKeyResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = C.ApiKeyResponse{err: wrapErr(fmt.Errorf("panic: %v", r))}
		}
	}()

	privateKeyStr, publicKeyStr, err := client.GenerateAPIKey()
	if err != nil {
		return C.ApiKeyResponse{err: wrapErr(err)}
	}

	return C.ApiKeyResponse{
		privateKey: C.CString(privateKeyStr),
		publicKey:  C.CString(publicKeyStr),
	}
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
	chainId = uint32(cChainId)
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

//export PrepareSignerNonces
func PrepareSignerNonces(cCount C.int) (ret *C.char) {
	defer func() {
		if r := recover(); r != nil {
			ret = wrapErr(fmt.Errorf("panic: %v", r))
		}
	}()
	return wrapErr(signer.PrepareSchnorrNonces(int(cCount)))
}

//export SignChangePubKey
func SignChangePubKey(cPubKey *C.char, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	pubKeyStr := C.GoString(cPubKey)
	pubKeyBytes, err := hexutil.Decode(pubKeyStr)
	if err != nil {
		return signedTxResponseErr(err)
	}
	if len(pubKeyBytes) != 40 {
		return signedTxResponseErr(fmt.Errorf("invalid pub key length. expected 40 but got %v", len(pubKeyBytes)))
	}
	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	tx := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetChangePubKeyTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignCreateOrder
func SignCreateOrder(cMarketIndex C.int, cClientOrderIndex C.longlong, cBaseAmount C.longlong, cPrice C.int, cIsAsk C.int, cOrderType C.int, cTimeInForce C.int, cReduceOnly C.int, cTriggerPrice C.int, cOrderExpiry C.longlong, cIntegratorAccountIndex C.longlong, cIntegratorTakerFee C.int, cIntegratorMakerFee C.int, cSelfTradeBehaviorMode C.uint8_t, cSelfTradeEqualityMode C.uint8_t, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	marketIndex := int16(cMarketIndex)
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
		orderExpiry = time.Now().Add(defaultOrderExpiryDuration).UnixMilli()
	}

	tx := &types.CreateOrderTxReq{
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
	ops := getIntegratorTransactOptsAll(cIntegratorAccountIndex, cIntegratorTakerFee, cIntegratorMakerFee, cSkipNonce, cNonce, cSelfTradeBehaviorMode, cSelfTradeEqualityMode)

	txInfo, err := c.GetCreateOrderTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

const (
	maxCreateOrderBatch        = 10_000
	defaultOrderExpiryDuration = 28 * 24 * time.Hour
)

type createOrderBatchOptions struct {
	integratorAccountIndex int64
	integratorTakerFee     uint32
	integratorMakerFee     uint32
	selfTradeBehaviorMode  uint8
	selfTradeEqualityMode  uint8
	skipNonce              uint8
	firstNonce             int64
	requestedAPIKeyIndex   uint8
	requestedAccountIndex  int64
}

func signCreateOrdersBatch(orders []types.CreateOrderTxReq, options createOrderBatchOptions) ([]packedSignedTxResponse, error) {
	length := len(orders)
	if length <= 0 || length > maxCreateOrderBatch {
		return nil, fmt.Errorf("batch length must be between 1 and %d", maxCreateOrderBatch)
	}
	if options.firstNonce < 0 {
		return nil, fmt.Errorf("batch signing requires an explicit non-negative first nonce; it does not perform network I/O")
	}
	if options.firstNonce > math.MaxInt64-int64(length-1) {
		return nil, fmt.Errorf("batch nonce range overflows int64")
	}

	c, err := client.GetClient(options.requestedAPIKeyIndex, options.requestedAccountIndex)
	if err != nil {
		return nil, err
	}

	results := make([]packedSignedTxResponse, length)
	keyManager := c.GetKeyManager()
	chainID := c.GetChainId()
	accountIndex := c.GetAccountIndex()
	apiKeyIndex := c.GetApiKeyIndex()
	expiredAt := time.Now().Add(client.DefaultExpireTime).UnixMilli()
	defaultOrderExpiry := time.Now().Add(defaultOrderExpiryDuration).UnixMilli()
	txAttributes := CreateIntegratorTxAttributes(
		options.integratorAccountIndex,
		options.integratorTakerFee,
		options.integratorMakerFee,
		options.skipNonce,
		options.selfTradeBehaviorMode,
		options.selfTradeEqualityMode,
	)

	// A batch call uses up to GOMAXPROCS workers and can therefore occupy every
	// available Go execution slot until the batch completes.
	workerCount := min(length, runtime.GOMAXPROCS(0))
	signRange := func(start, end int) {
		for i := start; i < end; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						results[i] = packedSignedTxResponse{err: fmt.Sprintf("panic: %v", r)}
					}
				}()

				order := orders[i]
				if order.OrderExpiry == -1 {
					order.OrderExpiry = defaultOrderExpiry
				}
				nonce := options.firstNonce + int64(i)
				ops := &types.TransactOpts{
					FromAccountIndex: &accountIndex,
					ApiKeyIndex:      &apiKeyIndex,
					ExpiredAt:        expiredAt,
					Nonce:            &nonce,
					TxAttributes:     txAttributes,
				}
				txInfo, signErr := types.ConstructCreateOrderTx(
					keyManager, chainID, &order, ops,
				)
				results[i] = convertTxInfoToPackedResponse(txInfo, signErr)
			}()
		}
	}

	if workerCount == 1 {
		signRange(0, length)
	} else {
		var workers sync.WaitGroup
		workers.Add(workerCount)
		for worker := 0; worker < workerCount; worker++ {
			start := worker * length / workerCount
			end := (worker + 1) * length / workerCount
			go func() {
				defer workers.Done()
				signRange(start, end)
			}()
		}
		workers.Wait()
	}
	return results, nil
}

//export SignCreateOrdersBatch
func SignCreateOrdersBatch(cOrders *C.CreateOrderTxReq, cLen C.int, cIntegratorAccountIndex C.longlong, cIntegratorTakerFee C.int, cIntegratorMakerFee C.int, cSelfTradeBehaviorMode C.uint8_t, cSelfTradeEqualityMode C.uint8_t, cSkipNonce C.uint8_t, cFirstNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxBatchResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxBatchResponseErr(fmt.Errorf("panic: %v", r))
		}
	}()

	length := int(cLen)
	if length <= 0 || length > maxCreateOrderBatch {
		return signedTxBatchResponseErr(fmt.Errorf("batch length must be between 1 and %d", maxCreateOrderBatch))
	}
	if cOrders == nil {
		return signedTxBatchResponseErr("batch input pointer is required")
	}

	cOrdersSlice := unsafe.Slice(cOrders, length)
	orders := make([]types.CreateOrderTxReq, length)
	for i, order := range cOrdersSlice {
		orders[i] = types.CreateOrderTxReq{
			MarketIndex:      int16(order.MarketIndex),
			ClientOrderIndex: int64(order.ClientOrderIndex),
			BaseAmount:       int64(order.BaseAmount),
			Price:            uint32(order.Price),
			IsAsk:            uint8(order.IsAsk),
			Type:             uint8(order.Type),
			TimeInForce:      uint8(order.TimeInForce),
			ReduceOnly:       uint8(order.ReduceOnly),
			TriggerPrice:     uint32(order.TriggerPrice),
			OrderExpiry:      int64(order.OrderExpiry),
		}
	}
	results, err := signCreateOrdersBatch(orders, createOrderBatchOptions{
		integratorAccountIndex: int64(cIntegratorAccountIndex),
		integratorTakerFee:     uint32(cIntegratorTakerFee),
		integratorMakerFee:     uint32(cIntegratorMakerFee),
		selfTradeBehaviorMode:  uint8(cSelfTradeBehaviorMode),
		selfTradeEqualityMode:  uint8(cSelfTradeEqualityMode),
		skipNonce:              uint8(cSkipNonce),
		firstNonce:             int64(cFirstNonce),
		requestedAPIKeyIndex:   uint8(cApiKeyIndex),
		requestedAccountIndex:  int64(cAccountIndex),
	})
	if err != nil {
		return signedTxBatchResponseErr(err)
	}
	return packSignedTxResponses(results)
}

//export SignCreateGroupedOrders
func SignCreateGroupedOrders(cGroupingType C.uint8_t, cOrders *C.CreateOrderTxReq, cLen C.int, cIntegratorAccountIndex C.longlong, cIntegratorTakerFee C.int, cIntegratorMakerFee C.int, cSelfTradeBehaviorMode C.uint8_t, cSelfTradeEqualityMode C.uint8_t, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	length := int(cLen)
	orders := make([]*types.CreateOrderTxReq, length)
	size := unsafe.Sizeof(*cOrders)
	for i := 0; i < length; i++ {
		order := (*C.CreateOrderTxReq)(unsafe.Pointer(uintptr(unsafe.Pointer(cOrders)) + uintptr(i)*uintptr(size)))

		orderExpiry := int64(order.OrderExpiry)
		if orderExpiry == -1 {
			orderExpiry = time.Now().Add(defaultOrderExpiryDuration).UnixMilli()
		}

		orders[i] = &types.CreateOrderTxReq{
			MarketIndex:      int16(order.MarketIndex),
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

	tx := &types.CreateGroupedOrdersTxReq{
		GroupingType: uint8(cGroupingType),
		Orders:       orders,
	}
	ops := getIntegratorTransactOptsAll(cIntegratorAccountIndex, cIntegratorTakerFee, cIntegratorMakerFee, cSkipNonce, cNonce, cSelfTradeBehaviorMode, cSelfTradeEqualityMode)

	txInfo, err := c.GetCreateGroupedOrdersTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignCancelOrder
func SignCancelOrder(cMarketIndex C.int, cOrderIndex C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	marketIndex := int16(cMarketIndex)
	orderIndex := int64(cOrderIndex)

	tx := &types.CancelOrderTxReq{
		MarketIndex: marketIndex,
		Index:       orderIndex,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetCancelOrderTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignWithdraw
func SignWithdraw(cAssetIndex C.int, cRouteType C.int, cAmount C.ulonglong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	assetIndex := int16(cAssetIndex)
	routeType := uint8(cRouteType)
	amount := uint64(cAmount)

	tx := &types.WithdrawTxReq{
		AssetIndex: assetIndex,
		RouteType:  routeType,
		Amount:     amount,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetWithdrawTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignCreateSubAccount
func SignCreateSubAccount(cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetCreateSubAccountTransaction(ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignCancelAllOrders
func SignCancelAllOrders(cTimeInForce C.int, cTime C.longlong, cCancelAllMarketIndex C.int, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	timeInForce := uint8(cTimeInForce)
	t := int64(cTime)

	tx := &types.CancelAllOrdersTxReq{
		TimeInForce: timeInForce,
		Time:        t,
	}
	ops := getCancelAllTransactOpts(cCancelAllMarketIndex, cSkipNonce, cNonce)

	txInfo, err := c.GetCancelAllOrdersTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignModifyOrder
func SignModifyOrder(cMarketIndex C.int, cIndex C.longlong, cBaseAmount C.longlong, cPrice C.longlong, cTriggerPrice C.longlong, cIntegratorAccountIndex C.longlong, cIntegratorTakerFee C.int, cIntegratorMakerFee C.int, cSelfTradeBehaviorMode C.uint8_t, cSelfTradeEqualityMode C.uint8_t, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	marketIndex := int16(cMarketIndex)
	index := int64(cIndex)
	baseAmount := int64(cBaseAmount)
	price := uint32(cPrice)
	triggerPrice := uint32(cTriggerPrice)

	tx := &types.ModifyOrderTxReq{
		MarketIndex:  marketIndex,
		Index:        index,
		BaseAmount:   baseAmount,
		Price:        price,
		TriggerPrice: triggerPrice,
	}
	ops := getIntegratorTransactOptsAll(cIntegratorAccountIndex, cIntegratorTakerFee, cIntegratorMakerFee, cSkipNonce, cNonce, cSelfTradeBehaviorMode, cSelfTradeEqualityMode)

	txInfo, err := c.GetModifyOrderTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignTransfer
func SignTransfer(cToAccountIndex C.longlong, cAssetIndex C.int16_t, cFromRouteType, cToRouteType C.uint8_t, cAmount, cUsdcFee C.longlong, cMemo *C.char, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	toAccountIndex := int64(cToAccountIndex)
	assetIndex := int16(cAssetIndex)
	fromRouteType := uint8(cFromRouteType)
	toRouteType := uint8(cToRouteType)
	amount := int64(cAmount)
	usdcFee := int64(cUsdcFee)
	memo := [32]byte{}
	memoStr := C.GoString(cMemo)
	if len(memoStr) == 66 {
		if memoStr[0:2] == "0x" {
			memoStr = memoStr[2:66]
		} else {
			return signedTxResponseErr(fmt.Sprintf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr)))
		}
	}

	// assume hex encoded here
	if len(memoStr) == 64 {
		b, err := hex.DecodeString(memoStr)
		if err != nil {
			return signedTxResponseErr(fmt.Sprintf("failed to decode hex string. err: %v", err))
		}

		for i := 0; i < 32; i += 1 {
			memo[i] = b[i]
		}
	} else if len(memoStr) == 32 {
		for i := 0; i < 32; i++ {
			memo[i] = byte(memoStr[i])
		}
	} else {
		return signedTxResponseErr(fmt.Sprintf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr)))
	}

	tx := &types.TransferTxReq{
		ToAccountIndex: toAccountIndex,
		AssetIndex:     assetIndex,
		FromRouteType:  fromRouteType,
		ToRouteType:    toRouteType,
		Amount:         amount,
		USDCFee:        usdcFee,
		Memo:           memo,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetTransferTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignCreatePublicPool
func SignCreatePublicPool(cOperatorFee C.longlong, cInitialTotalShares C.int, cMinOperatorShareRate C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	operatorFee := int64(cOperatorFee)
	initialTotalShares := int64(cInitialTotalShares)
	minOperatorShareRate := uint16(cMinOperatorShareRate)

	tx := &types.CreatePublicPoolTxReq{
		OperatorFee:          operatorFee,
		InitialTotalShares:   initialTotalShares,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetCreatePublicPoolTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignUpdatePublicPool
func SignUpdatePublicPool(cPublicPoolIndex C.longlong, cStatus C.int, cOperatorFee C.longlong, cMinOperatorShareRate C.int, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	status := uint8(cStatus)
	operatorFee := int64(cOperatorFee)
	minOperatorShareRate := uint16(cMinOperatorShareRate)

	tx := &types.UpdatePublicPoolTxReq{
		PublicPoolIndex:      publicPoolIndex,
		Status:               status,
		OperatorFee:          operatorFee,
		MinOperatorShareRate: minOperatorShareRate,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUpdatePublicPoolTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignMintShares
func SignMintShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)

	tx := &types.MintSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetMintSharesTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignBurnShares
func SignBurnShares(cPublicPoolIndex C.longlong, cShareAmount C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	publicPoolIndex := int64(cPublicPoolIndex)
	shareAmount := int64(cShareAmount)

	tx := &types.BurnSharesTxReq{
		PublicPoolIndex: publicPoolIndex,
		ShareAmount:     shareAmount,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetBurnSharesTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignUpdateLeverage
func SignUpdateLeverage(cMarketIndex C.int, cInitialMarginFraction C.int, cMarginMode C.int, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	marketIndex := int16(cMarketIndex)
	initialMarginFraction := uint16(cInitialMarginFraction)
	marginMode := uint8(cMarginMode)

	tx := &types.UpdateLeverageTxReq{
		MarketIndex:           marketIndex,
		InitialMarginFraction: initialMarginFraction,
		MarginMode:            marginMode,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUpdateLeverageTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
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
		return C.StrOrErr{err: wrapErr(err)}
	}

	deadline := int64(cDeadline)
	if deadline == 0 {
		deadline = time.Now().Add(time.Hour * 7).Unix()
	}

	authToken, err := c.GetAuthToken(time.Unix(deadline, 0))
	if err != nil {
		return C.StrOrErr{err: wrapErr(err)}
	}

	return C.StrOrErr{str: C.CString(authToken)}
}

//export SignUpdateMargin
func SignUpdateMargin(cMarketIndex C.int, cUSDCAmount C.longlong, cDirection C.int, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	marketIndex := int16(cMarketIndex)
	usdcAmount := int64(cUSDCAmount)
	direction := uint8(cDirection)

	tx := &types.UpdateMarginTxReq{
		MarketIndex: marketIndex,
		USDCAmount:  usdcAmount,
		Direction:   direction,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUpdateMarginTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignStakeAssets
func SignStakeAssets(cStakingPoolIndex C.longlong, cShareAmount C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	stakingPoolIndex := int64(cStakingPoolIndex)
	shareAmount := int64(cShareAmount)

	tx := &types.StakeAssetsTxReq{
		StakingPoolIndex: stakingPoolIndex,
		ShareAmount:      shareAmount,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetStakeAssetsTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignUnstakeAssets
func SignUnstakeAssets(cStakingPoolIndex C.longlong, cShareAmount C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	stakingPoolIndex := int64(cStakingPoolIndex)
	shareAmount := int64(cShareAmount)

	tx := &types.UnstakeAssetsTxReq{
		StakingPoolIndex: stakingPoolIndex,
		ShareAmount:      shareAmount,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUnstakeAssetsTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignApproveIntegrator
func SignApproveIntegrator(cIntegratorIndex C.longlong, cMaxPerpsTakerFee C.uint32_t, cMaxPerpsMakerFee C.uint32_t, cMaxSpotTakerFee C.uint32_t, cMaxSpotMakerFee C.uint32_t, cApprovalExpiry C.longlong, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()
	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	IntegratorIndex := int64(cIntegratorIndex)
	MaxPerpsMakerFee := uint32(cMaxPerpsMakerFee)
	MaxPerpsTakerFee := uint32(cMaxPerpsTakerFee)
	MaxSpotMakerFee := uint32(cMaxSpotMakerFee)
	MaxSpotTakerFee := uint32(cMaxSpotTakerFee)
	ApprovalExpiry := int64(cApprovalExpiry)

	tx := &types.ApproveIntegratorTxReq{
		IntegratorAccountIndex: IntegratorIndex,
		MaxPerpsTakerFee:       MaxPerpsTakerFee,
		MaxPerpsMakerFee:       MaxPerpsMakerFee,
		MaxSpotTakerFee:        MaxSpotTakerFee,
		MaxSpotMakerFee:        MaxSpotMakerFee,
		ApprovalExpiry:         ApprovalExpiry,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)
	txInfo, err := c.GetApproveIntegratorTx(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignUpdateAccountConfig
func SignUpdateAccountConfig(cAccountTradingMode C.uint8_t, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	accountTradingMode := uint8(cAccountTradingMode)

	tx := &types.UpdateAccountConfigTxReq{
		AccountTradingMode: accountTradingMode,
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUpdateAccountConfigTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export SignUpdateAccountAssetConfig
func SignUpdateAccountAssetConfig(cAssetIndex C.int16_t, cAssetMarginMode C.uint8_t, cSkipNonce C.uint8_t, cNonce C.longlong, cApiKeyIndex C.int, cAccountIndex C.longlong) (ret C.SignedTxResponse) {
	defer func() {
		if r := recover(); r != nil {
			ret = signedTxResponsePanic(r)
		}
	}()

	c, err := getClient(cApiKeyIndex, cAccountIndex)
	if err != nil {
		return signedTxResponseErr(err)
	}

	tx := &types.UpdateAccountAssetConfigTxReq{
		AssetIndex:      int16(cAssetIndex),
		AssetMarginMode: uint8(cAssetMarginMode),
	}
	ops := getTransactOpts(cSkipNonce, cNonce)

	txInfo, err := c.GetUpdateAccountAssetConfigTransaction(tx, ops)
	return convertTxInfoToResponse(txInfo, err)
}

//export Free
func Free(ptr unsafe.Pointer) {
	C.free(ptr)
}

func main() {}

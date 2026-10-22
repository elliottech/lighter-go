//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/elliottech/lighter-go/types/txtypes"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	c := make(chan struct{}, 0)

	println("Webassembly Started")

	register("_lighterTest", LighterTest)

	register("_createClient", CreateClient)
	register("_createClientByPrv", CreateClientByPrv)

	// Transactions
	register("_signChangePubKey", SignChangePubKey)
	register("_signRevokePubKey", SignRevokePubKey)
	register("_signCreateSubAccount", SignCreateSubAccount)
	register("_signCreatePublicPool", SignCreatePublicPool)
	register("_signUpdatePublicPool", SignUpdatePublicPool)
	register("_signCreateOrder", SignCreateOrder)
	register("_signCancelOrder", SignCancelOrder)
	register("_signWithdraw", SignWithdraw)
	register("_signCancelAllOrders", SignCancelAllOrders)
	register("_signModifyOrder", SignModifyOrder)
	register("_signTransfer", SignTransfer)
	register("_signMintShares", SignMintShares)
	register("_signBurnShares", SignBurnShares)
	register("_signUpdateLeverage", SignUpdateLeverage)
	register("_signUpdateMargin", SignUpdateMargin)
	register("_signUpdateAccountConfig", SignUpdateAccountConfig)
	register("_signUpdateAccountAssetConfig", SignUpdateAccountAssetConfig)
	register("_signCreateGroupedOrders", SignCreateGroupedOrders)
	register("_signApproveIntegrator", SignApproveIntegrator)
	register("_getApproveIntegratorTransaction", GetApproveIntegratorTransaction)
	register("_getChangePubKeyTransaction", GetChangePubKeyTransaction)
	register("_getRevokePubKeyTransaction", GetRevokePubKeyTransaction)
	register("_getTransferTransaction", GetTransferTransaction)
	register("_getAirdropAllocationMessage", GetAirdropAllocationMessage)
	register("_createAuthToken", CreateAuthToken)
	register("_signStakeAssets", SignStakeAssets)
	register("_signUnstakeAssets", SignUnstakeAssets)
	register("_signBinaryOptionsOperatorCreateMarket", SignBinaryOptionsOperatorCreateMarket)
	register("_signBinaryOptionsOperatorUpdateMarket", SignBinaryOptionsOperatorUpdateMarket)
	register("_signBinaryOptionsOperatorSettleMarket", SignBinaryOptionsOperatorSettleMarket)
	js.Global().Set("_signerVersion", 2)
	register("_resetClients", ResetClients)
	register("_selectClient", SelectClient)
	<-c
}

var clients = make(map[int64]*LightClient)

func LighterTest(this js.Value, args []js.Value) any {
	return js.Global().Get("Promise").Call("resolve", "Hello, World!")
}

func CreateClient(this js.Value, args []js.Value) any {
	seed := string(checkedString(args[0]))
	seed = strings.TrimPrefix(seed, "0x")
	chainId := checkedUint32(checkedInt(args[1]))
	accountIndex := int64(checkedInt(args[2]))
	nonce := int64(checkedInt(args[3]))
	apiKeyIndex := checkedUint8(checkedInt(args[4]))
	skipNonce := len(args) > 5 && checkedBool(args[5])

	return promiseResult(func(resolve js.Value) {
		sdkClient, err := NewLightClient(seed, chainId)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		sdkClient.SetAccountIndex(accountIndex)
		sdkClient.SetApiKeyIndex(apiKeyIndex)
		sdkClient.SetSkipNonce(skipNonce)

		pub := sdkClient.KeyManager().PubKeyBytes()
		txInfo := &ChangePubKeyReq{PubKey: pub}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		installClient(accountIndex, sdkClient)

		response := map[string]any{
			"success":       sdkClient != nil,
			"pk":            common.Bytes2Hex(pub[:]),
			"prv":           common.Bytes2Hex(sdkClient.KeyManager().PrvKeyBytes()),
			"pubKeySuccess": true,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func CreateClientByPrv(this js.Value, args []js.Value) any {
	prv := string(checkedString(args[0]))
	prv = strings.TrimPrefix(prv, "0x")
	chainId := checkedUint32(checkedInt(args[1]))
	accountIndex := int64(checkedInt(args[2]))
	nonce := int64(checkedInt(args[3]))
	apiKeyIndex := checkedUint8(checkedInt(args[4]))
	skipNonce := len(args) > 5 && checkedBool(args[5])

	return promiseResult(func(resolve js.Value) {
		sdkClient, err := NewLightClientByPrv(prv, chainId)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		sdkClient.SetAccountIndex(accountIndex)
		sdkClient.SetApiKeyIndex(apiKeyIndex)
		sdkClient.SetSkipNonce(skipNonce)

		pub := sdkClient.KeyManager().PubKeyBytes()
		txInfo := &ChangePubKeyReq{PubKey: pub}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		installClient(accountIndex, sdkClient)

		response := map[string]any{
			"success":       sdkClient != nil,
			"pk":            common.Bytes2Hex(pub[:]),
			"prv":           common.Bytes2Hex(sdkClient.KeyManager().PrvKeyBytes()), // Preserve the public SDK response contract.
			"pubKeySuccess": true,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func GetChangePubKeyTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(checkedInt(args[1]))
	apiKeyIndex := checkedUint8(checkedInt(args[2]))

	return promiseResult(func(resolve js.Value) {
		sdkClient := clients[accountIndex]
		pub := sdkClient.KeyManager().PubKeyBytes()
		txInfo := &ChangePubKeyReq{PubKey: pub}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		response := map[string]any{
			"success":       sdkClient != nil,
			"pk":            common.Bytes2Hex(pub[:]),
			"pubKeySuccess": true,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func GetRevokePubKeyTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(checkedInt(args[1]))
	apiKeyIndex := checkedUint8(checkedInt(args[2]))

	return promiseResult(func(resolve js.Value) {
		sdkClient := clients[accountIndex]
		pub := [40]byte{}
		txInfo := &ChangePubKeyReq{PubKey: pub}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		response := map[string]any{
			"success":       sdkClient != nil,
			"pk":            common.Bytes2Hex(pub[:]),
			"pubKeySuccess": true,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func SignChangePubKey(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(checkedString(args[1]))
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(checkedInt(args[2]))
	apiKeyIndex := checkedUint8(0)
	if len(args) > 3 {
		apiKeyIndex = checkedUint8(checkedInt(args[3]))
	}
	return promiseResult(func(resolve js.Value) {
		txInfo := &ChangePubKeyReq{PubKey: clients[accountIndex].KeyManager().PubKeyBytes()}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		tx, err := clients[accountIndex].GetChangePubKeyTransaction(txInfo, ops, signedMessage)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignRevokePubKey(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(checkedString(args[1]))
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(checkedInt(args[2]))
	apiKeyIndex := checkedUint8(0)
	if len(args) > 3 {
		apiKeyIndex = checkedUint8(checkedInt(args[3]))
	}
	return promiseResult(func(resolve js.Value) {
		pub := [40]byte{}
		txInfo := &ChangePubKeyReq{PubKey: pub}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		tx, err := clients[accountIndex].GetChangePubKeyTransaction(txInfo, ops, signedMessage)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCreateSubAccount(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	nonce := int64(checkedInt(args[1]))

	return promiseResult(func(resolve js.Value) {
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetCreateSubAccountTransaction(ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCreatePublicPool(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	operatorFee := int64(checkedInt(args[1]))
	initialTotalShares := int64(checkedInt(args[2]))
	minOperatorShareRate := checkedUint16(checkedInt(args[3]))
	nonce := int64(checkedInt(args[4]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &CreatePublicPoolTxReq{
			OperatorFee:          operatorFee,
			InitialTotalShares:   initialTotalShares,
			MinOperatorShareRate: minOperatorShareRate,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetCreatePublicPoolTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUpdatePublicPool(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	publicPoolIndex := int64(checkedInt(args[1]))
	status := checkedUint8(checkedInt(args[2]))
	operatorFee := int64(checkedInt(args[3]))
	minOperatorShareRate := checkedUint16(checkedInt(args[4]))
	nonce := int64(checkedInt(args[5]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UpdatePublicPoolTxReq{
			PublicPoolIndex:      publicPoolIndex,
			Status:               status,
			OperatorFee:          operatorFee,
			MinOperatorShareRate: minOperatorShareRate,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUpdatePublicPoolTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCreateOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int64(checkedInt(args[1]))
	clientOrderIndex := int64(checkedInt(args[2]))
	baseAmount, err := strconv.Atoi(checkedString(args[3]))
	if err != nil {
		return errPromise(fmt.Errorf("baseAmount is not an integer"))
	}
	price, err := strconv.Atoi(checkedString(args[4]))
	if err != nil {
		return errPromise(fmt.Errorf("price is not an integer"))
	}
	isAsk := checkedUint8(checkedInt(args[5]))
	orderType := checkedUint8(checkedInt(args[6]))
	timeInForce := checkedUint8(checkedInt(args[7]))
	reduceOnly := checkedUint8(checkedInt(args[8]))
	triggerPrice, err := strconv.Atoi(checkedString(args[9]))
	if err != nil {
		return errPromise(fmt.Errorf("triggerPrice is not an integer"))
	}
	orderExpiry := int64(checkedInt(args[10]))
	nonce := int64(checkedInt(args[11]))
	integratorAccountIndex, integratorTakerFee, integratorMakerFee := integratorAttributes(args, 12)

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	}

	validatedPrice, validatedTriggerPrice := checkedUint32(price), checkedUint32(triggerPrice)
	return promiseResult(func(resolve js.Value) {
		txInfo := &CreateOrderTxReq{
			MarketIndex:      marketIndex,
			ClientOrderIndex: clientOrderIndex,
			BaseAmount:       int64(baseAmount),
			Price:            validatedPrice,
			IsAsk:            isAsk,
			Type:             orderType,
			TimeInForce:      timeInForce,
			ReduceOnly:       reduceOnly,
			TriggerPrice:     validatedTriggerPrice,
			OrderExpiry:      orderExpiry,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.IntegratorAccountIndex = integratorAccountIndex
		ops.IntegratorTakerFee = integratorTakerFee
		ops.IntegratorMakerFee = integratorMakerFee
		tx, err := clients[accountIndex].GetCreateOrderTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCancelOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int64(checkedInt(args[1]))
	orderIndex, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("orderIndex is not an integer"))
	}
	nonce := int64(checkedInt(args[3]))
	return promiseResult(func(resolve js.Value) {
		txInfo := &CancelOrderTxReq{
			MarketIndex: marketIndex,
			Index:       orderIndex,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetCancelOrderTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignWithdraw(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	assetIndex := checkedInt16(checkedInt(args[1]))
	routeType := checkedUint8(checkedInt(args[2]))
	assetAmount, err := parseUnsignedAmount(checkedString(args[3]))
	if err != nil {
		return errPromise(err)
	}
	nonce := int64(checkedInt(args[4]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &WithdrawTxReq{
			AssetIndex: assetIndex,
			RouteType:  routeType,
			Amount:     assetAmount,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetWithdrawTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCancelAllOrders(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	timeInForce := checkedUint8(checkedInt(args[1]))
	time := int64(checkedInt(args[2]))
	nonce := int64(checkedInt(args[3]))
	cancelMarketIndex, hasCancelMarketIndex := cancelAllMarketIndexAttribute(args, 4)

	return promiseResult(func(resolve js.Value) {
		txInfo := &CancelAllOrdersTxReq{
			TimeInForce: timeInForce,
			Time:        time,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		if hasCancelMarketIndex {
			marketIndex := cancelMarketIndex
			ops.CancelAllMarketIndex = &marketIndex
		}
		tx, err := clients[accountIndex].GetCancelAllOrdersTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignModifyOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int64(checkedInt(args[1]))
	orderIndex, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("orderIndex is not an integer"))
	}
	baseAmount := int64(checkedInt(args[3]))
	price := checkedUint32(checkedInt(args[4]))
	triggerPrice := checkedUint32(checkedInt(args[5]))
	nonce := int64(checkedInt(args[6]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &ModifyOrderTxReq{
			MarketIndex:  marketIndex,
			Index:        orderIndex,
			BaseAmount:   baseAmount,
			Price:        price,
			TriggerPrice: triggerPrice,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetModifyOrderTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func GetTransferTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(checkedInt(args[1]))
	apiKeyIndex := checkedUint8(checkedInt(args[2]))
	toAccountIndex := int64(checkedInt(args[3]))
	assetIndex := checkedInt16(checkedInt(args[4]))
	fromRouteType := checkedUint8(checkedInt(args[5]))
	toRouteType := checkedUint8(checkedInt(args[6]))
	amount := int64(checkedInt(args[7]))
	USDCFee := int64(checkedInt(args[8]))

	memo := [32]byte{}
	if len(args) > 9 && !args[9].IsNull() && !args[9].IsUndefined() {
		memoArg := args[9]
		if !memoArg.InstanceOf(js.Global().Get("Array")) || memoArg.Length() > 32 {
			return errPromise(fmt.Errorf("memo must be an array of at most 32 bytes"))
		}
		{
			for i := 0; i < memoArg.Length() && i < 32; i++ {
				memo[i] = checkedByte(memoArg.Index(i))
			}
		}
	}

	return promiseResult(func(resolve js.Value) {
		txInfo := &TransferTxReq{
			ToAccountIndex: toAccountIndex,
			AssetIndex:     assetIndex,
			ToRouteType:    toRouteType,
			FromRouteType:  fromRouteType,
			Amount:         amount,
			USDCFee:        USDCFee,
			Memo:           memo,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		sdkClient := clients[accountIndex]
		txnBody, err := sdkClient.GenerateTransferSignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		response := map[string]any{
			"success":       sdkClient != nil,
			"pubKeySuccess": err == nil,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func SignTransfer(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(checkedString(args[1]))
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(checkedInt(args[2]))
	apiKeyIndex := checkedUint8(checkedInt(args[3]))
	toAccountIndex := int64(checkedInt(args[4]))
	assetIndex := checkedInt16(checkedInt(args[5]))
	fromRouteType := checkedUint8(checkedInt(args[6]))
	toRouteType := checkedUint8(checkedInt(args[7]))
	amount := int64(checkedInt(args[8]))
	USDCFee := int64(checkedInt(args[9]))

	memo := [32]byte{}
	if len(args) > 10 && !args[10].IsNull() && !args[10].IsUndefined() {
		memoArg := args[10]
		if !memoArg.InstanceOf(js.Global().Get("Array")) || memoArg.Length() > 32 {
			return errPromise(fmt.Errorf("memo must be an array of at most 32 bytes"))
		}
		{
			for i := 0; i < memoArg.Length() && i < 32; i++ {
				memo[i] = checkedByte(memoArg.Index(i))
			}
		}
	}

	return promiseResult(func(resolve js.Value) {
		txInfo := &TransferTxReq{
			ToAccountIndex: toAccountIndex,
			AssetIndex:     assetIndex,
			ToRouteType:    toRouteType,
			FromRouteType:  fromRouteType,
			Amount:         amount,
			USDCFee:        USDCFee,
			Memo:           memo,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		tx, err := clients[accountIndex].GetTransferTransaction(txInfo, ops, signedMessage)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func GetApproveIntegratorTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(checkedInt(args[1]))
	apiKeyIndex := checkedUint8(checkedInt(args[2]))
	integratorAccountIndex := int64(checkedInt(args[3]))
	maxPerpsTakerFee := checkedUint32(checkedInt(args[4]))
	maxPerpsMakerFee := checkedUint32(checkedInt(args[5]))
	maxSpotTakerFee := checkedUint32(checkedInt(args[6]))
	maxSpotMakerFee := checkedUint32(checkedInt(args[7]))
	approvalExpiry := int64(checkedInt(args[8]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &ApproveIntegratorTxReq{
			IntegratorAccountIndex: integratorAccountIndex,
			MaxPerpsTakerFee:       maxPerpsTakerFee,
			MaxPerpsMakerFee:       maxPerpsMakerFee,
			MaxSpotTakerFee:        maxSpotTakerFee,
			MaxSpotMakerFee:        maxSpotMakerFee,
			ApprovalExpiry:         approvalExpiry,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		sdkClient := clients[accountIndex]
		txnBody, err := sdkClient.GenerateApproveIntegratorSignBody(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}

		response := map[string]any{
			"success":       sdkClient != nil,
			"pubKeySuccess": err == nil,
			"body":          txnBody,
		}

		resolve.Invoke(response)
	})
}

func SignApproveIntegrator(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(checkedString(args[1]))
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(checkedInt(args[2]))
	apiKeyIndex := checkedUint8(checkedInt(args[3]))
	integratorAccountIndex := int64(checkedInt(args[4]))
	maxPerpsTakerFee := checkedUint32(checkedInt(args[5]))
	maxPerpsMakerFee := checkedUint32(checkedInt(args[6]))
	maxSpotTakerFee := checkedUint32(checkedInt(args[7]))
	maxSpotMakerFee := checkedUint32(checkedInt(args[8]))
	approvalExpiry := int64(checkedInt(args[9]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &ApproveIntegratorTxReq{
			IntegratorAccountIndex: integratorAccountIndex,
			MaxPerpsTakerFee:       maxPerpsTakerFee,
			MaxPerpsMakerFee:       maxPerpsMakerFee,
			MaxSpotTakerFee:        maxSpotTakerFee,
			MaxSpotMakerFee:        maxSpotMakerFee,
			ApprovalExpiry:         approvalExpiry,
		}

		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.ApiKeyIndex = &apiKeyIndex
		tx, err := clients[accountIndex].GetApproveIntegratorTransaction(txInfo, ops, signedMessage)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignStakeAssets(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	stakingPoolIndex := int64(checkedInt(args[1]))
	shareAmount, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(checkedInt(args[3]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &StakeAssetsTxReq{
			StakingPoolIndex: stakingPoolIndex,
			ShareAmount:      shareAmount,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetStakeAssetsTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUnstakeAssets(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	stakingPoolIndex := int64(checkedInt(args[1]))
	shareAmount, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(checkedInt(args[3]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UnstakeAssetsTxReq{
			StakingPoolIndex: stakingPoolIndex,
			ShareAmount:      shareAmount,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUnstakeAssetsTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignMintShares(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	publicPoolIndex := int64(checkedInt(args[1]))
	shareAmount, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(checkedInt(args[3]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &MintSharesTxReq{
			PublicPoolIndex: publicPoolIndex,
			ShareAmount:     shareAmount,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetMintSharesTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignBurnShares(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	publicPoolIndex := int64(checkedInt(args[1]))
	shareAmount, err := strconv.ParseInt(checkedString(args[2]), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(checkedInt(args[3]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &BurnSharesTxReq{
			PublicPoolIndex: publicPoolIndex,
			ShareAmount:     shareAmount,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetBurnSharesTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUpdateAccountConfig(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	accountTradingMode := checkedUint8(checkedInt(args[1]))
	nonce := int64(checkedInt(args[2]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UpdateAccountConfigTxReq{
			AccountTradingMode: accountTradingMode,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUpdateAccountConfigTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUpdateAccountAssetConfig(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	assetIndex := checkedInt16(checkedInt(args[1]))
	assetMarginMode := checkedUint8(checkedInt(args[2]))
	nonce := int64(checkedInt(args[3]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UpdateAccountAssetConfigTxReq{
			AssetIndex:      assetIndex,
			AssetMarginMode: assetMarginMode,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUpdateAccountAssetConfigTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUpdateLeverage(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	marketIndex := int64(checkedInt(args[1]))
	initialMarginFraction := checkedUint16(checkedInt(args[2]))
	marginMode := checkedUint8(checkedInt(args[3]))
	nonce := int64(checkedInt(args[4]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UpdateLeverageTxReq{
			MarketIndex:           marketIndex,
			InitialMarginFraction: initialMarginFraction,
			MarginMode:            marginMode,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUpdateLeverageTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignUpdateMargin(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	marketIndex := int64(checkedInt(args[1]))
	usdcAmount := int64(checkedInt(args[2]))
	direction := checkedUint8(checkedInt(args[3]))
	nonce := int64(checkedInt(args[4]))

	return promiseResult(func(resolve js.Value) {
		txInfo := &UpdateMarginTxReq{
			MarketIndex: marketIndex,
			USDCAmount:  usdcAmount,
			Direction:   direction,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetUpdateMarginTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignCreateGroupedOrders(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	defaultExpiryTime := time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	groupingType := checkedUint8(checkedInt(args[1]))
	orderCount := checkedUint8(checkedInt(args[2]))
	orders := make([]*CreateOrderTxReq, int(orderCount))
	currentIndex := 3
	for i := 0; i < int(orderCount); i++ {
		marketIndex := int64(checkedInt(args[currentIndex]))
		clientOrderIndex := int64(checkedInt(args[currentIndex+1]))
		baseAmount, err := strconv.Atoi(checkedString(args[currentIndex+2]))
		if err != nil {
			return errPromise(fmt.Errorf("baseAmount is not an integer"))
		}
		price, err := strconv.Atoi(checkedString(args[currentIndex+3]))
		if err != nil {
			return errPromise(fmt.Errorf("price is not an integer"))
		}
		isAsk := checkedUint8(checkedInt(args[currentIndex+4]))
		orderType := checkedUint8(checkedInt(args[currentIndex+5]))
		timeInForce := checkedUint8(checkedInt(args[currentIndex+6]))
		reduceOnly := checkedUint8(checkedInt(args[currentIndex+7]))
		triggerPrice, err := strconv.Atoi(checkedString(args[currentIndex+8]))
		if err != nil {
			return errPromise(fmt.Errorf("triggerPrice is not an integer"))
		}
		orderExpiry := int64(checkedInt(args[currentIndex+9]))
		if orderExpiry == -1 {
			orderExpiry = defaultExpiryTime
		}
		orders[i] = &CreateOrderTxReq{
			MarketIndex:      marketIndex,
			ClientOrderIndex: clientOrderIndex,
			BaseAmount:       int64(baseAmount),
			Price:            checkedUint32(price),
			IsAsk:            isAsk,
			Type:             orderType,
			TimeInForce:      timeInForce,
			ReduceOnly:       reduceOnly,
			TriggerPrice:     checkedUint32(triggerPrice),
			OrderExpiry:      orderExpiry,
		}
		currentIndex += 10
	}
	nonce := int64(checkedInt(args[currentIndex]))
	integratorAccountIndex, integratorTakerFee, integratorMakerFee := integratorAttributes(args, currentIndex+1)

	return promiseResult(func(resolve js.Value) {
		txInfo := &CreateGroupedOrdersTxReq{
			GroupingType: groupingType,
			Orders:       orders,
		}
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		ops.IntegratorAccountIndex = integratorAccountIndex
		ops.IntegratorTakerFee = integratorTakerFee
		ops.IntegratorMakerFee = integratorMakerFee
		tx, err := clients[accountIndex].GetCreateGroupedOrdersTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignBinaryOptionsOperatorCreateMarket(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	req := checkedObject(args[1])
	nonce := int64(checkedInt(args[2]))
	txInfo := &BinaryOptionsOperatorCreateMarketTxReq{
		MarketSlot:        checkedInt16(checkedInt(req.Get("MarketSlot"))),
		StartTimestampMs:  int64(checkedInt(req.Get("StartTimestampMs"))),
		EndTimestampMs:    int64(checkedInt(req.Get("EndTimestampMs"))),
		SettlementCap:     checkedUint32(checkedInt(req.Get("SettlementCap"))),
		SettlementType:    checkedUint8(checkedInt(req.Get("SettlementType"))),
		DefaultPrice:      checkedUint32(checkedInt(req.Get("DefaultPrice"))),
		TakerFee:          checkedUint32(checkedInt(req.Get("TakerFee"))),
		MakerFee:          checkedUint32(checkedInt(req.Get("MakerFee"))),
		MinBaseAmount:     int64(checkedInt(req.Get("MinBaseAmount"))),
		MinQuoteAmount:    int64(checkedInt(req.Get("MinQuoteAmount"))),
		OrderQuoteLimit:   int64(checkedInt(req.Get("OrderQuoteLimit"))),
		OpenInterestLimit: int64(checkedInt(req.Get("OpenInterestLimit"))),
		IsFrozen:          checkedUint8(checkedInt(req.Get("IsFrozen"))),
		Symbol:            checkedString(req.Get("Symbol")),
		SizeDecimals:      checkedUint8(checkedInt(req.Get("SizeDecimals"))),
		PriceDecimals:     checkedUint8(checkedInt(req.Get("PriceDecimals"))),
	}

	return promiseResult(func(resolve js.Value) {
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetBinaryOptionsOperatorCreateMarketTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignBinaryOptionsOperatorUpdateMarket(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	req := checkedObject(args[1])
	nonce := int64(checkedInt(args[2]))
	txInfo := &BinaryOptionsOperatorUpdateMarketTxReq{
		MarketSlot:        checkedInt16(checkedInt(req.Get("MarketSlot"))),
		MarketIndex:       int64(checkedInt(req.Get("MarketIndex"))),
		StartTimestampMs:  int64(checkedInt(req.Get("StartTimestampMs"))),
		EndTimestampMs:    int64(checkedInt(req.Get("EndTimestampMs"))),
		TakerFee:          checkedUint32(checkedInt(req.Get("TakerFee"))),
		MakerFee:          checkedUint32(checkedInt(req.Get("MakerFee"))),
		LiquidationFee:    optionalUint32(req.Get("LiquidationFee")),
		MinBaseAmount:     int64(checkedInt(req.Get("MinBaseAmount"))),
		MinQuoteAmount:    int64(checkedInt(req.Get("MinQuoteAmount"))),
		OrderQuoteLimit:   int64(checkedInt(req.Get("OrderQuoteLimit"))),
		OpenInterestLimit: int64(checkedInt(req.Get("OpenInterestLimit"))),
		IsFrozen:          checkedUint8(checkedInt(req.Get("IsFrozen"))),
	}

	return promiseResult(func(resolve js.Value) {
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetBinaryOptionsOperatorUpdateMarketTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

func SignBinaryOptionsOperatorSettleMarket(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	req := checkedObject(args[1])
	nonce := int64(checkedInt(args[2]))
	txInfo := &BinaryOptionsOperatorSettleMarketTxReq{
		MarketSlot:      checkedInt16(checkedInt(req.Get("MarketSlot"))),
		MarketIndex:     int64(checkedInt(req.Get("MarketIndex"))),
		SettlementPrice: checkedUint32(checkedInt(req.Get("SettlementPrice"))),
		IsRefund:        checkedUint8(checkedInt(req.Get("IsRefund"))),
	}

	return promiseResult(func(resolve js.Value) {
		ops := new(TransactOpts)
		ops.Nonce = &nonce
		tx, err := clients[accountIndex].GetBinaryOptionsOperatorSettleMarketTransaction(txInfo, ops)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
	})
}

// optionalUint32 treats a missing (undefined/null) object field as zero.
func optionalUint32(v js.Value) uint32 {
	if v.IsUndefined() || v.IsNull() {
		return 0
	}
	return checkedUint32(checkedInt(v))
}

func GetAirdropAllocationMessage(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	allocations := checkedString(args[1])
	return promiseResult(func(resolve js.Value) {
		message := fmt.Sprintf(txtypes.TemplateAirdropAllocation, allocations, getHex10FromUint64(uint64(clients[accountIndex].ChainId())))
		resolve.Invoke(map[string]any{
			"message": message,
		})
	})
}

func CreateAuthToken(this js.Value, args []js.Value) any {
	accountIndex := int64(checkedInt(args[0]))
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	apiKeyIndex := checkedUint8(0)
	if len(args) > 1 {
		apiKeyIndex = checkedUint8(checkedInt(args[1]))
	}
	return promiseResult(func(resolve js.Value) {
		deadline := time.Now().Add(time.Hour * 1).Unix()
		message := fmt.Sprintf("%v:%v:%v", deadline, accountIndex, apiKeyIndex)
		signature, err := clients[accountIndex].SignMessage(message)
		if err != nil {
			resolve.Invoke(errToJson(err))
			return
		}
		resolve.Invoke(map[string]any{
			"signature":    signature,
			"deadline":     deadline,
			"accountIndex": accountIndex,
			"apiKeyIndex":  apiKeyIndex,
			"token":        fmt.Sprintf("%v:%v:%v:%v", deadline, accountIndex, apiKeyIndex, signature),
		})
	})
}

func cancelAllMarketIndexAttribute(args []js.Value, idx int) (int64, bool) {
	if len(args) <= idx || args[idx].Type() != js.TypeNumber {
		return 0, false
	}
	marketIndex := int64(checkedInt(args[idx]))
	if marketIndex == txtypes.NilMarketIndex {
		return 0, false
	}
	return marketIndex, true
}

// integratorAttributes reads the optional trailing (integratorAccountIndex,
// integratorTakerFee, integratorMakerFee) args starting at idx. Missing,
// nil-valued args leave the corresponding attribute unset so
// existing callers keep signing identical hashes.
func integratorAttributes(args []js.Value, idx int) (accountIndex *int64, takerFee, makerFee *uint32) {
	if len(args) > idx && !args[idx].IsNull() && !args[idx].IsUndefined() {
		if v := int64(checkedInt(args[idx])); v != txtypes.NilIntegratorIndex {
			accountIndex = &v
		}
	}
	if len(args) > idx+1 && !args[idx+1].IsNull() && !args[idx+1].IsUndefined() {
		if v := checkedUint32(checkedInt(args[idx+1])); v != txtypes.NilIntegratorTakerFee {
			takerFee = &v
		}
	}
	if len(args) > idx+2 && !args[idx+2].IsNull() && !args[idx+2].IsUndefined() {
		if v := checkedUint32(checkedInt(args[idx+2])); v != txtypes.NilIntegratorMakerFee {
			makerFee = &v
		}
	}
	return
}

func errPromise(err error) js.Value {
	return promiseResult(func(resolve js.Value) {
		resolve.Invoke(errToJson(err))
	})
}

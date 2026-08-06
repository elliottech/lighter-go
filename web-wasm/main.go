//go:build js && wasm

package main

import (
	"fmt"
	"math/big"
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

	js.Global().Set("_createClient", js.FuncOf(CreateClient))
	js.Global().Set("_createClientByPrv", js.FuncOf(CreateClientByPrv))

	// Transactions
	js.Global().Set("_signChangePubKey", js.FuncOf(SignChangePubKey))
	js.Global().Set("_signRevokePubKey", js.FuncOf(SignRevokePubKey))
	js.Global().Set("_signCreateSubAccount", js.FuncOf(SignCreateSubAccount))
	js.Global().Set("_signCreatePublicPool", js.FuncOf(SignCreatePublicPool))
	js.Global().Set("_signUpdatePublicPool", js.FuncOf(SignUpdatePublicPool))
	js.Global().Set("_signCreateOrder", js.FuncOf(SignCreateOrder))
	js.Global().Set("_signCancelOrder", js.FuncOf(SignCancelOrder))
	js.Global().Set("_signWithdraw", js.FuncOf(SignWithdraw))
	js.Global().Set("_signCancelAllOrders", js.FuncOf(SignCancelAllOrders))
	js.Global().Set("_signModifyOrder", js.FuncOf(SignModifyOrder))
	js.Global().Set("_signTransfer", js.FuncOf(SignTransfer))
	js.Global().Set("_signMintShares", js.FuncOf(SignMintShares))
	js.Global().Set("_signBurnShares", js.FuncOf(SignBurnShares))
	js.Global().Set("_signUpdateLeverage", js.FuncOf(SignUpdateLeverage))
	js.Global().Set("_signUpdateMargin", js.FuncOf(SignUpdateMargin))
	js.Global().Set("_signUpdateAccountConfig", js.FuncOf(SignUpdateAccountConfig))
	js.Global().Set("_signUpdateAccountAssetConfig", js.FuncOf(SignUpdateAccountAssetConfig))
	js.Global().Set("_signCreateGroupedOrders", js.FuncOf(SignCreateGroupedOrders))
	js.Global().Set("_getChangePubKeyTransaction", js.FuncOf(GetChangePubKeyTransaction))
	js.Global().Set("_getRevokePubKeyTransaction", js.FuncOf(GetRevokePubKeyTransaction))
	js.Global().Set("_getTransferTransaction", js.FuncOf(GetTransferTransaction))
	js.Global().Set("_createAuthToken", js.FuncOf(CreateAuthToken))
	js.Global().Set("_signStakeAssets", js.FuncOf(SignStakeAssets))
	js.Global().Set("_signUnstakeAssets", js.FuncOf(SignUnstakeAssets))
	<-c
}

var clients = make(map[int64]*LightClient)

func CreateClient(this js.Value, args []js.Value) any {
	seed := string(args[0].String())
	seed = strings.TrimPrefix(seed, "0x")
	chainId := uint32(args[1].Int())
	accountIndex := int64(args[2].Int())
	nonce := int64(args[3].Int())
	apiKeyIndex := uint8(args[4].Int())
	skipNonce := len(args) > 5 && args[5].Truthy()

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				sdkClient, err := NewLightClient(seed, chainId)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				sdkClient.SetAccountIndex(accountIndex)
				sdkClient.SetApiKeyIndex(apiKeyIndex)
				sdkClient.SetSkipNonce(skipNonce)

				clients[accountIndex] = sdkClient
				pub := sdkClient.KeyManager().PubKeyBytes()
				prv := sdkClient.KeyManager().PrvKeyBytes()
				txInfo := &ChangePubKeyReq{PubKey: pub}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
				success := false
				if err == nil {
					success = true
				}

				response := map[string]any{
					"success":       sdkClient != nil,
					"pk":            common.Bytes2Hex(pub[:]),
					"prv":           common.Bytes2Hex(prv),
					"pubKeySuccess": success,
					"body":          txnBody,
				}

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func CreateClientByPrv(this js.Value, args []js.Value) any {
	prv := string(args[0].String())
	prv = strings.TrimPrefix(prv, "0x")
	chainId := uint32(args[1].Int())
	accountIndex := int64(args[2].Int())
	nonce := int64(args[3].Int())
	apiKeyIndex := uint8(args[4].Int())
	skipNonce := len(args) > 5 && args[5].Truthy()

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				sdkClient, err := NewLightClientByPrv(prv, chainId)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				sdkClient.SetAccountIndex(accountIndex)
				sdkClient.SetApiKeyIndex(apiKeyIndex)
				sdkClient.SetSkipNonce(skipNonce)

				clients[accountIndex] = sdkClient
				pub := sdkClient.KeyManager().PubKeyBytes()
				prv := sdkClient.KeyManager().PrvKeyBytes()
				txInfo := &ChangePubKeyReq{PubKey: pub}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
				success := false
				if err == nil {
					success = true
				}

				response := map[string]any{
					"success":       sdkClient != nil,
					"pk":            common.Bytes2Hex(pub[:]),
					"prv":           common.Bytes2Hex(prv),
					"pubKeySuccess": success,
					"body":          txnBody,
				}

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func GetChangePubKeyTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(args[1].Int())
	apiKeyIndex := uint8(args[2].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				sdkClient := clients[accountIndex]
				pub := sdkClient.KeyManager().PubKeyBytes()
				txInfo := &ChangePubKeyReq{PubKey: pub}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				ops.ApiKeyIndex = &apiKeyIndex
				txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
				success := false
				if err == nil {
					success = true
				}

				response := map[string]any{
					"success":       sdkClient != nil,
					"pk":            common.Bytes2Hex(pub[:]),
					"pubKeySuccess": success,
					"body":          txnBody,
				}

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func GetRevokePubKeyTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(args[1].Int())
	apiKeyIndex := uint8(args[2].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				sdkClient := clients[accountIndex]
				pub := [40]byte{}
				txInfo := &ChangePubKeyReq{PubKey: pub}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				ops.ApiKeyIndex = &apiKeyIndex
				txnBody, err := sdkClient.GenerateChangePubKeySignBody(txInfo, ops)
				success := false
				if err == nil {
					success = true
				}

				response := map[string]any{
					"success":       sdkClient != nil,
					"pk":            common.Bytes2Hex(pub[:]),
					"pubKeySuccess": success,
					"body":          txnBody,
				}

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignChangePubKey(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(args[1].String())
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(args[2].Int())
	apiKeyIndex := uint8(0)
	if len(args) > 3 {
		apiKeyIndex = uint8(args[3].Int())
	}
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignRevokePubKey(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(args[1].String())
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(args[2].Int())
	apiKeyIndex := uint8(0)
	if len(args) > 3 {
		apiKeyIndex = uint8(args[3].Int())
	}
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCreateSubAccount(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	nonce := int64(args[1].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				tx, err := clients[accountIndex].GetCreateSubAccountTransaction(ops)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCreatePublicPool(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	operatorFee := int64(args[1].Int())
	initialTotalShares := int64(args[2].Int())
	minOperatorShareRate := uint16(args[3].Int())
	nonce := int64(args[4].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUpdatePublicPool(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	publicPoolIndex := int64(args[1].Int())
	status := uint8(args[2].Int())
	operatorFee := int64(args[3].Int())
	minOperatorShareRate := uint16(args[4].Int())
	nonce := int64(args[5].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCreateOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int16(args[1].Int())
	clientOrderIndex := int64(args[2].Int())
	baseAmount, err := strconv.Atoi(args[3].String())
	if err != nil {
		return errPromise(fmt.Errorf("baseAmount is not an integer"))
	}
	price, err := strconv.Atoi(args[4].String())
	if err != nil {
		return errPromise(fmt.Errorf("price is not an integer"))
	}
	isAsk := uint8(args[5].Int())
	orderType := uint8(args[6].Int())
	timeInForce := uint8(args[7].Int())
	reduceOnly := uint8(args[8].Int())
	triggerPrice, err := strconv.Atoi(args[9].String())
	if err != nil {
		return errPromise(fmt.Errorf("triggerPrice is not an integer"))
	}
	orderExpiry := int64(args[10].Int())
	nonce := int64(args[11].Int())

	if orderExpiry == -1 {
		orderExpiry = time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	}

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				txInfo := &CreateOrderTxReq{
					MarketIndex:      marketIndex,
					ClientOrderIndex: clientOrderIndex,
					BaseAmount:       int64(baseAmount),
					Price:            uint32(price),
					IsAsk:            isAsk,
					Type:             orderType,
					TimeInForce:      timeInForce,
					ReduceOnly:       reduceOnly,
					TriggerPrice:     uint32(triggerPrice),
					OrderExpiry:      orderExpiry,
				}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				tx, err := clients[accountIndex].GetCreateOrderTransaction(txInfo, ops)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCancelOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int16(args[1].Int())
	orderIndex, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("orderIndex is not an integer"))
	}
	nonce := int64(args[3].Int())
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignWithdraw(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	assetIndex := int16(args[1].Int())
	routeType := uint8(args[2].Int())
	assetAmount := new(big.Int)
	assetAmount, ok := assetAmount.SetString(args[3].String(), 10)
	if !ok {
		return errPromise(fmt.Errorf("assetAmount is not an integer"))
	}
	nonce := int64(args[4].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				txInfo := &WithdrawTxReq{
					AssetIndex: assetIndex,
					RouteType:  routeType,
					Amount:     assetAmount.Uint64(),
				}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				tx, err := clients[accountIndex].GetWithdrawTransaction(txInfo, ops)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCancelAllOrders(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	timeInForce := uint8(args[1].Int())
	time := int64(args[2].Int())
	nonce := int64(args[3].Int())
	cancelMarketIndex, hasCancelMarketIndex := cancelAllMarketIndexAttribute(args, 4)

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignModifyOrder(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	marketIndex := int16(args[1].Int())
	orderIndex, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("orderIndex is not an integer"))
	}
	baseAmount := int64(args[3].Int())
	price := uint32(args[4].Int())
	triggerPrice := uint32(args[5].Int())
	nonce := int64(args[6].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func GetTransferTransaction(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	nonce := int64(args[1].Int())
	apiKeyIndex := uint8(args[2].Int())
	toAccountIndex := int64(args[3].Int())
	assetIndex := int16(args[4].Int())
	fromRouteType := uint8(args[5].Int())
	toRouteType := uint8(args[6].Int())
	amount := int64(args[7].Int())
	USDCFee := int64(args[8].Int())

	memo := [32]byte{}
	if len(args) > 9 {
		memoArg := args[9]
		if memoArg.InstanceOf(js.Global().Get("Array")) {
			for i := 0; i < memoArg.Length() && i < 32; i++ {
				memo[i] = byte(memoArg.Index(i).Int())
			}
		}
	}

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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

				response := map[string]any{
					"success":       sdkClient != nil,
					"pubKeySuccess": err == nil,
					"body":          txnBody,
				}

				resolve.Invoke(response)
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignTransfer(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	signedMessage := string(args[1].String())
	if !strings.HasPrefix(signedMessage, "0x") {
		signedMessage = "0x" + signedMessage
	}

	nonce := int64(args[2].Int())
	apiKeyIndex := uint8(args[3].Int())
	toAccountIndex := int64(args[4].Int())
	assetIndex := int16(args[5].Int())
	fromRouteType := uint8(args[6].Int())
	toRouteType := uint8(args[7].Int())
	amount := int64(args[8].Int())
	USDCFee := int64(args[9].Int())

	memo := [32]byte{}
	if len(args) > 10 {
		memoArg := args[10]
		if memoArg.InstanceOf(js.Global().Get("Array")) {
			for i := 0; i < memoArg.Length() && i < 32; i++ {
				memo[i] = byte(memoArg.Index(i).Int())
			}
		}
	}

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignStakeAssets(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	stakingPoolIndex := int64(args[1].Int())
	shareAmount, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(args[3].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUnstakeAssets(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	stakingPoolIndex := int64(args[1].Int())
	shareAmount, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(args[3].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignMintShares(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	publicPoolIndex := int64(args[1].Int())
	shareAmount, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(args[3].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignBurnShares(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	publicPoolIndex := int64(args[1].Int())
	shareAmount, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return errPromise(fmt.Errorf("shareAmount is not an integer"))
	}
	nonce := int64(args[3].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUpdateAccountConfig(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	accountTradingMode := uint8(args[1].Int())
	nonce := int64(args[2].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUpdateAccountAssetConfig(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	assetIndex := int16(args[1].Int())
	assetMarginMode := uint8(args[2].Int())
	nonce := int64(args[3].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUpdateLeverage(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	marketIndex := int16(args[1].Int())
	initialMarginFraction := uint16(args[2].Int())
	marginMode := uint8(args[3].Int())
	nonce := int64(args[4].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignUpdateMargin(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	marketIndex := int16(args[1].Int())
	usdcAmount := int64(args[2].Int())
	direction := uint8(args[3].Int())
	nonce := int64(args[4].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func SignCreateGroupedOrders(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}

	defaultExpiryTime := time.Now().Add(time.Hour * 24 * 28).UnixMilli() // 28 days
	groupingType := uint8(args[1].Int())
	orderCount := uint8(args[2].Int())
	orders := make([]*CreateOrderTxReq, int(orderCount))
	currentIndex := 3
	for i := 0; i < int(orderCount); i++ {
		marketIndex := int16(args[currentIndex].Int())
		clientOrderIndex := int64(args[currentIndex+1].Int())
		baseAmount, err := strconv.Atoi(args[currentIndex+2].String())
		if err != nil {
			return errPromise(fmt.Errorf("baseAmount is not an integer"))
		}
		price, err := strconv.Atoi(args[currentIndex+3].String())
		if err != nil {
			return errPromise(fmt.Errorf("price is not an integer"))
		}
		isAsk := uint8(args[currentIndex+4].Int())
		orderType := uint8(args[currentIndex+5].Int())
		timeInForce := uint8(args[currentIndex+6].Int())
		reduceOnly := uint8(args[currentIndex+7].Int())
		triggerPrice, err := strconv.Atoi(args[currentIndex+8].String())
		if err != nil {
			return errPromise(fmt.Errorf("triggerPrice is not an integer"))
		}
		orderExpiry := int64(args[currentIndex+9].Int())
		if orderExpiry == -1 {
			orderExpiry = defaultExpiryTime
		}
		orders[i] = &CreateOrderTxReq{
			MarketIndex:      marketIndex,
			ClientOrderIndex: clientOrderIndex,
			BaseAmount:       int64(baseAmount),
			Price:            uint32(price),
			IsAsk:            isAsk,
			Type:             orderType,
			TimeInForce:      timeInForce,
			ReduceOnly:       reduceOnly,
			TriggerPrice:     uint32(triggerPrice),
			OrderExpiry:      orderExpiry,
		}
		currentIndex += 10
	}
	nonce := int64(args[currentIndex].Int())

	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				txInfo := &CreateGroupedOrdersTxReq{
					GroupingType: groupingType,
					Orders:       orders,
				}
				ops := new(TransactOpts)
				ops.Nonce = &nonce
				tx, err := clients[accountIndex].GetCreateGroupedOrdersTransaction(txInfo, ops)
				if err != nil {
					resolve.Invoke(errToJson(err))
					return
				}
				resolve.Invoke(getTxResponse(tx, clients[accountIndex].ChainId()))
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func CreateAuthToken(this js.Value, args []js.Value) any {
	accountIndex := int64(args[0].Int())
	if clients[accountIndex] == nil {
		return errPromise(fmt.Errorf("client is not created for the account index %d", accountIndex))
	}
	apiKeyIndex := uint8(0)
	if len(args) > 1 {
		apiKeyIndex = uint8(args[1].Int())
	}
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
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
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

func cancelAllMarketIndexAttribute(args []js.Value, idx int) (int16, bool) {
	if len(args) <= idx || args[idx].Type() != js.TypeNumber {
		return 0, false
	}
	marketIndex := int16(args[idx].Int())
	if marketIndex == txtypes.NilMarketIndex {
		return 0, false
	}
	return marketIndex, true
}

func errPromise(err error) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			resolve := args[0]
			go func() {
				resolve.Invoke(errToJson(err))
			}()
			return nil
		})
		promiseCons := js.Global().Get("Promise")
		return promiseCons.New(handler)
	})
}

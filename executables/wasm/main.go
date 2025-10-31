// +build js

package main

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/elliottech/lighter-go/executables"
	"github.com/elliottech/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func wrapErr(err error) string {
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return ""
}

func main() {
	js.Global().Set("GenerateAPIKey", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return js.ValueOf(map[string]interface{}{"error": "GenerateAPIKey expects 1 arg: seed"})
		}
		seed := args[0].String()
		privateKey, publicKey, err := executables.GenerateAPIKey(seed)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"privateKey": privateKey, "publicKey": publicKey, "error": ""})
	}))

	js.Global().Set("CreateClient", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "CreateClient expects 5 args: url, privateKey, chainId, apiKeyIndex, accountIndex"})
		}
		url := args[0].String()
		privateKey := args[1].String()
		chainId := uint32(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())
		err := executables.CreateClient(url, privateKey, chainId, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"error": ""})
	}))

	js.Global().Set("CheckClient", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			return js.ValueOf(map[string]interface{}{"error": "CheckClient expects 2 args: apiKeyIndex, accountIndex"})
		}
		apiKeyIndex := uint8(args[0].Int())
		accountIndex := int64(args[1].Int())
		err := executables.CheckClient(apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"error": ""})
	}))

	js.Global().Set("CreateAuthToken", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return js.ValueOf(map[string]interface{}{"error": "CreateAuthToken expects 3 args: deadline, apiKeyIndex, accountIndex"})
		}
		deadline := int64(args[0].Int())
		apiKeyIndex := uint8(args[1].Int())
		accountIndex := int64(args[2].Int())
		token, err := executables.CreateAuthToken(deadline, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"authToken": token, "error": ""})
	}))

	js.Global().Set("SignChangePubKey", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 4 {
			return js.ValueOf(map[string]interface{}{"error": "SignChangePubKey expects 4 args: pubKeyHex, nonce, apiKeyIndex, accountIndex"})
		}
		pubKeyHex := args[0].String()
		nonce := int64(args[1].Int())
		apiKeyIndex := uint8(args[2].Int())
		accountIndex := int64(args[3].Int())

		pubKeyBytes, err := hexutil.Decode(pubKeyHex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		if len(pubKeyBytes) != 40 {
			return js.ValueOf(map[string]interface{}{"error": "invalid pub key length. expected 40 but got " + strconv.Itoa(len(pubKeyBytes))})
		}
		var pubKey [40]byte
		copy(pubKey[:], pubKeyBytes)

		txInfo, _, err := executables.GetChangePubKeyTransaction(pubKey, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCreateOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 13 {
			return js.ValueOf(map[string]interface{}{"error": "SignCreateOrder expects 13 args: marketIndex, clientOrderIndex, baseAmount, price, isAsk, orderType, timeInForce, reduceOnly, triggerPrice, orderExpiry, nonce, apiKeyIndex, accountIndex"})
		}
		marketIndex := uint8(args[0].Int())
		clientOrderIndex := int64(args[1].Int())
		baseAmount := int64(args[2].Int())
		price := uint32(args[3].Int())
		isAsk := uint8(args[4].Int())
		orderType := uint8(args[5].Int())
		timeInForce := uint8(args[6].Int())
		reduceOnly := uint8(args[7].Int())
		triggerPrice := uint32(args[8].Int())
		orderExpiry := int64(args[9].Int())
		nonce := int64(args[10].Int())
		apiKeyIndex := uint8(args[11].Int())
		accountIndex := int64(args[12].Int())

		txInfo, err := executables.GetCreateOrderTransaction(
			marketIndex, clientOrderIndex, baseAmount, price, isAsk,
			orderType, timeInForce, reduceOnly, triggerPrice, orderExpiry, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCancelOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "SignCancelOrder expects 5 args: marketIndex, orderIndex, nonce, apiKeyIndex, accountIndex"})
		}
		marketIndex := uint8(args[0].Int())
		orderIndex := int64(args[1].Int())
		nonce := int64(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())

		txInfo, err := executables.GetCancelOrderTransaction(marketIndex, orderIndex, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCancelAllOrders", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "SignCancelAllOrders expects 5 args: timeInForce, time, nonce, apiKeyIndex, accountIndex"})
		}
		timeInForce := uint8(args[0].Int())
		timeVal := int64(args[1].Int())
		nonce := int64(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())

		txInfo, err := executables.GetCancelAllOrdersTransaction(timeInForce, timeVal, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignTransfer", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 7 {
			return js.ValueOf(map[string]interface{}{"error": "SignTransfer expects 7 args: toAccount, usdcAmount, fee, memo, nonce, apiKeyIndex, accountIndex"})
		}
		toAccount := int64(args[0].Int())
		usdcAmount := int64(args[1].Int())
		fee := int64(args[2].Int())
		memoStr := args[3].String()
		nonce := int64(args[4].Int())
		apiKeyIndex := uint8(args[5].Int())
		accountIndex := int64(args[6].Int())

		var memoArr [32]byte
		bs := []byte(memoStr)
		for i := 0; i < len(bs) && i < 32; i++ {
			memoArr[i] = bs[i]
		}

		txInfo, err := executables.GetTransferTransaction(toAccount, usdcAmount, fee, nonce, memoArr, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignWithdraw", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 4 {
			return js.ValueOf(map[string]interface{}{"error": "SignWithdraw expects 4 args: usdcAmount, nonce, apiKeyIndex, accountIndex"})
		}
		usdcAmount := uint64(args[0].Int())
		nonce := int64(args[1].Int())
		apiKeyIndex := uint8(args[2].Int())
		accountIndex := int64(args[3].Int())

		txInfo, err := executables.GetWithdrawTransaction(usdcAmount, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignUpdateLeverage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 6 {
			return js.ValueOf(map[string]interface{}{"error": "SignUpdateLeverage expects 6 args: marketIndex, fraction, marginMode, nonce, apiKeyIndex, accountIndex"})
		}
		marketIndex := uint8(args[0].Int())
		fraction := uint16(args[1].Int())
		marginMode := uint8(args[2].Int())
		nonce := int64(args[3].Int())
		apiKeyIndex := uint8(args[4].Int())
		accountIndex := int64(args[5].Int())

		txInfo, err := executables.GetUpdateLeverageTransaction(marketIndex, marginMode, fraction, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignModifyOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 8 {
			return js.ValueOf(map[string]interface{}{"error": "SignModifyOrder expects 8 args: marketIndex, index, baseAmount, price, triggerPrice, nonce, apiKeyIndex, accountIndex"})
		}
		marketIndex := uint8(args[0].Int())
		index := int64(args[1].Int())
		baseAmount := int64(args[2].Int())
		price := uint32(args[3].Int())
		triggerPrice := uint32(args[4].Int())
		nonce := int64(args[5].Int())
		apiKeyIndex := uint8(args[6].Int())
		accountIndex := int64(args[7].Int())

		txInfo, err := executables.GetModifyOrderTransaction(marketIndex, index, baseAmount, price, triggerPrice, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCreateSubAccount", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return js.ValueOf(map[string]interface{}{"error": "SignCreateSubAccount expects 3 args: nonce, apiKeyIndex, accountIndex"})
		}
		nonce := int64(args[0].Int())
		apiKeyIndex := uint8(args[1].Int())
		accountIndex := int64(args[2].Int())

		txInfo, err := executables.GetCreateSubAccountTransaction(nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCreatePublicPool", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 6 {
			return js.ValueOf(map[string]interface{}{"error": "SignCreatePublicPool expects 6 args: operatorFee, initialTotalShares, minOperatorShareRate, nonce, apiKeyIndex, accountIndex"})
		}
		operatorFee := int64(args[0].Int())
		initialTotalShares := int64(args[1].Int())
		minOperatorShareRate := int64(args[2].Int())
		nonce := int64(args[3].Int())
		apiKeyIndex := uint8(args[4].Int())
		accountIndex := int64(args[5].Int())

		txInfo, err := executables.GetCreatePublicPoolTransaction(operatorFee, initialTotalShares, minOperatorShareRate, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignUpdatePublicPool", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 7 {
			return js.ValueOf(map[string]interface{}{"error": "SignUpdatePublicPool expects 7 args: publicPoolIndex, status, operatorFee, minOperatorShareRate, nonce, apiKeyIndex, accountIndex"})
		}
		publicPoolIndex := uint8(args[0].Int())
		status := uint8(args[1].Int())
		operatorFee := int64(args[2].Int())
		minOperatorShareRate := int64(args[3].Int())
		nonce := int64(args[4].Int())
		apiKeyIndex := uint8(args[5].Int())
		accountIndex := int64(args[6].Int())

		txInfo, err := executables.GetUpdatePublicPoolTransaction(publicPoolIndex, status, operatorFee, minOperatorShareRate, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignMintShares", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "SignMintShares expects 5 args: publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex"})
		}
		publicPoolIndex := int64(args[0].Int())
		shareAmount := int64(args[1].Int())
		nonce := int64(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())

		txInfo, err := executables.GetMintSharesTransaction(publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignBurnShares", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "SignBurnShares expects 5 args: publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex"})
		}
		publicPoolIndex := int64(args[0].Int())
		shareAmount := int64(args[1].Int())
		nonce := int64(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())

		txInfo, err := executables.GetBurnSharesTransaction(publicPoolIndex, shareAmount, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignUpdateMargin", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 6 {
			return js.ValueOf(map[string]interface{}{"error": "SignUpdateMargin expects 6 args: marketIndex, usdcAmount, direction, nonce, apiKeyIndex, accountIndex"})
		}
		marketIndex := uint8(args[0].Int())
		usdcAmount := int64(args[1].Int())
		direction := uint8(args[2].Int())
		nonce := int64(args[3].Int())
		apiKeyIndex := uint8(args[4].Int())
		accountIndex := int64(args[5].Int())

		txInfo, err := executables.GetUpdateMarginTransaction(marketIndex, direction, usdcAmount, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SwitchAPIKey", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 1 {
			return js.ValueOf(map[string]interface{}{"error": "SwitchAPIKey expects 1 arg: apiKeyIndex"})
		}
		apiKeyIndex := uint8(args[0].Int())
		err := executables.SwitchAPIKey(apiKeyIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"error": ""})
	}))

	js.Global().Set("SignCreateGroupedOrders", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 5 {
			return js.ValueOf(map[string]interface{}{"error": "SignCreateGroupedOrders expects 5 args: groupingType, orders array, nonce, apiKeyIndex, accountIndex"})
		}
		groupingType := uint8(args[0].Int())
		
		// Parse orders array from JS
		ordersArg := args[1]
		if ordersArg.Type() != js.TypeObject {
			return js.ValueOf(map[string]interface{}{"error": "orders must be an array"})
		}
		length := ordersArg.Length()
		orders := make([]*types.CreateOrderTxReq, length)
		
		for i := 0; i < length; i++ {
			orderObj := ordersArg.Index(i)
			if orderObj.Type() != js.TypeObject {
				return js.ValueOf(map[string]interface{}{"error": fmt.Sprintf("order %d must be an object", i)})
			}
			
			orders[i] = &types.CreateOrderTxReq{
				MarketIndex:      uint8(orderObj.Get("MarketIndex").Int()),
				ClientOrderIndex: int64(orderObj.Get("ClientOrderIndex").Int()),
				BaseAmount:       int64(orderObj.Get("BaseAmount").Int()),
				Price:            uint32(orderObj.Get("Price").Int()),
				IsAsk:            uint8(orderObj.Get("IsAsk").Int()),
				Type:             uint8(orderObj.Get("Type").Int()),
				TimeInForce:      uint8(orderObj.Get("TimeInForce").Int()),
				ReduceOnly:       uint8(orderObj.Get("ReduceOnly").Int()),
				TriggerPrice:     uint32(orderObj.Get("TriggerPrice").Int()),
				OrderExpiry:      int64(orderObj.Get("OrderExpiry").Int()),
			}
		}
		
		nonce := int64(args[2].Int())
		apiKeyIndex := uint8(args[3].Int())
		accountIndex := int64(args[4].Int())
		
		txInfo, err := executables.GetCreateGroupedOrdersTransaction(groupingType, orders, nonce, apiKeyIndex, accountIndex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	select {}
}


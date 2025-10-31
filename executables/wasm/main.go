// +build js

package main

import (
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/elliottech/lighter-go/executables"
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
		if len(args) < 4 {
			return js.ValueOf(map[string]interface{}{"error": "CreateClient expects 4 args: url, privateKey, chainId, apiKeyIndex, accountIndex"})
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
		var deadline int64
		if len(args) > 0 {
			deadline = int64(args[0].Int())
		}
		token, err := executables.CreateAuthToken(deadline)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"authToken": token, "error": ""})
	}))

	js.Global().Set("SignChangePubKey", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			return js.ValueOf(map[string]interface{}{"error": "SignChangePubKey expects 2 args: pubKeyHex, nonce"})
		}
		pubKeyHex := args[0].String()
		nonce := int64(args[1].Int())

		pubKeyBytes, err := hexutil.Decode(pubKeyHex)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		if len(pubKeyBytes) != 40 {
			return js.ValueOf(map[string]interface{}{"error": "invalid pub key length. expected 40 but got " + strconv.Itoa(len(pubKeyBytes))})
		}
		var pubKey [40]byte
		copy(pubKey[:], pubKeyBytes)

		txInfo, _, err := executables.GetChangePubKeyTransaction(pubKey, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCreateOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 11 {
			return js.ValueOf(map[string]interface{}{"error": "SignCreateOrder expects 11 args"})
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

		txInfo, err := executables.GetCreateOrderTransaction(
			marketIndex, clientOrderIndex, baseAmount, price, isAsk,
			orderType, timeInForce, reduceOnly, triggerPrice, orderExpiry, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCancelOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return js.ValueOf(map[string]interface{}{"error": "SignCancelOrder expects 3 args"})
		}
		marketIndex := uint8(args[0].Int())
		orderIndex := int64(args[1].Int())
		nonce := int64(args[2].Int())

		txInfo, err := executables.GetCancelOrderTransaction(marketIndex, orderIndex, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignCancelAllOrders", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 3 {
			return js.ValueOf(map[string]interface{}{"error": "SignCancelAllOrders expects 3 args"})
		}
		timeInForce := uint8(args[0].Int())
		timeVal := int64(args[1].Int())
		nonce := int64(args[2].Int())

		txInfo, err := executables.GetCancelAllOrdersTransaction(timeInForce, timeVal, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignTransfer", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 4 {
			return js.ValueOf(map[string]interface{}{"error": "SignTransfer expects at least 4 args"})
		}
		toAccount := int64(args[0].Int())
		usdcAmount := int64(args[1].Int())
		fee := int64(0)
		if len(args) > 2 {
			fee = int64(args[2].Int())
		}
		memoStr := ""
		if len(args) > 3 {
			memoStr = args[3].String()
		}
		nonce := int64(args[4].Int())

		var memoArr [32]byte
		bs := []byte(memoStr)
		for i := 0; i < len(bs) && i < 32; i++ {
			memoArr[i] = bs[i]
		}

		txInfo, err := executables.GetTransferTransaction(toAccount, usdcAmount, fee, nonce, memoArr)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignWithdraw", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			return js.ValueOf(map[string]interface{}{"error": "SignWithdraw expects 2 args: usdcAmount, nonce"})
		}
		usdcAmount := uint64(args[0].Int())
		nonce := int64(args[1].Int())

		txInfo, err := executables.GetWithdrawTransaction(usdcAmount, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignUpdateLeverage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 4 {
			return js.ValueOf(map[string]interface{}{"error": "SignUpdateLeverage expects 4 args"})
		}
		marketIndex := uint8(args[0].Int())
		fraction := uint16(args[1].Int())
		marginMode := uint8(args[2].Int())
		nonce := int64(args[3].Int())

		txInfo, err := executables.GetUpdateLeverageTransaction(marketIndex, marginMode, fraction, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	js.Global().Set("SignModifyOrder", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 6 {
			return js.ValueOf(map[string]interface{}{"error": "SignModifyOrder expects 6 args"})
		}
		marketIndex := uint8(args[0].Int())
		index := int64(args[1].Int())
		baseAmount := int64(args[2].Int())
		price := uint32(args[3].Int())
		triggerPrice := uint32(args[4].Int())
		nonce := int64(args[5].Int())

		txInfo, err := executables.GetModifyOrderTransaction(marketIndex, index, baseAmount, price, triggerPrice, nonce)
		if err != nil {
			return js.ValueOf(map[string]interface{}{"error": wrapErr(err)})
		}
		return js.ValueOf(map[string]interface{}{"txInfo": txInfo, "error": ""})
	}))

	select {}
}


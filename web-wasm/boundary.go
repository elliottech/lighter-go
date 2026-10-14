//go:build js && wasm

package main

import (
	"fmt"
	"math"
	"syscall/js"
)

type argumentError string

// Only controlled validation messages may cross the JS boundary.
func argumentFailure(recovered any) error {
	if reason, ok := recovered.(argumentError); ok {
		return fmt.Errorf("invalid signer arguments: %s", reason)
	}
	return fmt.Errorf("invalid signer arguments")
}

// Both callbacks are registered once per method. Each invocation returns a
// native JS bound function, preserving the legacy lazy callable -> Promise API
// without allocating or retaining a per-request Go callback.
func register(name string, fn func(js.Value, []js.Value) any) {
	execute := js.FuncOf(func(this js.Value, args []js.Value) (result any) {
		defer func() {
			if recovered := recover(); recovered != nil {
				result = errPromise(argumentFailure(recovered))
			}
		}()
		return fn(this, args)
	})
	switch name {
	case "_lighterTest", "_resetClients", "_selectClient":
		js.Global().Set(name, execute)
	default:
		js.Global().Set(name, js.FuncOf(func(this js.Value, args []js.Value) any {
			boundArgs := make([]any, len(args)+1)
			boundArgs[0] = js.Null()
			for i, arg := range args {
				boundArgs[i+1] = arg
			}
			return execute.Value.Call("bind", boundArgs...)
		}))
	}
}

func promiseResult(work func(js.Value)) js.Value {
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		defer func() {
			if recover() != nil {
				resolve.Invoke(errToJson(fmt.Errorf("signer operation failed")))
			}
		}()
		work(resolve)
		return nil
	})
	// The Promise constructor invokes its executor synchronously.
	defer handler.Release()
	return js.Global().Get("Promise").New(handler)
}

func checkedInt(v js.Value) int {
	if v.Type() != js.TypeNumber {
		panic(argumentError("expected number"))
	}
	n := v.Float()
	if math.IsNaN(n) || math.IsInf(n, 0) || math.Abs(n) > 9007199254740991 {
		panic(argumentError("expected safe integer"))
	}
	// Fractions truncate toward zero, as legacy callers send float results (e.g. 10000 / leverage).
	return int(n)
}

func checkedString(v js.Value) string {
	if v.Type() != js.TypeString {
		panic(argumentError("expected string"))
	}
	s := v.String()
	if len(s) > 1<<20 {
		panic(argumentError("string too long"))
	}
	return s
}

func checkedBool(v js.Value) bool {
	if v.Type() != js.TypeBoolean {
		panic(argumentError("expected boolean"))
	}
	return v.Bool()
}

func checkedUint8(v int) uint8 {
	if v < 0 || v > math.MaxUint8 {
		panic(argumentError("uint8 overflow"))
	}
	return uint8(v)
}
func checkedUint16(v int) uint16 {
	if v < 0 || v > math.MaxUint16 {
		panic(argumentError("uint16 overflow"))
	}
	return uint16(v)
}
func checkedUint32(v int) uint32 {
	if v < 0 || uint64(v) > math.MaxUint32 {
		panic(argumentError("uint32 overflow"))
	}
	return uint32(v)
}
func checkedInt16(v int) int16 {
	if v < math.MinInt16 || v > math.MaxInt16 {
		panic(argumentError("int16 overflow"))
	}
	return int16(v)
}

// Memo bytes are opaque data, not quantities: a fractional element is rejected, not truncated.
func checkedByte(v js.Value) byte {
	if v.Type() == js.TypeNumber && math.Trunc(v.Float()) != v.Float() {
		panic(argumentError("expected integer byte"))
	}
	return checkedUint8(checkedInt(v))
}

func ResetClients(this js.Value, args []js.Value) any {
	for identity, client := range clientRegistry {
		// Drop the adapter's reference; runtime termination is still required.
		client.keyManager = nil
		delete(clientRegistry, identity)
	}
	clear(clients)
	return js.Global().Get("Promise").Call("resolve", map[string]any{"success": true})
}

type clientIdentity struct {
	account int64
	chain   uint32
	apiKey  uint8
}

var clientRegistry = make(map[clientIdentity]*LightClient)

func installClient(account int64, client *LightClient) {
	identity := clientIdentity{account, client.chainId, client.apiKeyIndex}
	if old := clientRegistry[identity]; old != nil {
		old.keyManager = nil
	}
	clientRegistry[identity] = client
	clients[account] = client
}

func SelectClient(this js.Value, args []js.Value) any {
	identity := clientIdentity{int64(checkedInt(args[0])), checkedUint32(checkedInt(args[1])), checkedUint8(checkedInt(args[2]))}
	client := clientRegistry[identity]
	if client == nil {
		return errPromise(fmt.Errorf("signer identity not initialized"))
	}
	clients[identity.account] = client
	return js.Global().Get("Promise").Call("resolve", map[string]any{"success": true})
}

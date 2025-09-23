//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func requireArgs(args []js.Value, want int) error {
	if len(args) < want {
		return fmt.Errorf("expected %d arguments, got %d", want, len(args))
	}
	return nil
}

func runJS(fn func() (interface{}, error)) interface{} {
	var (
		result interface{}
		err    error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		result, err = fn()
	}()
	if err != nil {
		return map[string]any{"err": err.Error()}
	}
	return result
}

func toUint8(v js.Value) uint8   { return uint8(v.Int()) }
func toUint16(v js.Value) uint16 { return uint16(v.Int()) }
func toUint32(v js.Value) uint32 { return uint32(v.Int()) }
func toInt64(v js.Value) int64   {
	var n int64
   	fmt.Sscan(v.String(), &n)
    return n
}
func toUint64(v js.Value) uint64 {
	var n uint64
   	fmt.Sscan(v.String(), &n)
    return n
}

func registerStrFunc(name string, handler func([]js.Value) (string, error)) {
	js.Global().Set(name, js.FuncOf(func(_ js.Value, args []js.Value) any {
		return runJS(func() (interface{}, error) {
			str, err := handler(args)
			if err != nil {
				return nil, err
			}
			return map[string]any{"str": str}, nil
		})
	}))
}

func registerErrFunc(name string, handler func([]js.Value) error) {
	js.Global().Set(name, js.FuncOf(func(_ js.Value, args []js.Value) any {
		return runJS(func() (interface{}, error) {
			if err := handler(args); err != nil {
				return nil, err
			}
			return nil, nil
		})
	}))
}

func registerAPIKeyFunc() {
	js.Global().Set("GenerateAPIKey", js.FuncOf(func(_ js.Value, args []js.Value) any {
		return runJS(func() (interface{}, error) {
			seed := ""
			if len(args) > 0 {
				seed = args[0].String()
			}
			privateKey, publicKey, err := generateAPIKey(seed)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"privateKey": privateKey,
				"publicKey":  publicKey,
			}, nil
		})
	}))
}

func main() {
	registerAPIKeyFunc()

	registerErrFunc("CreateClient", func(args []js.Value) error {
		if err := requireArgs(args, 5); err != nil {
			return err
		}
		return createClient(
			args[0].String(),
			args[1].String(),
			toUint32(args[2]),
			toUint8(args[3]),
			toInt64(args[4]),
		)
	})

	registerStrFunc("SignChangePubKey", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 2); err != nil {
			return "", err
		}
		return signChangePubKey(args[0].String(), toInt64(args[1]))
	})

	registerStrFunc("SignCreateOrder", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 11); err != nil {
			return "", err
		}
		return signCreateOrder(
			toUint8(args[0]),
			toInt64(args[1]),
			toInt64(args[2]),
			toUint32(args[3]),
			toUint8(args[4]),
			toUint8(args[5]),
			toUint8(args[6]),
			toUint8(args[7]),
			toUint32(args[8]),
			toInt64(args[9]),
			toInt64(args[10]),
		)
	})

	registerStrFunc("SignCancelOrder", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 3); err != nil {
			return "", err
		}
		return signCancelOrder(toUint8(args[0]), toInt64(args[1]), toInt64(args[2]))
	})

	registerStrFunc("SignWithdraw", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 2); err != nil {
			return "", err
		}
		return signWithdraw(toUint64(args[0]), toInt64(args[1]))
	})

	registerStrFunc("SignCreateSubAccount", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 1); err != nil {
			return "", err
		}
		return signCreateSubAccount(toInt64(args[0]))
	})

	registerStrFunc("SignCancelAllOrders", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 3); err != nil {
			return "", err
		}
		return signCancelAllOrders(toUint8(args[0]), toInt64(args[1]), toInt64(args[2]))
	})

	registerStrFunc("SignModifyOrder", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 6); err != nil {
			return "", err
		}
		return signModifyOrder(
			toUint8(args[0]),
			toInt64(args[1]),
			toInt64(args[2]),
			toUint32(args[3]),
			toUint32(args[4]),
			toInt64(args[5]),
		)
	})

	registerStrFunc("SignTransfer", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 5); err != nil {
			return "", err
		}
		return signTransfer(
			toInt64(args[0]),
			toInt64(args[1]),
			toInt64(args[2]),
			args[3].String(),
			toInt64(args[4]),
		)
	})

	registerStrFunc("SignCreatePublicPool", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 4); err != nil {
			return "", err
		}
		return signCreatePublicPool(
			toInt64(args[0]),
			toInt64(args[1]),
			toInt64(args[2]),
			toInt64(args[3]),
		)
	})

	registerStrFunc("SignUpdatePublicPool", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 5); err != nil {
			return "", err
		}
		return signUpdatePublicPool(
			toInt64(args[0]),
			toUint8(args[1]),
			toInt64(args[2]),
			toInt64(args[3]),
			toInt64(args[4]),
		)
	})

	registerStrFunc("SignMintShares", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 3); err != nil {
			return "", err
		}
		return signMintShares(toInt64(args[0]), toInt64(args[1]), toInt64(args[2]))
	})

	registerStrFunc("SignBurnShares", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 3); err != nil {
			return "", err
		}
		return signBurnShares(toInt64(args[0]), toInt64(args[1]), toInt64(args[2]))
	})

	registerStrFunc("SignUpdateLeverage", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 4); err != nil {
			return "", err
		}
		return signUpdateLeverage(
			toUint8(args[0]),
			toUint16(args[1]),
			toUint8(args[2]),
			toInt64(args[3]),
		)
	})

	registerStrFunc("CreateAuthToken", func(args []js.Value) (string, error) {
		deadline := int64(0)
		if len(args) > 0 {
			deadline = toInt64(args[0])
		}
		return createAuthToken(deadline)
	})

	registerErrFunc("SwitchAPIKey", func(args []js.Value) error {
		if err := requireArgs(args, 1); err != nil {
			return err
		}
		return switchAPIKey(toUint8(args[0]))
	})

	registerStrFunc("SignUpdateMargin", func(args []js.Value) (string, error) {
		if err := requireArgs(args, 4); err != nil {
			return "", err
		}
		return signUpdateMargin(
			toUint8(args[0]),
			toInt64(args[1]),
			toUint8(args[2]),
			toInt64(args[3]),
		)
	})

	select {}
}

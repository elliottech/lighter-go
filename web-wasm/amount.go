//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
)

func parseUnsignedAmount(value string) (uint64, error) {
	if len(value) == 0 || len(value) > 20 || value[0] == '0' {
		return 0, fmt.Errorf("assetAmount must be a canonical positive uint64")
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("assetAmount must be a canonical positive uint64")
		}
	}
	amount, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("assetAmount exceeds uint64")
	}
	return amount, nil
}

package types

import (
	"encoding/hex"
	"fmt"
)

func NewTxAttributesFromSkipNonce(skipNonce uint8) *L2TxAttributes {
	attr := L2TxAttributes{}
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func NewIntegratorTxAttributes(integratorAccountIndex int64, integratorTakerFee uint32, integratorMakerFee uint32, skipNonce uint8) *L2TxAttributes {
	attr := L2TxAttributes{}
	attr.IntegratorAccountIndex = &integratorAccountIndex
	attr.IntegratorTakerFee = &integratorTakerFee
	attr.IntegratorMakerFee = &integratorMakerFee
	if skipNonce == 1 {
		attr.SkipNonce = &skipNonce
	}
	return &attr
}

func NewTransactOpts(skipNonce uint8, nonce int64) *TransactOpts {
	txAttributes := NewTxAttributesFromSkipNonce(skipNonce)
	return &TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func NewIntegratorTransactOpts(integratorAccountIndex int64, integratorTakerFee uint32, integratorMakerFee uint32, skipNonce uint8, nonce int64) *TransactOpts {
	txAttributes := NewIntegratorTxAttributes(integratorAccountIndex, integratorTakerFee, integratorMakerFee, skipNonce)
	return &TransactOpts{
		Nonce:        &nonce,
		TxAttributes: txAttributes,
	}
}

func ParseMemo(memoStr string) ([32]byte, error) {
	var memo [32]byte

	if len(memoStr) == 66 {
		if memoStr[0:2] == "0x" {
			memoStr = memoStr[2:66]
		} else {
			return memo, fmt.Errorf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr))
		}
	}

	if len(memoStr) == 64 {
		b, err := hex.DecodeString(memoStr)
		if err != nil {
			return memo, fmt.Errorf("failed to decode hex string. err: %v", err)
		}
		for i := 0; i < 32; i++ {
			memo[i] = b[i]
		}
	} else if len(memoStr) == 32 {
		for i := 0; i < 32; i++ {
			memo[i] = byte(memoStr[i])
		}
	} else {
		return memo, fmt.Errorf("memo expected to be 32 bytes or 64 hex encoded or 66 if 0x hex encoded -- long but received %v", len(memoStr))
	}

	return memo, nil
}

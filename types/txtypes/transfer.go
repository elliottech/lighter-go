package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
)

var _ TxInfo = (*L2TransferTxInfo)(nil)

type L2TransferTxInfo struct {
	FromAccountIndex int64
	ApiKeyIndex      uint8

	ToAccountIndex int64
	USDCAmount     int64 // USDCAmount is given with 6 decimals

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`
}

func (txInfo *L2TransferTxInfo) Validate() error {
	// plus one for treasury account
	if txInfo.FromAccountIndex < MinAccountIndex+1 {
		return ErrFromAccountIndexTooLow
	}
	if txInfo.FromAccountIndex > MaxAccountIndex {
		return ErrFromAccountIndexTooHigh
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}

	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.ToAccountIndex < MinAccountIndex+1 {
		return ErrToAccountIndexTooLow
	}
	if txInfo.ToAccountIndex > MaxAccountIndex {
		return ErrToAccountIndexTooHigh
	}

	if txInfo.USDCAmount <= 0 {
		return ErrTransferAmountTooLow
	}
	if txInfo.USDCAmount > MaxTransferAmount {
		return ErrTransferAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2TransferTxInfo) GetTxType() uint8 {
	return TxTypeL2Transfer
}

func (txInfo *L2TransferTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2TransferTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2TransferTxInfo) Hash(lighterChainId uint32, extra ...g.Element) (msgHash []byte, err error) {
	elems := make([]g.Element, 0, 9)

	elems = append(elems, g.FromUint32(lighterChainId))
	elems = append(elems, g.FromUint32(TxTypeL2Transfer))
	elems = append(elems, g.FromInt64(txInfo.Nonce))
	elems = append(elems, g.FromInt64(txInfo.ExpiredAt))

	elems = append(elems, g.FromInt64(txInfo.FromAccountIndex))
	elems = append(elems, g.FromUint32(uint32(txInfo.ApiKeyIndex)))
	elems = append(elems, g.FromInt64(txInfo.ToAccountIndex))
	elems = append(elems, g.FromUint64(uint64(txInfo.USDCAmount)&0xFFFFFFFF)) //nolint:gosec
	elems = append(elems, g.FromUint64(uint64(txInfo.USDCAmount)>>32))        //nolint:gosec

	return p2.HashToQuinticExtension(elems).ToLittleEndianBytes(), nil
}

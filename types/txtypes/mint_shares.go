package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
)

var _ TxInfo = (*L2MintSharesTxInfo)(nil)

type L2MintSharesTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	PublicPoolIndex int64
	ShareAmount     int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string
}

func (txInfo *L2MintSharesTxInfo) GetTxType() uint8 {
	return TxTypeL2MintShares
}

func (txInfo *L2MintSharesTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2MintSharesTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2MintSharesTxInfo) Validate() error {
	if txInfo.AccountIndex < MinAccountIndex {
		return ErrFromAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrFromAccountIndexTooHigh
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	// PublicPoolIndex
	if txInfo.PublicPoolIndex < MinAccountIndex {
		return ErrPublicPoolIndexTooLow
	}
	if txInfo.PublicPoolIndex > MaxAccountIndex {
		return ErrPublicPoolIndexTooHigh
	}

	if txInfo.ShareAmount < MinPoolSharesToMintOrBurn {
		return ErrPoolMintShareAmountTooLow
	}
	if txInfo.ShareAmount > MaxPoolSharesToMintOrBurn {
		return ErrPoolMintShareAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2MintSharesTxInfo) Hash(lighterChainId uint32, extra ...g.Element) (msgHash []byte, err error) {
	elems := make([]g.Element, 0, 8)

	elems = append(elems, g.FromUint32(lighterChainId))
	elems = append(elems, g.FromUint32(TxTypeL2MintShares))
	elems = append(elems, g.FromInt64(txInfo.Nonce))
	elems = append(elems, g.FromInt64(txInfo.ExpiredAt))

	elems = append(elems, g.FromInt64(txInfo.AccountIndex))
	elems = append(elems, g.FromUint32(uint32(txInfo.ApiKeyIndex)))
	elems = append(elems, g.FromInt64(txInfo.PublicPoolIndex))
	elems = append(elems, g.FromInt64(txInfo.ShareAmount))

	return p2.HashToQuinticExtension(elems).ToLittleEndianBytes(), nil
}

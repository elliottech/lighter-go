package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2CancelOrderTxInfo)(nil)

type L2CancelOrderTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex int16
	Index       int64 // Client Order Index or Order Index of the order to cancel

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2CancelOrderTxInfo) GetTxType() uint8 {
	return TxTypeL2CancelOrder
}

func (txInfo *L2CancelOrderTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2CancelOrderTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2CancelOrderTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	// AccountIndex
	if txInfo.AccountIndex < MinAccountIndex {
		return ErrAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrAccountIndexTooHigh
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	// MarketIndex
	isSpotMarket := txInfo.MarketIndex >= MinSpotMarketIndex && txInfo.MarketIndex <= MaxSpotMarketIndex
	isPerpsMarket := txInfo.MarketIndex >= MinPerpsMarketIndex && txInfo.MarketIndex <= MaxPerpsMarketIndex
	if !isSpotMarket && !isPerpsMarket {
		return ErrInvalidMarketIndex
	}

	// Index
	if txInfo.Index < MinClientOrderIndex && txInfo.Index < MinOrderIndex {
		return ErrOrderIndexTooLow
	}
	if txInfo.Index > MaxClientOrderIndex && txInfo.Index > MaxOrderIndex {
		return ErrOrderIndexTooHigh
	}

	// Nonce
	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2CancelOrderTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 8)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2CancelOrder))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.Index))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

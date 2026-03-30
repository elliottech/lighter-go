package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2UpdateLeverageTxInfo)(nil)

type L2UpdateLeverageTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex           int16
	InitialMarginFraction uint16
	MarginMode            uint8

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2UpdateLeverageTxInfo) GetTxType() uint8 {
	return TxTypeL2UpdateLeverage
}

func (txInfo *L2UpdateLeverageTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2UpdateLeverageTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2UpdateLeverageTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

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
	// MarketIndex
	if txInfo.MarketIndex == NilMarketIndex {
		return ErrInvalidMarketIndex
	}

	if txInfo.MarginMode != CrossMargin && txInfo.MarginMode != IsolatedMargin {
		return ErrInvalidMarginMode
	}

	// InitialMarginFraction
	if txInfo.InitialMarginFraction <= 0 {
		return ErrInitialMarginFractionTooLow
	}
	if txInfo.InitialMarginFraction > uint16(MarginFractionTick) { //nolint:gosec
		return ErrInitialMarginFractionTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2UpdateLeverageTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 9)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2UpdateLeverage))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.InitialMarginFraction))
	elems = append(elems, g.GoldilocksField(txInfo.MarginMode))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

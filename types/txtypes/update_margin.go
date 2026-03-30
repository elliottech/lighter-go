package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2UpdateMarginTxInfo)(nil)

type L2UpdateMarginTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex int16
	USDCAmount  int64
	Direction   uint8

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2UpdateMarginTxInfo) GetTxType() uint8 {
	return TxTypeL2UpdateMargin
}

func (txInfo *L2UpdateMarginTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2UpdateMarginTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2UpdateMarginTxInfo) Validate() error {
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
	if txInfo.MarketIndex < MinPerpsMarketIndex || txInfo.MarketIndex > MaxPerpsMarketIndex {
		return ErrInvalidMarketIndex
	}

	if txInfo.USDCAmount == 0 {
		return ErrTransferAmountTooLow
	}
	if txInfo.USDCAmount > MaxTransferAmount {
		return ErrTransferAmountTooHigh
	}
	if txInfo.Direction != RemoveFromIsolatedMargin && txInfo.Direction != AddToIsolatedMargin {
		return ErrInvalidUpdateMarginDirection
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2UpdateMarginTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 10)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2UpdateMargin))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.USDCAmount&0xFFFFFFFF))
	elems = append(elems, g.GoldilocksField(txInfo.USDCAmount>>32))
	elems = append(elems, g.GoldilocksField(txInfo.Direction))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

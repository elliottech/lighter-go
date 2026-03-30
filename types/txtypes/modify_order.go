package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2ModifyOrderTxInfo)(nil)

type L2ModifyOrderTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex  int16
	Index        int64 // Client Order Index or Order Index of the order to modify
	BaseAmount   int64
	Price        uint32
	TriggerPrice uint32

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2ModifyOrderTxInfo) GetTxType() uint8 {
	return TxTypeL2ModifyOrder
}

func (txInfo *L2ModifyOrderTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2ModifyOrderTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2ModifyOrderTxInfo) Validate() error {
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
		return ErrClientOrderIndexTooLow
	}
	if txInfo.Index > MaxClientOrderIndex && txInfo.Index > MaxOrderIndex {
		return ErrClientOrderIndexTooHigh
	}

	// BaseAmount
	if txInfo.BaseAmount != NilOrderBaseAmount && txInfo.BaseAmount < MinOrderBaseAmount {
		return ErrBaseAmountTooLow
	}
	if txInfo.BaseAmount > MaxOrderBaseAmount {
		return ErrBaseAmountTooHigh
	}

	// Price
	if txInfo.Price < MinOrderPrice {
		return ErrPriceTooLow
	}
	if txInfo.Price > MaxOrderPrice {
		return ErrPriceTooHigh
	}

	// TriggerPrice
	if (txInfo.TriggerPrice < MinOrderTriggerPrice || txInfo.TriggerPrice > MaxOrderTriggerPrice) && txInfo.TriggerPrice != NilOrderTriggerPrice {
		return ErrOrderTriggerPriceInvalid
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

func (txInfo *L2ModifyOrderTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 11)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2ModifyOrder))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.Index))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount))
	elems = append(elems, g.GoldilocksField(txInfo.Price))
	elems = append(elems, g.GoldilocksField(txInfo.TriggerPrice))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

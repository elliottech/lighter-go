package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*BinaryOptionsOperatorSettleMarketTxInfo)(nil)

// BinaryOptionsOperatorSettleMarketTxInfo settles a binary options order book by providing the settlement price.
// Only the market operator can send this transaction. For discrete markets the price must be
// 0 or the settlement cap, for continuous markets it can be any price in [0, cap].
// A refund sends IsRefund with a zero settlement price and settles at the market's default price.
type BinaryOptionsOperatorSettleMarketTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex       int16
	PublicMarketIndex int64
	SettlementPrice   uint32
	IsRefund          uint8

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *BinaryOptionsOperatorSettleMarketTxInfo) GetTxType() uint8 {
	return TxTypeBinaryOptionsOperatorSettleMarket
}

func (txInfo *BinaryOptionsOperatorSettleMarketTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *BinaryOptionsOperatorSettleMarketTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *BinaryOptionsOperatorSettleMarketTxInfo) Validate() error {
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
	if txInfo.MarketIndex < MinBinaryOptionsMarketIndex || txInfo.MarketIndex > MaxBinaryOptionsMarketIndex {
		return ErrInvalidMarketIndex
	}

	// PublicMarketIndex
	if txInfo.PublicMarketIndex < MinMarketIndex || txInfo.PublicMarketIndex > MaxMarketIndex ||
		txInfo.PublicMarketIndex == NilMarketIndex {
		return ErrInvalidMarketIndex
	}

	if txInfo.IsRefund > 1 {
		return ErrInvalidIsRefund
	}
	if txInfo.IsRefund > 0 && txInfo.SettlementPrice != 0 {
		return ErrInvalidSettlementPrice
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *BinaryOptionsOperatorSettleMarketTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 10)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeBinaryOptionsOperatorSettleMarket))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))
	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.PublicMarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.SettlementPrice))
	elems = append(elems, g.GoldilocksField(txInfo.IsRefund))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

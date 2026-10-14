package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*BinaryOptionsOperatorUpdateMarketSlotTxInfo)(nil)

// BinaryOptionsOperatorUpdateMarketSlotTxInfo updates the slot-level settings of a binary options market slot. Only the
// insurance fund operator can send this transaction. The slot may hold an active market, a settled
// market or no market at all; slot settings stay on the slot across settlements. The only setting for
// now is the market operator, which authorizes CreateMarket and UpdateMarket on the slot.
type BinaryOptionsOperatorUpdateMarketSlotTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	MarketIndex          int16
	OperatorAccountIndex int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *BinaryOptionsOperatorUpdateMarketSlotTxInfo) GetTxType() uint8 {
	return TxTypeBinaryOptionsOperatorUpdateMarketSlot
}

func (txInfo *BinaryOptionsOperatorUpdateMarketSlotTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *BinaryOptionsOperatorUpdateMarketSlotTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *BinaryOptionsOperatorUpdateMarketSlotTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	if txInfo.AccountIndex != InsuranceFundOperatorAccountIndex {
		return ErrAccountIndexMustBtInsuranceFundOperator
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

	// OperatorAccountIndex
	if txInfo.OperatorAccountIndex < MinAccountIndex {
		return ErrToAccountIndexTooLow
	}
	if txInfo.OperatorAccountIndex > MaxAccountIndex {
		return ErrToAccountIndexTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *BinaryOptionsOperatorUpdateMarketSlotTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 8)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeBinaryOptionsOperatorUpdateMarketSlot))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))
	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MarketIndex))
	elems = append(elems, g.GoldilocksField(txInfo.OperatorAccountIndex))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

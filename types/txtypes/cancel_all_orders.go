package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2CancelAllOrdersTxInfo)(nil)

type L2CancelAllOrdersTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	TimeInForce uint8
	Time        int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2CancelAllOrdersTxInfo) GetTxType() uint8 {
	return TxTypeL2CancelAllOrders
}

func (txInfo *L2CancelAllOrdersTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2CancelAllOrdersTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2CancelAllOrdersTxInfo) Validate() error {
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

	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}
	if txInfo.ApiKeyIndex > MaxApiKeyIndex && txInfo.ApiKeyIndex != NilApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	// Nonce
	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	// TimeInForce and Time
	switch txInfo.TimeInForce {
	case ImmediateCancelAll:
		if txInfo.Time != NilOrderExpiry {
			return ErrCancelAllTimeisNotNill
		}
	case ScheduledCancelAll:
		if txInfo.Time < MinOrderExpiry || txInfo.Time > MaxOrderExpiry {
			return ErrCancelAllTimeIsNotInRange
		}
	case AbortScheduledCancelAll:
		if txInfo.Time != 0 {
			return ErrCancelAllTimeisNotNill
		}
	default:
		return ErrInvalidCancelAllTimeInForce
	}

	return nil
}

func (txInfo *L2CancelAllOrdersTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 8)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2CancelAllOrders))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.TimeInForce))
	elems = append(elems, g.GoldilocksField(txInfo.Time))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

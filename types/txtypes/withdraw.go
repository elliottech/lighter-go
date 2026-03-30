package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2WithdrawTxInfo)(nil)

type L2WithdrawTxInfo struct {
	FromAccountIndex int64
	ApiKeyIndex      uint8
	AssetIndex       int16
	RouteType        uint8
	Amount           uint64
	ExpiredAt        int64
	Nonce            int64
	Sig              []byte
	SignedHash       string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2WithdrawTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}

	if txInfo.FromAccountIndex < MinAccountIndex {
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

	// AssetIndex
	if txInfo.AssetIndex < MinAssetIndex {
		return ErrAssetIndexTooLow
	}
	if txInfo.AssetIndex > MaxAssetIndex {
		return ErrAssetIndexTooHigh
	}

	// RouteType
	if txInfo.RouteType != AssetRouteType_Perps && txInfo.RouteType != AssetRouteType_Spot {
		return ErrRouteTypeInvalid
	}

	// Amount
	if txInfo.Amount == 0 {
		return ErrWithdrawalAmountTooLow
	}
	if txInfo.Amount > MaxWithdrawalAmount {
		return ErrWithdrawalAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2WithdrawTxInfo) GetTxType() uint8 {
	return TxTypeL2Withdraw
}

func (txInfo *L2WithdrawTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2WithdrawTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2WithdrawTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 10)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2Withdraw))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.FromAccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.AssetIndex))
	elems = append(elems, g.GoldilocksField(txInfo.RouteType))
	elems = append(elems, g.GoldilocksField(txInfo.Amount&0xFFFFFFFF))
	elems = append(elems, g.GoldilocksField(txInfo.Amount>>32))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

package txtypes

import (
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ TxInfo = (*L2StakeAssetsTxInfo)(nil)

type L2StakeAssetsTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	StakingPoolIndex int64
	ShareAmount      int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2StakeAssetsTxInfo) GetTxType() uint8 {
	return TxTypeL2StakeAssets
}

func (txInfo *L2StakeAssetsTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2StakeAssetsTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2StakeAssetsTxInfo) Validate() error {
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

	// StakingPoolIndex
	if txInfo.StakingPoolIndex < MinSubAccountIndex {
		return ErrStakingPoolIndexTooLow
	}
	if txInfo.StakingPoolIndex > MaxAccountIndex {
		return ErrStakingPoolIndexTooHigh
	}

	if txInfo.ShareAmount < MinStakingSharesToMintOrBurn {
		return ErrPoolStakeAssetsAmountTooLow
	}
	if txInfo.ShareAmount > MaxStakingSharesToMintOrBurn {
		return ErrPoolStakeAssetsAmountTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2StakeAssetsTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 8)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2StakeAssets))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.StakingPoolIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ShareAmount))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

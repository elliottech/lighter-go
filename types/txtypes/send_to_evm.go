package txtypes

import (
	"fmt"

	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
	"github.com/ethereum/go-ethereum/common"
)

var _ TxInfo = (*L2SendToEvmTxInfo)(nil)

type L2SendToEvmTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8
	AssetIndex   uint32
	EvmRecipient string
	RouteType    uint8
	BaseAmount   uint64
	USDCFee      int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	OwnerSig   string
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2SendToEvmTxInfo) Validate() error {
	if err := txInfo.L2TxAttributes.Validate(); err != nil {
		return err
	}
	if !txInfo.L2TxAttributes.IsEmpty() {
		return ErrInvalidAttributeType
	}

	// Core's minimum is zero; -1 is only a client-selection sentinel.
	if txInfo.AccountIndex < 0 {
		return ErrInteropAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrAccountIndexTooHigh
	}

	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.AssetIndex < MinAssetIndex {
		return ErrAssetIndexTooLow
	}
	if txInfo.AssetIndex > MaxAssetIndex {
		return ErrAssetIndexTooHigh
	}

	if !common.IsHexAddress(txInfo.EvmRecipient) || common.HexToAddress(txInfo.EvmRecipient) == (common.Address{}) {
		return ErrEvmRecipientInvalid
	}

	if txInfo.RouteType != AssetRouteType_Perps && txInfo.RouteType != AssetRouteType_Spot {
		return ErrRouteTypeInvalid
	}

	if txInfo.BaseAmount == 0 {
		return ErrTransferAmountTooLow
	}
	if txInfo.BaseAmount > uint64(MaxTransferAmount) {
		return ErrTransferAmountTooHigh
	}

	if txInfo.USDCFee < 0 {
		return ErrTransferFeeNegative
	}
	if txInfo.USDCFee > MaxTransferAmount {
		return ErrTransferFeeTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2SendToEvmTxInfo) GetTxType() uint8 {
	return TxTypeL2SendToEvm
}

func (txInfo *L2SendToEvmTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2SendToEvmTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

// GetOwnerSignatureBody returns the EIP-191 personal-sign message, not its digest.
func (txInfo *L2SendToEvmTxInfo) GetOwnerSignatureBody(chainId uint32) string {
	signatureBody := fmt.Sprintf(
		TemplateSendToEvm,
		getHex10FromUint64(uint64(txInfo.Nonce)),
		getHex10FromUint64(uint64(txInfo.AccountIndex)),
		getHex10FromUint64(uint64(txInfo.RouteType)),
		getHex10FromUint64(uint64(txInfo.ApiKeyIndex)),
		common.HexToAddress(txInfo.EvmRecipient).Hex(),
		getHex10FromUint64(uint64(txInfo.AssetIndex)),
		getHex10FromUint64(txInfo.BaseAmount),
		getHex10FromUint64(uint64(txInfo.USDCFee)),
		getHex10FromUint64(uint64(chainId)),
	)
	return signatureBody
}

func (txInfo *L2SendToEvmTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 17)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2SendToEvm))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.AssetIndex))
	elems = appendAddressLimbs(elems, txInfo.EvmRecipient)
	elems = append(elems, g.GoldilocksField(txInfo.RouteType))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount&0xFFFFFFFF))
	elems = append(elems, g.GoldilocksField(txInfo.BaseAmount>>32))
	elems = append(elems, g.GoldilocksField(uint64(txInfo.USDCFee)&0xFFFFFFFF))
	elems = append(elems, g.GoldilocksField(uint64(txInfo.USDCFee)>>32))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

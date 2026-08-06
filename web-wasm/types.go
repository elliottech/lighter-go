package main

import (
	"fmt"

	"github.com/elliottech/lighter-go/signer"
	"github.com/elliottech/lighter-go/types/txtypes"
	gFp5 "github.com/elliottech/poseidon_crypto/field/goldilocks_quintic_extension"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

type TransactOpts struct {
	FromAccountIndex *int64
	ApiKeyIndex      *uint8
	ExpiredAt        int64
	Nonce            *int64
	DryRun           bool

	SkipNonce            bool
	CancelAllMarketIndex *int16
}

func (ops *TransactOpts) L2TxAttributes() txtypes.L2TxAttributes {
	attributes := txtypes.L2TxAttributes{}
	if ops.SkipNonce {
		attributes[txtypes.AttributeTypeSkipTxNonce] = 1
	}
	if ops.CancelAllMarketIndex != nil {
		attributes[txtypes.AttributeTypeCancelAllMarketIndex] = int(*ops.CancelAllMarketIndex)
	}
	if len(attributes) == 0 {
		return nil
	}
	return attributes
}

type ChangePubKeyReq struct {
	PubKey [40]byte
}

type TransferTxReq struct {
	ToAccountIndex int64
	AssetIndex     int16
	FromRouteType  uint8
	ToRouteType    uint8
	Amount         int64
	USDCFee        int64
	Memo           [32]byte
}

type WithdrawTxReq struct {
	AssetIndex int16
	RouteType  uint8
	Amount     uint64
}

type CreateOrderTxReq struct {
	MarketIndex      int16
	ClientOrderIndex int64
	BaseAmount       int64
	Price            uint32
	IsAsk            uint8
	Type             uint8
	TimeInForce      uint8
	ReduceOnly       uint8
	TriggerPrice     uint32
	OrderExpiry      int64
}

type CreateGroupedOrdersTxReq struct {
	GroupingType uint8
	Orders       []*CreateOrderTxReq
}

type ModifyOrderTxReq struct {
	MarketIndex  int16
	Index        int64
	BaseAmount   int64
	Price        uint32
	TriggerPrice uint32
}

type CancelOrderTxReq struct {
	MarketIndex int16
	Index       int64
}

type CancelAllOrdersTxReq struct {
	TimeInForce uint8
	Time        int64
}

type StakeAssetsTxReq struct {
	StakingPoolIndex int64
	ShareAmount      int64
}

type UnstakeAssetsTxReq struct {
	StakingPoolIndex int64
	ShareAmount      int64
}

type CreatePublicPoolTxReq struct {
	OperatorFee          int64
	InitialTotalShares   int64
	MinOperatorShareRate uint16
}

type UpdatePublicPoolTxReq struct {
	PublicPoolIndex      int64
	Status               uint8
	OperatorFee          int64
	MinOperatorShareRate uint16
}

type MintSharesTxReq struct {
	PublicPoolIndex int64
	ShareAmount     int64
}

type BurnSharesTxReq struct {
	PublicPoolIndex int64
	ShareAmount     int64
}

type UpdateLeverageTxReq struct {
	MarketIndex           int16
	InitialMarginFraction uint16
	MarginMode            uint8
}

type UpdateAccountConfigTxReq struct {
	AccountTradingMode uint8
}

type UpdateAccountAssetConfigTxReq struct {
	AssetIndex      int16
	AssetMarginMode uint8
}

type UpdateMarginTxReq struct {
	MarketIndex int16
	USDCAmount  int64
	Direction   uint8
}

type PublicKey = gFp5.Element

func ConvertChangePubKeyTx(tx *ChangePubKeyReq, ops *TransactOpts) *txtypes.L2ChangePubKeyTxInfo {
	return &txtypes.L2ChangePubKeyTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		PubKey:         tx.PubKey[:],
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertCreateSubAccountTx(ops *TransactOpts) *txtypes.L2CreateSubAccountTxInfo {
	return &txtypes.L2CreateSubAccountTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertCreatePublicPoolTx(tx *CreatePublicPoolTxReq, ops *TransactOpts) *txtypes.L2CreatePublicPoolTxInfo {
	return &txtypes.L2CreatePublicPoolTxInfo{
		AccountIndex:         *ops.FromAccountIndex,
		ApiKeyIndex:          *ops.ApiKeyIndex,
		OperatorFee:          tx.OperatorFee,
		InitialTotalShares:   tx.InitialTotalShares,
		MinOperatorShareRate: tx.MinOperatorShareRate,
		ExpiredAt:            ops.ExpiredAt,
		Nonce:                *ops.Nonce,
		L2TxAttributes:       ops.L2TxAttributes(),
	}
}

func ConvertUpdatePublicPoolTx(tx *UpdatePublicPoolTxReq, ops *TransactOpts) *txtypes.L2UpdatePublicPoolTxInfo {
	return &txtypes.L2UpdatePublicPoolTxInfo{
		AccountIndex:         *ops.FromAccountIndex,
		ApiKeyIndex:          *ops.ApiKeyIndex,
		PublicPoolIndex:      tx.PublicPoolIndex,
		Status:               tx.Status,
		OperatorFee:          tx.OperatorFee,
		MinOperatorShareRate: tx.MinOperatorShareRate,
		ExpiredAt:            ops.ExpiredAt,
		Nonce:                *ops.Nonce,
		L2TxAttributes:       ops.L2TxAttributes(),
	}
}

func ConvertCreateOrderTx(tx *CreateOrderTxReq, ops *TransactOpts) *txtypes.L2CreateOrderTxInfo {
	return &txtypes.L2CreateOrderTxInfo{
		AccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:  *ops.ApiKeyIndex,
		OrderInfo: &txtypes.OrderInfo{
			MarketIndex:      tx.MarketIndex,
			ClientOrderIndex: tx.ClientOrderIndex,
			BaseAmount:       tx.BaseAmount,
			Price:            tx.Price,
			IsAsk:            tx.IsAsk,
			Type:             tx.Type,
			TimeInForce:      tx.TimeInForce,
			ReduceOnly:       tx.ReduceOnly,
			TriggerPrice:     tx.TriggerPrice,
			OrderExpiry:      tx.OrderExpiry,
		},
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertCancelOrderTx(tx *CancelOrderTxReq, ops *TransactOpts) *txtypes.L2CancelOrderTxInfo {
	return &txtypes.L2CancelOrderTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		MarketIndex:    tx.MarketIndex,
		Index:          tx.Index,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertCancelAllOrdersTx(tx *CancelAllOrdersTxReq, ops *TransactOpts) *txtypes.L2CancelAllOrdersTxInfo {
	return &txtypes.L2CancelAllOrdersTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		TimeInForce:    tx.TimeInForce,
		Time:           tx.Time,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertModifyOrderTx(tx *ModifyOrderTxReq, ops *TransactOpts) *txtypes.L2ModifyOrderTxInfo {
	return &txtypes.L2ModifyOrderTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		MarketIndex:    tx.MarketIndex,
		Index:          tx.Index,
		BaseAmount:     tx.BaseAmount,
		Price:          tx.Price,
		TriggerPrice:   tx.TriggerPrice,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertTransferTx(tx *TransferTxReq, ops *TransactOpts) *txtypes.L2TransferTxInfo {
	return &txtypes.L2TransferTxInfo{
		FromAccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:      *ops.ApiKeyIndex,
		ToAccountIndex:   tx.ToAccountIndex,
		AssetIndex:       tx.AssetIndex,
		FromRouteType:    tx.FromRouteType,
		ToRouteType:      tx.ToRouteType,
		Amount:           tx.Amount,
		USDCFee:          tx.USDCFee,
		Memo:             tx.Memo,
		ExpiredAt:        ops.ExpiredAt,
		Nonce:            *ops.Nonce,
		L2TxAttributes:   ops.L2TxAttributes(),
	}
}

func ConvertMintSharesTx(tx *MintSharesTxReq, ops *TransactOpts) *txtypes.L2MintSharesTxInfo {
	return &txtypes.L2MintSharesTxInfo{
		AccountIndex:    *ops.FromAccountIndex,
		ApiKeyIndex:     *ops.ApiKeyIndex,
		PublicPoolIndex: tx.PublicPoolIndex,
		ShareAmount:     tx.ShareAmount,
		ExpiredAt:       ops.ExpiredAt,
		Nonce:           *ops.Nonce,
		L2TxAttributes:  ops.L2TxAttributes(),
	}
}

func ConvertBurnSharesTx(tx *BurnSharesTxReq, ops *TransactOpts) *txtypes.L2BurnSharesTxInfo {
	return &txtypes.L2BurnSharesTxInfo{
		AccountIndex:    *ops.FromAccountIndex,
		ApiKeyIndex:     *ops.ApiKeyIndex,
		PublicPoolIndex: tx.PublicPoolIndex,
		ShareAmount:     tx.ShareAmount,
		ExpiredAt:       ops.ExpiredAt,
		Nonce:           *ops.Nonce,
		L2TxAttributes:  ops.L2TxAttributes(),
	}
}

func ConvertStakeAssetsTx(tx *StakeAssetsTxReq, ops *TransactOpts) *txtypes.L2StakeAssetsTxInfo {
	return &txtypes.L2StakeAssetsTxInfo{
		AccountIndex:     *ops.FromAccountIndex,
		ApiKeyIndex:      *ops.ApiKeyIndex,
		StakingPoolIndex: tx.StakingPoolIndex,
		ShareAmount:      tx.ShareAmount,
		ExpiredAt:        ops.ExpiredAt,
		Nonce:            *ops.Nonce,
		L2TxAttributes:   ops.L2TxAttributes(),
	}
}

func ConvertUnstakeAssetsTx(tx *UnstakeAssetsTxReq, ops *TransactOpts) *txtypes.L2UnstakeAssetsTxInfo {
	return &txtypes.L2UnstakeAssetsTxInfo{
		AccountIndex:     *ops.FromAccountIndex,
		ApiKeyIndex:      *ops.ApiKeyIndex,
		StakingPoolIndex: tx.StakingPoolIndex,
		ShareAmount:      tx.ShareAmount,
		ExpiredAt:        ops.ExpiredAt,
		Nonce:            *ops.Nonce,
		L2TxAttributes:   ops.L2TxAttributes(),
	}
}

func ConvertUpdateLeverageTx(tx *UpdateLeverageTxReq, ops *TransactOpts) *txtypes.L2UpdateLeverageTxInfo {
	return &txtypes.L2UpdateLeverageTxInfo{
		AccountIndex:          *ops.FromAccountIndex,
		ApiKeyIndex:           *ops.ApiKeyIndex,
		MarketIndex:           tx.MarketIndex,
		InitialMarginFraction: tx.InitialMarginFraction,
		MarginMode:            tx.MarginMode,
		ExpiredAt:             ops.ExpiredAt,
		Nonce:                 *ops.Nonce,
		L2TxAttributes:        ops.L2TxAttributes(),
	}
}

func ConvertUpdateAccountConfigTx(tx *UpdateAccountConfigTxReq, ops *TransactOpts) *txtypes.L2UpdateAccountConfigTxInfo {
	return &txtypes.L2UpdateAccountConfigTxInfo{
		AccountIndex:       *ops.FromAccountIndex,
		ApiKeyIndex:        *ops.ApiKeyIndex,
		AccountTradingMode: tx.AccountTradingMode,
		ExpiredAt:          ops.ExpiredAt,
		Nonce:              *ops.Nonce,
		L2TxAttributes:     ops.L2TxAttributes(),
	}
}

func ConvertUpdateAccountAssetConfigTx(tx *UpdateAccountAssetConfigTxReq, ops *TransactOpts) *txtypes.L2UpdateAccountAssetConfigTxInfo {
	return &txtypes.L2UpdateAccountAssetConfigTxInfo{
		AccountIndex:    *ops.FromAccountIndex,
		ApiKeyIndex:     *ops.ApiKeyIndex,
		AssetIndex:      tx.AssetIndex,
		AssetMarginMode: tx.AssetMarginMode,
		ExpiredAt:       ops.ExpiredAt,
		Nonce:           *ops.Nonce,
		L2TxAttributes:  ops.L2TxAttributes(),
	}
}

func ConvertUpdateMarginTx(tx *UpdateMarginTxReq, ops *TransactOpts) *txtypes.L2UpdateMarginTxInfo {
	return &txtypes.L2UpdateMarginTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		MarketIndex:    tx.MarketIndex,
		USDCAmount:     tx.USDCAmount,
		Direction:      tx.Direction,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}
}

func ConvertWithdrawTx(tx *WithdrawTxReq, ops *TransactOpts) *txtypes.L2WithdrawTxInfo {
	return &txtypes.L2WithdrawTxInfo{
		FromAccountIndex: *ops.FromAccountIndex,
		ApiKeyIndex:      *ops.ApiKeyIndex,
		AssetIndex:       tx.AssetIndex,
		RouteType:        tx.RouteType,
		Amount:           tx.Amount,
		ExpiredAt:        ops.ExpiredAt,
		Nonce:            *ops.Nonce,
		L2TxAttributes:   ops.L2TxAttributes(),
	}
}

func ConstructChangePubKeyTx(key signer.KeyManager, lighterChainId uint32, tx *ChangePubKeyReq, ops *TransactOpts) (*txtypes.L2ChangePubKeyTxInfo, error) {
	convertedTx := ConvertChangePubKeyTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature

	if !IsZeroByteSlice(convertedTx.PubKey) {
		pk := key.PubKeyBytes()
		msgHash, _ := convertedTx.Hash(lighterChainId)

		if err := schnorr.Validate(pk[:], msgHash, convertedTx.Sig); err != nil {
			return nil, fmt.Errorf("failed to validate signature. error: %v", err)
		}
	}
	return convertedTx, nil
}

func ConstructCreateSubAccountTx(key signer.KeyManager, lighterChainId uint32, ops *TransactOpts) (*txtypes.L2CreateSubAccountTxInfo, error) {
	convertedTx := ConvertCreateSubAccountTx(ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructCreatePublicPoolTx(key signer.KeyManager, lighterChainId uint32, tx *CreatePublicPoolTxReq, ops *TransactOpts) (*txtypes.L2CreatePublicPoolTxInfo, error) {
	convertedTx := ConvertCreatePublicPoolTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdatePublicPoolTx(key signer.KeyManager, lighterChainId uint32, tx *UpdatePublicPoolTxReq, ops *TransactOpts) (*txtypes.L2UpdatePublicPoolTxInfo, error) {
	convertedTx := ConvertUpdatePublicPoolTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructCreateOrderTx(key signer.KeyManager, lighterChainId uint32, tx *CreateOrderTxReq, ops *TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	convertedTx := ConvertCreateOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2CreateGroupedOrdersTx(key signer.KeyManager, lighterChainId uint32, tx *CreateGroupedOrdersTxReq, ops *TransactOpts) (*txtypes.L2CreateGroupedOrdersTxInfo, error) {
	orderCount := len(tx.Orders)
	convertedOrders := make([]*txtypes.OrderInfo, orderCount)
	for i, order := range tx.Orders {
		convertedOrders[i] = ConvertCreateOrderTx(order, ops).OrderInfo
	}

	convertedTx := &txtypes.L2CreateGroupedOrdersTxInfo{
		AccountIndex:   *ops.FromAccountIndex,
		ApiKeyIndex:    *ops.ApiKeyIndex,
		GroupingType:   tx.GroupingType,
		Orders:         convertedOrders,
		ExpiredAt:      ops.ExpiredAt,
		Nonce:          *ops.Nonce,
		L2TxAttributes: ops.L2TxAttributes(),
	}

	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2CancelOrderTx(key signer.KeyManager, lighterChainId uint32, tx *CancelOrderTxReq, ops *TransactOpts) (*txtypes.L2CancelOrderTxInfo, error) {
	convertedTx := ConvertCancelOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2CancelAllOrdersTx(key signer.KeyManager, lighterChainId uint32, tx *CancelAllOrdersTxReq, ops *TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	convertedTx := ConvertCancelAllOrdersTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructL2ModifyOrderTx(key signer.KeyManager, lighterChainId uint32, tx *ModifyOrderTxReq, ops *TransactOpts) (*txtypes.L2ModifyOrderTxInfo, error) {
	convertedTx := ConvertModifyOrderTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructTransferTx(key signer.KeyManager, lighterChainId uint32, tx *TransferTxReq, ops *TransactOpts) (*txtypes.L2TransferTxInfo, error) {
	convertedTx := ConvertTransferTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructMintSharesTx(key signer.KeyManager, lighterChainId uint32, tx *MintSharesTxReq, ops *TransactOpts) (*txtypes.L2MintSharesTxInfo, error) {
	convertedTx := ConvertMintSharesTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructBurnSharesTx(key signer.KeyManager, lighterChainId uint32, tx *BurnSharesTxReq, ops *TransactOpts) (*txtypes.L2BurnSharesTxInfo, error) {
	convertedTx := ConvertBurnSharesTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructStakeAssetsTx(key signer.KeyManager, lighterChainId uint32, tx *StakeAssetsTxReq, ops *TransactOpts) (*txtypes.L2StakeAssetsTxInfo, error) {
	convertedTx := ConvertStakeAssetsTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUnstakeAssetsTx(key signer.KeyManager, lighterChainId uint32, tx *UnstakeAssetsTxReq, ops *TransactOpts) (*txtypes.L2UnstakeAssetsTxInfo, error) {
	convertedTx := ConvertUnstakeAssetsTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateAccountConfigTx(key signer.KeyManager, lighterChainId uint32, tx *UpdateAccountConfigTxReq, ops *TransactOpts) (*txtypes.L2UpdateAccountConfigTxInfo, error) {
	convertedTx := ConvertUpdateAccountConfigTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateAccountAssetConfigTx(key signer.KeyManager, lighterChainId uint32, tx *UpdateAccountAssetConfigTxReq, ops *TransactOpts) (*txtypes.L2UpdateAccountAssetConfigTxInfo, error) {
	convertedTx := ConvertUpdateAccountAssetConfigTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateLeverageTx(key signer.KeyManager, lighterChainId uint32, tx *UpdateLeverageTxReq, ops *TransactOpts) (*txtypes.L2UpdateLeverageTxInfo, error) {
	convertedTx := ConvertUpdateLeverageTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructUpdateMarginTx(key signer.KeyManager, lighterChainId uint32, tx *UpdateMarginTxReq, ops *TransactOpts) (*txtypes.L2UpdateMarginTxInfo, error) {
	convertedTx := ConvertUpdateMarginTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func ConstructWithdrawTx(key signer.KeyManager, lighterChainId uint32, tx *WithdrawTxReq, ops *TransactOpts) (*txtypes.L2WithdrawTxInfo, error) {
	convertedTx := ConvertWithdrawTx(tx, ops)
	err := convertedTx.Validate()
	if err != nil {
		return nil, err
	}

	msgHash, err := convertedTx.Hash(lighterChainId)
	if err != nil {
		return nil, err
	}

	signature, err := key.Sign(msgHash, p2.NewPoseidon2())
	if err != nil {
		return nil, err
	}

	convertedTx.SignedHash = ethCommon.Bytes2Hex(msgHash)
	convertedTx.Sig = signature
	return convertedTx, nil
}

func IsZeroByteSlice(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
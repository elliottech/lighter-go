/*
 * Copyright © 2023 ZkLighter Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package txtypes

import (
	"fmt"

	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

var _ (TxInfo) = (*L2ApproveIntegratorTxInfo)(nil)

type L2ApproveIntegratorTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	IntegratorAccountIndex int64
	MaxPerpsTakerFee       uint32
	MaxPerpsMakerFee       uint32
	MaxSpotTakerFee        uint32
	MaxSpotMakerFee        uint32
	ApprovalExpiry         int64

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	L1Sig      string
	SignedHash string `json:"-"`

	L2TxAttributes
}

func (txInfo *L2ApproveIntegratorTxInfo) Validate() error {
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

	// IntegratorAccountIndex
	if txInfo.IntegratorAccountIndex < MinAccountIndex {
		return ErrIntegratorAccountIndexTooLow
	}
	if txInfo.IntegratorAccountIndex > MaxAccountIndex {
		return ErrIntegratorAccountIndexTooHigh
	}

	// Fees
	if int64(txInfo.MaxPerpsTakerFee) > FeeTick || int64(txInfo.MaxPerpsMakerFee) > FeeTick ||
		int64(txInfo.MaxSpotTakerFee) > FeeTick || int64(txInfo.MaxSpotMakerFee) > FeeTick {
		return ErrFeeTooHigh
	}

	isRevokingApproval := txInfo.MaxPerpsTakerFee == 0 && txInfo.MaxPerpsMakerFee == 0 && txInfo.MaxSpotTakerFee == 0 && txInfo.MaxSpotMakerFee == 0
	if isRevokingApproval != (txInfo.ApprovalExpiry == 0) {
		return ErrApprovalExpiryZeroOnRevocation
	}

	// ApprovalExpiry
	if txInfo.ApprovalExpiry < 0 || txInfo.ApprovalExpiry > MaxTimestamp {
		return ErrApprovalExpiryInvalid
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	return nil
}

func (txInfo *L2ApproveIntegratorTxInfo) GetTxType() uint8 {
	return TxTypeL2ApproveIntegrator
}

func (txInfo *L2ApproveIntegratorTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2ApproveIntegratorTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2ApproveIntegratorTxInfo) GetL1SignatureBody(chainId uint32) string {
	signatureBody := fmt.Sprintf(
		TemplateL2ApproveIntegrator,
		getHex10FromUint64(uint64(txInfo.Nonce)),        //nolint:gosec
		getHex10FromUint64(uint64(txInfo.AccountIndex)), //nolint:gosec
		getHex10FromUint64(uint64(txInfo.ApiKeyIndex)),
		getHex10FromUint64(uint64(txInfo.IntegratorAccountIndex)), //nolint:gosec
		getHex10FromUint64(uint64(txInfo.MaxPerpsTakerFee)),       //nolint:gosec
		getHex10FromUint64(uint64(txInfo.MaxPerpsMakerFee)),       //nolint:gosec
		getHex10FromUint64(uint64(txInfo.MaxSpotTakerFee)),        //nolint:gosec
		getHex10FromUint64(uint64(txInfo.MaxSpotMakerFee)),        //nolint:gosec
		getHex10FromUint64(uint64(txInfo.ApprovalExpiry)),         //nolint:gosec
		getHex10FromUint64(uint64(chainId)),                       //nolint:gosec
	)
	return signatureBody
}

func (txInfo *L2ApproveIntegratorTxInfo) Hash(lighterChainId uint32) (msgHash []byte, err error) {
	elems := make([]g.GoldilocksField, 0, 12)

	elems = append(elems, g.GoldilocksField(lighterChainId))
	elems = append(elems, g.GoldilocksField(TxTypeL2ApproveIntegrator))
	elems = append(elems, g.GoldilocksField(txInfo.Nonce))
	elems = append(elems, g.GoldilocksField(txInfo.ExpiredAt))

	elems = append(elems, g.GoldilocksField(txInfo.AccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.ApiKeyIndex))
	elems = append(elems, g.GoldilocksField(txInfo.IntegratorAccountIndex))
	elems = append(elems, g.GoldilocksField(txInfo.MaxPerpsTakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MaxPerpsMakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MaxSpotTakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.MaxSpotMakerFee))
	elems = append(elems, g.GoldilocksField(txInfo.ApprovalExpiry))

	txHash := p2.HashToQuinticExtension(elems)
	return txInfo.L2TxAttributes.AggregateTxHash(txHash)
}

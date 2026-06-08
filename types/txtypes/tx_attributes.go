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
	"sort"

	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	"github.com/elliottech/poseidon_crypto/field/goldilocks_quintic_extension"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
)

const (
	_                                   = iota
	AttributeTypeIntegratorAccountIndex = 1
	AttributeTypeIntegratorTakerFee     = 2
	AttributeTypeIntegratorMakerFee     = 3
	AttributeTypeSkipTxNonce            = 4
	AttributeTypeCancelAllMarketIndex   = 5
	AttributeTypeSelfTradeBehaviorMode  = 6
	AttributeTypeSelfTradeEqualityMode  = 7

	MaxAttributeType = AttributeTypeSelfTradeEqualityMode
)

type AttibuteConfig struct {
	ByteSize          int
	MinValue          int
	MaxValue          int
	NilValue          int
	InvalidRangeError error
}

var AttributeTypeToConfig = map[uint8]*AttibuteConfig{
	AttributeTypeIntegratorAccountIndex: {
		ByteSize:          6,
		MinValue:          0,
		MaxValue:          int(MaxAccountIndex),
		NilValue:          NilIntegratorIndex,
		InvalidRangeError: ErrIntegratorAccountIndexInvalidRange,
	},
	AttributeTypeIntegratorTakerFee: {
		ByteSize:          4,
		MinValue:          0,
		MaxValue:          int(FeeTick),
		NilValue:          NilIntegratorTakerFee,
		InvalidRangeError: ErrIntegratorFeeInvalidRange,
	},
	AttributeTypeIntegratorMakerFee: {
		ByteSize:          4,
		MinValue:          0,
		MaxValue:          int(FeeTick),
		NilValue:          NilIntegratorMakerFee,
		InvalidRangeError: ErrIntegratorFeeInvalidRange,
	},
	AttributeTypeSkipTxNonce: {
		ByteSize:          1,
		MinValue:          1,
		MaxValue:          1,
		NilValue:          0,
		InvalidRangeError: ErrNonceSkipAttributeInvalid,
	},
	AttributeTypeCancelAllMarketIndex: {
		ByteSize:          2,
		MinValue:          int(MinPerpsMarketIndex),
		MaxValue:          int(NilMarketIndex),
		NilValue:          int(NilMarketIndex),
		InvalidRangeError: ErrCancelAllMarketIndexInvalidRange,
	},
	AttributeTypeSelfTradeBehaviorMode: {
		ByteSize:          1,
		MinValue:          SelfTradeBehaviorExpireMaker,
		MaxValue:          SelfTradeBehaviorReduce,
		NilValue:          SelfTradeBehaviorExpireMaker,
		InvalidRangeError: ErrSelfTradeBehaviorModeInvalidRange,
	},
	AttributeTypeSelfTradeEqualityMode: {
		ByteSize:          1,
		MinValue:          SelfTradeEqualityAccountIndex,
		MaxValue:          SelfTradeEqualityMasterAccountIndex,
		NilValue:          SelfTradeEqualityAccountIndex,
		InvalidRangeError: ErrSelfTradeEqualityModeInvalidRange,
	},
}

type L2TxAttributes map[uint8]int // Type to value

type L2TxAttributeIsNilMap map[uint8]bool

// Caller must make sure attrType is valid
func (attr L2TxAttributes) GetValueOrDefault(attrType uint8) int {
	value, ok := attr[attrType]
	if !ok {
		return AttributeTypeToConfig[attrType].NilValue
	}
	return value
}

func (attr L2TxAttributes) getIsNilInfos() L2TxAttributeIsNilMap {
	helpers := make(L2TxAttributeIsNilMap)
	helpers[0] = true
	for typ := uint8(1); typ <= MaxAttributeType; typ++ {
		helpers[typ] = attr.GetValueOrDefault(typ) == AttributeTypeToConfig[typ].NilValue
	}
	return helpers
}

func (attr L2TxAttributes) Validate() error {
	if len(attr) == 0 {
		return nil
	}

	if len(attr) > NbAttributesPerTx {
		return ErrTooManyAttributes
	}

	for typ, val := range attr {
		config, ok := AttributeTypeToConfig[typ]
		if !ok {
			return fmt.Errorf("%w: %d", ErrInvalidAttributeType, typ)
		}
		minValue, maxValue := config.MinValue, config.MaxValue
		if val < minValue || val > maxValue { // ErrAttributeValueOutOfRange
			return config.InvalidRangeError
		}
	}
	isNil := attr.getIsNilInfos()

	// Fees and integrator index must be specified together
	hasFees := !isNil[AttributeTypeIntegratorTakerFee] || !isNil[AttributeTypeIntegratorMakerFee]
	hasIntegratorIndex := !isNil[AttributeTypeIntegratorAccountIndex]
	if hasFees && !hasIntegratorIndex {
		return ErrIntegratorAccountIndexRequiredForNonZeroFees
	}

	// Disallow self-trade specification if any fee is specified
	hasSelfTradeSpec := !isNil[AttributeTypeSelfTradeBehaviorMode] || !isNil[AttributeTypeSelfTradeEqualityMode]
	if hasSelfTradeSpec && hasFees {
		return ErrSelfTradeSpecificationNotAllowedWithNonZeroFees
	}

	// Disallow reduce mode with master account index equality mode
	isMaiEqualityMode := attr[AttributeTypeSelfTradeEqualityMode] == SelfTradeEqualityMasterAccountIndex
	isReduceBehaviorMode := attr[AttributeTypeSelfTradeBehaviorMode] == SelfTradeBehaviorReduce
	if isReduceBehaviorMode && isMaiEqualityMode {
		return ErrReduceModeNotAllowedWithMasterAccountIndexEqualityMode
	}

	return nil
}

func (attr L2TxAttributes) IsEmpty() bool {
	for key, value := range attr {
		if value != AttributeTypeToConfig[key].NilValue {
			return false
		}
	}
	return true
}

// Nonzero types in ascending order, padded with zeroes to NbAttributesPerTx
func (attr L2TxAttributes) getNormalizedTypes() (attrTypes [NbAttributesPerTx]uint8) {
	i := 0
	for key, value := range attr {
		if value == AttributeTypeToConfig[key].NilValue {
			continue
		}
		attrTypes[i] = key
		i++
	}
	sort.Slice(attrTypes[:i], func(i, j int) bool {
		return attrTypes[i] < attrTypes[j]
	})
	return attrTypes
}

func (attr L2TxAttributes) Hash() (msgHash goldilocks_quintic_extension.Element, err error) {
	elems := make([]g.GoldilocksField, 0, NbAttributesPerTx*2)
	for _, attrType := range attr.getNormalizedTypes() {
		attrValue := 0
		if attrType != 0 {
			attrValue = attr[attrType]
		}
		elems = append(elems, g.GoldilocksField(attrType))
		elems = append(elems, g.GoldilocksField(attrValue))
	}
	return p2.HashToQuinticExtension(elems), nil
}

func (attr L2TxAttributes) AggregateTxHash(txHash goldilocks_quintic_extension.Element) ([]byte, error) {
	if attr.IsEmpty() {
		return txHash.ToLittleEndianBytes(), nil
	}

	attributesHash, err := attr.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to compute attributes hash. error: %w", err)
	}
	combinedElements := make([]g.GoldilocksField, 0, 10)
	combinedElements = append(combinedElements, txHash[:]...)
	combinedElements = append(combinedElements, attributesHash[:]...)

	return p2.HashToQuinticExtension(combinedElements).ToLittleEndianBytes(), nil
}

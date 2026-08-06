package signer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math/big"

	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	gFp5 "github.com/elliottech/poseidon_crypto/field/goldilocks_quintic_extension"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
)

type Signer interface {
	Sign(message []byte, hFunc hash.Hash) ([]byte, error)
}

type KeyManager interface {
	Signer
	PubKey() gFp5.Element
	PubKeyBytes() [40]byte
	PrvKeyBytes() []byte
}

type keyManager struct {
	key curve.ECgFp5Scalar
}

// hex string seed should be at least 32 bytes when decoded
func NewSeedKeyManager(seed string) (KeyManager, error) {
	return &keyManager{key: GetScalarFromSeed(seed)}, nil
}


func NewKeyManager(b []byte) (KeyManager, error) {
	if len(b) != 40 {
		return nil, fmt.Errorf("invalid private key length. expected: 40 got: %v", len(b))
	}
	return &keyManager{key: curve.ScalarElementFromLittleEndianBytes(b)}, nil
}

func (key *keyManager) Sign(hashedMessage []byte, hFunc hash.Hash) ([]byte, error) {
	hashedMessageAsQuinticExtension, err := gFp5.FromCanonicalLittleEndianBytes(hashedMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message while signing. message: %v err: %w", hashedMessage, err)
	}
	return schnorr.SchnorrSignHashedMessage(hashedMessageAsQuinticExtension, key.key).ToBytes(), nil
}

func (key *keyManager) PubKey() gFp5.Element {
	return schnorr.SchnorrPkFromSk(key.key)
}

func (key *keyManager) PubKeyBytes() (res [40]byte) {
	bytes := key.PubKey().ToLittleEndianBytes()
	copy(res[:], bytes[:])
	return
}

func (key *keyManager) PrvKeyBytes() []byte {
	return key.key.ToLittleEndianBytes()
}


func GetScalarFromSeed(seed string) curve.ECgFp5Scalar {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		panic(fmt.Sprintf("failed to decode seed hex string: %v", err))
	}

	if len(seedBytes) < 32 {
		panic("seed too short, should be at least 32 bytes")
	}

	hasher := sha256.New()
	hasher.Write([]byte{1})
	hasher.Write(seedBytes)

	part1 := hasher.Sum(nil)

	hasher.Reset()
	hasher.Write([]byte{2})
	hasher.Write(seedBytes)
	part2 := hasher.Sum(nil)

	combined := make([]byte, 40)
	copy(combined[0:32], part1)
	copy(combined[32:40], part2)

	return curve.FromNonCanonicalBigInt(new(big.Int).SetBytes(combined))
}
package signer

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

// A secp256k1 api key. Txs are signed over the personal_sign digest of the 40-byte tx hash,
// keccak256("\x19Ethereum Signed Message:\n40" || txHash), as the circuit verifies them.
type ECDSAKeyManager struct {
	key *ecdsa.PrivateKey
}

var _ KeyManager = (*ECDSAKeyManager)(nil)

func NewECDSAKeyManager(privateKeyHex string) (*ECDSAKeyManager, error) {
	key, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid ecdsa private key: %w", err)
	}
	return &ECDSAKeyManager{key: key}, nil
}

// Derives the secp256k1 key from the same hex seed the Schnorr key manager uses (sha256 of the seed).
func NewECDSAKeyManagerFromSeed(seed string) (*ECDSAKeyManager, error) {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode seed hex string: %w", err)
	}
	if len(seedBytes) < 32 {
		return nil, fmt.Errorf("seed too short, should be at least 32 bytes")
	}
	digest := sha256.Sum256(append([]byte{3}, seedBytes...))
	key, err := crypto.ToECDSA(digest[:])
	if err != nil {
		return nil, fmt.Errorf("failed to derive ecdsa key from seed: %w", err)
	}
	return &ECDSAKeyManager{key: key}, nil
}

// The 32-byte digest signed for an L2 tx hash.
func ECDSASigningHash(txHash []byte) []byte {
	return accounts.TextHash(txHash)
}

// Returns the 64-byte `r || s` signature (no recovery id).
func (key *ECDSAKeyManager) Sign(hashedMessage []byte, _ hash.Hash) ([]byte, error) {
	sig, err := crypto.Sign(ECDSASigningHash(hashedMessage), key.key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with ecdsa key: %w", err)
	}
	padded := make([]byte, TxSignatureLength)
	copy(padded, sig[:ECDSASignatureLength])
	return padded, nil
}

func (key *ECDSAKeyManager) KeyType() uint8 {
	return KeyTypeECDSA
}

// Padded 40-byte representation of the compressed public key.
func (key *ECDSAKeyManager) PubKeyBytes() (res [40]byte) {
	copy(res[:], crypto.CompressPubkey(&key.key.PublicKey))
	return res
}

// Compressed SEC1 encoding (02|03 || x), as stored in the api key.
func (key *ECDSAKeyManager) ECDSAPubKeyBytes() (res [33]byte) {
	copy(res[:], crypto.CompressPubkey(&key.key.PublicKey))
	return res
}

func (key *ECDSAKeyManager) PrvKeyBytes() []byte {
	return crypto.FromECDSA(key.key)
}

package signer

import (
	"crypto/ecdsa"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

// A secp256k1 api key. Txs are signed over the personal_sign digest of the 40-byte tx hash,
// keccak256("\x19Ethereum Signed Message:\n40" || txHash).
type ECDSAKeyManager struct {
	key *ecdsa.PrivateKey
}

func NewECDSAKeyManager(b []byte) (KeyManager, error) {
	key, err := crypto.ToECDSA(b)
	if err != nil {
		return nil, fmt.Errorf("invalid ecdsa private key: %w", err)
	}
	return &ECDSAKeyManager{key: key}, nil
}

func ECDSASigningHash(txHash []byte) []byte {
	return accounts.TextHash(txHash)
}

// Returns the 64-byte `r || s` signature padded to the 80-byte tx signature length.
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

// Compressed public key (02|03 || x) padded to 40 bytes.
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

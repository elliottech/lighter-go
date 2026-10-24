package signer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"runtime"
	"sync"

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
	key       curve.ECgFp5Scalar
	publicKey gFp5.Element
}

const maxPreparedNoncePoolSize = 10_000

var (
	preparedNoncePool = make(chan *schnorr.PreparedNonce, maxPreparedNoncePoolSize)
	prepareNonceMu    sync.Mutex
)

// PrepareSchnorrNonces moves the expensive fixed-generator multiplication out
// of the latency-sensitive signing path. The pool is process-global: any key
// manager may consume its entries. Each nonce is process-bound, held in memory
// until it is used, and consumed exactly once. Prepared nonces must never be
// serialized, persisted, or shared across a fork.
func PrepareSchnorrNonces(count int) error {
	if count < 0 {
		return fmt.Errorf("prepared nonce count must be non-negative")
	}
	if count == 0 {
		return nil
	}

	prepareNonceMu.Lock()
	defer prepareNonceMu.Unlock()
	if available := cap(preparedNoncePool) - len(preparedNoncePool); count > available {
		return fmt.Errorf("prepared nonce pool has room for %d nonces, got %d", available, count)
	}

	nonces := make([]*schnorr.PreparedNonce, count)
	workerCount := min(count, runtime.GOMAXPROCS(0))
	var workers sync.WaitGroup
	workers.Add(workerCount)
	var prepareErr error
	var prepareErrOnce sync.Once
	for worker := 0; worker < workerCount; worker++ {
		start := worker * count / workerCount
		end := (worker + 1) * count / workerCount
		go func() {
			defer workers.Done()
			defer func() {
				if recovered := recover(); recovered != nil {
					prepareErrOnce.Do(func() {
						prepareErr = fmt.Errorf("failed to prepare Schnorr nonce: %v", recovered)
					})
				}
			}()
			for i := start; i < end; i++ {
				nonces[i] = schnorr.PrepareNonce()
			}
		}()
	}
	workers.Wait()
	if prepareErr != nil {
		return prepareErr
	}
	for _, nonce := range nonces {
		preparedNoncePool <- nonce
	}
	return nil
}

func takePreparedSchnorrNonce() *schnorr.PreparedNonce {
	select {
	case nonce := <-preparedNoncePool:
		return nonce
	default:
		return nil
	}
}

// hex string seed should be at least 32 bytes when decoded
func NewSeedKeyManager(seed string) (KeyManager, error) {
	return newKeyManager(GetScalarFromSeed(seed)), nil
}

func NewKeyManager(b []byte) (KeyManager, error) {
	if len(b) != 40 {
		return nil, fmt.Errorf("invalid private key length. expected: 40 got: %v", len(b))
	}
	return newKeyManager(curve.ScalarElementFromLittleEndianBytes(b)), nil
}

func newKeyManager(key curve.ECgFp5Scalar) *keyManager {
	curve.WarmGeneratorTable()
	return &keyManager{
		key:       key,
		publicKey: schnorr.SchnorrPkFromSk(key),
	}
}

func (key *keyManager) Sign(hashedMessage []byte, _ hash.Hash) ([]byte, error) {
	return key.SignHash(hashedMessage)
}

// SignHash signs a canonical Fp5 message hash directly. Keeping this separate
// from the legacy Sign method lets transaction construction avoid allocating a
// hash.Hash value that the built-in signer does not use.
func (key *keyManager) SignHash(hashedMessage []byte) ([]byte, error) {
	hashedMessageAsQuinticExtension, err := gFp5.FromCanonicalLittleEndianBytes(hashedMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse message while signing. message: %v err: %w", hashedMessage, err)
	}
	if nonce := takePreparedSchnorrNonce(); nonce != nil {
		signature, err := schnorr.SchnorrSignHashedMessagePrepared(
			hashedMessageAsQuinticExtension, key.key, nonce,
		)
		if err == nil {
			return signature.ToBytes(), nil
		}
		if !errors.Is(err, schnorr.ErrPreparedNonceForked) {
			return nil, fmt.Errorf("failed to sign with prepared nonce: %w", err)
		}
	}
	return schnorr.SchnorrSignHashedMessage(hashedMessageAsQuinticExtension, key.key).ToBytes(), nil
}

func (key *keyManager) PubKey() gFp5.Element {
	return key.publicKey
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

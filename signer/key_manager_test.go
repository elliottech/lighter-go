package signer

import (
	"strings"
	"sync"
	"testing"

	curve "github.com/elliottech/poseidon_crypto/curve/ecgfp5"
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks_plonky2"
	schnorr "github.com/elliottech/poseidon_crypto/signature/schnorr"
)

func TestNewSeedKeyManagerCachesMatchingPublicKey(t *testing.T) {
	seed := strings.Repeat("01", 32)
	key, err := NewSeedKeyManager(seed)
	if err != nil {
		t.Fatalf("new seed key manager: %v", err)
	}
	want := schnorr.SchnorrPkFromSk(GetScalarFromSeed(seed))
	if got := key.PubKey(); got != want {
		t.Fatalf("seed public key = %v, want %v", got, want)
	}
}

func drainPreparedNoncePool() {
	for {
		select {
		case <-preparedNoncePool:
		default:
			return
		}
	}
}

func TestPreparedNoncePoolConcurrentSignatures(t *testing.T) {
	drainPreparedNoncePool()
	t.Cleanup(drainPreparedNoncePool)

	const signatureCount = 256
	if err := PrepareSchnorrNonces(signatureCount); err != nil {
		t.Fatalf("prepare nonces: %v", err)
	}
	key, err := NewKeyManager(curve.SampleScalar().ToLittleEndianBytes())
	if err != nil {
		t.Fatalf("new key manager: %v", err)
	}
	publicKey := key.PubKeyBytes()

	var workers sync.WaitGroup
	for i := range signatureCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			message := p2.HashToQuinticExtension([]g.GoldilocksField{
				g.GoldilocksField(i + 1),
			}).ToLittleEndianBytes()
			signature, signErr := key.Sign(message, nil)
			if signErr != nil {
				t.Errorf("sign message %d: %v", i, signErr)
				return
			}
			if validateErr := schnorr.Validate(publicKey[:], message, signature); validateErr != nil {
				t.Errorf("validate message %d: %v", i, validateErr)
			}
		}()
	}
	workers.Wait()

	if remaining := len(preparedNoncePool); remaining != 0 {
		t.Fatalf("prepared nonce pool contains %d unused entries", remaining)
	}
}

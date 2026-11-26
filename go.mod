module github.com/elliottech/lighter-go

go 1.26.0

require (
	github.com/elliottech/poseidon_crypto v0.0.15
	github.com/ethereum/go-ethereum v1.17.0
)

require (
	github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime v0.0.0-20251001021608-1fe7b43fc4d6 // indirect
	github.com/bits-and-blooms/bitset v1.20.0 // indirect
	github.com/consensys/gnark-crypto v0.18.1 // indirect
	github.com/crate-crypto/go-eth-kzg v1.4.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/ethereum/c-kzg-4844/v2 v2.1.5 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/supranational/blst v0.3.16-0.20250831170142-f48500c1fdbe // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

retract v1.0.1 // GenerateAPIKey with a non-empty seed produces weak API keys; use v1.0.2+

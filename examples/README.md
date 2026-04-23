# C++ 

Compile the example, using the following command (select the correct shared library)
```
clang++ -std=c++20 -O3 ./examples/cpp/example.cpp ./build/lighter-signer-darwin-arm64.dylib -o ./build/example-cpp
```

Run the example from the `./build` folder as `./example-cpp`

# Java

JNA bindings for the lighter-go shared library, with a benchmark.

### Prerequisites

- Java 21+
- Maven
- Go (to build the shared library)

Install on macOS:
```
brew install --cask temurin
brew install maven
```

### Build

**1. Build the shared library** from the repo root:

```
go build -buildmode=c-shared -o sharedlib/lighter.dylib ./sharedlib   # macOS
go build -buildmode=c-shared -o sharedlib/lighter.so   ./sharedlib   # Linux
```

**2. Compile** from the `examples/java/` directory:

```
cd examples/java
mvn compile
```

### Run

```
mvn exec:java
```

### What the benchmark does

Spawns 5 threads, each of which:
1. Generates a fresh API key pair
2. Creates a client on chain 304
3. Obtains an auth token (7-hour expiry)
4. Signs 100 create-order + cancel-order pairs back to back
5. Prints elapsed time for the signing loop


# Rust

FFI bindings for the lighter-go shared library, with a benchmark.

### Prerequisites

- Rust (stable, 1.70+)
- Go (to build the shared library)

Install Rust via [rustup](https://rustup.rs/):
```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

### Build

**1. Build the shared library** from the repo root:

```
go build -buildmode=c-shared -o sharedlib/lighter.dylib ./sharedlib   # macOS
go build -buildmode=c-shared -o sharedlib/lighter.so   ./sharedlib   # Linux
```

**2. Compile** from the `examples/rust/` directory:

```
cd examples/rust
cargo build --release
```

### Run

```
cargo run --release
```

### What the benchmark does

Spawns 5 threads, each of which:
1. Generates a fresh API key pair
2. Creates a client on chain 304
3. Obtains an auth token (7-hour expiry)
4. Signs 100 create-order + cancel-order pairs back to back
5. Prints elapsed time for the signing loop

### Sample output

```
[3] publicKey=0xdb20326defefbe7671156f61d927c2c6...
[1] publicKey=0x152911bbbcb8353c98502905fa8e6339...
[4] publicKey=0x51c684fa2248dfc8e58c74ec9d3414a5...
[2] publicKey=0xbf262215de6727d495b3f6f07b224229...
[0] publicKey=0xf0fea739719c12d1e62235b8a546c313...
[2] authToken=1776311597710:100:2:60d78eb8...
[0] authToken=1776311597710:100:0:da74b9d3...
[1] authToken=1776311597710:100:1:40463366...
[3] authToken=1776311597710:100:3:4051768060...
[4] authToken=1776311597710:100:4:4e37a696...
[1] 100 create+cancel pairs in 47.19 ms
[0] 100 create+cancel pairs in 47.22 ms
[4] 100 create+cancel pairs in 47.20 ms
[2] 100 create+cancel pairs in 47.43 ms
[3] 100 create+cancel pairs in 47.55 ms
```

# WASM

Node smoke test that loads the lighter-go WASM build and exercises the signer globals.

### Prerequisites

- Node.js 22+
- Go 1.23+ (to build the WASM artifact)
- [just](https://github.com/casey/just) (optional, for the build recipe)

Install on macOS:
```
brew install node go just
```

### Build

**1. Build the WASM artifact** from the repo root:

```
just build-wasm
# or, equivalently:
GOOS=js GOARCH=wasm go build -trimpath -o ./build/lighter-signer.wasm ./wasm/
```

**2. Install `wasm_exec.js`** into `./build` (the script loads it from there):

```
SRC="$(go env GOROOT)/lib/wasm/wasm_exec.js"
[ -f "$SRC" ] || SRC="$(go env GOROOT)/misc/wasm/wasm_exec.js"
cp "$SRC" ./build/wasm_exec.js
```

### Run

From the repo root:

```
node ./examples/wasm/test_wasm.mjs
```

### What the script does

1. Instantiates `build/lighter-signer.wasm` via `wasm_exec.js` and starts the Go runtime
2. Calls `GenerateAPIKey()` and asserts a valid hex keypair is returned
3. Calls `CreateClient(...)` on chain 304 with the generated private key
4. Signs a cancel-order, cancel-all-orders, create-order, create-sub-account and update-leverage transaction
5. For each signed tx, asserts the `txType`, `txHash` and decoded `txInfo` fields match the inputs
6. Verifies that toggling the `skipNonce` flag changes the resulting tx hash and populates the `L2TxAttributes` accordingly
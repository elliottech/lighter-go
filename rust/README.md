# Rust

FFI bindings for the lighter-go shared library, with a benchmark.

## Prerequisites

- Rust (stable, 1.70+)
- Go (to build the shared library)

Install Rust via [rustup](https://rustup.rs/):
```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

## Build

**1. Build the shared library** from the repo root:

```
go build -buildmode=c-shared -o sharedlib/lighter.dylib ./sharedlib   # macOS
go build -buildmode=c-shared -o sharedlib/lighter.so   ./sharedlib   # Linux
```

**2. Compile** from the `rust/` directory:

```
cd rust
cargo build --release
```

## Run

```
cargo run --release
```

## What the benchmark does

Spawns 5 threads, each of which:
1. Generates a fresh API key pair
2. Creates a client on chain 304
3. Obtains an auth token (7-hour expiry)
4. Signs 100 create-order + cancel-order pairs back to back
5. Prints elapsed time for the signing loop

## Sample output

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

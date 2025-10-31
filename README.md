# lighter-go

This repository serves as the reference implementation of signing & hashing of Lighter transactions. 

## Architecture

The repository is organized with a clear separation between business logic and platform-specific wrappers:

```
executables/
├── shared.go              # Pure Go business logic (shared by all platforms)
├── sharedlib/
│   └── main.go            # CGO/FFI wrappers for C libraries (.so, .dylib, .dll)
└── wasm/
    └── main.go            # WASM/JS wrappers for WebAssembly
```

This structure eliminates code duplication and makes maintenance easier.

## Build Targets

### C Shared Libraries
Compiled for a variety of platforms:
- macOS (darwin) dynamic library (.dylib) for arm architecture (M processor, not Intel)
- linux shared object (.so) for both amd64 and arm architectures
- windows .dll (dynamic-link library) for amd64 architecture

### WebAssembly
WASM build for frontend/browser usage:
- `.wasm` file for JavaScript interop (works on all platforms - one file for everything)

All generated shared libraries follow the naming convention `lighter_signer_{os}_{arch}` where os is linux/windows/darwin and arch is amd64(x86) or arm64.

## Features

The go SDK implements just the core signing, as well as a small HTTP client so that users can:
- not specify the nonce of the transaction (this will result in an HTTP call, so beware)
- check that a client was initialized correctly, by verifying that the given API key matches the one on the server

The [Python SDK](https://github.com/elliottech/lighter-python) offers support for HTTP and WebSocket functionality as well as [examples](https://github.com/elliottech/lighter-python/tree/main/examples) on how to generate the API keys, how to create and cancel orders, generate AUTH tokens for various HTTP/WS endpoints which require them. 

## Building

Builds can be found in the release notes. To compile your own binaries, use the commands in the `justfile`:

- `just build-darwin-local` - Build macOS shared library (.dylib)
- `just build-linux-local` - Build Linux shared library (.so)
- `just build-windows-local` - Build Windows shared library (.dll)
- `just build-wasm` - Build WebAssembly (.wasm - single file for all platforms)
- `just build-*-docker` - Docker-based cross-platform builds
- `just build-all` - Build all sharedlib and WASM
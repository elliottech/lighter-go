# Building Windows Signer DLL

## Quick Start

### Using `just` (Recommended)
```bash
just build-windows-docker  # Docker build (works anywhere)
just build-windows-local   # Local build (requires MinGW)
```

### Using PowerShell Script
```powershell
.\build-windows.ps1 -Method docker  # Docker build
.\build-windows.ps1 -Method local   # Local build
```

## Manual Commands

### PowerShell
```powershell
# Docker build
docker run --rm -v "${PWD}:/go/src/sdk" -w /go/src/sdk golang:1.23.2-bullseye /bin/sh -c "apt-get update -qq && apt-get install -y -qq mingw-w64 && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/sharedlib.go"

# Local build (requires MinGW/GCC)
$env:CGO_ENABLED='1'; go build -buildmode=c-shared -trimpath -o .\build\signer-amd64.dll .\sharedlib\sharedlib.go
```

### Bash/Git Bash
```bash
# Docker build
docker run --rm -v $(pwd):/go/src/sdk -w /go/src/sdk golang:1.23.2-bullseye /bin/sh -c "apt-get update -qq && apt-get install -y -qq mingw-w64 && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/sharedlib.go"

# Local build
CGO_ENABLED=1 go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/sharedlib.go
```

## Key Differences

| Shell | Current Directory | Environment Variable |
|-------|-------------------|---------------------|
| **PowerShell** | `${PWD}` | `$env:VAR='value'; command` |
| **Bash/sh** | `$(pwd)` | `VAR=value command` |

## Installation

Copy built DLL to Python package:
```powershell
copy .\build\signer-amd64.dll ..\signers\signer-amd64.dll
```


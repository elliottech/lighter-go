build-darwin-local:
    go mod vendor
    go build -buildmode=c-shared -trimpath -o ./build/signer-arm64.dylib ./sharedlib

build-linux-local:
    go mod vendor
    go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.so ./sharedlib

build-linux-docker:
    go mod vendor
    docker run --platform linux/amd64 -v $(pwd):/go/src/sdk golang:1.23.2-bullseye /bin/sh -c "cd /go/src/sdk && go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.so ./sharedlib"

build-js-wasm:
    go mod vendor
    GOOS=js GOARCH=wasm go build -trimpath -o ./build/signer.wasm ./sharedlib
    cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./build/wasm_exec.js

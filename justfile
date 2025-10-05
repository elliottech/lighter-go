build-darwin-local:
    go mod vendor
    go build -buildmode=c-shared -trimpath -o ./build/signer-arm64.dylib ./sharedlib/sharedlib.go

build-linux-local:
    go mod vendor
    go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.so ./sharedlib/sharedlib.go

build-linux-docker:
    go mod vendor
    docker run --platform linux/amd64 -v $(pwd):/go/src/sdk golang:1.23.2-bullseye /bin/sh -c "cd /go/src/sdk && go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.so ./sharedlib/sharedlib.go"

build-windows-local:
    go mod vendor
    set GOOS=windows && set GOARCH=amd64 && go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/sharedlib.go

build-windows-docker:
    go mod vendor
    docker run --platform windows/amd64 -v $(pwd):/go/src/sdk golang:1.23.2-bullseye /bin/sh -c "cd /go/src/sdk && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -trimpath -o ./build/signer-amd64.dll ./sharedlib/sharedlib.go"


#!/bin/bash

echo "Building WASM..."
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm

# Check/Compress Brotli
if command -v brotli &> /dev/null; then
    echo "Compressing with Brotli..."
    brotli -Z -f -k main.wasm
else
    echo "Brotli not installed. Skipping."
fi

# Check/Compress Gzip
if command -v gzip &> /dev/null; then
    echo "Compressing with Gzip..."
    gzip -9 -f -k main.wasm
else
    echo "Gzip not installed. Skipping."
fi

echo "Done."

#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_WASM_DIR="$SCRIPT_DIR/../web-wasm"
OUTPUT_FILE="$SCRIPT_DIR/dist/wasm-wrapper.standalone.html"

echo "Building main.wasm and sdk.html..."
"$WEB_WASM_DIR/build.sh" --no-copy > /dev/null

mkdir -p "$(dirname "$OUTPUT_FILE")"
cp "$WEB_WASM_DIR/sdk.html" "$OUTPUT_FILE"
cp "$WEB_WASM_DIR/main.manifest.json" "$SCRIPT_DIR/dist/wasm-wrapper.manifest.json"

echo "Done: $OUTPUT_FILE"

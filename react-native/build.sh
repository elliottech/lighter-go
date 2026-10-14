#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Builds web-wasm/main.wasm (+ .gz/.br/manifest) and writes dist/wasm-wrapper.standalone.html.
python3 "$SCRIPT_DIR/../web-wasm/build.py" "$@"

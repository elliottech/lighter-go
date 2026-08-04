#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_WASM_DIR="$SCRIPT_DIR/../web-wasm"
WASM_FILE="$WEB_WASM_DIR/main.wasm"
TEMPLATE_FILE="$SCRIPT_DIR/src/wasm-template.html"
OUTPUT_FILE="$SCRIPT_DIR/dist/wasm-wrapper.standalone.html"

echo "Building main.wasm..."
(cd "$WEB_WASM_DIR" && ./build.sh)

mkdir -p "$(dirname "$OUTPUT_FILE")"

B64_FILE="$(mktemp)"
trap 'rm -f "$B64_FILE"' EXIT

echo "Encoding main.wasm to base64..."
base64 -i "$WASM_FILE" | tr -d '\n' > "$B64_FILE"

echo "Generating $OUTPUT_FILE..."
awk -v b64file="$B64_FILE" '
  index($0, "{{WASM_BASE64}}") {
    split($0, parts, "{{WASM_BASE64}}")
    printf "%s", parts[1]
    while ((getline line < b64file) > 0) printf "%s", line
    printf "%s\n", parts[2]
    next
  }
  { print }
' "$TEMPLATE_FILE" > "$OUTPUT_FILE"

echo "Done."

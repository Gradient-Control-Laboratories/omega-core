#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

mkdir -p "${ROOT_DIR}/bin"

echo "Building gslc..."
go build -o "${ROOT_DIR}/bin/gslc" "${ROOT_DIR}/cmd/gslc"
echo "Build complete: ${ROOT_DIR}/bin/gslc"

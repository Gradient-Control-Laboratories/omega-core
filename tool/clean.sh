#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Cleaning build artifacts and test cache..."
rm -rf "${ROOT_DIR}/bin" "${ROOT_DIR}/scratch" "${ROOT_DIR}/tmp"
(cd "${ROOT_DIR}" && go clean -testcache)
echo "  ✓ Clean complete."

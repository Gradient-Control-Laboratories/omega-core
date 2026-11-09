#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Regenerating test goldens..."
UPDATE_GOLDENS=true (cd "${ROOT_DIR}" && go test ./...)
echo "  ✓ Goldens updated."

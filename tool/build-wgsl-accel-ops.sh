#!/usr/bin/env bash
set -euo pipefail

# Regenerate WGSL hardware acceleration ops table and inliner registry from stdlib sidecars.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "Generating WGSL ops table..."
(cd "${ROOT_DIR}" && go generate ./backend/wgsl/internal/accel/generate.go)


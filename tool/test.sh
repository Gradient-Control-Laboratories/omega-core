#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Running omega_core test suite..."
(cd "${ROOT_DIR}" && go test ./cmd/gslc/... ./parser/... ./assembler/... ./backend/wgsl ./backend/mesh ./backend/tape ./backend/audio ./stdlib/... ./program/... ./tool/analyze-deps/...)

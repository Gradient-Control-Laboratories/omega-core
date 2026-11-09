#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJ_DIR="$(cd "${ROOT_DIR}/.." && pwd)"
HISTORY_FILE="${PROJ_DIR}/.test-history/omega-core-test.jsonl"

echo "==> Running omega-core Go test suite..."
go run "${ROOT_DIR}/tool/run-go-tests" \
    --repo-root "${ROOT_DIR}" \
    --history-file "${HISTORY_FILE}" \
    --wgsl-testdata "backend/wgsl/internal/codegen/testdata" \
    --fidget-testdata "internal/fidget/internal/convert/testdata" \
    "${ROOT_DIR}/parser" \
    "${ROOT_DIR}/program" \
    "${ROOT_DIR}/stdlib" \
    "${ROOT_DIR}/internal" \
    "${ROOT_DIR}/assembler" \
    "${ROOT_DIR}/eval" \
    "${ROOT_DIR}/backend" \
    "${ROOT_DIR}/cmd" \
    "${ROOT_DIR}/tool"


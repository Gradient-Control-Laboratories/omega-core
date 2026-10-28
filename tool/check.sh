#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Running gofmt check..."
UNFORMATTED=$(gofmt -l "${ROOT_DIR}")
if [ -n "${UNFORMATTED}" ]; then
  echo "The following files are not properly formatted:"
  echo "${UNFORMATTED}"
  exit 1
fi
echo "  ✓ gofmt passed."

echo "==> Running go vet..."
(cd "${ROOT_DIR}" && go vet ./...)
echo "  ✓ go vet passed."

echo "==> Running tests..."
(cd "${ROOT_DIR}" && go test ./cmd/gslc/... ./parser/... ./assembler/... ./backend/wgsl ./backend/mesh ./backend/tape ./backend/audio ./stdlib/... ./program/... ./tool/analyze-deps/...)
echo "  ✓ Tests passed."

echo "==> All quality checks passed!"

#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

"${SCRIPT_DIR}/build-stdlib.sh"

echo "==> Running gofmt check..."
UNFORMATTED=$(gofmt -l "${ROOT_DIR}")
if [ -n "${UNFORMATTED}" ]; then
  echo "The following files are not properly formatted:"
  echo "${UNFORMATTED}"
  exit 1
fi
echo "  ✓ gofmt passed."

echo "==> Running architectural dependency check..."
(cd "${ROOT_DIR}" && go run ./tool/analyze-deps -rules)
echo "  ✓ Architectural dependency check passed."

echo "==> Running go vet..."
(cd "${ROOT_DIR}" && go vet ./...)
echo "  ✓ go vet passed."

echo "==> Running tests..."
(cd "${ROOT_DIR}" && go test ./...)
echo "  ✓ Tests passed."

echo "==> All quality checks passed!"

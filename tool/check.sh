#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

SHARD="all"
while [ $# -gt 0 ]; do
  case "$1" in
    --shard)
      if [ $# -lt 2 ]; then
        echo "Error: --shard requires a value (all, lint, eval-core, eval-heavy, other)" >&2
        exit 1
      fi
      SHARD="$2"
      shift 2
      ;;
    --shard=*)
      SHARD="${1#--shard=}"
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

run_lint() {
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
}

run_tests() {
  "${SCRIPT_DIR}/build-stdlib.sh"
  echo "==> Running tests (${SHARD})..."
  case "${SHARD}" in
    all)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./...)
      ;;
    eval-core)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./eval ./eval/joints ./eval/internal/libflattest ./eval/internal/environment)
      ;;
    eval-heavy)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./eval/internal/libboxtest ./eval/internal/boxsweep/...)
      ;;
    other)
      (cd "${ROOT_DIR}" && go test -timeout 20m $(go list ./... | grep -v '^gradientcontrol.com/omega/core/eval'))
      ;;
    *)
      echo "Unknown shard: ${SHARD}" >&2
      exit 1
      ;;
  esac
  echo "  ✓ Tests passed."
}

case "${SHARD}" in
  all)
    "${SCRIPT_DIR}/build-stdlib.sh"
    run_lint
    echo "==> Running tests..."
    (cd "${ROOT_DIR}" && go test -timeout 20m ./...)
    echo "  ✓ Tests passed."
    ;;
  lint)
    run_lint
    ;;
  eval-core|eval-heavy|other)
    run_tests
    ;;
  *)
    echo "Unknown shard: ${SHARD}" >&2
    exit 1
    ;;
esac

echo "==> All quality checks passed!"

#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

SHARD="all"
while [ $# -gt 0 ]; do
  case "$1" in
    --shard)
      if [ $# -lt 2 ]; then
        echo "Error: --shard requires a value (all, lint, eval-0, eval-1, eval-2, eval-sub, eval-heavy, other-gpu, other-swatch, other)" >&2
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

run_eval_shard() {
  local shard_idx="$1"
  local num_shards="$2"
  local pattern
  pattern=$(cd "${ROOT_DIR}" && go test -list '^Test' ./eval | grep '^Test' | sort | awk -v idx="${shard_idx}" -v n="${num_shards}" '(NR - 1) % n == idx' | paste -sd '|' -)
  if [ -z "${pattern}" ]; then
    echo "No tests matched eval shard ${shard_idx}/${num_shards}" >&2
    exit 1
  fi
  (cd "${ROOT_DIR}" && go test -timeout 20m -run "^(${pattern})$" ./eval)
}

run_tests() {
  "${SCRIPT_DIR}/build-stdlib.sh"
  echo "==> Running tests (${SHARD})..."
  case "${SHARD}" in
    all)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./...)
      ;;
    eval-0)
      run_eval_shard 0 3
      ;;
    eval-1)
      run_eval_shard 1 3
      ;;
    eval-2)
      run_eval_shard 2 3
      ;;
    eval-sub)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./eval/joints ./eval/internal/libflattest ./eval/internal/environment)
      ;;
    eval-heavy)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./eval/internal/libboxtest ./eval/internal/boxsweep/...)
      ;;
    other-gpu)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./internal/fidget/internal/gpuparity)
      ;;
    other-swatch)
      (cd "${ROOT_DIR}" && go test -timeout 20m ./tool/materialswatch)
      ;;
    other)
      (cd "${ROOT_DIR}" && go test -timeout 20m $(go list ./... | grep -v -E '^gradientcontrol\.com/omega/core/(eval|internal/fidget/internal/gpuparity|tool/materialswatch)'))
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
  eval-0|eval-1|eval-2|eval-sub|eval-heavy|other-gpu|other-swatch|other)
    run_tests
    ;;
  *)
    echo "Unknown shard: ${SHARD}" >&2
    exit 1
    ;;
esac

echo "==> All quality checks passed!"

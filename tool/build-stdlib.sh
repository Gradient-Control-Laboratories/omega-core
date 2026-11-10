#!/usr/bin/env bash
set -euo pipefail

#
# usage: build-stdlib.sh [--force]
#   Ensures every stdlib/<ImportPath>/package.pb file is up to date by
#   delegating all protobuf generation to the single-source-of-truth Go
#   generator (`go run ./stdlib/internal/gen`).
#
#   Why shell mtime guards precede `go run`:
#   Build and test scripts invoke this wrapper on every run. Checking file
#   mtimes in shell takes ~10 ms and avoids the ~150 ms `go run` startup cost
#   when nothing changed:
#     1. If --force is passed, or if global compiler inputs (non-test
#        parser/**/*.go, program/program.pb.go, stdlib/pkg_info.go, or
#        stdlib/internal/gen/main.go) are newer than the oldest package.pb, run
#        `go run ./stdlib/internal/gen --force` so all 50 .pb files rebuild.
#     2. Else if any stdlib package directory is missing package.pb or its
#        local *.gsl / package.yaml files are newer than its package.pb, run
#        incremental `go run ./stdlib/internal/gen` (re-parsing only modified
#        packages).
#     3. Otherwise exit immediately with zero overhead.
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
STDLIB_DIR="${ROOT_DIR}/stdlib"

FORCE=false
if [ "${1:-}" = "--force" ]; then
  FORCE=true
elif [ $# -gt 0 ]; then
  echo "usage: $0 [--force]" >&2
  exit 1
fi

if [ "${FORCE}" = "true" ]; then
  echo "==> Regenerating all stdlib package.pb files (--force)..."
  (cd "${ROOT_DIR}" && go run ./stdlib/internal/gen --force)
  exit 0
fi

NEEDS_INCREMENTAL=false
OLDEST_PB=""

# Check per-package .gsl and package.yaml files against that directory's own
# package.pb, and track the oldest package.pb across the tree for global checks.
while IFS= read -r gsl_file; do
  pkg_dir="${gsl_file%/*}"
  pb_file="${pkg_dir}/package.pb"
  if [ ! -f "${pb_file}" ]; then
    NEEDS_INCREMENTAL=true
    continue
  fi
  if [ "${gsl_file}" -nt "${pb_file}" ]; then
    NEEDS_INCREMENTAL=true
  fi
  if [ -f "${pkg_dir}/package.yaml" ] && [ "${pkg_dir}/package.yaml" -nt "${pb_file}" ]; then
    NEEDS_INCREMENTAL=true
  fi
  if [ -z "${OLDEST_PB}" ] || [ "${pb_file}" -ot "${OLDEST_PB}" ]; then
    OLDEST_PB="${pb_file}"
  fi
done < <(find "${STDLIB_DIR}" -name "*.gsl" -not -path "*/_*/*")

# If any global parser/schema/generator source is newer than OLDEST_PB, rebuild all.
if [ -n "${OLDEST_PB}" ]; then
  if [ "${ROOT_DIR}/program/program.pb.go" -nt "${OLDEST_PB}" ] ||
     [ "${STDLIB_DIR}/pkg_info.go" -nt "${OLDEST_PB}" ] ||
     [ "${STDLIB_DIR}/internal/gen/main.go" -nt "${OLDEST_PB}" ]; then
    echo "==> Compiler schema or stdlib registry changed; regenerating all stdlib package.pb files..."
    (cd "${ROOT_DIR}" && go run ./stdlib/internal/gen --force)
    exit 0
  fi

  NEWER_PARSER_GO="$(find "${ROOT_DIR}/parser" -name "*.go" -not -name "*_test.go" -newer "${OLDEST_PB}" -print -quit)"
  if [ -n "${NEWER_PARSER_GO}" ]; then
    echo "==> Parser implementation changed; regenerating all stdlib package.pb files..."
    (cd "${ROOT_DIR}" && go run ./stdlib/internal/gen --force)
    exit 0
  fi
fi

if [ "${NEEDS_INCREMENTAL}" = "true" ]; then
  echo "==> Updating modified stdlib package.pb files..."
  (cd "${ROOT_DIR}" && go run ./stdlib/internal/gen)
fi

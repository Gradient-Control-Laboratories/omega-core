#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

mkdir -p "${ROOT_DIR}/bin"
rm -f "${ROOT_DIR}/bin/gslc" "${ROOT_DIR}/bin/gslserve"

for tool in gsl compile serve fix; do
    echo "Building ${tool}..."
    go build -o "${ROOT_DIR}/bin/${tool}" "${ROOT_DIR}/cmd/${tool}"
    echo "Build complete: ${ROOT_DIR}/bin/${tool}"
done

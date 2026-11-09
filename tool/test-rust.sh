#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "==> Running omega-core Rust test suite..."

cd "${ROOT_DIR}"
cargo test --release \
    --target-dir "${ROOT_DIR}/target" \
    -p render \
    -p tape-core \
    -p manifold_glue \
    -p fidget-core

cd "${ROOT_DIR}/internal/fidget/internal/gpu-eval"
cargo test --release --target-dir "${ROOT_DIR}/target"

cd "${ROOT_DIR}/internal/fidget/internal/wgpu-render"
cargo test --release --target-dir "${ROOT_DIR}/target" -p wgpu-gbuf -p wgpu-render

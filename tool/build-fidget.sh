#!/usr/bin/env bash
set -euo pipefail

#
# usage: build-fidget.sh [--no-prompt]
#   build fidget and manifold Rust/C++ dependencies in omega-core
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

NO_PROMPT=false
for arg in "$@"; do
  if [ "$arg" = "--no-prompt" ]; then
    NO_PROMPT=true
  fi
done

confirmOrExit() {
  if [ "$NO_PROMPT" = "true" ]; then
    exit 1
  fi
  printf '%s' "$1"
  read -r response
  if [ "$response" = "n" ] || [ "$response" = "N" ]; then
    exit 1
  fi
}

if [ ! -f "${ROOT_DIR}/vendor/manifold/CMakeLists.txt" ]; then
  echo ""
  echo "vendor/manifold appears empty. Run:"
  echo "    (cd ${ROOT_DIR} && git submodule update --init --recursive)"
  confirmOrExit "Continue anyway? (Y/n): "
fi

# Release profile is mandatory for fidget: the meshSdf oracle runs ~13x slower on
# the hot BSP + Fast GWN loop in debug, and the cgo LDFLAGS point at
# target/release.
cd "${ROOT_DIR}/internal/fidget/internal/rust"
set +e
cargo build --release --target-dir "${ROOT_DIR}/target"
RESULT=$?
set -e
if [ $RESULT -ne 0 ]; then
  echo ""
  echo "Failed to build fidget"
  confirmOrExit "Continue anyway? (Y/n): "
fi
echo ""

# fidget-wasm: fidget for the browser (dual contouring, and the field oracle
# the viewer's debug API compares against), served from
# examples/viewer/vendor/fidget-wasm.
if rustup target list --installed 2>/dev/null | grep -q '^wasm32-unknown-unknown$'; then
  cd "${ROOT_DIR}/internal/fidget/internal/wasm"
  set +e
  cargo build --release --target wasm32-unknown-unknown --target-dir "${ROOT_DIR}/target"
  RESULT=$?
  set -e
  if [ $RESULT -ne 0 ]; then
    echo ""
    echo "Failed to build fidget-wasm"
    confirmOrExit "Continue anyway? (Y/n): "
  else
    mkdir -p "${ROOT_DIR}/examples/viewer/vendor/fidget-wasm"
    cp "${ROOT_DIR}/target/wasm32-unknown-unknown/release/fidget_wasm.wasm" \
      "${ROOT_DIR}/examples/viewer/vendor/fidget-wasm/fidget.wasm"
  fi
else
  echo "Skipping fidget-wasm: rustup target add wasm32-unknown-unknown to build it"
fi
echo ""

# fidget-wgpu-web: fidget's all-GPU pipeline for the browser, served from
# examples/viewer/vendor/fidget-wgpu.
WB_VERSION="$(grep -A1 '^name = "wasm-bindgen"$' "${ROOT_DIR}/internal/fidget/internal/wgpu-web/Cargo.lock" | sed -n 's/version = "\(.*\)"/\1/p')"
if ! rustup target list --installed 2>/dev/null | grep -q '^wasm32-unknown-unknown$'; then
  echo "Skipping fidget-wgpu-web: rustup target add wasm32-unknown-unknown to build it"
elif [ "$(wasm-bindgen --version 2>/dev/null || true)" != "wasm-bindgen ${WB_VERSION}" ]; then
  echo "Skipping fidget-wgpu-web: cargo install wasm-bindgen-cli --version ${WB_VERSION} to build it"
else
  cd "${ROOT_DIR}/internal/fidget/internal/wgpu-web"
  set +e
  cargo build --release --target wasm32-unknown-unknown --target-dir "${ROOT_DIR}/target" &&
    wasm-bindgen --target web --out-dir "${ROOT_DIR}/examples/viewer/vendor/fidget-wgpu" --out-name fidget_wgpu \
      "${ROOT_DIR}/target/wasm32-unknown-unknown/release/fidget_wgpu_web.wasm"
  RESULT=$?
  set -e
  if [ $RESULT -ne 0 ]; then
    echo ""
    echo "Failed to build fidget-wgpu-web"
    confirmOrExit "Continue anyway? (Y/n): "
  fi
fi
echo ""

# Release profile is mandatory for manifold: backend/mesh/manifold cgo LDFLAGS
# link against target/release/libmanifold_glue, and build.rs CMake-builds the
# vendor/manifold C++ submodule.
cd "${ROOT_DIR}/backend/mesh/manifold/internal/rust"
set +e
cargo build --release --target-dir "${ROOT_DIR}/target"
RESULT=$?
set -e
if [ $RESULT -ne 0 ]; then
  echo ""
  echo "Failed to build manifold"
  confirmOrExit "Continue anyway? (Y/n): "
fi
echo ""

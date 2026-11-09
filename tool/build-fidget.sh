#!/usr/bin/env bash
set -euo pipefail

#
# usage: build-fidget.sh [--no-prompt] [--wasm-out <dir>]
#   build fidget and manifold Rust/C++ dependencies in omega-core.
#   when --wasm-out <dir> is provided, also build fidget-wasm and fidget-wgpu-web
#   into <dir>/fidget-wasm and <dir>/fidget-wgpu.
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

NO_PROMPT=false
WASM_OUT=""
while [ $# -gt 0 ]; do
  case "$1" in
    --no-prompt)
      NO_PROMPT=true
      shift
      ;;
    --wasm-out)
      if [ $# -lt 2 ]; then
        echo "Error: --wasm-out requires a directory argument" >&2
        exit 1
      fi
      WASM_OUT="$2"
      shift 2
      ;;
    --wasm-out=*)
      WASM_OUT="${1#--wasm-out=}"
      shift
      ;;
    *)
      shift
      ;;
  esac
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

if [ ! -f "${ROOT_DIR}/backend/mesh/manifold/vendor/manifold/CMakeLists.txt" ]; then
  echo ""
  echo "backend/mesh/manifold/vendor/manifold appears empty. Run:"
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

# gpu-eval: the headless WGSL runner the GPU field-parity test drives
# (internal/fidget/internal/gpuparity), which skips when the binary is missing.
cd "${ROOT_DIR}/internal/fidget/internal/gpu-eval"
set +e
cargo build --release --target-dir "${ROOT_DIR}/target"
RESULT=$?
set -e
if [ $RESULT -ne 0 ]; then
  echo ""
  echo "Failed to build gpu-eval"
  confirmOrExit "Continue anyway? (Y/n): "
fi
echo ""

# fidget-wasm and fidget-wgpu-web: browser WebAssembly bundles, built only when
# --wasm-out <dir> is provided (used by omega web applications).
if [ -n "$WASM_OUT" ]; then
  WB_VERSION="$(grep -A1 '^name = "wasm-bindgen"$' "${ROOT_DIR}/internal/fidget/internal/wgpu-web/Cargo.lock" | sed -n 's/version = "\(.*\)"/\1/p')"
  if ! rustup target list --installed 2>/dev/null | grep -q '^wasm32-unknown-unknown$'; then
    echo "WARNING: Rust target 'wasm32-unknown-unknown' is not installed."
    echo "  Skipping fidget-wasm (required for 'Browser DC' mesh export in omega web apps)"
    echo "  and fidget-wgpu-web (optional 'Fidget' WebGPU render device)."
    echo ""
    echo "  To enable them, install the Rust WASM target and re-run:"
    echo "    rustup target add wasm32-unknown-unknown"
    confirmOrExit "Continue without building WASM targets? (Y/n): "
  else
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
      mkdir -p "${WASM_OUT}/fidget-wasm"
      cp "${ROOT_DIR}/target/wasm32-unknown-unknown/release/fidget_wasm.wasm" \
        "${WASM_OUT}/fidget-wasm/fidget.wasm"
    fi
    echo ""

    if [ "$(wasm-bindgen --version 2>/dev/null || true)" != "wasm-bindgen ${WB_VERSION}" ]; then
      echo "WARNING: 'wasm-bindgen-cli' ${WB_VERSION} is not installed."
      echo "  Skipping fidget-wgpu-web (optional 'Fidget' WebGPU render device in omega web apps)."
      echo ""
      echo "  To enable it, install wasm-bindgen-cli ${WB_VERSION} and re-run:"
      echo "    cargo install wasm-bindgen-cli --version ${WB_VERSION}"
      confirmOrExit "Continue without fidget-wgpu-web? (Y/n): "
    else
      cd "${ROOT_DIR}/internal/fidget/internal/wgpu-web"
      set +e
      cargo build --release --target wasm32-unknown-unknown --target-dir "${ROOT_DIR}/target" &&
        wasm-bindgen --target web --out-dir "${WASM_OUT}/fidget-wgpu" --out-name fidget_wgpu \
          "${ROOT_DIR}/target/wasm32-unknown-unknown/release/fidget_wgpu_web.wasm"
      RESULT=$?
      set -e
      if [ $RESULT -ne 0 ]; then
        echo ""
        echo "Failed to build fidget-wgpu-web"
        confirmOrExit "Continue anyway? (Y/n): "
      fi
    fi
  fi
  echo ""
fi

# Release profile is mandatory for manifold: backend/mesh/manifold cgo LDFLAGS
# link against target/release/libmanifold_glue, and build.rs CMake-builds the
# backend/mesh/manifold/vendor/manifold C++ submodule.
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


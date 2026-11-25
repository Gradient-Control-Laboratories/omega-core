#!/usr/bin/env bash
set -euo pipefail

#
# usage: ensure-native-libs.sh [--version <tag-or-sha>] [--out <dir>] [--force-build]
#   Ensure omega-core cgo native libraries (librender, libmanifold_glue) exist in <dir>
#   for the host OS/arch. When --version is a SemVer tag (v*) and --force-build is not set,
#   attempts to download and SHA256-verify the prebuilt GitHub Release tarball first,
#   falling back transparently to building from source via tool/build-fidget.sh --no-prompt.
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO="Gradient-Control-Laboratories/omega-core"

VERSION=""
OUT_DIR="${ROOT_DIR}/target/release"
FORCE_BUILD=false

while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      if [ $# -lt 2 ]; then
        echo "Error: --version requires an argument" >&2
        exit 2
      fi
      VERSION="$2"
      shift 2
      ;;
    --version=*)
      VERSION="${1#--version=}"
      shift
      ;;
    --out)
      if [ $# -lt 2 ]; then
        echo "Error: --out requires a directory argument" >&2
        exit 2
      fi
      OUT_DIR="$2"
      shift 2
      ;;
    --out=*)
      OUT_DIR="${1#--out=}"
      shift
      ;;
    --force-build)
      FORCE_BUILD=true
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      echo "usage: ensure-native-libs.sh [--version <tag-or-sha>] [--out <dir>] [--force-build]" >&2
      exit 2
      ;;
  esac
done

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux" ;;
    *)
      echo "Error: unsupported OS '$(uname -s)'" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    arm64|aarch64) echo "arm64" ;;
    x86_64|amd64)  echo "amd64" ;;
    *)
      echo "Error: unsupported architecture '$(uname -m)'" >&2
      exit 1
      ;;
  esac
}

verify_sha256() {
  local file="$1"
  local expected="$2"
  local actual=""
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${file}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${file}" | awk '{print $1}')"
  else
    echo "Warning: neither sha256sum nor shasum found; skipping download verification" >&2
    return 1
  fi
  [ "${actual}" = "${expected}" ]
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
LIB_EXT="so"
if [ "${OS}" = "darwin" ]; then
  LIB_EXT="dylib"
fi

RENDER_LIB="librender.${LIB_EXT}"
MANIFOLD_LIB="libmanifold_glue.${LIB_EXT}"
mkdir -p "${OUT_DIR}"

if [ "${FORCE_BUILD}" = "false" ] && [[ "${VERSION}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  ASSET="omega-core-native-${VERSION}-${OS}-${ARCH}.tar.gz"
  BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
  TMP_DIR="$(mktemp -d)"
  trap 'rm -rf "${TMP_DIR}"' EXIT

  echo "==> Attempting to download prebuilt native libs (${ASSET})..."
  if command -v curl >/dev/null 2>&1 \
    && curl -fsSL -o "${TMP_DIR}/${ASSET}" "${BASE_URL}/${ASSET}" 2>/dev/null \
    && curl -fsSL -o "${TMP_DIR}/${ASSET}.sha256" "${BASE_URL}/${ASSET}.sha256" 2>/dev/null; then
    EXPECTED_SHA="$(awk '{print $1}' "${TMP_DIR}/${ASSET}.sha256")"
    if verify_sha256 "${TMP_DIR}/${ASSET}" "${EXPECTED_SHA}"; then
      tar -xzf "${TMP_DIR}/${ASSET}" -C "${OUT_DIR}"
      if [ -f "${OUT_DIR}/${RENDER_LIB}" ] && [ -f "${OUT_DIR}/${MANIFOLD_LIB}" ]; then
        echo "  ✓ Downloaded and verified ${RENDER_LIB} and ${MANIFOLD_LIB} into ${OUT_DIR}."
        exit 0
      fi
    else
      echo "Warning: SHA256 mismatch for ${ASSET}; falling back to local source build." >&2
    fi
  else
    echo "==> Prebuilt tarball not available for ${VERSION} (${OS}/${ARCH}); falling back to local source build."
  fi
elif [ "${FORCE_BUILD}" = "true" ]; then
  echo "==> --force-build specified; building native libs from source for ${OS}/${ARCH}..."
else
  echo "==> Non-release version '${VERSION:-local}'; building native libs from source for ${OS}/${ARCH}..."
fi

"${SCRIPT_DIR}/build-fidget.sh" --no-prompt

SRC_DIR="${ROOT_DIR}/target/release"
if [ ! -f "${SRC_DIR}/${RENDER_LIB}" ] || [ ! -f "${SRC_DIR}/${MANIFOLD_LIB}" ]; then
  echo "Error: expected ${RENDER_LIB} and ${MANIFOLD_LIB} in ${SRC_DIR} after build." >&2
  exit 1
fi

ABS_SRC="$(cd "${SRC_DIR}" && pwd)"
ABS_OUT="$(cd "${OUT_DIR}" && pwd)"
if [ "${ABS_SRC}" != "${ABS_OUT}" ]; then
  cp "${ABS_SRC}/${RENDER_LIB}" "${ABS_OUT}/${RENDER_LIB}"
  cp "${ABS_SRC}/${MANIFOLD_LIB}" "${ABS_OUT}/${MANIFOLD_LIB}"
fi

echo "  ✓ Native libraries ready in ${ABS_OUT} (${RENDER_LIB}, ${MANIFOLD_LIB})."

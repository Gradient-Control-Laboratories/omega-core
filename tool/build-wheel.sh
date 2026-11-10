#!/usr/bin/env bash
set -euo pipefail

#
# usage: tool/build-wheel.sh [--programs-only]
#   build the omega-gsl wheel for this machine: gsl and compile into target/static/bin, compile linked
#   with the render library as a static library (the omega_static tag), so it loads only the system's
#   libraries; check that it does; write the licences of everything they link
#   (tool/third_party_licenses.py); then package them (python/omega-gsl) into dist/.
#   --programs-only stops before packaging.
#
#   env: OMEGA_GSL_TAG    the wheel's platform tag (CI: manylinux_2_28_<arch>); default this machine's
#        OMEGA_GSL_GLIBC  the newest glibc symbol version compile may need (default 2.28), Linux only;
#                         libstdc++'s are capped as manylinux_2_28 caps them (GLIBCXX 3.4.25, CXXABI 1.3.11)
#
# The development build (tool/build.sh, target/release) is not touched: this builds in target/static.
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT="${ROOT_DIR}/target/static/bin"
PROGRAMS_ONLY=false
[ "${1:-}" = "--programs-only" ] && PROGRAMS_ONLY=true

fail() { echo "build-wheel: $*" >&2; exit 1; }

for sub in internal/fidget/vendor/openvdb internal/fidget/vendor/tbb internal/fidget/vendor/imath \
           internal/fidget/vendor/blosc internal/fidget/vendor/zlib; do
  [ -f "${ROOT_DIR}/${sub}/CMakeLists.txt" ] ||
    fail "${sub} is empty: run (cd ${ROOT_DIR} && git submodule update --init --recursive)"
done

case "$(uname -s)" in
  Darwin) export MACOSX_DEPLOYMENT_TARGET=11.0 ;;
  Linux) ;;
  *) fail "no wheel build for $(uname -s) yet" ;;
esac

cd "${ROOT_DIR}"
echo "==> render as a static library (target/static)"
cargo rustc -p render --release --crate-type staticlib --target-dir "${ROOT_DIR}/target/static" \
  -- --print native-static-libs

mkdir -p "${OUT}"
"${SCRIPT_DIR}/build-stdlib.sh"
echo "==> compile (omega_static) and gsl"
go build -tags omega_static -trimpath -ldflags "-s -w" -o "${OUT}/compile" ./cmd/compile
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "${OUT}/gsl" ./cmd/gsl

echo "==> what compile loads"
case "$(uname -s)" in
  Darwin)
    otool -L "${OUT}/compile"
    foreign="$(otool -L "${OUT}/compile" | tail -n +2 | grep -vE '^[[:space:]]+/(usr/lib|System/Library)/' || true)"
    [ -z "${foreign}" ] || fail "compile loads libraries a Mac does not have: ${foreign}"
    minos="$(otool -l "${OUT}/compile" | awk '/LC_BUILD_VERSION/{f=1} f&&/minos/{print $2; exit}')"
    echo "minimum macOS: ${minos}"
    [ "${minos%%.*}" -le 11 ] || fail "compile needs macOS ${minos}, the wheel says 11.0"
    ;;
  Linux)
    readelf -d "${OUT}/compile" | grep NEEDED
    foreign="$(readelf -d "${OUT}/compile" | awk '/NEEDED/{print $NF}' | tr -d '[]' |
      grep -vE '^(libc|libm|libdl|libpthread|librt|libstdc\+\+|ld-linux[-a-z0-9_]*)\.so' || true)"
    [ -z "${foreign}" ] || fail "compile loads libraries outside glibc and libstdc++: ${foreign}"
    # The newest version of each symbol family compile needs, against manylinux_2_28's.
    for family_limit in "GLIBC ${OMEGA_GSL_GLIBC:-2.28}" "GLIBCXX 3.4.25" "CXXABI 1.3.11"; do
      set -- ${family_limit}
      newest="$(objdump -T "${OUT}/compile" | grep -o "$1_[0-9][0-9.]*" | sed "s/^$1_//" | sort -V | tail -1)"
      echo "newest $1 symbol: ${newest:-none} (limit $2)"
      [ -z "${newest}" ] || [ "$(printf '%s\n%s\n' "${newest}" "$2" | sort -V | tail -1)" = "$2" ] ||
        fail "compile needs $1 ${newest}, newer than $2"
    done
    ;;
esac
"${OUT}/gsl" --version
"${OUT}/compile" --version

echo "==> the licences of everything they link"
python3 "${SCRIPT_DIR}/third_party_licenses.py" "${OUT}/THIRD_PARTY_LICENSES.txt"

${PROGRAMS_ONLY} && exit 0

echo "==> the wheel"
OMEGA_GSL_BIN="${OUT}" python3 -m build --wheel --outdir "${ROOT_DIR}/dist" "${ROOT_DIR}/python/omega-gsl"
ls -l "${ROOT_DIR}/dist"

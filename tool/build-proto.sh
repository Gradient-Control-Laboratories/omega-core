#!/usr/bin/env bash
set -euo pipefail

#
# usage: build-proto.sh
#   build protobuf files in omega-core (program/program.proto and backend/tape/tape.proto)
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

if ! command -v protoc > /dev/null 2>&1; then
  echo "Add protoc to the path"
  echo "  See https://developers.google.com/protocol-buffers/"
  exit 1
fi
echo "found protoc on path"

if ! command -v protoc-gen-go > /dev/null 2>&1; then
  echo "Add protoc-gen-go to the path"
  echo "  See https://pkg.go.dev/google.golang.org/protobuf"
  exit 1
fi
echo "found protoc-gen-go on path"
echo ""

echo "building ${ROOT_DIR}/program/program.proto"
protoc \
  --go_out="${ROOT_DIR}" \
  --go_opt=module=gradientcontrol.com/omega/core \
  -I "${ROOT_DIR}/program" \
  "${ROOT_DIR}/program/program.proto"

echo "building ${ROOT_DIR}/backend/tape/tape.proto"
protoc \
  --go_out="${ROOT_DIR}" \
  --go_opt=module=gradientcontrol.com/omega/core \
  -I "${ROOT_DIR}/backend/tape" \
  "${ROOT_DIR}/backend/tape/tape.proto"

echo "build complete"


#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${SCRIPT_DIR}/build-proto.sh"
echo ""
"${SCRIPT_DIR}/build-stdlib.sh"
echo ""
"${SCRIPT_DIR}/build-wgsl-accel-ops.sh"
echo ""
"${SCRIPT_DIR}/build-fidget.sh" "$@"
echo ""
"${SCRIPT_DIR}/build-cli.sh"


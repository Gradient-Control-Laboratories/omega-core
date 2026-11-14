#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

SKIP_RUST=false
for arg in "$@"; do
    case "$arg" in
        --skip-rust)
            SKIP_RUST=true
            ;;
        *)
            echo "Unknown option: $arg" >&2
            echo "Usage: $0 [--skip-rust]" >&2
            exit 2
            ;;
    esac
done

"${SCRIPT_DIR}/test-go.sh"
if [ "$SKIP_RUST" = false ]; then
    echo ""
    "${SCRIPT_DIR}/test-rust.sh"
fi

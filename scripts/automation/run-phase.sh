#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  preflight)
    exec bash scripts/cycle/preflight.sh
    ;;
  dispatch)
    exec bash scripts/cycle/dispatch.sh
    ;;
  closeout)
    exec bash scripts/cycle/closeout.sh
    ;;
  *)
    echo "usage: $0 {preflight|dispatch|closeout}" >&2
    exit 1
    ;;
esac

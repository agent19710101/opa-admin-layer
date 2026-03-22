#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"

run_closeout() {
  log "closeout: running nightly validation"
  go test ./... >> "$LOG_DIR/closeout.log" 2>&1
  set_validation "go test ./... passed"
  if git diff --quiet && git diff --cached --quiet; then
    set_git_sync "clean local repo; nightly closeout clean"
  else
    set_git_sync "local changes present at closeout; sync required"
  fi
  set_next_action "2026-03-23T19:55:00+01:00 dev-preflight"
  update_summary
}

with_lock run_closeout

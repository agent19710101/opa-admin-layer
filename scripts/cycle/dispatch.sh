#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"

run_phase() {
  local phase
  phase=$(json_get currentPhase)
  log "dispatch: running phase ${phase}"
  case "$phase" in
    preflight)
      git status -sb >> "$LOG_DIR/dispatch.log"
      ;;
    ingest)
      find "$REPO_ROOT/docs/understanding" -maxdepth 1 -type f | sort >> "$LOG_DIR/dispatch.log"
      ;;
    architecture)
      find "$REPO_ROOT" -maxdepth 2 -type d | sort >> "$LOG_DIR/dispatch.log"
      ;;
    openspec)
      find "$REPO_ROOT/openspec" -type f | sort >> "$LOG_DIR/dispatch.log"
      ;;
    implement)
      go build ./cmd/opa-admin-layer >> "$LOG_DIR/dispatch.log" 2>&1
      ;;
    verify)
      go test ./... >> "$LOG_DIR/dispatch.log" 2>&1
      set_validation "go test ./... passed"
      ;;
    sync)
      if git diff --quiet && git diff --cached --quiet; then
        set_git_sync "clean local repo; no sync action needed"
      else
        set_git_sync "changes present locally; ready for commit/push"
      fi
      ;;
    persist)
      update_summary
      set_next_action "2026-03-22T20:15:00+01:00 dev-cycle-dispatch"
      ;;
    *)
      echo "unknown phase: $phase" >&2
      exit 1
      ;;
  esac
  advance_phase
  update_summary
}

with_lock run_phase

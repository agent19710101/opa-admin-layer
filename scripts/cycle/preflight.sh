#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "$0")/lib.sh"

run_preflight() {
  log "preflight: checking control plane and repo health"
  if python3 - "$SWITCH_REQUEST_FILE" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    data = json.load(f)
raise SystemExit(0 if data.get('status') == 'pending' else 1)
PY
  then
    log "preflight: pending switch request detected"
    python3 - "$ACTIVE_PROJECT_FILE" "$SWITCH_REQUEST_FILE" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    active = json.load(f)
with open(sys.argv[2], 'r', encoding='utf-8') as f:
    switch = json.load(f)
active['status'] = 'paused'
active['pausedReason'] = switch.get('notes', 'switch requested')
with open(sys.argv[1], 'w', encoding='utf-8') as f:
    json.dump(active, f, indent=2)
    f.write('\n')
switch['status'] = 'acknowledged'
with open(sys.argv[2], 'w', encoding='utf-8') as f:
    json.dump(switch, f, indent=2)
    f.write('\n')
PY
  fi
  git status -sb >> "$LOG_DIR/preflight.log"
  set_validation "pending"
  set_git_sync "pending"
  set_next_action "2026-03-22T20:00:00+01:00 dev-cycle-dispatch"
  update_summary
}

with_lock run_preflight

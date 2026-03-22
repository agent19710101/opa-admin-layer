#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
STATE_DIR="${OPA_ADMIN_LAYER_STATE_DIR:-${XDG_STATE_HOME:-$HOME/.local/state}/opa-admin-layer}"
STATE_FILE="$STATE_DIR/status/current-cycle.json"
SUMMARY_FILE="$STATE_DIR/status/summary.md"
BLOCKERS_FILE="$STATE_DIR/status/blockers.md"
ACTIVE_PROJECT_FILE="$STATE_DIR/control/active-project.json"
PROJECT_QUEUE_FILE="$STATE_DIR/control/project-queue.json"
SWITCH_REQUEST_FILE="$STATE_DIR/control/switch-request.json"
LOG_DIR="$STATE_DIR/logs"
LOCK_DIR="$STATE_DIR/.cycle.lock"
PHASE_ORDER=(preflight ingest architecture openspec implement verify sync persist)

ensure_state_layout() {
  mkdir -p "$STATE_DIR/control" "$STATE_DIR/status" "$LOG_DIR"

  if [[ ! -f "$ACTIVE_PROJECT_FILE" ]]; then
    cat > "$ACTIVE_PROJECT_FILE" <<'JSON'
{
  "id": "opa-administration-layer",
  "name": "OPA administration layer",
  "status": "active",
  "goal": "Build and operate a continuous workflow-first delivery system for the OPA administration layer.",
  "topology": "opa-only",
  "adminSurface": ["cli", "rest"],
  "tenantModel": "multi-tenant topic-scoped"
}
JSON
  fi

  if [[ ! -f "$PROJECT_QUEUE_FILE" ]]; then
    cat > "$PROJECT_QUEUE_FILE" <<'JSON'
{
  "projects": []
}
JSON
  fi

  if [[ ! -f "$SWITCH_REQUEST_FILE" ]]; then
    cat > "$SWITCH_REQUEST_FILE" <<'JSON'
{
  "status": "none",
  "requestedAt": null,
  "targetProject": null,
  "notes": "No pending switch request."
}
JSON
  fi

  if [[ ! -f "$STATE_FILE" ]]; then
    cat > "$STATE_FILE" <<'JSON'
{
  "phaseOrder": [
    "preflight",
    "ingest",
    "architecture",
    "openspec",
    "implement",
    "verify",
    "sync",
    "persist"
  ],
  "currentPhase": "preflight",
  "lastCompletedPhase": null,
  "lastRunAt": null,
  "lastValidation": "pending",
  "gitSync": "pending",
  "nextScheduledAction": "pending"
}
JSON
  fi

  if [[ ! -f "$BLOCKERS_FILE" ]]; then
    cat > "$BLOCKERS_FILE" <<'EOF_BLOCKERS'
# Blockers

## Current blockers

- OpenSpec CLI is not installed in this repo bootstrap flow; mirrored OpenSpec structure is in use.
EOF_BLOCKERS
  fi

  if [[ ! -f "$SUMMARY_FILE" ]]; then
    update_summary
  fi
}

log() {
  printf '%s %s\n' "$(date -Iseconds)" "$*" | tee -a "$LOG_DIR/cycle.log"
}

with_lock() {
  if ! mkdir "$LOCK_DIR" 2>/dev/null; then
    echo "cycle lock already held" >&2
    exit 1
  fi
  trap 'rmdir "$LOCK_DIR"' EXIT
  "$@"
}

json_get() {
  python3 - "$STATE_FILE" "$1" <<'PY'
import json, sys
path = sys.argv[2].split('.')
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    data = json.load(f)
value = data
for key in path:
    value = value.get(key)
print(value if value is not None else "")
PY
}

write_state() {
  python3 - "$STATE_FILE" "$1" "$2" <<'PY'
import json, sys
state_path, key, value = sys.argv[1], sys.argv[2], sys.argv[3]
with open(state_path, 'r', encoding='utf-8') as f:
    data = json.load(f)
parts = key.split('.')
ref = data
for part in parts[:-1]:
    ref = ref.setdefault(part, {})
ref[parts[-1]] = value
with open(state_path, 'w', encoding='utf-8') as f:
    json.dump(data, f, indent=2)
    f.write('\n')
PY
}

advance_phase() {
  local current next i
  current=$(json_get currentPhase)
  next=${PHASE_ORDER[0]}
  for ((i=0; i<${#PHASE_ORDER[@]}; i++)); do
    if [[ "${PHASE_ORDER[$i]}" == "$current" ]]; then
      next=${PHASE_ORDER[$(((i+1) % ${#PHASE_ORDER[@]}))]}
      break
    fi
  done
  write_state lastCompletedPhase "$current"
  write_state currentPhase "$next"
  write_state lastRunAt "$(date -Iseconds)"
}

set_validation() {
  write_state lastValidation "$1"
}

set_git_sync() {
  write_state gitSync "$1"
}

set_next_action() {
  write_state nextScheduledAction "$1"
}

update_summary() {
  cat > "$SUMMARY_FILE" <<EOF_SUMMARY
# Current workflow summary

- active project: $(python3 - "$ACTIVE_PROJECT_FILE" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    print(json.load(f)["name"])
PY
)
- state directory: $STATE_DIR
- current phase: $(json_get currentPhase)
- last completed phase: $(json_get lastCompletedPhase)
- last run at: $(json_get lastRunAt)
- latest validation: $(json_get lastValidation)
- git sync: $(json_get gitSync)
- next scheduled action: $(json_get nextScheduledAction)
EOF_SUMMARY
}

ensure_state_layout

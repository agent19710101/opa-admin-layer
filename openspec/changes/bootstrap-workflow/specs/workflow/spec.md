## ADDED Requirements

### Requirement: persisted workflow control plane
The system MUST maintain active project state, project queue state, switch-request state, cycle status, summary, and blockers in a runtime state directory outside the repository.

#### Scenario: initialize control plane
- **WHEN** the repository is bootstrapped
- **THEN** the runtime state directory contains `control/active-project.json`, `control/project-queue.json`, `control/switch-request.json`, `status/current-cycle.json`, `status/summary.md`, and `status/blockers.md`

### Requirement: one dispatcher slice per trigger
The dispatcher MUST execute exactly one ordered workflow phase for each trigger and persist the next phase.

#### Scenario: dispatch advances phase
- **WHEN** `scripts/cycle/dispatch.sh` runs
- **THEN** it records the current phase result and advances to the next phase in the fixed order

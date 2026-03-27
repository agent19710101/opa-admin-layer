# Proposal: reject repeated effective ServiceAccount names

## Summary

Reject repeated effective `serviceAccountName` values across topics so the renderer never emits ambiguous ownership for the same Kubernetes `ServiceAccount` object.

## Why now

The admin layer now renders a `ServiceAccount` manifest whenever a workload resolves an explicit effective `serviceAccountName`, and generated ServiceAccounts carry topic-specific identity labels. That means two topics that converge on the same effective name would attempt to own one Kubernetes object with conflicting metadata, even when operators only intended to share a workload identity binding.

## Scope

- detect repeated effective `serviceAccountName` values across all topics during validation
- report which topic first claimed the repeated name so operators can rename or centralize the binding deliberately
- document the contract in architecture, understanding docs, README, and changelog

## Out of scope

- deduplicating repeated ServiceAccounts into a shared-object plan model
- merge semantics for repeated ServiceAccount annotations or labels
- RBAC generation or broader ServiceAccount customization

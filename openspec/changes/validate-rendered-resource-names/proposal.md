# Proposal: validate rendered resource names

## Why

The renderer already validates propagated topic labels, but generated Kubernetes object names still depend on free-form spec, tenant, and topic names. Today the code lowercases and replaces spaces/underscores, which means punctuation, overly long identifiers, or unlucky combinations can still produce invalid Deployment, ConfigMap, and Service names.

## Change

- validate the rendered Kubernetes resource names used for tenant/topic workloads before plan rendering
- reject spec, tenant, and topic combinations that produce invalid or oversized object names
- return the same failures through CLI and REST validation/build paths
- document the stricter naming contract

## Impact

Operators get earlier, clearer feedback when tenant/topic identifiers cannot safely become Kubernetes object names, and generated plans stay applyable without downstream manifest repair.

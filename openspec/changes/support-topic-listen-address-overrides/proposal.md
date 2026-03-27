# Proposal: support topic listen-address overrides

## Summary

Allow a topic to override the shared `controlPlane.defaultListenAddress` with its own `listenAddress` so one workload can render a different OPA socket and Service port.

## Why now

The current shared-only listen-address contract forces operators to fork otherwise-identical specs when one workload needs a different bind port. That is unnecessary operational friction for a field the renderer already validates and fans out into Deployment args, probes, and Service ports.

## Scope

- add optional topic-level `listenAddress`
- validate topic values with the same strict `:port`, `host:port`, or `[ipv6]:port` contract used by `controlPlane.defaultListenAddress`
- render each topic's Deployment args, probes, and Service ports from its effective listen address
- document the override behavior in README, architecture notes, understanding docs, and changelog

## Out of scope

- protocol-specific probe customization
- per-topic host networking or container port naming changes
- multiple listener ports per topic

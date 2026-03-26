## Why

The REST API already accepts strict JSON and YAML admin specs, but the handler currently ignores request `Content-Type` and will try to decode any posted body. That makes the contract fuzzy for operators and weakens error signaling for bad integrations.

## What changes

- require `/v1/validate` and `/v1/plans` POST requests to use a supported JSON or YAML media type when `Content-Type` is provided
- keep empty `Content-Type` accepted for simple curl and internal callers, using the existing payload sniffing/decode path
- return a clear `415 Unsupported Media Type` error for unsupported media types
- document the accepted media types in README and architecture notes

## Impact

Operators and client integrations get a clearer REST contract, better failure modes, and less ambiguity about how YAML support should be used over HTTP.

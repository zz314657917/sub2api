### PASS: cf3577a3c-request

## Scope

- Implemented request-side Responses compatibility in the assigned service owners only: OAuth prompt/null-input and commands normalization, non-Astra reasoning mode removal (`pro` maps to `effort=max`), format-schema and lookaround pattern sanitization, orphan outputs, Unicode-safe input truncation, native `fc`/`ctc`/`tsc` IDs and item references, rejected-field retry coverage, and `gpt-image-2*` `input_fidelity` removal.
- Added OAuth reserved `python` aliasing plus selective response restoration. `allowed_tools` is kept intact; restoration is restricted to call-shaped objects so arbitrary content/nested application objects are not rewritten.
- Full JSON decoding introduced by this scope uses `Decoder.UseNumber`; compaction test also verifies a value above IEEE-754 precision is retained.
- Tool-schema null-type normalization accepts legal whitespace such as `"type": null` while still leaving omitted types untouched.

## Tests

- `go test <all service GoFiles> internal/service/cfport_request_test.go -count=1 -v` — PASS (nine focused cases, including null-content, exact cache-field, ambiguous/error-status, shared-budget, alias-collision and native-reference boundaries).
- `go build ./...` — PASS.
- `git diff --check` — PASS.

## Risks / handoff

- HTTP gateway and WS/bridge call-site wiring belongs to the controller and WS owner; helpers are same-package service APIs and have been communicated.
- No real provider, database, container, deployment, or browser validation was run.

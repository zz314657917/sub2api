### DONE: upstream-openai-ws-replay-s243

Implemented the approved behavior-level adaptation from upstream `25da02ddd`
and `66808413d`.

- Added exact tool output/context coverage analysis for array and object input.
- HTTP bridge replay now triggers for incomplete tool context, not complete
  tool-call/output pairs.
- Historical replay removes only unpaired historical tool-call context; paired
  function/custom calls and current-turn calls remain intact.
- Changed paths are limited to the S243 allowlist, including the bridge
  request-count regression test.

Evidence:

- Focused service tests passed.
- Full `go test ./internal/service` passed.
- `go test ./cmd/server -run '^$' -count=1` passed.
- `gofmt` and `git diff --check` passed.
- No provider, database, container, deployment, or push operation was run.

### DONE: upstream-v0177-remote-compaction-v2-s218

# S218 Generator Result

## Changed Files

- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_gateway_compact_body_signal_test.go`
- `backend/internal/handler/openai_gateway_handler_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_openai_compact_test.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_account_scheduler_compact_test.go`
- `backend/internal/service/openai_agent_identity_compat_test.go`
- `backend/internal/service/openai_compact_body_signal.go`
- `backend/internal/service/openai_compact_probe.go`
- `backend/internal/service/openai_compact_probe_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_forwarder_success_test.go`
- `backend/internal/service/openai_channel_restriction_compaction_test.go`
- `backend/internal/service/openai_compaction_context.go`

## Implementation

- A bare streaming Responses request containing `compaction_trigger` remains on
  `/responses` with its original body and does not set the legacy compact
  stream marker or keepalive bridge. Explicit `/responses/compact` requests and
  non-streaming body-signal requests retain legacy compact behavior.
- Native v2 account selection requires the local `openai_capabilities`
  `responses` gate, keeps legacy compact eligibility/mapping separate, and
  applies channel restrictions to the mapped forward model.
- The `x-codex-beta-features` allowlist and HTTP, passthrough, and WebSocket
  forwarding add `remote_compaction_v2` for native v2. OAuth default requests
  receive it only without a non-empty client declaration.
- Compact probes now use streaming `/responses`, an actual trigger,
  `text/event-stream`, OAuth `store:false`, a stable UUID-shaped session ID,
  and require a real compaction output item before persisting support.

## Verification

All commands ran from `E:/codex-worktrees/sub2api/s218-remote-compaction-v2/backend` unless noted.

- Default-tag discovery: each of the 4 handler and 11 service contract test
  names was found by `go test -list`.
- `go test ./internal/handler -run <S218 handler pattern> -count=10`: PASS.
- `go test ./internal/service -run <S218 service pattern> -count=10`: PASS.
- `go test ./internal/service -run <legacy compact pattern> -count=1`: PASS.
- `go test ./internal/service -count=1`: PASS (63.302s).
- `go test ./internal/handler -count=1`: PASS (59.756s).
- `go test ./internal/server -count=1`: PASS.
- `go test ./cmd/server -run '^$' -count=0`: PASS.
- From the worktree: `gofmt -d` on every changed and untracked Go file,
  `git diff --check`, conflict-index checks, and provenance checks for
  `9662cff2e`, `a8b9ea22b`, and `8ae6d8f67`: PASS.

## Boundary And Risks

- All regressions use in-process fakes, `httptest`, or local loopback fixtures.
  No OpenAI/ChatGPT/provider, production, shared PostgreSQL/Redis, deployment,
  container, or push request was made.
- This checkout predates the upstream `Responses` capability enum. The scoped
  scheduler helper uses the existing `openai_capabilities` map and defaults
  legacy-unconfigured OpenAI accounts to supported; explicitly configured
  chat-only accounts are rejected. Independent QA should recheck this local
  compatibility decision together with native/legacy mapping isolation.
- Changes are limited to contract allowlisted source and tests plus this result;
  no denied frontend, migration, dependency, generated, deployment, or
  workflow-controller path was modified.

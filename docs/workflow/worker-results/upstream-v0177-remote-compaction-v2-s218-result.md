### DONE: upstream-v0177-remote-compaction-v2-s218

## R1 Follow-up

- `Responses` account selection for API-key accounts now also honors the
  existing `openai_responses_supported` / `force_chat_completions` decision:
  unknown and explicit support remain compatible, while explicit unsupported
  and force-chat accounts are excluded before native v2 forwarding.
- `Forward` defensively preserves native-v2 requests on `/responses` when it
  is called directly, so a marked compaction trigger cannot be converted to raw
  chat even if a force-chat account bypasses normal scheduling.
- Added selector and real local-upstream regression coverage for both guards;
  the default-tag service regression is the acceptance gate, while the
  equivalent regression remains in the requested unit-tag fallback suite.

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
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
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
  `responses` gate and the existing API-key Responses support decision,
  keeps legacy compact eligibility/mapping separate, and applies channel
  restrictions to the mapped forward model.
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

### R1 Reverification

- Default-tag discovery found all original contract tests and
  `TestOpenAIGatewayServiceForwardNativeCompactionV2DoesNotUseRawChatFallback`.
- The explicit Responses-unsupported/force-chat selector regression and the
  direct native-v2 Forward regression passed together with `-count=10`.
- The handler and service contract patterns, extended with the default-tag
  direct-forward regression, passed with `-count=10`; legacy compact
  compatibility passed again.
- `go test ./internal/service -count=1`: PASS (65.416s).
- `go test ./internal/handler -count=1`: PASS (59.822s).
- `go test ./internal/server -count=1` and
  `go test ./cmd/server -run '^$' -count=0`: PASS.
- `go test -tags=unit ./internal/service -run
  '^TestForwardResponses_NativeCompactionV2DoesNotFallBackToRawChat$' -count=10`
  could not compile before the selected test because existing unit-tag files
  have unrelated duplicate `stringPtr`, stale billing signatures, and absent
  proxy/runtime members. The same local-upstream request assertion is present
  and passing in the default-tag service suite; no baseline unit-tag failures
  were modified.
- R1 format, diff, allowlist, conflict-index, unmerged-index, and upstream
  provenance checks passed.

## Boundary And Risks

- All regressions use in-process fakes, `httptest`, or local loopback fixtures.
  No OpenAI/ChatGPT/provider, production, shared PostgreSQL/Redis, deployment,
  container, or push request was made.
- This checkout predates the upstream `Responses` capability enum. The scoped
  scheduler helper combines the existing `openai_capabilities` map with the
  API-key `ShouldUseResponsesAPI` decision: unknown/explicit support remains
  compatible; an explicit unsupported or force-chat account is rejected.
  Independent QA should recheck this local compatibility decision together
  with native/legacy mapping isolation.
- Changes are limited to contract allowlisted source and tests plus this result;
  no denied frontend, migration, dependency, generated, deployment, or
  workflow-controller path was modified.

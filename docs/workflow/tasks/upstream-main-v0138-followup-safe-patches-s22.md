---
task_id: upstream-main-v0138-followup-safe-patches-s22
phase: contract-approved
owner: Codex
qa_mode: runtime
created_at: 2026-06-26 16:20 +08:00
---

# Task Contract: upstream v0.1.138 follow-up safe patches S22

## Goal

Port the next small, locally useful backend fixes from post-`v0.1.138` `upstream/main` without wholesale merging upstream or changing local product surfaces.

## Success Criteria

- Chat Completions -> Responses stream conversion does not duplicate tool call arguments when a single upstream chunk includes id, name, and arguments together.
- OpenAI Responses passthrough events dedupe accidentally doubled `function_call` / `custom_tool_call` `arguments` in done/item/completed payloads while preserving normal argument deltas.
- `refresh_token_invalidated` is classified as a non-retryable OpenAI OAuth token refresh error.
- `/v1/chat/completions` transport errors return `UpstreamFailoverError` instead of writing a terminal 502 response directly, so normal failover can continue.
- Email auth identity creation errors are not swallowed by an inner shadowed `err`.
- Targeted backend tests, denied-path audit, and `git diff --check` pass.

## Allowed Paths

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_passthrough_function_args_test.go`
- `backend/internal/service/openai_upstream_transport_error_handle_test.go`
- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/auth_service.go`
- `backend/internal/service/auth_service_identity_shadow_test.go`
- `backend/internal/service/auth_service_identity_sync_test.go`
- `docs/workflow/tasks/upstream-main-v0138-followup-safe-patches-s22.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/worker-results/upstream-main-v0138-followup-safe-patches-s22-result.md`
- `docs/workflow/qa-reports/upstream-main-v0138-followup-safe-patches-s22-qa.md`

## Denied Paths

- Ent schema/generated files and migrations.
- Payment fulfillment, affiliate rebate, subscription package logic, balance preflight, order currency UI/API changes.
- Scheduler strategy/config, channel monitor, ops dashboard layout, CI, deploy, README, VERSION, sponsor assets.
- Public pages, Studio/Canvas, model market, payment pages, frontend surfaces, and unrelated knowledge files.
- GPT-5.5 Codex instructions fallback, Antigravity standard-tier project fallback, payment provider supported-types changes.

## Constraints

- Do not cherry-pick upstream commits blindly; adapt the patches to local helpers and test file layout.
- Keep argument dedupe narrowly scoped to exact repeated JSON argument strings; do not mutate arbitrary model text.
- Do not write client responses on transport errors in chat-completions paths before failover has a chance to run.
- Do not stage or commit unless explicitly requested.

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/pkg/apicompat -run "TestStream_ToolCallArgumentsInFirstChunkNotDoubled" -count=1
go test ./internal/service -run "TestHandleStreamingResponsePassthroughDeduplicatesFunctionCallArguments|TestForwardResponsesChatCompletionsFallbackKeepsFunctionArgumentsSingle|TestForwardAsChatCompletions_TransportErrorReturnsFailover|TestForwardAsRawChatCompletions_TransportErrorReturnsFailover|TestIsNonRetryableRefreshError|TestEnsureEmailAuthIdentityCreateErrorReturnsFalse" -count=1

cd F:/mcplugins/sub2api
git diff --check
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|backend/internal/service/payment_|backend/internal/service/subscription_|backend/internal/service/gateway_billing_|deploy/|README|README_|assets/partners/|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment|knowledge/05-current-focus.md)" || echo NO_DENIED_PATHS
```

## Output

- Code diff in the allowed paths only.
- Worker-style result, QA report, and final report with `Findings / Executed Checks / Unverified Risks / Recommendation`.
- Candidate accounting that marks each reviewed upstream patch as `ported`, `equivalent`, or `skipped`.

## Stop Rules

- Stop if any selected fix requires Ent/migration/wire generation or product-surface changes.
- Stop if argument dedupe would need a parser that rewrites non-exact repeated JSON content.
- Stop if chat-completions failover changes would alter already-committed response semantics after client output has started.

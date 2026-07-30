---
task_id: openai-model-capacity-retry-s126
repo: F:/mcplugins/sub2api
phase: done
owner: codex
source: upstream/issue-2223-followup
---

# Task Contract

## Task ID

`openai-model-capacity-retry-s126`

## Role

Planner/Generator by Codex; no worker delegation is needed for this narrow
service-layer correction. The final review remains the Evaluator gate.

## Goal

Complete the local follow-up to upstream PR `#2481` so the precise OpenAI
`Selected model is at capacity. Please try a different model.` failure receives
bounded same-account retries before the existing account failover path, without
changing the requested model or broadening retry behavior for other 400 errors.

## Success Criteria

- A normal non-passthrough OpenAI HTTP 400 capacity response returns an
  `UpstreamFailoverError` with `RetryableOnSameAccount=true` for both OAuth and
  API-key accounts, including accounts that are not in pool mode.
- An OpenAI passthrough HTTP 400 capacity response is not written to the client;
  it returns the same retryable failover signal. A non-capacity passthrough 400
  remains an immediate passthrough response, while existing 429/529 failover
  behavior remains unchanged.
- Standard and passthrough `response.failed` capacity events received before
  client output return a retryable failover signal. Generic processing errors,
  policy/safety failures, context-window failures, disconnects, and failures
  after client output retain their current behavior.
- The handler continues to use the existing bounded retry loop: non-pool
  accounts inherit the existing default retry limit of three and 500 ms delay;
  after exhaustion, the existing account-switch path runs. No model is selected
  or substituted by the gateway.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/openai-capacity-retry-s126`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Upstream source: `9f07741c1` and `ed7ef8634` from merged PR `#2481`
- Local first-layer port: `a5b420f09` and `64629b422`
- Issue `#2223` was still open at the 2026-07-29 live check. Independently,
  local inspection confirms passthrough and non-pool same-account eligibility
  are not fully covered by the upstream patch.

## Allowed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_oauth_passthrough_test.go`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/openai-model-capacity-retry-s126.md`
- `docs/workflow/qa-reports/openai-model-capacity-retry-s126-qa.md`

## Denied Paths

- `backend/internal/handler/**`
- `backend/internal/service/**` except the four explicitly allowed files
- `backend/internal/repository/**`
- `backend/internal/server/**`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- Any path not listed under Allowed Paths

## Constraints

- Introduce one narrow capacity classifier and reuse it in the existing HTTP,
  passthrough, and stream failover construction paths.
- Do not make every transient processing error retryable on the same account;
  only the precise model-capacity error receives the non-pool retry flag.
- Preserve upstream response bodies, operation-error events, redaction,
  rate-limit side effects, model routing, billing, scheduling, and account
  selection.
- Do not add configuration, retry counters, delays, migrations, routes, or
  frontend controls. Do not automatically change the requested model.
- Keep all changes in the isolated worktree until final source-level QA passes.
  Do not push, deploy, or update containers.

## Acceptance Commands

Run from `E:/codex-worktrees/sub2api/openai-capacity-retry-s126/backend`:

```powershell
go test ./internal/service -run "Test(IsOpenAIModelCapacityError|IsOpenAITransientProcessingError|OpenAIGatewayService_Forward_(TransientProcessingErrorTriggersFailover|ModelCapacityErrorTriggersFailoverAndSameAccountRetry)|OpenAIGatewayService_OpenAIPassthrough_(Capacity400RetriesBeforeFailover|429And529TriggerFailover)|OpenAIGatewayService_OAuthPassthrough_UpstreamErrorIncludesPassthroughFlag|OpenAIStreaming(ResponseFailedBeforeOutput|PassthroughResponseFailedBeforeOutput))" -count=1
go test ./... -run "^$"
gofmt -d internal/service/openai_gateway_service.go internal/service/openai_gateway_service_codex_cli_only_test.go internal/service/openai_gateway_service_test.go internal/service/openai_oauth_passthrough_test.go
```

Run from the worktree root:

```powershell
git diff --check
rg -n "^(<<<<<<<|=======|>>>>>>>)" backend/internal/service docs/workflow
git diff --name-only HEAD
```

The final path audit must contain only Allowed Paths. Focused tests must prove
the client writer remains untouched for retryable capacity errors and that
non-capacity 400/policy/context behavior does not broaden.

## Output

- Narrow service change, focused regressions, workflow QA report, and final
  Evaluator verdict as `PASS`, `FAIL`, or `BLOCKED`.
- No worker report is required because Codex owns this narrow implementation.

## Stop Rules

- Stop before implementation if bounded retries require a handler, scheduler,
  account-selection, persistence, route, billing, or configuration change.
- Stop if the capacity condition cannot be separated from generic 400,
  context-window, or policy/safety failures.
- Stop if any primary-worktree dirty path or any path outside Allowed Paths
  appears in the S126 diff.

## Budget

- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree: `E:/codex-worktrees/sub2api/openai-capacity-retry-s126`

---
task_id: openai-model-capacity-retry-five-s127
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

For the exact OpenAI error `Selected model is at capacity. Please try a
different model.`, raise the bounded same-account retry limit from three to
five attempts before the existing account failover path. All other errors keep
their current retry limits and account-pool behavior.

## Success Criteria

- Normal HTTP, passthrough HTTP, standard pre-output streaming, and passthrough
  pre-output streaming capacity failures carry `RetryableOnSameAccount=true`
  and an explicit same-account retry limit of `5`.
- The generic handler failover loop performs five same-account retries for an
  explicit limit of `5`; the sixth failure runs the existing unschedule and
  account-switch path.
- OpenAI Responses, Messages, Chat Completions, and Images loops prefer the
  capacity-specific limit when present, otherwise preserve
  `account.GetPoolModeRetryCount()`.
- Generic transient errors, pool-mode 429/529, overloaded failures, context
  failures, and all non-capacity errors keep their existing retry limits.
- The retry delay remains 500 ms and no model is substituted.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/openai-capacity-retry-five-s127`
- S126 already introduced capacity classification and same-account eligibility.
  Its generic handler default remains three retries, which is the narrow
  follow-up addressed here.

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_oauth_passthrough_test.go`
- `backend/internal/handler/failover_loop.go`
- `backend/internal/handler/failover_loop_test.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_images.go`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/**`
- `frontend/**`
- `deploy/**`
- `Dockerfile*`
- `docker-compose*.yml`
- `knowledge/**`
- `outputs/**`
- `C:/Users/Administrator/.codex/memories/**`
- Any path not listed under Allowed Paths

## Constraints

- Do not change `Account.GetPoolModeRetryCount()` or a global retry setting.
- Only the exact capacity classifier may set the explicit limit; generic
  transient/overload/rate-limit errors must leave it unset.
- Preserve account selection, cooldown, unscheduling, billing, request body,
  route, persistence, model routing, and scheduling behavior.
- Do not push, deploy, update containers, reset, clean, or modify the primary
  worktree.

## Acceptance Commands

Run from `backend`:

```powershell
go test ./internal/service -run "Test(OpenAIGatewayService_Forward_ModelCapacityErrorTriggersFailoverAndSameAccountRetry|OpenAIGatewayService_OpenAIPassthrough_Capacity400RetriesBeforeFailover|OpenAIStreaming(ResponseFailedBeforeOutputCapacityErrorReturnsFailover|PassthroughResponseFailedBeforeOutputCapacityErrorRetriesSameAccount)|OpenAIGatewayService_Forward_TransientProcessingErrorTriggersFailover|OpenAIGatewayService_OpenAIPassthrough_429And529TriggerFailover)" -count=1
go test ./internal/handler -run "TestHandleFailoverError_(SameAccountRetry|BasicSwitch|IntegrationScenario)" -count=1
go test ./... -run "^$"
```

Run from the worktree root:

```powershell
gofmt -d <changed Go files>
git diff --check
rg -n "^(<<<<<<<|=======|>>>>>>>)" <changed paths>
git diff --name-only HEAD
```

## Output

- A narrow implementation, focused regressions, and
  `docs/workflow/qa-reports/openai-model-capacity-retry-five-s127-qa.md`.
- Codex directly owns implementation and QA; no worker is invoked because the
  user did not authorize sub-agent work.

## Stop Rules

- Stop if raising this limit requires a global pool retry default, a config
  change, a scheduler/account-selection change, persistence, deployment, or a
  model-routing change.
- Stop if a non-capacity path needs to receive the new explicit limit.
- Stop if an out-of-scope path appears in the diff.

## Contract Review

`PASS`: the existing `UpstreamFailoverError` reaches all required handler
loops, so an optional capacity-specific limit can preserve the generic fallback
without changing the pool account configuration or scheduler semantics.

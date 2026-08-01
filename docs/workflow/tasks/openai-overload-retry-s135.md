---
task_id: openai-overload-retry-s135
status: done
role: Generator
qa_mode: runtime
---

# Task Contract

## Task ID

`openai-overload-retry-s135`

## Role

Planner/Generator by Codex; no worker delegation is authorized for this bounded
follow-up. Codex also owns focused QA and the final Evaluator verdict.

## Goal

Retry narrowly identified OpenAI overload failures on the same account three
times with linear delays of 1s, 2s, and 3s before existing account failover,
while preserving the model-capacity policy of five retries at 500ms.

## Success Criteria

- Structured OpenAI error codes `server_is_overloaded` and `slow_down`, or the
  exact phrase `Our servers are currently overloaded`, receive
  `RetryableOnSameAccount=true`, retry limit `3`, and a 1-second linear
  backoff base.
- Normal HTTP, passthrough HTTP, standard pre-output streaming, and passthrough
  pre-output streaming paths carry the overload retry policy before writing a
  client response.
- Existing OpenAI compatibility/fallback constructors used by Chat
  Completions, Messages, Responses fallback, and Images carry the same policy
  so handler behavior does not depend on endpoint adaptation mode.
- Handler retry attempt 1 waits 1s, attempt 2 waits 2s, and attempt 3 waits 3s;
  the fourth overload failure uses the existing unschedule/account-switch path.
- `Selected model is at capacity` remains five retries with the established
  fixed 500ms delay.
- Generic `overloaded` text, ordinary 400/500/503 responses, context/policy
  failures, and generic passthrough 429/529 responses do not gain the overload
  policy.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `E:/codex-worktrees/sub2api/openai-overload-retry-s135`
- Base: local `main@80e564a63`; the primary worktree contains unrelated S134
  frontend/workflow changes and must remain untouched.
- S127 introduced a capacity-only explicit retry limit. S135 extends the same
  failover metadata boundary with a distinct overload delay profile.

## Allowed Paths

- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
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

- Match overload only by the two structured codes or the agreed exact phrase;
  do not classify generic `overloaded`, generic HTTP 5xx, or arbitrary server
  text as same-account overload retries.
- Preserve the existing capacity classifier, limit, fixed delay, account
  selection, cooldown, unscheduling, billing, routing, request payload,
  persistence, and model behavior.
- Implement linear overload delay as a per-error policy carried through
  `UpstreamFailoverError`; do not change a global retry setting or account pool
  configuration.
- During implementation and QA, do not commit, merge, push, deploy, update
  containers, reset, clean, or alter the primary S134 worktree. Git publication
  is a separate post-QA action authorized by the user.

## Acceptance Commands

Run from `backend`:

```powershell
go test ./internal/service -run "Test(IsOpenAIServerOverloadedError|OpenAISameAccountRetryPolicy|OpenAIGatewayService_Forward_(ModelCapacityError|ServerOverloaded)|OpenAIGatewayService_OpenAIPassthrough_(Capacity400|ServerOverloaded|429And529)|OpenAIStreaming(ResponseFailedBeforeOutputCapacity|ResponseFailedBeforeOutputServerOverloaded|PassthroughResponseFailedBeforeOutputCapacity|PassthroughResponseFailedBeforeOutputServerOverloaded))" -count=1
go test ./internal/handler -run "Test(SameAccountRetryLimit|SameAccountRetryDelay|HandleFailoverError_SameAccountRetry)" -count=1
go test ./... -run "^$"
```

Run from the worktree root:

```powershell
gofmt -d <changed Go files>
git diff --check
rg -n "^(<<<<<<<|=======|>>>>>>>)" <changed paths>
git diff --name-only HEAD
```

The amended constructor coverage also requires a source audit confirming that
every in-scope OpenAI capacity failover constructor now uses
`openAISameAccountRetryPolicy`; Embeddings and Videos remain on their existing
capacity-only behavior until a pinned-account retry seam is designed.

## Output

- Narrow implementation, focused regressions, and
  `docs/workflow/qa-reports/openai-overload-retry-s135-qa.md`.
- Final verdict must be `PASS`, `FAIL`, or `BLOCKED` with executed evidence and
  unverified runtime risks.

## Stop Rules

- Stop if the behavior requires a global retry default, scheduler/account
  selection change, schema/configuration, persistence, frontend, deployment,
  or container change.
- Stop if a generic overload message or generic HTTP status would receive the
  explicit overload retry policy.
- Stop if any required path overlaps primary S134 dirt or any out-of-scope path
  appears in the isolated diff.

## Contract Review

`PASS`: the existing failover error already carries a capacity-only retry limit;
adding a separate overload limit and backoff base preserves the generic pool
fallback while covering the four in-scope Responses, Messages, Chat
Completions, and Images handler paths.

`PASS / amended`: the initial implementation audit found eight existing OpenAI
compatibility/fallback constructors that still emitted the capacity-only
metadata. The amendment covers only the four in-scope endpoint families;
Embeddings and Videos were withdrawn after review because their current
selection loops do not pin retries to the original account.

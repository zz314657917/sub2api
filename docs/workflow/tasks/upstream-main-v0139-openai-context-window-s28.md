---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-06-30 13:49 +08:00
---

# Task Contract

## Task ID
upstream-main-v0139-openai-context-window-s28

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small follow-up port. No external worker is used.

## Goal
Port the safe part of upstream `7cbf82ed6` so OpenAI context-window errors are returned to the caller instead of being classified as retryable upstream failures that trigger account failover.

## Success Criteria
- OpenAI HTTP 502/5xx context-window responses do not trigger `UpstreamFailoverError`.
- Raw Responses streaming `response.failed` context-window events before client output are passed through to the client.
- Chat Completions buffered and streaming Responses-bridge `response.failed` context-window events return client-visible errors without failover.
- Transient OpenAI failures such as `server_is_overloaded` still trigger failover.
- No frontend, migrations, Ent, wire, deploy, VERSION, README, or `knowledge/*` files are included in the commit.

## Context
- Repo: `F:/mcplugins/sub2api`
- Upstream reference: `7cbf82ed6 修复 OpenAI 上下文窗口错误误触发账号切换`
- Upstream also touched `openai_account_runtime_block_fastpath.go`; this local repo currently has no matching file or fastpath entry, so S28 only ports the gateway classification and pass-through behavior that exists locally.
- Current worktree contains unrelated dirty proxy/account/frontend/knowledge files. They are outside this sprint.

## Allowed Paths
- `backend/internal/service/error_passthrough_runtime_test.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_cli_only_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-main-v0139-openai-context-window-s28.md`
- `docs/workflow/worker-results/upstream-main-v0139-openai-context-window-s28-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-openai-context-window-s28-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/server/routes/**`
- `backend/internal/service/wire.go`
- `deploy/**`
- `README*`
- `assets/partners/**`
- Payment, subscription, keys UI, ops UI, Grok routing, risk-control, OAuth email flow, proxy/account ownership work, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep this sprint to OpenAI context-window error classification and pass-through behavior only.
- Do not create local equivalents of upstream files that do not exist here unless required by compiled code.
- Preserve existing transient-error failover behavior.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, or knowledge files.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestIsOpenAIContextWindowError|TestShouldFailoverOpenAIUpstreamResponseContextWindow502|TestOpenAIHandleErrorResponse_ContextWindow502KeepsMessageWithoutFailover|TestForwardAsChatCompletions_BufferedContextWindowResponseFailedReturnsErrorWithoutFailover|TestForwardAsChatCompletions_BufferedTransientResponseFailedTriggersFailover|TestForwardAsChatCompletions_StreamContextWindowResponseFailedReturnsErrorWithoutFailover|TestOpenAIStreamingContextWindowResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedBeforeOutputServerOverloadedCodeReturnsFailover" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/service/error_passthrough_runtime_test.go backend/internal/service/openai_gateway_chat_completions.go backend/internal/service/openai_gateway_chat_completions_test.go backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_codex_cli_only_test.go backend/internal/service/openai_gateway_service_test.go docs/workflow/tasks/upstream-main-v0139-openai-context-window-s28.md docs/workflow/worker-results/upstream-main-v0139-openai-context-window-s28-result.md docs/workflow/qa-reports/upstream-main-v0139-openai-context-window-s28-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0139-openai-context-window-s28-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0139-openai-context-window-s28-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementing this requires frontend, Ent, migration, wire, route, handler, repository, deploy, README, or `knowledge/*` changes.
- Stop if the context-window classifier would suppress genuine transient errors.
- Stop if local code lacks a safe corresponding path for an upstream hunk.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`

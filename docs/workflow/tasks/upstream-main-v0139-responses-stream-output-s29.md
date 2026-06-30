---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-06-30 14:06 +08:00
---

# Task Contract

## Task ID
upstream-main-v0139-responses-stream-output-s29

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small follow-up port. No external worker is used.

## Goal
Port the local-relevant part of upstream `e9a2db8e80` so OpenAI Responses streaming terminal events with `response.output:null` or missing usable output are normalized to SDK-parseable `response.output` arrays.

## Success Criteria
- Streaming `response.completed` / `response.done` / incomplete/cancel terminal events rebuild `response.output` from accumulated text, function-call, reasoning-summary, and image output events when terminal output is null or empty.
- If no content was accumulated, terminal `response.output` is normalized to an empty array instead of `null`.
- Existing non-empty terminal `response.output` remains unchanged.
- Model replacement still happens after terminal normalization so mapped-model values do not leak to clients.
- No frontend, migrations, Ent, wire, deploy, VERSION, README, or `knowledge/*` files are included in the commit.

## Context
- Repo: `F:/mcplugins/sub2api`
- Upstream reference: `e9a2db8e80 fix: normalize responses streaming terminal output`
- Candidate scan before S29:
  - `b9509e823a` / `ed2aac25a` billing long-context cache read/creation multipliers are already local-equivalent.
  - `8a999f438d` WS terminal events excluded from token events is already local-equivalent.
  - `0a521f09fb` Gemini messages tool-use block closure is already local-equivalent.
  - `03ae510c68` ops count_tokens metrics exclusion is already local-equivalent.
- Current worktree contains unrelated dirty proxy/account/frontend/knowledge files. They are outside this sprint.

## Allowed Paths
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `docs/workflow/tasks/upstream-main-v0139-responses-stream-output-s29.md`
- `docs/workflow/worker-results/upstream-main-v0139-responses-stream-output-s29-result.md`
- `docs/workflow/qa-reports/upstream-main-v0139-responses-stream-output-s29-qa.md`
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
- Payment, subscription, keys UI, ops UI, Grok routing, risk-control, OAuth email flow, proxy/account ownership work, and production configuration paths.

## Constraints
- Do not merge or rebase `upstream/main`.
- Keep this sprint to OpenAI Responses streaming terminal output normalization only.
- Do not change OpenAI error/failover classification in this sprint.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, or knowledge files.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/service -run "TestOpenAIStreamingNormalizesTerminalOutputFromDeltas|TestOpenAIStreamingNormalizesTerminalOutputToEmptyArray|TestOpenAIStreamingPreambleKeepaliveUsesDownstreamIdle|TestOpenAIStreamingPolicyResponseFailedBeforeOutputPassesThrough|TestOpenAIStreamingResponseFailedAfterOutputSanitizesVerboseResponseForClient" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_test.go docs/workflow/tasks/upstream-main-v0139-responses-stream-output-s29.md docs/workflow/worker-results/upstream-main-v0139-responses-stream-output-s29-result.md docs/workflow/qa-reports/upstream-main-v0139-responses-stream-output-s29-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0139-responses-stream-output-s29-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0139-responses-stream-output-s29-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementing this requires frontend, Ent, migration, wire, route, handler, repository, deploy, README, or `knowledge/*` changes.
- Stop if normalization would rewrite terminal events that already contain non-empty output.
- Stop if a corresponding local streaming path cannot be identified safely.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`

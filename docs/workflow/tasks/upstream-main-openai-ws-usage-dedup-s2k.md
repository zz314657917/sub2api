# Task Contract

## Task ID
upstream-main-openai-ws-usage-dedup-s2k

## Role
Codex acts as Planner, Generator, and Final Evaluator for this tiny backend service Sprint. Implement only the OpenAI WS usage request-id deduplication fix selected here.

## Goal
Port the safe service subset of upstream `1e2193c3d fix: avoid websocket usage dedup conflicts` onto the current upstream-sync branch. Normal OpenAI HTTP usage should keep preferring stable `client_request_id`, but OpenAI WebSocket turns must use the upstream/response request id so multiple turns on one WS connection do not collapse into one billing/usage dedup key.

## Success Criteria
- `OpenAIGatewayService.RecordUsage` keeps existing request-id behavior for non-WS OpenAI requests.
- When `OpenAIForwardResult.OpenAIWSMode` is true and `Result.RequestID` is non-empty, billing and usage log request id use `Result.RequestID`, even if `ctxkey.ClientRequestID` exists.
- No OpenAI WS forwarder, WS bridge, scheduler, gateway routing, schema, migration, public API, config, frontend, payment, or subscription behavior is changed.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-openai-oauth-refresh-enrichment-s2j`
- Work branch: `codex/upstream-main-openai-ws-usage-dedup-s2k`
- Upstream source commit: `1e2193c3d fix: avoid websocket usage dedup conflicts`
- Main worktree `F:/mcplugins/sub2api` must not be modified.
- Already equivalent locally, not part of this Sprint:
  - `8a999f438 fix(ws): exclude terminal events from first-token detection`
  - `2bd3125d Preserve usage request context`

## Allowed Paths
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `docs/workflow/tasks/upstream-main-openai-ws-usage-dedup-s2k.md`
- `docs/workflow/worker-results/upstream-main-openai-ws-usage-dedup-s2k-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-ws-usage-dedup-s2k-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/ratelimit_service.go`
- `frontend/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS bridge redesign, Responses bridge redesign, OpenAI gateway routing redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Keep this as a surgical service/test patch; do not cherry-pick if it attempts to widen beyond allowed paths.
- Preserve `RequestIDOverride` behavior if present in local code.
- Preserve non-WS preference for stable client request id.
- If the fix requires changing WS forwarder payload handling, gateway handlers, billing schema, or API shape, stop and defer to a later WS/gateway Sprint.

## Candidate Commit
- Primary: `1e2193c3d fix: avoid websocket usage dedup conflicts`

## Explicitly Deferred
- OpenAI WS bridge/failover/image tool injection commits remain deferred.
- Large gateway request-body/view refactors remain deferred.
- OpenAI endpoint capability UI/API and account/admin behavior changes remain out of this Sprint unless already handled in earlier local commits.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test ./internal/service -run "OpenAIGatewayServiceRecordUsage_(PrefersClientRequestIDOverUpstreamRequestID|WSModePrefersUpstreamRequestIDOverClientRequestID|GeneratesRequestIDWhenAllSourcesMissing)" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-ws-usage-dedup-s2k-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-openai-ws-usage-dedup-s2k-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval and QA events.
- Update `knowledge/tasks/current-task.md` with the current handoff snapshot after QA.

## Stop Rules
- Stop if implementation touches denied paths.
- Stop if target tests require changing billing schema, repositories, request DTOs, or WS forwarder behavior.
- Stop if non-WS request-id preference changes or `RequestIDOverride` is broken.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`

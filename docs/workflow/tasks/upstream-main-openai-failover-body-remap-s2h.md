# Task Contract

## Task ID
upstream-main-openai-failover-body-remap-s2h

## Role
Codex acts as Planner, Generator, and Final Evaluator for this small backend failover Sprint. Implement only the OpenAI failover request-body remap behavior and regression coverage selected here.

## Goal
Port the safe subset behind upstream `c8cd91e3c test(openai): 覆盖 failover 请求体重映射` onto the current upstream-sync branch. OpenAI failover retries must reparse the original request body for each candidate account so per-account `model_mapping` is applied from the client model, not from a previously rewritten cached request map.

## Success Criteria
- `OpenAIGatewayService.Forward` failover attempts can apply different `credentials.model_mapping` values for different OpenAI OAuth accounts.
- `getOpenAIRequestBodyMap` parses the supplied `body` argument and ignores legacy `OpenAIParsedRequestBodyKey` context cache values.
- `getOpenAIRequestBodyMap` does not write a parsed body back into gin context.
- Existing JSON parse error behavior remains intact.
- Handler-side use of `OpenAIParsedRequestBodyKey` for validation/Claude Code helpers is not removed or broadened.
- No schema, migration, config, public API, frontend, OpenAI WS bridge, Responses bridge redesign, or gateway routing redesign is introduced.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-claude-count-tokens-s2g`
- Work branch: `codex/upstream-main-openai-failover-body-remap-s2h`
- Upstream source commit: `c8cd91e3c`
- Main worktree `F:/mcplugins/sub2api` must not be modified.
- Local inspection confirmed current `getOpenAIRequestBodyMap` still reads and writes `OpenAIParsedRequestBodyKey`, while upstream now reparses `body`.

## Allowed Paths
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/service/openai_failover_cached_body_test.go`
- `docs/workflow/tasks/upstream-main-openai-failover-body-remap-s2h.md`
- `docs/workflow/worker-results/upstream-main-openai-failover-body-remap-s2h-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-failover-body-remap-s2h-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `frontend/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS/Responses bridge redesign, OpenAI gateway routing redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Do not cherry-pick if it attempts to widen the patch beyond allowed paths.
- Keep the implementation limited to service-layer request map parsing semantics and focused tests.
- Preserve `OpenAIParsedRequestBodyKey` for existing handler-side validation and Claude Code helper usage.
- Do not add live upstream smoke tests or require credentials.
- If the selected patch requires touching handlers, schema, frontend, config, WS/Responses bridge, or routing architecture, stop and split a new Sprint.

## Candidate Commit
- Primary: `c8cd91e3c test(openai): 覆盖 failover 请求体重映射`

## Explicitly Deferred
- `08e19bb15`, `d7bed40dd`, `08061717b`: OpenAI WS bridge/failover-sized changes.
- `f10bca815`: Codex Responses bridge redesign.
- `003b2786d`: apicompat bridge test chain.
- `a39163519`: OpenAI key generated config default model policy.
- `5fd9a3509`: pricing resource/test update mismatch in current local branch.
- Any migration, payment/subscription/redeem/channel-monitor/DingTalk/user quota feature.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test ./internal/service -run "FailoverReparsesCachedBody|GetOpenAIRequestBodyMap" -count=1
go test ./internal/service -run "OpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|GetOpenAIRequestBodyMap" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-openai-failover-body-remap-s2h-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-openai-failover-body-remap-s2h-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval and QA events.
- Update `knowledge/tasks/current-task.md` with the current handoff snapshot after QA.

## Stop Rules
- Stop if implementation requires touching denied paths.
- Stop if test failures require gateway handler, schema, config, frontend, WS bridge, Responses bridge, or routing redesign changes.
- Stop if reparsing the original body would break handler-side validation cache behavior outside service failover/body remapping.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`

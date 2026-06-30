---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-01 01:50 +08:00
---

# Task Contract

## Task ID
upstream-main-v0141-antigravity-system-role-s34

## Role
Codex acts as Planner, Generator, QA, and Final Evaluator for this small upstream port. No external worker is used.

## Goal
Port upstream `65559ac58993c5eb42eb14d9f889ec76f2f44c8e` so Antigravity Claude-to-Gemini conversion moves `messages[].role == "system"` content into Gemini `systemInstruction` instead of emitting an invalid `system` role in `contents`.

## Success Criteria
- `buildContents` returns normal Gemini contents plus message-level system parts separately.
- `TransformClaudeToGeminiWithOptions` appends message-level system parts after top-level system instructions.
- Existing assistant-to-model role mapping remains unchanged.
- Ordinary user/assistant conversations remain unchanged.
- A targeted regression test covers message-level system movement, top-level + message system merge order, assistant role mapping, and ordinary conversation stability.
- No frontend, Ent, migration, wire, deploy, README, VERSION, user proxy/account, payment, admin, or `knowledge/*` files are included.

## Context
- Repo: `F:/mcplugins/sub2api`
- Local anchor before S34: `aef339c07 fix(payment): support plural subscription validity units`
- Upstream anchor after refresh: `v0.1.141-1-gdc1bc1545`
- Upstream reference:
  - `65559ac58993c5eb42eb14d9f889ec76f2f44c8e fix(antigravity): merge system role messages`
- Screening notes:
  - `271aba1abe` IP-denied SLA exclusion: local equivalent already present.
  - `930326116` subscription payment display: local direct-plan-price behavior already present.
  - `0ae3329613`, `04deb819b0`, `1e2193c3d2`, `bf1a2d6dc2`: local equivalent already present.
  - `c40a74d98`, `55655b865`, `727ac3f68`, `f6e0ebc6b`, `bf3787de1`, `20f534078`: local equivalent already present.

## Allowed Paths
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `docs/workflow/tasks/upstream-main-v0141-antigravity-system-role-s34.md`
- `docs/workflow/worker-results/upstream-main-v0141-antigravity-system-role-s34-result.md`
- `docs/workflow/qa-reports/upstream-main-v0141-antigravity-system-role-s34-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `knowledge/**`
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/server/**`
- `backend/internal/service/**`
- Payment, admin credentials redaction, Ops classification, image billing metadata, Gemini chat-completions routing, channel-monitor, email notification, user proxy/account ownership work, production configuration paths, README, deploy, Dockerfile, and assets.

## Constraints
- Do not merge or rebase `upstream/main`.
- Do not change Antigravity tools, thinking, identity patch, MCP XML, billing header, or web-search fallback behavior outside message-role handling.
- Do not stage existing dirty Ent, migration, proxy/account, frontend, service, handler, or knowledge files unrelated to this task.
- Keep the port compatible with the current local Antigravity transformer API and tests.

## Acceptance Commands
```powershell
cd F:/mcplugins/sub2api/backend
go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_MessageRoles|TestTransformClaudeToGeminiWithOptions_PreservesBillingHeaderSystemBlock" -count=1
cd F:/mcplugins/sub2api
git diff --check -- backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go docs/workflow/tasks/upstream-main-v0141-antigravity-system-role-s34.md docs/workflow/worker-results/upstream-main-v0141-antigravity-system-role-s34-result.md docs/workflow/qa-reports/upstream-main-v0141-antigravity-system-role-s34-qa.md docs/workflow/status.md docs/workflow/main-log.md
```

## Output
- Write worker result to `docs/workflow/worker-results/upstream-main-v0141-antigravity-system-role-s34-result.md`.
- Write QA report to `docs/workflow/qa-reports/upstream-main-v0141-antigravity-system-role-s34-qa.md`.
- Update `docs/workflow/status.md` and append `docs/workflow/main-log.md`.

## Stop Rules
- Stop if implementation requires frontend, Ent, migration, wire, service, route, repository, broader Antigravity gateway routing, user proxy/account work, or `knowledge/*`.
- Stop if tests cannot compile because of this patch's allowed paths.

## Budget
- worker_mode: `local-codex`
- qa_worker_mode: `local-codex`
- max_budget_usd: `0`

---
task_id: upstream-v0168-small-fixes-s125
status: contract-approved
role: Generator
qa_mode: runtime
---

# Task Contract

## Goal

Manually adapt five independently reviewable upstream reliability and usability
behaviors to the local topology: ChatGPT Live store-outage finalization,
OpenAI passthrough model-support short-circuiting, Caddy non-SSE compression,
Grok manual-test HTTP 402 cooldown, and model-ID copying.

## Success Criteria

- A Live observer does not silently abandon usage finalization or lease release
  when the Live store temporarily fails; bounded retries and expiry fallback
  remain idempotent.
- OpenAI passthrough accounts accept requested models even when credentials
  retain a stale non-empty `model_mapping`; non-passthrough allowlists keep
  their existing behavior.
- Caddy continues compressing the approved text/application/image types without
  matching `text/event-stream`, while reverse proxy routing and cache ownership
  remain unchanged.
- A Grok account manual connectivity test that receives HTTP 402 persists the
  same 30-minute `grok payment required` cooldown used by real forwarding.
- Copying a model ID from the whitelist selector writes only to the clipboard;
  selecting and deselecting a model retain their existing behavior.

## Context

- Repo: `E:/codex-worktrees/sub2api/upstream-v0168-small-fixes-s125`
- Upstream references: `1c26dc7ad`, `83b368553`, `c81191b46`,
  `e46d55bc5`, `2db0cbd29`, and `d8ae153ae`.
- The primary `F:/mcplugins/sub2api` worktree is user-dirty and must remain
  unchanged until the isolated branch passes all source-level gates.

## Allowed Paths

- `backend/internal/service/openai_live.go`
- `backend/internal/service/openai_live_lifecycle_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_model_passthrough_s125_test.go`
- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_grok_s125_test.go`
- `deploy/Caddyfile`
- `deploy/test-caddyfile-cache.sh`
- `frontend/src/components/account/ModelWhitelistSelector.vue`
- `frontend/src/components/account/__tests__/ModelWhitelistSelector.spec.ts`
- `docs/workflow/**`

## Denied Paths

- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/domain/constants.go`
- `backend/internal/pkg/claude/constants.go`
- `backend/internal/service/billing_service.go`
- `backend/internal/service/bedrock_request.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/pricing_service.go`
- `backend/resources/model-pricing/**`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/components/account/AccountStatusIndicator.vue`
- `frontend/src/views/admin/group-buy/**`
- `frontend/src/views/public/TutorialView.vue`
- `frontend/src/views/public/__tests__/TutorialView.spec.ts`
- `knowledge/**`
- `outputs/**`
- `Dockerfile*`

## Constraints

- Adapt behavior without merging or cherry-picking upstream history. Preserve
  local split files, helpers, scheduler policy, and existing error semantics.
- Live fallback must not shorten a valid session, duplicate usage rows, or
  release another controller's active lease before `ExpiresAt`.
- The passthrough short circuit applies only when
  `IsOpenAIPassthroughEnabled()` is true; ordinary OAuth and API-key mappings
  remain allowlists.
- Caddy must keep exactly one explicit encode block, must not configure
  `flush_interval`, and must not add imported policy outside the checked file.
- Grok HTTP 402 persistence is best effort and must use a bounded detached
  context so client cancellation cannot discard the cooldown. Other
  manual-test status codes remain unchanged.
- The copy control must be keyboard accessible and must not trigger the model
  selection action.
- Frontend dependencies may be exposed through a temporary directory junction
  to the primary worktree's existing `frontend/node_modules`; remove the
  junction after QA. Do not modify package manifests or lockfiles.
- Do not push, deploy, update containers, clean the primary worktree, or modify
  unrelated branches/worktrees.

## Acceptance Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v0168-small-fixes-s125/backend
go test ./internal/service -run "Test(WaitForLiveObserverRetry|ObserveLiveCallStoreOutage|FinalizeLiveCallUsageLog|FinalizeLiveCallIsIdempotent|IsModelSupported_OpenAIPassthroughIgnoresLeftoverMapping|AccountTestServiceGrokOAuthPaymentRequired)" -count=1
go test ./internal/service -run "Test(AccountIsModelSupported|Account_IsOpenAIPassthroughEnabled|HandleGrokAccountUpstreamErrorPaymentRequired|Live)" -count=1
go test ./... -run "^$"
cd ..
D:/Git/bin/bash.exe deploy/test-caddyfile-cache.sh
cd frontend
corepack.cmd pnpm exec vitest run src/components/account/__tests__/ModelWhitelistSelector.spec.ts
corepack.cmd pnpm run typecheck
corepack.cmd pnpm run build
cd ..
git diff --check
```

## Output

- A focused implementation, worker-style result, S125 QA report, and local
  branch commits only after all source-level acceptance gates pass.
- No external worker is invoked; Codex performs implementation, QA, and final
  evaluation directly.

## Stop Rules

- Stop if any slice requires a migration, billing change, account scheduler
  rewrite, Caddy deployment, package update, or denied path.
- Stop if Live outage handling cannot retain idempotent finalization or if the
  Grok test path cannot reuse the existing cooldown semantics.
- Stop before primary-worktree integration while any S125 path overlaps an
  uncommitted user change or if a validation failure cannot be attributed.

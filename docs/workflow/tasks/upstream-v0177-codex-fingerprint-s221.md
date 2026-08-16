---
task_id: upstream-v0177-codex-fingerprint-s221
phase: draft-pending-s220
role: Generator
worker_model: gpt-5.6-terra
qa_worker_model: gpt-5.6-terra
---

# Upstream v0.1.177 Codex Fingerprint Convergence S221

## Task ID

upstream-v0177-codex-fingerprint-s221

## Role

You are the independent `gpt-5.6-terra` Generator worker. Execute only the
approved contract after S220 integration; do not expand scope.

## Goal

Behaviorally port Codex OAuth fingerprint convergence from upstream
`c0ab3a00e` plus the remaining fingerprint/default/passthrough behavior of
`fce41e318`. Complete the deferred EditAccountModal control for the S220
OpenAI long-context account veto, preserve the already integrated S219 HTTP
turn-state behavior, and preserve the user's existing account-modal layout
patch.

## Success Criteria

- Missing, empty, invalid, or explicit `off` fingerprint mode preserves client
  identifiers. `device`, `session`, and `full` remain explicit administrator
  opt-ins with the upstream header/body convergence semantics.
- Each attempt resolves IDs once and stages them unconditionally, including
  nil. Failover from a converged account to an off/non-OAuth account cannot
  reuse the previous account's IDs.
- Normal and OpenAI passthrough paths apply staged headers after local session
  isolation and before identity enforcement. Existing S219 turn-state echo
  guarding remains independent and keeps its established ordering.
- Passthrough rewrites `client_metadata` on raw JSON with gjson/sjson semantics,
  preserves unrelated and multi-MB body fields, does not fully unmarshal the
  hot path, and matches the map variant.
- Header and body carriers share installation/session/thread/turn/window IDs;
  `off` remains a no-op except for existing local session isolation.
- Create, edit, and bulk account controls expose the four modes only for OpenAI
  OAuth accounts. `off` deletes the extra key; every opt-in mode persists it.
- EditAccountModal also exposes and persists the S220
  `openai_long_context_billing_enabled` boolean for eligible OpenAI accounts,
  retaining the upstream default-off and Spark-shadow exclusion behavior.
- The final `EditAccountModal.vue` keeps the user-owned `extra-wide` dialog and
  asymmetric group/availability grid. Its existing test remains valid while
  adding fingerprint assertions.

## Context

- Repo: `F:/mcplugins/sub2api`.
- Product base: S220-integrated `main` approval commit supplied at dispatch.
- Upstream: `upstream/main@baeac1f3de21d37b129405f092ef86c24b3f203d`.
- Source commits: `c0ab3a00ea733cc0559a5a949c28fb5d9d7c5d16` and
  `fce41e318f147cca42d04944754ed4e693714ddf`.
- S219 already integrated the turn-state relay/guard hook ideas from
  `fce41e318`; do not duplicate or replace that implementation.
- User baseline patch ID:
  `5d316e5b6935fdc5dbf825f940feaf231d79ac0f`. The dispatch worktree must be
  based on a temporary local baseline commit containing exactly that two-file
  patch. Worker commits must contain only S221 changes above that baseline.

## Allowed Paths

- `backend/internal/service/openai_codex_fingerprint.go`
- `backend/internal/service/openai_codex_fingerprint_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/openai_tool_namespace_normalization_s92_test.go`
- `frontend/src/components/account/BulkEditAccountModal.vue`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/BulkEditAccountModal.spec.ts`
- `frontend/src/components/account/__tests__/CreateAccountModal.timeAvailability.spec.ts`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/worker-results/upstream-v0177-codex-fingerprint-s221-result.md`

## Denied Paths

- Migration/schema/Ent changes, group pricing, group daily rollups, dependencies,
  CI/release workflows, VERSION, providers, shared databases, production data,
  containers, deployment, push, and `outputs/**`.
- S219 turn-state storage/provenance redesign, WS protocol changes, unrelated
  account options, and all files outside Allowed Paths.
- `docs/workflow/status.md`, `docs/workflow/main-log.md`, QA reports,
  `knowledge/**`, and global memories.

## Constraints

- Work only in the isolated S221 worktree after contract approval. Do not amend,
  squash, or include the temporary user-baseline commit in worker commits.
- Adapt upstream split `openai_gateway_forward.go` and
  `openai_gateway_passthrough.go` behavior into local `gateway_service.go`.
- Preserve the exact user baseline patch and test meaning. S221 additions must
  be extractable as a patch relative to that baseline and applicable to the
  clean main index independently from the still-uncommitted user patch.
- Preserve newer local remote-compaction, turn-state, identity enforcement,
  session isolation, Responses capability, and billing behavior.

## Acceptance Commands

```powershell
Set-Location E:/codex-worktrees/sub2api/upstream-v0177-codex-fingerprint-s221/backend
$focused = '^(' + (@(
  'TestGetCodexFingerprintMode',
  'TestResolveCodexFingerprintIDsFromRequest_DefaultIsOff',
  'TestResolveCodexFingerprintIDsFromRequest_ExplicitOptInHonored',
  'TestFingerprintIDs_HeaderAndBody_TurnID_Consistent',
  'TestApplyCodexFingerprintClientMetadataRaw_MatchesMapVariant',
  'TestApplyCodexFingerprintClientMetadataRaw_PreservesUnrelatedFields',
  'TestStageCodexFingerprintIDs_NilOverwritesPreviousAccount',
  'TestBuildUpstreamRequestOpenAIPassthrough_AppliesStagedFingerprint',
  'TestBuildUpstreamRequestOpenAIPassthrough_OffModeKeepsIsolatedSession'
) -join '|') + ')$'
go test ./internal/service -run $focused -count=10
if ($LASTEXITCODE -ne 0) { throw 'S221 focused backend failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S221 complete service failed' }
go test ./internal/handler -count=1
if ($LASTEXITCODE -ne 0) { throw 'S221 complete handler failed' }
go test ./internal/server -count=1
if ($LASTEXITCODE -ne 0) { throw 'S221 complete server failed' }
go test ./cmd/server -run '^$' -count=0
if ($LASTEXITCODE -ne 0) { throw 'S221 server compile failed' }

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-codex-fingerprint-s221/frontend
pnpm.cmd exec vitest run src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/account/__tests__/CreateAccountModal.timeAvailability.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts
if ($LASTEXITCODE -ne 0) { throw 'S221 modal regressions failed' }
pnpm.cmd run typecheck
if ($LASTEXITCODE -ne 0) { throw 'S221 frontend typecheck failed' }
pnpm.cmd run build
if ($LASTEXITCODE -ne 0) { throw 'S221 frontend build failed' }

Set-Location E:/codex-worktrees/sub2api/upstream-v0177-codex-fingerprint-s221
git diff --check
if ((git diff --name-only --diff-filter=U) -or (git ls-files -u)) {
  throw 'S221 conflict or unmerged index found'
}
foreach ($commit in @('c0ab3a00e','fce41e318')) {
  git merge-base --is-ancestor $commit upstream/main
  if ($LASTEXITCODE -ne 0) { throw "missing upstream provenance: $commit" }
}
```

## Output

- Write
  `docs/workflow/worker-results/upstream-v0177-codex-fingerprint-s221-result.md`
  with the required first-line verdict.
- Commit only S221 deltas above the supplied temporary user-baseline commit.
  Report changed files, real commands, user-patch preservation evidence, risks,
  and contract compliance.

## Stop Rules

- Stop if the user patch is absent/different, if S219 ordering cannot be
  preserved, or if migration/dependency/provider/shared-runtime work is needed.
- Stop if an S221 commit includes the temporary user baseline or if the S221
  patch cannot be separated from it cleanly.
- Stop after two failed implementation rounds; do not integrate, push, deploy,
  update containers, or clean worktrees/branches.

## Contract Review

Pending S220 integration and Evaluator review.

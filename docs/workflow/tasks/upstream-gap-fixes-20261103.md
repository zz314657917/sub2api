---
type: task-contract
status: approved
review_verdict: PASS
task_id: upstream-gap-fixes-20261103
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 7ea675c020ff2c5d5e4124ac5e249c9dc65034c5
spec_ref: docs/workflow/spec.md
last_verified: 2026-11-03
---

# Task Contract

## Task ID
upstream-gap-fixes-20261103

## Role
Planner/Final Evaluator=current controller. Independent Contract Reviewer, Developer and QA use the repository matrix model gpt-6.1-sol. Approval metadata is effective only together with an independent PASS report binding the exact final contract SHA256. Developer and QA are separate invocations.

## Goal
Adapt six verified upstream fixes to the local architecture: deleted upstream account settlement, nonstream image counting, admin subscription extension serialization, future-reset quota pause, account long-context cost gate and API key quota-reset UI synchronization.

## Context
- Main repository: F:/mcplugins/sub2api.
- Read-only audit: E:/codex-runtime/sub2api/upstream-gap-audit-20261103/gap-report.md; original patches and probes are adjacent.
- Runtime/evidence root: E:/codex-runtime/sub2api/upstream-gap-fixes-20261103.
- Pin local base_commit above and upstream 4a4a83ee5ec86b19d3a2dd9b6471de5128963bec.
- Original tracked file hashes, dirty status and index are in runtime/main-before.json. Preserve concurrent Pelican frontend changes and all unrelated documentation.
- Developer uses E:/codex-worktrees/sub2api/upstream-gap-fixes-20261103; QA uses a separately frozen candidate copy. Native independent-agent dispatch is used because the configured external CLI gateway rejects its existing credential group; the matrix model stays gpt-6.1-sol.
- Six references: d6120b5d8, 1487fa511, e3cce574d, 02088be2d, fd7af1d46, a65d476f2.

## Success Criteria
1. Deleting a quota-limited APIKey/Bedrock upstream account in flight must not roll back user balance/subscription settlement. Skip only ErrAccountNotFound in its account counter; real SQL errors still roll back. Retain already-fixed client API Key deletion behavior, idempotency and share-ledger order. Cover balance, subscription, active account, missing account, real database error and duplicate settlement.
2. Nonstream image count must not fall back to requested n. Use actual returned images and retain the existing local single-image fallback where applicable; do not claim or introduce a new zero-charge policy. Cover request n=4 with empty/one/two returned images; retain streaming and gpt-image-2 splitting behavior.
3. Admin ExtendSubscription must read expiry with GetByIDForUpdate inside the existing withSubscriptionUpdateTx and update expiry/status in the same transaction. Commit before cache invalidation and final reread. Two concurrent +5 adjustments must accumulate +10. Preserve bounds, expired-negative rejection, errors, existing redeem/assignment locks and no-client test compatibility.
4. Stale quota usage must keep the account paused until a valid known future absolute or relative reset. Preserve self-healing for stale snapshots with missing/invalid/past reset and existing missing/invalid update-time behavior. Absolute reset takes precedence; equality means reset reached. Retain both 5h and 7d windows and do not extend a relative reset by re-anchoring it on every check.
5. OpenAI account-cost statistics must honor the account long-context gate, independently of the group selling-price gate. Preserve model-specific OAuth exemption, custom-rule precedence, ApplyPricingToAccountStats precedence, normal token/cache/image pricing and other-platform callers. APIKey account-off example should change account stats 1.545 to .78 without changing customer charge .78; account-on remains 1.545.
6. ResetQuotaUsed must capture the request target before await and apply returned quota_used/status to that target. Update form status only when the current selected key still has the target ID. Preserve server-authoritative values, error handling and quota/rate-limit separation. Mounted Vue tests cover exhausted->active, nonzero returned quota, selection change during request, cleared selection, inactive response, failure without mutation and subsequent edit save using the new status.
7. Minimal diff, no unrelated formatting or dependency changes. Existing assertions cannot be weakened to obtain PASS. New tests should expose broken behavior before the fix or be independently baseline-checked.

## Allowed Paths
- `backend/internal/repository/usage_billing_repo.go`
- `backend/internal/repository/usage_billing_deleted_account_test.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/subscription_service.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/account_stats_pricing.go`
- `backend/internal/service/upstream_gap_fixes_test.go`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/__tests__/KeysView.quotaReset.spec.ts`
- `docs/workflow/worker-results/upstream-gap-fixes-20261103-result.md`

Controller only: task contract, independent contract review, task QA report, own bounded status/spec/current-task additions and external runtime tools/evidence. Developer must not edit controller-owned paths.

## Denied Paths
- `knowledge/**`
- `backend/migrations/**`
- All paths outside the allowlist and global memories/**.
- Pricing resolver/billing-service image-price semantics (a2007db70); this seventh audit item is a separate compatibility batch.
- Dependency declarations/locks, database schema/migrations, deployment/configuration/secrets, frontend Pelican files.
- Inflight-balance reservation, SQL performance refactors, Vertex feature expansion, unrelated failover/protocol work.
- git add/commit/push/merge/reset/clean and production/container operations.

## Constraints
- Use existing local interfaces; do not import upstream's broader architecture or unrelated features.
- Do not change shared main source. Work only in the isolated candidate.
- No dependency installation or network provider/API/DB access. Use installed tools and local test fixtures.
- Go: E:/tools/go1.26.10/go/bin/go.exe with process-local GOROOT=E:/tools/go1.26.10/go and GOTOOLCHAIN=local; use -mod=readonly. Do not change shell-wide environment.
- Frontend dependencies may be linked to the verified installed main frontend/node_modules; do not change its bytes. Build output must go under runtime, never backend embedded assets.
- Log real exit codes. Do not call expected baseline failures a passing full suite.

## Acceptance Commands
Run from backend with the fixed toolchain:
```text
go test -mod=readonly -tags unit ./internal/service ./internal/repository -run 'Test(GapFix|UsageBillingRepositoryApply_(DeletedAPIKey|ActiveAPIKey|APIKeyDatabase)|AssignOrExtendSubscription|OpenAI.*(Quota|LongContext)|AccountStats)' -count=1
go test -mod=readonly ./internal/service -run 'Test(GapFix|AssignOrExtendSubscription|OpenAI.*(Quota|LongContext)|AccountStats)' -count=1
go build -mod=readonly ./...
```
Also run affected existing image, account-stat, subscription and protocol tests after identifying their exact names. Keep preexisting unrelated failures attributable.
Run from frontend with installed dependency executables:
```text
vitest run src/views/user/__tests__/KeysView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/views/user/__tests__/KeysView.quotaReset.spec.ts
vue-tsc --noEmit
eslint src/views/user/KeysView.vue src/views/user/__tests__/KeysView.quotaReset.spec.ts
vite build --outDir E:/codex-runtime/sub2api/upstream-gap-fixes-20261103/frontend-dist
```
Use native node.exe with each package's installed CLI if necessary. Run gofmt on changed Go files and scoped git diff --check.

## Output
- Developer report: docs/workflow/worker-results/upstream-gap-fixes-20261103-result.md, first line `### DONE: upstream-gap-fixes-20261103`, `### FAILED: ...` or `### BLOCKED: ...`.
- List changed files, actual commands/logs and exit codes, regressions, risks, contract compliance and knowledge_candidates.
- Independent QA report starts `### PASS: upstream-gap-fixes-20261103`, `### FAIL: ...` or `### BLOCKED: ...`, binds all candidate source SHA256 and actual evidence hashes.
- Controller integrates exact approved source only after independent QA and local review; verify target baseline and preserve every other original tracked-file byte, HEAD and index. Concurrent drift requires exact reconciliation or stopping before copy.
- Final controller report distinguishes local verification from live database/provider/browser acceptance.

## Stop Rules
- No independent contract PASS on the final contract hash: do not implement.
- Missing model, permission/toolchain inability, target drift or needed denied-path edit: report precise BLOCKED, do not improvise broader changes.
- Two failed Developer iterations: stop the worker loop and return to controller diagnosis.
- No secrets, global process killing, production access, publishing or deployment.

## Budget
- worker_mode: native-independent-agent
- qa_worker_mode: independent-qa
- worker_model: gpt-6.1-sol
- qa_worker_model: gpt-6.1-sol
- max_budget_usd: 15.00
- worktree_root: E:/codex-worktrees

## Worker Output
Same as Output above.

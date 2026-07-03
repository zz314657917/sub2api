---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 13:14 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-group-peak-rate-impl-s44

## Role
Codex acts as Planner and Final Evaluator. Implementation and fixes are performed in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `v0.1.143` subscription group peak-rate capability into the local fork without overwriting local payment, welfare, Studio Bridge, image, or modular i18n customizations.

## Success Criteria
- Groups support `peak_rate_enabled`, `peak_start`, `peak_end`, and `peak_rate_multiplier`.
- Peak rate is valid only for subscription groups; non-subscription groups clear peak config through the service Normalize -> Validate path.
- `peak_rate_multiplier=0` is accepted.
- Membership/user/group token multiplier is applied before peak multiplier.
- Peak multiplier affects token billing only, including token-mode image output tokens.
- Image/per-request billing keeps image multiplier behavior and is not peak-multiplied.
- API key auth cache snapshots include peak fields and are versioned.
- Frontend admin/user/payment/channel views display peak windows with the server UTC offset and keep local modular i18n.
- Migrations use local `181_add_group_peak_rate_multiplier.sql`; do not introduce upstream `158` or touch local voucher migration numbering.

## Allowed Paths
- `backend/ent/**`
- `backend/internal/handler/admin/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `backend/internal/service/**`
- `backend/migrations/181_add_group_peak_rate_multiplier.sql`
- `frontend/src/components/channels/**`
- `frontend/src/components/common/**`
- `frontend/src/components/payment/**`
- `frontend/src/i18n/locales/**`
- `frontend/src/types/**`
- `frontend/src/utils/**`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/user/AvailableChannelsView.vue`
- `frontend/src/views/user/KeysView.vue`
- `frontend/src/views/user/PaymentView.vue`
- `frontend/src/views/user/SubscriptionsView.vue`
- `docs/workflow/tasks/upstream-main-v0143-group-peak-rate-impl-s44.md`
- `docs/workflow/worker-results/upstream-main-v0143-group-peak-rate-impl-s44-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-group-peak-rate-impl-s44-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/internal/payment/**`
- `backend/internal/service/welfare_*`
- `backend/internal/repository/welfare_*`
- `backend/migrations/158_add_group_peak_rate_multiplier.sql`
- `backend/migrations/179_welfare_vouchers.sql`
- `backend/cmd/server/wire_gen.go`
- `deploy/**`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`
- `frontend/src/views/studio/**`
- `frontend/src/views/user/WelfareView.vue`
- `frontend/src/views/user/OpenAIImagesView.vue`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of upstream `v0.1.143` or `upstream/main`.
- Keep this in the isolated worktree until staged diff and QA pass.
- Do not use `git add .`.
- Preserve local modular i18n and payment display conventions.
- Do not overwrite user changes in the main worktree.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "Test.*Peak.*|Test.*Group.*Peak.*|Test.*Billing.*Peak.*|Test.*Gateway.*Peak.*|Test.*RecordUsage.*Peak.*" -count=1
go test ./internal/handler -run "Test.*AvailableChannel.*Peak.*|Test.*Payment.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1
go test ./internal/handler/admin -run "Test.*Group.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1

cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"
cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/utils/apiKeyCapabilities.spec.ts"
git diff --check
git diff --cached --name-only | rg "^(backend/internal/payment/|backend/internal/service/welfare_|backend/internal/repository/welfare_|backend/migrations/(158|179)_|backend/cmd/server/wire_gen.go|deploy/|frontend/src/i18n/locales/(en|zh)\\.ts|frontend/src/views/studio/|frontend/src/views/user/(WelfareView|OpenAIImagesView)\\.vue|knowledge/|\\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-group-peak-rate-impl-s44`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-group-peak-rate-impl-s44-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-group-peak-rate-impl-s44-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if staged diff includes denied paths.
- Stop if token-mode image output tokens are billed or logged with image multiplier instead of token peak multiplier.
- Stop if frontend typecheck fails after peak fields are added.
- Stop if workflow evidence cannot be separated from unrelated main worktree dirty files.

## Review Result
- Reviewed at: 2026-07-03 13:14 +08:00.
- Verdict: approved.
- Reason: this contract scopes the upstream peak-rate capability, explicitly protects local custom surfaces, and defines runtime/backend/frontend acceptance checks.

### DONE: upstream-main-v0143-group-peak-rate-impl-s44

## Summary
- Ported subscription group peak-rate fields through Ent schema, migration, repository, service, auth cache, admin handler, payment, available channels, and frontend display surfaces.
- Preserved local migration numbering with `backend/migrations/181_add_group_peak_rate_multiplier.sql`.
- Kept local modular i18n by placing peak-rate strings in `frontend/src/i18n/locales/en/common.ts` and `frontend/src/i18n/locales/zh/common.ts`.
- Fixed post-review issues found by subagents:
  - token-mode image output tokens now keep token billing mode and token peak multiplier in usage logs.
  - generic and OpenAI gateway cost selection no longer treat every `ImageCount > 0` request as image/per-request billing when channel pricing is token mode.
  - admin group create handler no longer pre-validates peak fields before service normalization.
  - Keys smart-routing summaries now display route group badges, including peak windows.
  - frontend tests/fixtures and admin type state were synchronized with peak fields.

## Changed Files
- Backend service, handler, repository, Ent, and migration files for group peak-rate support.
- Frontend group/payment/channel/key views, shared peak-rate display utility, and modular i18n fragments.
- Focused backend and frontend tests.
- Workflow evidence files for S44.

## Commands Run
- `go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/upsert,intercept,sql/execquery,sql/lock --idtype int64 ./schema`
  - Result: PASS.
- `go test ./internal/service -run "Test.*Peak.*|Test.*Group.*Peak.*|Test.*Billing.*Peak.*|Test.*Gateway.*Peak.*|Test.*RecordUsage.*Peak.*" -count=1`
  - Result: PASS.
- `go test ./internal/handler -run "Test.*AvailableChannel.*Peak.*|Test.*Payment.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - Result: PASS, no matching non-admin handler tests for the peak patterns.
- `go test ./internal/handler/admin -run "Test.*Group.*Peak.*|TestGroupHandlerCreate_LeavesPeakRateNormalizationToService|TestGroupHandlerEndpoints" -count=1`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend install --frozen-lockfile"`
  - Result: PASS; installed ignored `frontend/node_modules` for isolated worktree verification.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend run typecheck"`
  - Result: PASS.
- `cmd.exe /d /s /c "corepack.cmd pnpm --dir frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/KeysView.createQuery.spec.ts src/utils/apiKeyCapabilities.spec.ts"`
  - Result: PASS, 4 files / 37 tests.
- `git diff --check`
  - Result: PASS.

## Contract Compliance
- Did not merge all upstream release content.
- Did not introduce upstream migration `158` or local voucher migration `179`.
- Did not touch payment internals, welfare service/repository, deploy, `.github`, README, knowledge, Studio views, or OpenAI Images view.
- Kept the implementation in the isolated worktree and did not modify the main worktree.

## Risks
- Full backend/frontend suites were not run; verification targeted S44 paths and subagent findings.
- Real browser visual smoke was not run.
- `frontend/node_modules` remains ignored in the isolated worktree after verification.

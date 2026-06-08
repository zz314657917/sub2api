### DONE: upstream-main-ops-alert-temp-unscheduled-s14

## Summary

- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-ops-alert-temp-unscheduled-s14`.
- Created branch `codex/upstream-main-ops-alert-temp-unscheduled-s14` from local `main@074dc565a`.
- Ported the approved Ops alert metric candidate without directly merging `upstream/main`.
- Kept changes inside approved Ops alert backend, frontend metric UI/type, modular i18n, and workflow paths.

## Candidate Results

- `f20e6bf76`: `CHERRY_PICKED` as `0fb09933c`.
  - Backend alert rule validation now accepts `account_temp_unscheduled_count`.
  - Ops alert evaluator counts accounts whose `TempUnschedulableUntil` is currently active.
  - Frontend admin Ops alert rule editor exposes the metric with modular en/zh locale strings.

## Local Adaptations

- Upstream modified monolithic `frontend/src/i18n/locales/en.ts` and `frontend/src/i18n/locales/zh.ts`.
- Local frontend uses modular locale files, so those root aggregators were preserved and the new strings were placed in:
  - `frontend/src/i18n/locales/en/admin/ops.ts`
  - `frontend/src/i18n/locales/zh/admin/ops.ts`

## Deferred / Skipped

- `af19d4432`: `DEFERRED`. Proxy expiry/fallback requires schema, migration, frontend, and API contract work.
- README/sponsors/VERSION/docs-only upstream commits: `SKIPPED`.

## Commits

- `0fb09933c` feat(ops): 新增 account_temp_unscheduled_count 告警指标

## Changed Files

- `backend/internal/handler/admin/ops_alerts_handler.go`
- `backend/internal/service/ops_alert_evaluator_service.go`
- `backend/internal/service/ops_alert_evaluator_service_test.go`
- `frontend/src/api/admin/ops.ts`
- `frontend/src/i18n/locales/en/admin/ops.ts`
- `frontend/src/i18n/locales/zh/admin/ops.ts`
- `frontend/src/views/admin/ops/components/OpsAlertRulesCard.vue`
- `docs/workflow/tasks/upstream-main-ops-alert-temp-unscheduled-s14.md`
- `docs/workflow/worker-results/upstream-main-ops-alert-temp-unscheduled-s14-result.md`
- `docs/workflow/qa-reports/upstream-main-ops-alert-temp-unscheduled-s14-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check main...HEAD`
- denied path audit against `main...HEAD`
- `go test ./internal/service -run "OpsAlert|TempUnscheduled|AccountTemp|RuleMetric" -count=1` -> PASS but did not run `unit`-tagged tests; rerun below.
- `go test -tags unit ./internal/service -run "ComputeRuleMetric|TempUnscheduled|OpsAlert" -count=1`
- `go test ./internal/handler/admin -run "OpsAlert|Metric" -count=1`
- `go test ./internal/service ./internal/handler/admin -count=1`
- `corepack.cmd pnpm --dir frontend install --frozen-lockfile`
- `corepack.cmd pnpm --dir frontend run typecheck`

## Notes

- No `backend/ent/`, `backend/migrations/`, `.github/`, `deploy/`, `assets/`, `README*`, `skills/`, `knowledge/`, `docs/workflow/status.md`, or `docs/workflow/spec.md` changes were made.
- The first frontend typecheck attempt failed because the isolated worktree had no `frontend/node_modules`; frozen-lockfile install fixed the environment and the rerun passed.
- Local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, and workflow docs were preserved.

### DONE: upstream-main-usage-window-s13

## Summary

- Created branch `codex/upstream-main-usage-window-s13` from local `main@b905c03a2`.
- Ported the approved 5h usage window candidate without directly merging `upstream/main`.
- Kept changes inside approved backend usage/account, frontend usage progress, i18n, and workflow paths.

## Candidate Results

- `16bc87693`: `CHERRY_PICKED` as `8a0d0ed63`. Active usage poll now writes 5h `ResetsAt` into `SessionWindowEnd`, expired windows are zeroed during estimate display, and `UsageProgressBar` distinguishes "Now" from "Pending refresh".

## Local Adaptations

- Upstream modified monolithic `frontend/src/i18n/locales/en.ts` and `zh.ts`; local frontend uses modular locale files.
- Preserved local `en.ts` / `zh.ts` module aggregator structure and placed new strings in:
  - `frontend/src/i18n/locales/en/usage.ts`
  - `frontend/src/i18n/locales/zh/usage.ts`
- Updated the Sprint contract allowed paths to include those modular i18n files.

## Deferred / Skipped

- `f20e6bf76`: `DEFERRED`. `account_temp_unscheduled_count` alert metric remains a separate ops Sprint candidate.
- `f5cecea5b`: `DEFERRED`. Pure frontend Select dropdown height fix should stay separate.
- `af19d4432`: `DEFERRED`. Proxy expiry/fallback feature requires schema, migration, frontend, and API contract work.
- README/VERSION-only upstream commits: `SKIPPED`.

## Commits

- `fbc6bdd29` docs: add usage window s13 contract
- `8a0d0ed63` fix(usage): sync 5h ResetsAt to SessionWindowEnd and zero expired window
- `aa73565c2` docs: update usage window s13 allowed paths

## Changed Files

- `backend/internal/repository/account_repo.go`
- `backend/internal/server/api_contract_test.go`
- `backend/internal/service/account_service.go`
- `backend/internal/service/account_service_delete_test.go`
- `backend/internal/service/account_usage_service.go`
- `backend/internal/service/account_usage_session_window_test.go`
- `backend/internal/service/gateway_multiplatform_test.go`
- `backend/internal/service/gemini_multiplatform_test.go`
- `backend/internal/service/ratelimit_session_window_test.go`
- `frontend/src/components/account/UsageProgressBar.vue`
- `frontend/src/components/account/__tests__/UsageProgressBar.spec.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `docs/workflow/tasks/upstream-main-usage-window-s13.md`
- `docs/workflow/worker-results/upstream-main-usage-window-s13-result.md`
- `docs/workflow/qa-reports/upstream-main-usage-window-s13-qa.md`
- `docs/workflow/main-log.md`

## Verification

- `git status --short --branch`
- `git diff --check main..HEAD`
- denied path audit against `main..HEAD`
- `go test ./internal/service -run "SessionWindow|Usage|ResetsAt|RateLimit|Gateway|Gemini|Delete" -count=1`
- `go test ./internal/repository ./internal/server -run "SessionWindow|Usage|Contract|Account" -count=1`
- `corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/UsageProgressBar.spec.ts`
- `go test ./internal/service ./internal/repository ./internal/server -count=1`
- `corepack.cmd pnpm --dir frontend run typecheck`

## Notes

- No `backend/ent/`, `backend/migrations/`, `skills/`, `deploy/`, `knowledge/`, `docs/workflow/status.md`, or `docs/workflow/spec.md` changes were made.
- Local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, and workflow docs were preserved.

---
status: completed
review_verdict: pending_qa
task_id: pelican-multi-model-20261007
worker_model: gpt-6.1-sol
---

# Developer Result: pelican-multi-model-20261007

## Changed files
- `backend/internal/service/pelican_test_service.go`: additive `model_ids` normalization/legacy alias, per-model validation, ordered serial execution, per-model timeout and release, model history routing, call estimation.
- `backend/internal/repository/pelican_test_repo.go`: model-aware plan persistence/scanning, lease budget scaling, model-separated latest/history/retention queries, group pagination, artwork generation timestamp.
- `backend/internal/handler/pelican_test_handler.go`: optional server-side `model_id` history filter validation/routing.
- `backend/migrations/251_pelican_multi_model.sql`: additive array storage, legacy backfill, trigger validation/first-model alias, focused index.
- `frontend/src/api/pelicanTests.ts`: `model_ids`, artwork timestamp and history filter types.
- `frontend/src/components/pelican/PelicanPlansPanel.vue`: ordered checkbox multi-select, edit roundtrip, estimate display and model list.
- `frontend/src/views/user/PelicanTestsView.vue`: group cards with vertically stacked model entries, ModelIcon, visible generation time/duration, model-aware history and group pagination layout.
- Existing Pelican frontend tests and PostgreSQL fixture migration lists were updated for the additive contract.

## Verification
- `GOMAXPROCS=2 GOMEMLIMIT=1GiB go test -p 1 ./internal/service ./internal/repository ./internal/handler -run Pelican -count=1` PASS.
- Disposable PostgreSQL: `PELICAN_TEST_POSTGRES_DSN` and `PELICAN_COST_TEST_DSN` against task-owned PostgreSQL 18.6; focused Pelican repository integration tests PASS.
- `vitest run ...PelicanTestsView.spec.ts ...PelicanPlansPanel.spec.ts --maxWorkers=1 --minWorkers=1` PASS: 25 tests.
- `git diff --check` PASS.

## Limitations / risks
- Full frontend production build/typecheck was not completed because pnpm attempted a noninteractive dependency rebuild and reported ignored `esbuild`/`vue-demi` build scripts; direct focused Vitest passed.
- Real provider calls, production database, deployment and browser runtime acceptance were not performed.
- No commit, push, container replacement or paid provider call was performed.

## Canonical test follow-up (2026-10-07)

- Extracted the independently validated QA overlay into `backend/internal/service/pelican_multi_model_test.go` and `backend/internal/repository/pelican_multi_model_postgres_test.go`, with `gofmt` formatting and task-specific test/helper names. PostgreSQL tests explicitly skip when `PELICAN_TEST_POSTGRES_DSN` is absent.
- Added interaction coverage only to the existing `PelicanPlansPanel.spec.ts` and `PelicanTestsView.spec.ts`: selection-order create/edit roundtrip, deselection/replacement, clearing multiple selections on group switch with overlapping IDs, ten-model limit, same-group distinct-plan stacking, model-scoped history and dialog identity after refresh/model removal.
- `GOMAXPROCS=2 GOMEMLIMIT=1GiB go test -p 1 ./internal/service ./internal/repository -run TestPelicanMultiModel -count=1 -v`: exit 0; service input/run cases PASS; PostgreSQL case explicitly SKIP because DSN is absent. This follow-up did not restart PostgreSQL or rerun its live cases; the independent QA report records the equivalent overlay PostgreSQL PASS.
- `vitest run src/components/pelican/__tests__/PelicanPlansPanel.spec.ts src/components/pelican/__tests__/PelicanTestsView.spec.ts --maxWorkers=1 --minWorkers=1`: PASS, 2 files and 29 tests.
- Correction to the earlier limitation: direct `vue-tsc --noEmit` completed with exit 0 during the initial implementation. Independent QA subsequently recorded full typecheck/build and the original 43 frontend tests PASS.
- Browser acceptance remains BLOCKED with the outstanding operations documented in the independent QA report. Unit interaction tests do not replace those browser checks. No business source changes, service startup, commit, push or deployment occurred in this follow-up.

## Complete timestamp tooltip fix (2026-10-12)

- Addressed the sole remaining finding in `qa-reports/pelican-multi-model-20261012.md`: generation-time `title` now uses a separate timestamp formatter returning an ISO UTC timestamp with seconds/milliseconds and explicit `(UTC)`. Compact visible time still uses the existing formatter.
- Preserved artwork timestamp selection: failed-round fallback uses `artwork_finished_at`, not the newer failed attempt timestamp. Missing/invalid timestamps retain `时间未知`.
- Added a view interaction assertion covering successful artwork, old-success fallback with differing failure time, UTC conversion from `+08:00`, seconds, explicit timezone and the missing-time fallback.
- `NODE_OPTIONS=--max-old-space-size=4096 node node_modules/vitest/vitest.mjs run src/components/pelican/__tests__/PelicanTestsView.spec.ts --maxWorkers=1 --minWorkers=1`: PASS, 18 tests.
- Scoped ESLint for the view and its spec with existing `NODE_PATH=node_modules/.pnpm/node_modules`: exit 0.
- No browser, PostgreSQL or application service was started; no commit/push/deployment. Independent QA must perform the targeted tooltip recheck before changing its FAIL verdict.

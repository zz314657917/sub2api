---
status: approved
review_verdict: PASS
task_id: pelican-radar-20261030
worker_model: gpt-6.1-sol
base_commit: 424b96a6e9602d2922a1b8f524644ca2a427fa5b
spec_ref: docs/workflow/tasks/pelican-radar-20261030.md
---

# Task ID: pelican-radar-20261030

## Role

Planner/final evaluator: controller. Independent Contract Reviewer, Developer and QA: gpt-6.1-sol, separate dispatches. Implement only after independent contract PASS.

## Goal

Embed administrator-selected third-party Codex subscription reference scores above filters in the original `/pelican-tests` right content area. Reuse AppLayout/navigation/header and preserve the concurrent preview-only horizontal gallery.

## Source inspection and scoring contract

- Fixed upstream JSON: `https://codexradar.com/data/codex-four-task-results.json`; reference link `https://codexradar.com/#stationmaster-round` (verified source section ID).
- Captured live source and homepage in `E:/codex-runtime/sub2api/pelican-radar-20261030/source.json`, `source-home.html`, `source-ui.mjs`. Schema `codex-four-task-results/1`, protocol `codex-four-weighted-150-v2`.
- Top-level `score_max`, `task_weights`, `round_id`, `author_test_interval.started_at/completed_at`, `score_snapshot_at`, `models`, `model_names`, `model_efforts`, `efforts`, `rows`. Actual rows have model/effort plus four task objects; no precomputed total.
- Match exact canonical model and exact effort. Explicit aliases limited to the source's exact `model_names` display labels → canonical IDs; no stripping suffixes, fuzzy/case-insensitive matching, dated-model fallback, highest-score or highest-effort fallback.
- Total = `combined.score/100 * weights.combined + pompeii.score100/100 * weights.pompeii + pelican.score/100 * weights.pelican + (candy.correct ? 1 : 0) * weights.candy`. Use final `score`, never `raw_score_total` or reward. Homepage confirms four-task weighting and no zero-fill for missing parts.
- All weights finite positive and sum to source score_max (tolerance 1e-6); task scores finite within [0,100], correct strict bool. Missing/null task values yield `not_recorded` for that exact row, no total. Wrong non-null type/out-of-range rejects entire refresh.
- Validate schema/protocol, nonempty bounded model/name/effort lists, unique row model+effort, rows consistent with declared model_efforts, valid RFC3339 dates and ordered test interval. Reject absent critical metadata, duplicate identities, invalid weights/types/ranges and unknown schema/protocol; preserve last good cache.
- `round_id` is source batch/round identity, not a guessed numeric repetition count. UI labels “来源批次”. Source does not publish per-combination repeated-round count; do not invent it.
- Status: available; not_recorded (unknown model or supported pair missing/incomplete row); unsupported (known source model but effort not in its declared model_efforts); unavailable (no successful snapshot). Scores null unless available. Return selected pairs only, in configured order, plus source-supported catalog for admin configuration via existing admin settings response or separate admin-only catalog.

## Architecture and settings

- Extend existing `pelican_test_settings` JSON with an optional pointer/object `radar: {enabled, selections:[{model, effort}]}`; no new tables/migrations/dependencies. Legacy absence means disabled and empty selection. Omitted radar on older PUT preserves saved radar; explicitly empty selections clears choices. Existing DisplayName/Prompt/Enabled behavior must remain.
- Maximum 20 pairs, string lengths bounded, no blank identities/duplicates after alias normalization, known effort enum low/medium/high/xhigh/max/ultra. Allow syntactically valid unknown models to surface not_recorded rather than inventing a mapping. Store array order as display order. Existing comparable settings tests must stay valid or adjust only task-owned necessary expectations.
- Ordinary users have only authenticated read endpoint `GET /api/v1/pelican-tests/reference-scores`; do not expose prompt/settings or enable user writes. Reuse plaza-enabled/admin authorization semantics; admin config remains existing admin GET/PUT settings. No group/account scores or source data forwarded as HTML.
- Admin-only `GET /api/v1/admin/pelican-reference-catalog` may fetch the shared cache explicitly even when reference display is disabled, so administrators can configure before enabling; this is the sole exception to disabled no-fetch. Returns `{models:[{model,display_name,efforts:[]}], status, synced_at}`; unavailable catalog must not erase saved/custom pairs.
- Stable shared DTO for read endpoint: `{enabled:boolean,status:"ready"|"stale"|"unavailable"|"disabled",source_url:string,test_started_at:string|null,test_completed_at:string|null,source_snapshot_at:string|null,synced_at:string|null,round_id:string|null,score_max:number|null,items:[{model:string,display_name:string,effort:string,status:"available"|"not_recorded"|"unsupported"|"unavailable",score:number|null}]}`. No raw upstream error text. Settings `radar` optional for old clients; all dates RFC3339 or null.
- Add process-scoped cache to existing PelicanTestService, initialized once with constructor; no DI/wire generator expansion. Thread-safe immutable snapshots, clone exported slices. Fetch fixed HTTPS URL only, timeout <=10s, body <=2MiB, disallow redirects. No source URL parameter or arbitrary user-controlled fetch.
- Lazy first access may await one bounded fetch; all concurrent callers share one request. Successful cache TTL 30 min; ordinary requests hit cache. Expired cache returns previous snapshot immediately and starts one background refresh, independent of the caller cancellation. Failures retain good snapshot, set stale/update-delayed; retry at most once/minute. First failure returns usable unavailable DTO with retry backoff. No timer/per-user refresh loops.
- Snapshot synced_at is local successful fetch time, not source snapshot time. Stale if refresh failed or TTL expired. Never replace valid snapshot on fetch/decode/validation error. Disabled config does not fetch. Configuration changes apply immediately to selection/order over same source snapshot.

## UI

- Section heading “降智雷达 · 第三方参考评分”; compact horizontal cards, source model icon via existing ModelIcon, display name, effort, prominent total/max.
- Source test date/interval, source batch, last successful sync, safe “查看来源” external anchor. Explicit “第三方 Codex 订阅测试，不代表本站分组或账号的实测成绩。” Explain four-task weighted total; round to at most one decimal for display only.
- Desktop cards flex to fill width; narrow screens overflow inside score strip only, keyboard-scrollable with accessible label; no page horizontal overflow. Handle off/empty/loading/unavailable/not_recorded/unsupported/stale without disrupting original results.
- Existing admin settings gains switch and ordered model+effort list with add/remove/up/down. Catalog reflects exact source pairs; unknown configured values remain visible. Save failures preserve draft; older settings/API mocks without radar gracefully default.
- Put section outside gallery overlay replacement (before toolbar). It is real production component, not a separate page. Do not implement/change production horizontal gallery, retention rules, prompt or other plan behavior.

## Allowed Paths

Backend Developer:
- `backend/internal/service/pelican_test_service.go` (settings/cache field/constructor only)
- `backend/internal/service/pelican_reference_scores.go` (new adapter/cache/DTO/config helpers)
- `backend/internal/service/pelican_reference_scores_test.go` (new)
- `backend/internal/handler/pelican_test_handler.go` (read handler only, optional capability interface to preserve stubs)
- `backend/internal/handler/pelican_reference_scores_test.go` (new)
- existing route owner for authenticated pelican endpoints (`backend/internal/server/routes/user.go`, exact owner to verify before editing)
- `backend/internal/server/routes/admin.go` (single admin catalog GET route only)
- new `backend/internal/service/testdata/pelican-radar-*.json` if needed, small deterministic fixture, not complete upstream assets.

Frontend Developer:
- `frontend/src/api/pelicanTests.ts`
- `frontend/src/views/user/PelicanTestsView.vue` (settings and top section only)
- new `frontend/src/components/pelican/PelicanReferenceScores.vue`
- new `frontend/src/components/pelican/PelicanReferenceSettings.vue`
- new associated `frontend/src/components/pelican/__tests__/PelicanReference*.spec.ts`
- existing `frontend/src/components/pelican/__tests__/PelicanTestsView.spec.ts` only for necessary API mock/regression additions.

Controller: task-specific workflow contract/review/QA/report files; minimal concurrent status/main-log/current-task additions; preview directory `outputs/pelican-row-preview-20261030/` only for additive radar mock/fixture wiring and evidence, preserving gallery behavior. External evidence/scripts under `E:/codex-runtime/sub2api/pelican-radar-20261030`.

## Denied Paths

Other business modules, dependency manifests/lockfiles, database schema/migrations, generated wire, backend embedded dist, other workers' targets, global memory/config, production credentials, git index/history/remotes, Docker/container/deployment. Do not touch ongoing gallery files except controller's additive preview wiring.

## Constraints and frozen scope

Candidate `E:/codex-worktrees/sub2api/pelican-radar-20261030` at base plus current tracked dirty prerequisites and listed necessary untracked files. `main-baseline.json` binds 4805 tracked main files; preserve unknown drift. Freeze hashes of task targets before implementation and before exact main integration. Before copying back compare baseline vs main target bytes; changed target means stop integration/reconcile, never overwrite. Unrelated observed changes are preserved and reported. No git commit/push/deploy/container update.

## Success Criteria / Acceptance Commands

1. Source fixture exact total oracle, dynamic weights/max, real date/batch, strict matching/aliases/no fallback, unknown/unsupported/missing/null vs malformed, duplicate rows, schema/weights/date validation.
2. Clock/HTTP-controlled cache tests: 30min TTL, first-fetch dedup, expired reads + one refresh, caller cancellation independence, failure backoff/first unavailable, stale retention and malformed-cache rejection. Disabled causes no HTTP.
3. Settings old JSON/old PUT compatibility, clear/order/max/duplicate validation, admin-only PUT, authenticated read/plaza-disabled access, public response redaction.
4. Go focused commands from candidate backend: `go test ./internal/service -run 'TestPelican' -count=1`; `go test ./internal/handler -run 'TestPelican' -count=1`; `go build ./cmd/server`; gofmt task Go files and diff-check. Use GOTOOLCHAIN=local if necessary. Preserve/document pre-existing failures; no weakening old tests.
5. Frontend candidate: `npm.cmd run test:run -- src/components/pelican/__tests__`; `npm.cmd run typecheck`; targeted ESLint of changed Vue/TS; `npm.cmd exec vite -- build --outDir E:/codex-runtime/sub2api/pelican-radar-20261030/dist --emptyOutDir`. No output into backend/internal/web/dist.
6. Real browser original AppLayout preview at 127.0.0.1:62080/pelican-tests, live-source-derived fixture for radar (clearly mock preview, no provider/DB requests): desktop/mobile light/dark, sidebar open/collapsed, no global overflow, selected exact score/name/effort/date/batch, off/unknown/unsupported/unavailable/stale, admin editor, original plans/filter/enlarge/gallery.
7. Runtime local authenticated handler + fixture fetch tests prove backend HTTP/DTO independent from preview mock. Real upstream GET adapter parse test can run explicitly, never part of offline default tests.
8. Browser task-specific profile/session/PID evidence; close and verify owned Chrome/daemon zero. Keep preview server for user. Independent QA after frozen implementation, final controller diff/integration review.

## Output

Worker first-line verdict, exact changed files, commands/logs, limitations. Independent review `docs/workflow/contract-reviews/pelican-radar-20261030-review.md`; QA `docs/workflow/qa-reports/pelican-radar-20261030.md`, source hashes/evidence under external directory. Final original-page preview URL, scoped change list and tests; explicitly no commit/push/container/deployment.

## Stop Rules

Contract not PASS; model unavailable; unknown target drift; need schema/dependency/auth architecture expansion; unsupported source scoring semantics; cannot preserve gallery/dirty scope; independent QA reveals unresolved correctness issue. Retain failed evidence, fix narrowly and retest. No fabricated approvals, live-provider test or deployment.

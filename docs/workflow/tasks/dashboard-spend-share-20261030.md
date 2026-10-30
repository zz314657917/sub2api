---
task_id: dashboard-spend-share-20261030
status: approved
review_verdict: PASS
worker_model: gpt-6.1-sol
base_commit: 8f202b73282465a9b4e9d5bcfeb1c267362e72c4
spec_ref: dashboard-spend-share-20261030.md
---

# Task ID / Role / Goal

dashboard-spend-share-20261030. Controller Planner/Final Evaluator; independent Contract Reviewer, Developer and QA use gpt-6.1-sol. User requests actual credit consumption percentages for the dashboard input/output/cache-read/cache-write composition. Preserve previous dashboard/header/token/cache-rate changes.

## Baseline / Constraints

Isolated candidate E:/codex-worktrees/sub2api/dashboard-spend-share-20261030 at base plus current backend/frontend source snapshot. Evidence and SHA256 originals: E:/codex-runtime/sub2api/dashboard-spend-share-20261030/baseline-files.json. Dependency junction is read-only. All existing dirty work must survive; main file hashes checked before exact integration. No blanket copy, staging, commit, push, production/provider/live DB/container actions, migrations or dependency changes. No budget cap requested.

## Allowed Paths

- backend/internal/pkg/usagestats/usage_log_types.go
- backend/internal/repository/usage_log_repo.go
- New backend/internal/repository/usage_log_spend_breakdown.go and usage_log_spend_breakdown_test.go if needed.
- frontend/src/types/index.ts
- frontend/src/components/user/dashboard/UserDashboardCharts.vue
- frontend/src/components/user/dashboard/__tests__/UserDashboardCharts.spec.ts
- frontend/src/styles/console-design.css
- frontend/src/i18n/locales/{zh,en}/dashboard.ts
- DESIGN.md
- Controller owns this contract, contract-reviews/dashboard-spend-share-20261030-review.md, qa-reports/dashboard-spend-share-20261030.md and only this concurrent section/status/main-log entries.
- Workers may write external evidence, independent QA harnesses and source manifests under this task evidence root. Workers do not modify main.

## Denied Paths

All other product paths, billing/pricing/write flows, schema, dependencies, lockfiles, unrelated dashboard behavior, admin aggregation semantics, user authorization, shared preview/processes, knowledge and memories. If an existing test needs change outside allowlist, stop and request a narrow amendment.

## Success Criteria

1. Existing user trend read path returns an optional nested actual_cost_breakdown with input, output, cache_creation, cache_read, other costs. Values are actual credits, not standard costs or token counts. Reuse one grouped raw query and existing user/filter/date/timezone scope, no per-log API/N+1 or extra user-wide fetch.
2. Allocate EACH billing_status='applied' record's actual_cost in proportion to its stored reference cost components and total_cost, THEN sum by bucket. Never multiply aggregate component costs by one averaged multiplier. pending/failed contribute zero. Preserve actual_cost total (including existing subscription consumption semantics) and existing stats/count semantics. Use safe SQL division; total_cost<=0, null component metadata, negative components, or component-sum>total_cost leaves the entire actual spend in other. Strict fallback is acceptable even for historical storage-rounding inconsistencies: do not invent corrected prices by normalizing invalid weights. Normal valid rows: each category=actual_cost*reference_category/total_cost, other=actual_cost-sum(category), with a nonnegative numerical-residue guard only within machine arithmetic precision. Image/flat/request/video and all unclassified reference costs remain in other; never double count image costs as text token categories. No ledger mutation or fabricated pricing.
3. Protect original nine-column scanner and preaggregated administrator day/hour query path. Use explicit extended scanning only for user-scoped raw trends (including GetUserUsageTrendByUserID and filtered user trend); nil optional fields elsewhere retain backward compatibility. Both dashboard trend and snapshot use same DTO naturally.
4. UI keeps token counts, shows actual credit amount and share of selected-range actual_cost for all four categories; progress width equals consumption percentage, not count/max. Clear localised heading/description and per-request allocation explanation. Display other when nonzero; total denominator includes it, never force four categories to 100%. Missing/incomplete backend fields show unavailable, never infer spend from tokens or standard cost. Valid zero actual spend shows 0% (with existing zero-cost explanation). Loading/error/date changes preserve existing behavior.
5. Finite/nonnegative validation and sensible bounded precision for percentages/widths; tiny nonzero shares retain an honest indication. Minor independent rounding can sum about 100%; do not silently renormalize wrong denominators. Desktop1440/tablet768/mobile390 light/dark, keyboard/readability and no horizontal overflow.
6. Meaningful focused tests cover mixed requests/multipliers and costs unequal to token ratios, applied vs pending/failed, zero/missing costs, image/flat residual, malformed/rounding policy, both user raw entrypoints, administrator original scanning, propagation/shape, zero/missing/partial frontend fields and range changes. Independent QA executes arithmetic against generated SQL using a disposable local SQL harness where available; distinguish such evidence from real production PostgreSQL/backend acceptance.

## Acceptance Commands / Output

- In candidate/backend: GOTOOLCHAIN=local with existing Go1.26.10 binary; go test -tags unit ./internal/repository -run 'Test.*(SpendBreakdown|Trend)' -count=1; go test ./internal/pkg/usagestats; go build ./... . Record toolchain path and baseline failures honestly.
- In candidate/frontend: npm.cmd run test:run -- src/components/user/dashboard src/views/user/__tests__/DashboardView.spec.ts src/components/layout/__tests__; npm.cmd run build -- --outDir E:/codex-runtime/sub2api/dashboard-spend-share-20261030/dist (vue-tsc included; never default backend dist).
- Targeted ESLint using existing parser resolution; gofmt check, git diff --check and SHA scope checks against baseline (not unrelated dirty diff).
- QA real Chromium using dedicated session/profile, recorded PIDs; deterministic API fixtures for shares/ranges/missing/zero/other cases; desktop/tablet/mobile light/dark screenshots and settled overflow checks. Read playwright skill before browser tools; close browser and verify exact task session/profile browser/daemon processes zero. Retain user's preview server.
- Developer freezes changed sources with path/baselineSha256/sha256 and first-line PASS/FAIL/BLOCKED report. Developer stops writes before QA. Independent QA uses frozen sources and own assertions; starts report ### PASS/FAIL/BLOCKED: dashboard-spend-share-20261030. Controller reviews diff/evidence before exact integration.
- Controller updates local fixture preview after integration and preserves today's Token, cache rate and direct header entries. Preview is demonstrative; source tests/build do not prove a live DB/provider deployment.

## Stop Rules

Stop for missing contract approval, denied paths, shared-source drift, unavailable assigned model, unsafe/broken cost semantics or mandatory runtime unavailability. Return findings to original owner; scope ambiguity to Planner. Final acceptance independent of contract review.

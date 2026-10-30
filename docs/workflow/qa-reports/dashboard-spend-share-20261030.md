### PASS: dashboard-spend-share-20261030

## Scope and findings

User requested actual credit consumption percentages for input/output/cache read/cache write. Exactly11 approved product files integrated; no clear implementation issue found in independent candidate QA. Consumption is allocated per applied record using stored reference weights, then summed, with unclassified/invalid reference weights retained in Other. Existing billed actual total, subscription semantics, today's Tokens, cache rate and header shortcuts are preserved.

## Executed checks

- Independent contract review PASS: draft85ace248, contract-reviews/dashboard-spend-share-20261030-review.md.
- Developer source freeze and independent QA PASS: E:/codex-runtime/sub2api/dashboard-spend-share-20261030/qa/report.md. Go focused tests/build, Vue88 tests/build, targeted ESLint/gofmt/diff,13 independent SQL arithmetic cases and58 Chromium assertions PASS. All11 frozen hashes match before/after.
- Exact main integration checked original target hashes immediately before copying. All11 integrated hashes match;12303 non-target files, HEAD and index unchanged. Originals retained externally in baseline-targets. Evidence: integration-result.json.
- Fresh main: Vue13files/88tests PASS, Go repository focused tests PASS and full Go/frontend builds PASS. Scoped diff-check PASS. Logs: main-go-tests.log, main-go-build.log, main-frontend-tests.log, main-frontend-build.log.
- Initial strict equality comparison of main-built assets versus candidate assets failed and is retained as main-final-verification.json. The source comparison identifies exactly four preserved concurrent Pelican source/test changes; all11 task source hashes still match. Build hash changes cascade through shared imports. combination-differences.json and main-final-verification-r2.json record this explanation without claiming byte-identical output.
- Independent main-built preview follow-up PASS: qa/main-smoke/report.md under the evidence root. All58 original browser assertions rerun against actual current-main build;1440/768/390 light/dark, shares/amounts/progress widths, Other, dates, keyboard, missing/partial/zero/tiny and previous header/token/cache behavior pass. The four concurrent Pelican diffs were independently reviewed as unrelated and preserved. All11 source hashes rechecked after browser testing.

## Limits and recommendation

SQL arithmetic executed generated expressions in disposable SQLite with an independent Decimal oracle, not production PostgreSQL. Browser APIs are deterministic fixtures; no live DB/provider/backend deployment acceptance. Existing chunk-size warnings remain. No commit, push, migration, dependency, pricing/write-flow or container/deployment change.

Final Evaluator: local integration accepted. Preview: http://127.0.0.1:62084/dashboard, loopback fixture-only server PID5932 retained. Both independent QA sessions/profiles were closed; original and main-built follow-up process reports show owned_processes=[]. Source, exact integration, fresh-main build/tests and current-main browser evidence agree. No commit/push/deployment.

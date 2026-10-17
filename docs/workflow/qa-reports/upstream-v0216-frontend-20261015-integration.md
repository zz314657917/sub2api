### PASS: upstream-v0216-frontend-20261015-integration

2026-10-15. Independent contract review, Developer and separate QA completed; controller integrated the exact sealed 15 product/test files into the existing main workspace. No retrospective approval of the early provisional candidate commits is claimed.

## Scope

- Payment countdowns in three views use elapsed wall time and expire polling after suspended timers resume.
- Group multiplier edits reject negatives/nonfinite values while preserving local zero overrides.
- Select does not consume IME composition keys.
- Ops datetime-local defaults use local time and retain UTC API serialization.
- Account group names wrap in cells/popovers; default GroupBadge truncation elsewhere is preserved.
- Gemini OAuth URL errors use the existing API-error extraction utility.
- GPT-6.1 Sol 500k and upstream-subscription long-context exemption already exist; regression only.

No commits, push or deployment authorized in this batch. Original unrelated dirty files, UsageProgressBar changes and digital-store isolation must remain intact.

## Findings

No blocking frontend issue found in the accepted scope. Upstream rejection of zero multipliers was incompatible with the existing local regression; the accepted adaptation rejects negative/nonfinite values and preserves zero.

Gateway execution remains BLOCKED by existing canonical `unit` test compilation failures. That separate batch is not covered by this PASS.

## Executed Checks

- Independent QA: mandatory 7 suites / 41 tests and related 25 suites / 139 tests PASS, zero skip/fail. Nine tests overlap between runs: 171 different tests, 180 executions. Full vue-tsc, production Vite build, scoped 15-file ESLint and committed/uncommitted/base-working diff checks PASS. Report: `docs/workflow/qa-reports/upstream-v0216-frontend-20261015.md`.
- Independent seal manifest SHA-256 `46d500c35b56db9450f7a530c6f9d9de83821287ad07ee95229ab88d7e5b53f1`. All 15 source bytes rechecked against the seal before copy and on main after copy.
- External integration script completed successfully. Main HEAD remains `de33a18798a62759f5ce5e03cdd4ba6cdf9d6d1d`; index unchanged. All 12102 non-target files captured immediately before integration retained identical hashes afterward; no unexpected path drift. Original dirty work was not replaced.
- The original 12111-file starting snapshot differs only in the nine accepted existing product files and controller-owned workflow/handoff updates; six new accepted tests were added. Generated assets and dependency contents were not copied to main.
- GPT-6.1 Sol existing behavior regression: 21 top-level / 364 including subtests PASS in the isolated unchanged backend, OpenAI package PASS. Main workspace full production `go build -p 1 ./...` PASS before frontend source copy. No backend product files changed in this batch.

Evidence root: `E:/codex-runtime/upstream-v0216-20261015/`; `integration-result.json`, `main-before.json`, `qa-sealed-files.json`, `qa-checks.json`, and main target byte backups under `main-target-backup/`.

## Unverified Risks

Frontend acceptance was executed on the isolated candidate; the integrated 15 files are byte-identical. No full test run against all unrelated main dirty changes is claimed. No real browser layout/IME, real payment/Gemini provider, production database or deployment acceptance. Existing gateway tagged baseline compilation failures remain visible in `gateway-baseline-unit.log`.

## Recommendation

Frontend local integration PASS. Leave files uncommitted for review. Resolve the separately documented gateway test compatibility blocker under a new reviewed scope before implementing that batch.

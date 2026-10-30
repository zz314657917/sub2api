# Console design — final acceptance

Verdict: **PASS — locally integrated**, 2026-10-30.

## Delivered

- Warm ivory/terracotta light theme and warm charcoal dark theme; scoped shared shell, flat surfaces, grouped user navigation and keyboard-accessible help menu.
- Dashboard four core metrics, 7/30/90-day actual-credit spend bars and interval summaries, actual-cost-ranked models, Token trends and compact performance details.
- Explicit loading/error/empty/zero states; stale-response protection; existing onboarding, feature flags, custom/admin navigation and credit semantics preserved.
- Project `DESIGN.md` documents adapted reference principles and reusable tokens/components. No new library or remote font.

## Source and integration

- Approved contract: `tasks/console-design-20261030.md`; independent contract review PASS binds draft SHA256 `2db0e5531990928af05120171fa9f2932c48b67a4ba09bfb96214edb5b0de9ad`.
- Final R3 manifest: `E:/codex-runtime/sub2api/console-design-20261030/developer-freeze-r3.json`, SHA256 `fd0b94bc3154e530fa17895e997174f85c4dbbd3d749e7b03807d471a53201c1`.
- Exactly 28 allowed files integrated from the stopped, independently accepted candidate. Original current-workspace bytes were checked immediately before copying. All integrated hashes match; unrelated frontend bytes, HEAD and index unchanged.
- Evidence: `integration-before.json`, `integration-result.json`, `main-final-verification.json` under the evidence root above. No commit, push, deployment, backend/API/dependency/router change performed by this task.

## Executed verification

- Fresh main checkout: 13 files / **75 tests PASS**, including existing onboarding coverage; `main-tests.log`.
- Fresh main build including `vue-tsc -b` PASS, external output `main-dist`; `main-build.log`. Scoped diff-check and final file/HEAD/index protection PASS.
- Independent R2: standalone typecheck, 75 tests, ESLint and external build PASS; desktop 1440, tablet 768 and mobile 390 light/dark views with no document horizontal overflow. Range dates/timezone, delayed-response protection, retry, first-use/range-empty/zero/error states, onboarding persistence and explicit navigation verified with synthetic fixtures.
- Independent R2 also verifies header keyboard/focus, feature-gated and custom/admin navigation, representative keys/usage/profile shell, actual-cost ordering, balance-link fallback, chart fonts and reduced motion.
- R2 found a genuine dark DataTable header color override. Original failure retained. R3 changes only `console-ui.css`; other 27 hashes match R2. Independent R3 build and affected real-browser checks PASS: warm header, sticky/default/hover surfaces, mobile drawer settled closure and representative light/dark dashboard screenshots. R3 explicitly inherits unchanged R2 functional checks; it does not claim they were rerun in that phase.
- Both QA rounds verified source hashes before/after. Final dedicated browser, daemon and Vite cleanup reports zero task-owned remaining processes. No unrelated process closed.

## Evidence and limitations

- Independent final report: `E:/codex-runtime/sub2api/console-design-20261030/qa/r3/report.md`; broader functional details: `qa/r2/report.md`.
- Final screenshots: `qa/r3/top-desktop-light.png`, `top-desktop-dark.png`, `after-mobile-light.png`, `after-mobile-dark.png`, `keys-dark.png`. Tablet matrix evidence remains in `qa/r2/`.
- Real Chromium used deterministic intercepted synthetic APIs, a dedicated profile and dead-loopback backend proxy. This proves frontend rendering/interactions, not real provider/database/backend or deployed acceptance. Unknown/external/mutating requests were denied; final R3 request blocks, console errors and page errors are empty.
- The keys table fixture fits its desktop viewport; actual nonzero horizontal scrolling was not exercised. Broader production feature combinations/dense editors are not covered by these representative checks.
- Existing Browserslist/Node deprecation/chunk-size build warnings remain; no dependency upgrade or bundle-size rework was included.

## User follow-up: direct header shortcuts

User requested Tutorial, Model plaza and Contact support remain directly visible in the top-right header. Updated AppHeader and its focused tests, renamed the shared shortcut style, and synchronized DESIGN.md. Desktop shows icon/text; narrow screens retain labeled icons. Original model/support visibility gates and support popup event are preserved; documentation remains in the account panel.

Focused header tests 2/2 PASS, targeted ESLint PASS (existing pnpm parser resolved through process-local NODE_PATH; no dependency change), external build including vue-tsc PASS, scoped diff-check PASS. Real Chromium verifies all three shortcuts at 1440/768/390 with no settled document overflow, plus the original support event. Initial immediate resize sample was superseded by waiting for responsive chart resizing. CLI close and owned-process scan report zero. Evidence: `E:/codex-runtime/sub2api/console-design-20261030/preview/shortcut-browser-results-settled.log`, `shortcut-browser-cleanup.json`, `shortcut-build.log`. Current fixture preview is http://127.0.0.1:62083/dashboard; previous 62082 preview is retained. No commit/push/deployment.

## User follow-up: today's Tokens and selected-range cache rate

The third summary card now displays today_tokens; the cost panel side summary displays aggregate cache_read_tokens / (input_tokens + cache_read_tokens), with unavailable denominator shown as a dash. Token counts, date/range behavior and actual spend remain based on their original data fields.

This follow-up executed 9 files / 51 tests PASS (separate from the original 75-test design pass), targeted ESLint and external build including vue-tsc PASS. Dedicated Chromium verified 1440/390 layout, today's Tokens, cache rate 16.7% and 30-day switching. CLI close and owned-process cleanup report zero. Evidence: preview/token-cache-tests.log, token-cache-build.log, token-cache-browser-results.log and token-cache-browser-cleanup.json under the console-design evidence root. Fixture evidence does not establish live backend/DB acceptance. Actual credit consumption composition is a separate follow-up under dashboard-spend-share-20261030.

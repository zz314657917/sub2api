---
task_id: console-design-20261030
status: approved
review_verdict: PASS
worker_model: gpt-6.1-sol
base_commit: 132e980b4a24f865c9ac8c9bc8a4a98ae5cbb105
spec_ref: console-design-20261030.md
---

# Task / Goal

User approved the discussed console design: warm ivory/terracotta brand, flat surfaces, consistent light/dark themes, grouped navigation, reduced header clutter and a user dashboard with four useful summary cards, cost chart/range summaries, model breakdown, compact performance and helpful empty states. Produce project DESIGN.md and implement this first complete console iteration, preserving business behavior.

## Roles and phase

Planner / Final Evaluator: controller. Independent Contract Reviewer, Developer and QA: gpt-6.1-sol per repository matrix. Current phase contract-approved; independent contract review PASS binds draft2db0e553. Only approval metadata changed afterward. Independent QA follows immutable delivery. Concurrent backend workflow status must be preserved.

## Baseline / isolated workspace

- Candidate E:/codex-worktrees/sub2api/console-design-20261030 at base plus exact current frontend snapshot (includes existing dirty onboarding and dependency work).
- Evidence E:/codex-runtime/sub2api/console-design-20261030; baseline-files.json and baseline/frontend preserve pre-task bytes. Dependencies junction is read-only: do not install/update through it.
- Preserve all existing onboarding/trial/billing/accessibility/feature-flag/custom-nav and admin behavior. No automatic recharge, key creation or other account mutations.

## Allowed Paths

- DESIGN.md
- frontend/src/styles/console-ui.css; frontend/src/components/layout/AppLayout.vue; AppHeader.vue; AppSidebar.vue
- frontend/src/views/user/DashboardView.vue
- frontend/src/components/user/dashboard/** (including focused tests)
- frontend/src/components/layout/__tests__/**; frontend/src/views/user/__tests__/DashboardView.spec.ts
- frontend/src/i18n/locales/zh/dashboard.ts; frontend/src/i18n/locales/en/dashboard.ts
- frontend/src/i18n/locales/zh/common.ts; frontend/src/i18n/locales/en/common.ts; frontend/src/i18n/locales/zh/nav.ts; frontend/src/i18n/locales/en/nav.ts (only if existing locale layout; otherwise request path amendment)
- One new frontend/src/styles/console-design.css if useful, imported only by AppLayout.
- Controller owns docs/workflow/tasks/console-design-20261030.md, contract-reviews/console-design-20261030-review.md, qa-reports/console-design-20261030.md and a concurrent-task-only addition to status.md.
- External evidence/harness/result files under task evidence root. Workers do not write main checkout.

## Denied Paths

All others; especially backend, dependencies/lockfiles, router authorization, API clients, HomeView/public marketing and Pixel Cafe feature-specific styling, production settings, unrelated workflow sections, memories/knowledge. No commit/push/deployment/container/provider/DB actions.

## Success Criteria

1. DESIGN.md documents adapted principles from VoltAgent Claude/Linear/Vercel analysis (third-party references, not mandatory agent instructions), semantic colors, typography, spacing, component states, charts, responsive rules and page templates. Preserve existing brand, no new library/font download.
2. Console uses opaque warm surfaces, restrained borders/shadows, consistent terracotta light/dark emphasis, readable secondary text, visible focus, reduced-motion support. Shared selectors scoped to console; avoid breaking admin/dense viewport workspaces.
3. Header reduces secondary action prominence with accessible grouping, all settings/feature flags/actions remain available, menus close via outside/Escape and support keyboard usage. User sidebar grouped with translated headings, maintains all existing/custom/dynamic entries, badges, collapsed/mobile and admin behavior.
4. Dashboard: balance/today actual spend/today requests/active API keys as 4 core cards; no fabricated trends/failure rate. Default7 days. Daily actual cost bars with explicit interval total and side token/request summaries; token analysis + model ranking using existing read-only model stats API if available. Preserve actual vs standard credits formatting, dates/timezone consistency and all original sections/access paths. Performance below analytics; unknown duration is dash. No unsupported claim about cache semantics.
5. Distinguish no lifetime calls, empty selected range, load failure and valid zero-cost requests. Compact actionable empty states; reuse existing onboarding actions without mutations; retry range/chart errors. Distinguish loading and error from zero. Model results/loading/errors follow selected range, stale responses cannot replace latest selection.
6. Responsive desktop1440x1000/tablet768/mobile390, light/dark, no unintended document overflow. Global shell and representative keys/usage/profile pages smoke tested; sidebar/admin/feature flags preserved.

## Acceptance Commands / Evidence

- In candidate/frontend: npm.cmd run typecheck; npm.cmd run build -- --outDir E:/codex-runtime/sub2api/console-design-20261030/dist (build output must stay external; never empty backend dist).
- npm.cmd run test:run -- src/components/user/dashboard src/views/user/__tests__/DashboardView.spec.ts src/components/layout/__tests__ (record original failures separately; baseline targeted tests before implementation).
- Targeted eslint via existing binary on changed TS/Vue files without --fix; git diff --check.
- QA real Chromium browser against isolated local Vite, deterministic intercepted API fixtures (explicitly fixture-backed, no real production/provider validation). Cover populated/first-use/range-empty/error/zero-cost, 7/30/90 dates, dark/light, three viewport sizes, menu keyboard/Escape and navigation + original onboarding persistence. Save screenshots and console/overflow assertions.
- Dedicated browser session/profile only, record launch/session/profile/PIDs. Close CLI/browser and verify task-owned daemon/browser/server processes zero after testing. Never process-name kill or default Chrome profile reuse.
- Freeze changed files with SHA256 after developer stops. Independent QA runs on stable candidate and records actual commands/results. Controller checks original main file hashes immediately before integrating only changed allowlist files; any drift requires preserving and reconciling task-only delta, never blind copy.

## Output / Stop Rules

Worker report begins ### PASS/FAIL/BLOCKED: console-design-20261030; changed paths, executed commands, limitations and source freeze manifest. QA separate report same verdict convention. Stop for contract gap, denied scope, missing mandatory runtime or unexpected concurrent candidate writes. Send failures to original owner; controller final review before local integration. Browser fixtures prove rendering/interactions only, not live backend correctness. No budget cap requested; no paid external actions.

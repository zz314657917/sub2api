# Task Contract: upstream-latency-health-column-s65

## Task ID

`upstream-latency-health-column-s65`

## Status

`approved`

## Role

Codex acts as Planner, Generator, QA executor, and Final Evaluator for this narrow selective upstream port.

## Goal

Port only the latency-health presentation introduced by upstream commit `1a3cc2a78`: merge first-token and total-duration display into one latency column with severity colors and a two-ended gradient bar, while preserving the locally customized usage-page layout.

## Success Criteria

- Admin and user usage tables expose one latency column instead of separate first-token and duration columns.
- First-token thresholds are 10s / 30s / 60s; total-duration thresholds are 1m / 3m / 5m.
- Missing first-token data renders `-` and uses a solid bar based on total duration.
- Chinese and English labels cover latency, first token, and total duration.
- Utility boundary tests, admin/user usage tests, typecheck, production build, and `git diff --check` pass.
- No upstream ranking, filter, tab, error-log, or broad page-layout refactor is included.

## Allowed Paths

- `frontend/src/utils/latencyHealth.ts`
- `frontend/src/utils/__tests__/latencyHealth.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `frontend/src/views/admin/UsageView.vue`
- `frontend/src/views/admin/__tests__/UsageView.spec.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `docs/workflow/tasks/upstream-latency-health-column-s65.md`
- `docs/workflow/qa-reports/upstream-latency-health-column-s65-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- backend, migrations, billing, routing, deployment, and production configuration
- upstream `UserTokenRanking`, `UsageFilters`, `OpsErrorLogTable`, or whole-page layout changes
- `knowledge/**` and global memories

## Constraints

- Selective port only; do not cherry-pick upstream commit `1a3cc2a78` wholesale.
- Preserve local warm console styling and existing CSV/export semantics.
- Do not change API payloads or duration data semantics.

## Acceptance Commands

```powershell
Push-Location frontend
npm.cmd run test:run -- src/utils/__tests__/latencyHealth.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/views/user/__tests__/UsageView.spec.ts
npm.cmd run typecheck
npm.cmd run build
Pop-Location
git diff --check
```

## Output

- QA report: `docs/workflow/qa-reports/upstream-latency-health-column-s65-qa.md`
- Final verdict: `PASS`, `FAIL`, or `BLOCKED`

## Stop Rules

- Stop if implementation requires backend/API changes or broad upstream layout replacement.
- Stop on out-of-scope paths, missing required tests, or unresolved regression failures.

# Task Contract: usage-model-reasoning-effort-s82

## Task ID

`usage-model-reasoning-effort-s82`

## Status

`approved`

## Role

Direct Codex frontend implementation with an evidence-first Evaluator pass.

## Goal

In user and administrator usage-record tables, append the normalized reasoning
effort to the requested model name when the record contains a meaningful
effort value.

## Success Criteria

- User usage rows show `gpt-5.5 (XHigh)` for a row whose model is `gpt-5.5`
  and whose `reasoning_effort` is `x-high`/`XHigh`.
- Admin usage rows apply the same suffix to the requested model.
- For an admin mapping chain, the suffix appears only on the first/requested
  model, never on each upstream model in the chain.
- Empty, whitespace-only, `none`, and `minimal` effort values add no suffix.
- Existing standalone reasoning-effort columns remain available and unchanged.
- Existing upstream-brand sanitization still applies before the suffix.

## Allowed Paths

- `frontend/src/utils/modelDisplay.ts`
- `frontend/src/utils/__tests__/modelDisplay.spec.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/usage-model-reasoning-effort-s82.md`
- `docs/workflow/worker-results/usage-model-reasoning-effort-s82-result.md`
- `docs/workflow/qa-reports/usage-model-reasoning-effort-s82-qa.md`

## Denied Paths

- Backend, API response types, migrations, generated files, persistence, and
  reasoning-effort capture/normalization behavior.
- Dashboard recent usage, balance-history modal, model plaza, charts, exports,
  filters, column visibility defaults, i18n copy, and public pages.
- Dependencies, lockfiles, deploy/Docker, VERSION, `knowledge/**`, and global
  memories.

## Constraints

- Reuse the existing `formatReasoningEffort` normalization and
  `displayModelLabel` sanitization; do not duplicate effort mappings.
- Append the suffix only when the formatter produces a meaningful value rather
  than `-`.
- Preserve the user's three unrelated dirty `knowledge/**` files exactly.
- Do not commit, push, deploy, update containers, or modify database state.

## Acceptance Commands

```powershell
Push-Location frontend
npm.cmd run test:run -- src/utils/__tests__/modelDisplay.spec.ts src/views/user/__tests__/UsageView.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts
npm.cmd run typecheck
npm.cmd run build
Pop-Location

git diff --check
git status --short
```

Evaluator additionally reviews the rendered branches for simple model,
requested/upstream pair, and mapping-chain rows; audits all changed paths; and
checks that the existing dirty `knowledge/**` diffs were not touched.

## Output

- Worker result: `docs/workflow/worker-results/usage-model-reasoning-effort-s82-result.md`
- QA report: `docs/workflow/qa-reports/usage-model-reasoning-effort-s82-qa.md`
- Workflow status/log entries through final PASS/FAIL/BLOCKED.

## Stop Rules

- Stop if the API lacks `reasoning_effort` on `UsageLog` or either target table
  does not receive it.
- Stop if implementation requires backend/schema changes or modifies model
  routing, billing, exports, filters, or column defaults.
- Stop if any changed business path leaves the six frontend-path allowlist.

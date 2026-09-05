### DONE: prompt-audit-policy-matrix-s293-r3-admin

## Changed Files

- `frontend/src/features/prompt-audit/PromptAuditView.vue`
- `frontend/src/features/prompt-audit/components/PolicyPanel.vue`
- `frontend/src/features/prompt-audit/viewModel.ts`

## Implemented

- Separated saved policy fingerprint from local edits; publish is disabled after
  editing until the draft is saved again.
- Initial load restores a stored draft when its base configuration matches.
- Editing invalidates preview/shadow output.
- Added safety, multi-category, group, model and provider rule fields.
- Backend policy preview now derives examples through `ParseQwen3Guard`.

## Checks

- Prompt Audit Vitest: 32/32 PASS
- `npm.cmd run typecheck` PASS
- `npm.cmd run build` PASS

## Risks / Unverified

- Real authenticated browser against a test backend and Qwen3Guard runtime were
  not run; mock/component tests do not prove backend lifecycle convergence.
- Legacy normalized shadow remains compatible; the admin editor now uses the
  optional bounded guard_output path with the active scanners and policies.

## Continuation 2026-09-05

- Saved response fingerprints no longer mark in-flight edits clean. Stale drafts
  remain editable/visible but cannot publish; publish/rollback update local config
  versions, and ordinary config saves immediately expose draft base conflicts.
- Sample/context/candidate changes invalidate late preview/shadow results. Group,
  model and provider context are editable; invalid group IDs do not silently become
  global policy scope. Independent Terra recheck completed.
- Preview/shadow compare active and candidate independently from parser results;
  a custom escalation can be removed without lowering Unsafe/unknown floors.
- Latest focused Vitest 43/43, typecheck and production build PASS. Runtime and
  browser limitations are recorded in the continuation QA report.

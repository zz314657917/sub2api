---
task_id: ops-log-cleanup-i18n-s144
phase: contract-approved
owner: codex
contract_review: approved
---

# Ops System Log Cleanup Error Localization S144

## Role

Planner/Generator/Evaluator (direct small fix; no worker required).

## Goal

When the administrator Ops system-log cleanup endpoint rejects a request with
the stable `OPS_SYSTEM_LOG_CLEANUP_FILTER_REQUIRED` error code, show the
localized Chinese or English explanation instead of the generic cleanup
failure message. Preserve the backend-provided detail for all other errors.

## Scope

Allowed paths:

- `frontend/src/views/admin/ops/components/OpsSystemLogTable.vue`
- `frontend/src/i18n/locales/zh/admin/ops.ts`
- `frontend/src/i18n/locales/en/admin/ops.ts`
- `frontend/src/views/admin/ops/**/__tests__/**` (only if focused coverage is
  needed)
- this contract and the corresponding workflow receipt/QA report

Denied paths:

- backend, API contracts, migrations, generated code, deployment, containers,
  and production/runtime state
- `frontend/src/views/admin/AuditLogView.vue` and its operation-audit locale
  namespace
- unrelated Ops hardcoded labels or upstream feature chains

## Constraints

- Reuse the existing `extractApiErrorMessage` helper and error-code mapping
  convention.
- Do not change request payloads, cleanup confirmation behavior, or success
  handling.
- Keep existing user-facing fallback text for unrecognized errors.
- Do not cherry-pick the upstream commit blindly; adapt only the three-file
  behavior slice to the current tree.

## Acceptance Commands

- `corepack.cmd pnpm --dir frontend exec vitest run <focused Ops tests>`
- `corepack.cmd pnpm --dir frontend run typecheck`
- `corepack.cmd pnpm --dir frontend run build`
- `git diff --check`
- conflict-marker and unmerged-index scans
- exact allowed-path audit

## Output

- One focused implementation commit on the isolated branch.
- A QA report with executed commands, evidence, and explicit runtime/deploy
  limitations.

## Stop Rules

- Stop if the backend error code or response shape differs from the stated
  contract; do not modify backend code in this Sprint.
- Stop if the diff leaves the allowed paths or touches the operation-audit
  page.
- Do not deploy, update containers, or force-push.

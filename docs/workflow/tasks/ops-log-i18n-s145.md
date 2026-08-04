---
task_id: ops-log-i18n-s145
phase: contract-approved
owner: codex
contract_review: approved
---

# Ops System Log UI Localization S145

## Role

Planner/Generator/Evaluator (bounded frontend port; no worker required).

## Goal

Complete the existing `OpsSystemLogTable` display localization so English
sessions no longer show the component's historical Chinese labels, while
Chinese wording remains unchanged and all current request, filtering, cleanup,
runtime-config, and table behavior is preserved.

## Scope

Allowed paths:

- `frontend/src/views/admin/ops/components/OpsSystemLogTable.vue`
- `frontend/src/i18n/locales/zh/admin/ops.ts`
- `frontend/src/i18n/locales/en/admin/ops.ts`
- `frontend/src/views/admin/ops/components/**/__tests__/**`
- this contract and the corresponding workflow QA report

Denied paths:

- backend, API contracts, migrations, generated code, deployment, containers,
  and production/runtime state
- host-filter, API-key-filter, mobile-card, account-label, or other upstream
  behavior changes not required for localization
- `frontend/src/views/admin/AuditLogView.vue` and `admin.audit` locales
- unrelated hardcoded labels outside `OpsSystemLogTable.vue`

## Constraints

- Reuse the existing `useI18n` and `admin.ops.systemLogs` namespace.
- Keep the Chinese strings behaviorally equivalent to the current UI text.
- Keep English and Chinese `systemLogs` key sets identical.
- Preserve request payloads, confirmation/success/error behavior, filter reset,
  pagination, and runtime configuration semantics.
- Keep the S144 stable cleanup-error mapping intact.
- Do not cherry-pick upstream `b4f38b092` or `d9e514f98` wholesale; adapt only
  the localization slice to the current modular locale tree.

## Acceptance Commands

- `corepack.cmd pnpm --dir frontend exec vitest run <focused Ops tests>`
- `corepack.cmd pnpm --dir frontend run typecheck`
- changed-file ESLint
- `corepack.cmd pnpm --dir frontend run build`
- `git diff --check`
- conflict-marker and unmerged-index scans
- exact allowed-path audit

## Output

- One focused implementation commit on the isolated branch.
- A QA report with locale parity evidence and explicit runtime/deploy limits.

## Stop Rules

- Stop if localization requires changing API payloads or unrelated upstream
  behavior; open a separate contract instead.
- Stop if the diff leaves the allowed paths or changes the audit-log page.
- Do not deploy, update containers, or force-push.

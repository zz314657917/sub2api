---
task_id: audit-log-i18n-s141
status: contract-approved
role: Generator
qa_mode: browser
---

# Task Contract

## Goal

补齐管理员操作日志页的显示层本地化：角色、认证方式和审计动作在当前语言
下显示可读标签，同时保留原始值用于筛选、悬浮提示和问题追踪。

## Success Criteria

- 中文界面不再直接显示常见的 `admin`、`user`、`admin_api_key` 和审计动作
  标识；英文界面保持自然英文标签。
- 常见动作按已知动作键或动作段落翻译；未知值安全回退为原始标识，不影响
  日志展示或 API 查询。
- 列表和详情使用同一套显示逻辑，原始动作仍保留在 `title` 属性中。
- 审计 locale 结构保持单层 `admin.audit`，不重新引入重复命名空间。
- focused locale/component checks、typecheck、changed-file ESLint、production
  build、diff、冲突标记和 allowlist 检查通过。

## Allowed Paths

- `frontend/src/views/admin/AuditLogView.vue`
- `frontend/src/i18n/locales/en/admin/audit.ts`
- `frontend/src/i18n/locales/zh/admin/audit.ts`
- `frontend/src/i18n/__tests__/auditLocales.spec.ts`
- `frontend/src/views/admin/__tests__/AuditLogView.i18n.spec.ts`
- `docs/workflow/tasks/audit-log-i18n-s141.md`
- `docs/workflow/qa-reports/audit-log-i18n-s141-qa.md`
- `docs/workflow/worker-results/audit-log-i18n-s141-result.md`

## Denied Paths

- `backend/**`, API contracts, migrations, generated output, deployment and
  containers
- unrelated views, global locale refactors, audit storage or action generation
- `knowledge/**`, primary worktree dirty files, and production runtime changes

## Constraints

- Do not translate or mutate the value sent in `filters.action` or any API query.
- Keep raw identifiers available through the existing action title/metadata.
- Keep unknown roles, auth methods and action segments visible as raw fallback.
- Do not add a new dependency or change the locale namespace shape.

## Acceptance Commands

```powershell
corepack.cmd pnpm --dir frontend exec vitest run src/i18n/__tests__/auditLocales.spec.ts src/views/admin/__tests__/AuditLogView.i18n.spec.ts
corepack.cmd pnpm --dir frontend run typecheck
corepack.cmd pnpm --dir frontend exec eslint src/views/admin/AuditLogView.vue src/i18n/locales/en/admin/audit.ts src/i18n/locales/zh/admin/audit.ts src/i18n/__tests__/auditLocales.spec.ts src/views/admin/__tests__/AuditLogView.i18n.spec.ts
corepack.cmd pnpm --dir frontend run build
git diff --check
git ls-files -u
```

## Stop Rules

- Stop if translating requires changing backend action values, query semantics,
  schema, or unrelated pages.
- Stop on any out-of-scope path, locale namespace regression, or unresolved test
  failure that is not isolated to this contract.

## Output

- Implementation and focused regressions within the allowlist.
- QA and worker-result documents with explicit PASS/FAIL/BLOCKED first lines,
  executed checks, unverified risks and recommendation.

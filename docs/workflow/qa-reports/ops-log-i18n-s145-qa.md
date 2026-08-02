### PASS: ops-log-i18n-s145

# QA Report

## Task ID

ops-log-i18n-s145

## Verdict

`PASS / source-level + production build`

## Contract Checked

- `docs/workflow/tasks/ops-log-i18n-s145.md`

## Findings

- 未发现明确问题。`OpsSystemLogTable` 的标题、健康摘要、运行时配置、筛选器、
  操作按钮、空态和表头均改用 `admin.ops.systemLogs` 或现有 `common` locale；
  英文会话不再落回历史中文文本。
- 中文 locale 保留当前界面语义和用户文案；英文 locale 与中文 locale 具有完全
  一致的 43 个 `systemLogs` 键，`cleanupSuccess` 保留 `{count}` 插值。
- S144 的稳定清理错误码映射、后端 detail fallback、请求字段、过滤语义、成功
  后刷新和运行时配置行为均保持不变。
- upstream `b4f38b092`/`d9e514f98` 未整提交移植；host、API-key、移动端卡片和
  账户展示等独立行为明确排除。

## Executed Checks

- diff reviewed: `yes`
- user-visible hardcoded-text scan: PASS（仅保留 2 处中文注释）
- real locale-module parity test: PASS（zh/en exact 43 keys，2 tests）
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsSystemLogTable.i18n.spec.ts` -> PASS (1 file, 2 tests)
- `corepack.cmd pnpm --dir frontend run typecheck` -> PASS
- changed-file ESLint -> PASS (0 warnings)
- `corepack.cmd pnpm --dir frontend run build` -> PASS (1127 modules)
- `git diff --check` -> PASS
- conflict-marker and unmerged-index check -> PASS
- exact staged allowlist audit -> PASS

## Unverified Risks

- 未在真实管理员登录态或部署实例中切换中英文并操作系统日志；本 Sprint 不包含
  部署、容器更新或生产运行态验证。
- 构建生成的 `backend/internal/web/dist` 和依赖目录均为忽略产物，未提交。

## Recommendation

可与 S144 一起合入发布分支并正常推送；部署后建议在中英文管理员会话分别打开
运维系统日志页，检查运行时配置、筛选、清理确认和清理错误提示。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 部署后的中英文系统日志页面 smoke。

## Knowledge Promotion

`none`

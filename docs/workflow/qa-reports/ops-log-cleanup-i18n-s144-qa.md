### PASS: ops-log-cleanup-i18n-s144

# QA Report

## Task ID

ops-log-cleanup-i18n-s144

## Verdict

`PASS / source-level + production build`

## Contract Checked

- `docs/workflow/tasks/ops-log-cleanup-i18n-s144.md`

## Findings

- 未发现明确问题。`OPS_SYSTEM_LOG_CLEANUP_FILTER_REQUIRED` 现在通过统一
  `extractApiErrorMessage` 错误码映射显示本地化文案；其它错误仍优先保留后端
  `detail`，没有改变请求参数、确认流程或成功处理。
- 中文和英文新增文案使用相同的 `admin.ops.systemLogs` 键。
- 本地实现是对 upstream `61a80114e` 的行为级适配，不直接 cherry-pick 上游
  上下文；当前本地组件的历史硬编码中文和截图对应的 `admin.audit` 审计页不在
  本 Sprint 范围内。

## Executed Checks

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsSystemLogTable.i18n.spec.ts` -> PASS (1 file, 1 test)
- `corepack.cmd pnpm --dir frontend run typecheck` -> PASS
- changed-file ESLint -> PASS (0 warnings)
- `corepack.cmd pnpm --dir frontend run build` -> PASS (1127 modules)
- `git diff --check` -> PASS
- conflict-marker scan and unmerged-index check -> PASS
- exact staged allowlist audit -> PASS

## Unverified Risks

- 未在真实管理员登录态、部署实例或容器中执行清理操作 smoke；本 Sprint 不包含
  部署、容器更新或生产运行态验证。
- 已生成的 `backend/internal/web/dist` 仍是忽略构建产物，未提交。

## Recommendation

可合入发布分支并正常推送；部署后建议在中英文管理员会话中分别触发无筛选清理，
确认稳定错误码显示对应本地化文案，再验证带筛选条件的成功路径。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 部署后的管理员系统日志清理中英文 smoke。

## Knowledge Promotion

`none`

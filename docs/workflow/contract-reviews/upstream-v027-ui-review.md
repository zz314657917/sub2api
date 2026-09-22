### PASS: upstream-v027-ui

# Contract Review

合同将 15 个低风险 UI 行为收束为 30 条明确的生产/测试路径，并要求每项在结果中标记为
已实现、已等价或不适用且附源代码证据。它排除了 backend、锁文件、依赖、支付计算、认证
边界和共享服务写入，允许的 `outputs/upstream-v027-ui/`、`v027-ui-smoke.html`、
`v027-ui-smoke.ts` 仅用于任务自身的 mock 浏览器 harness，且在最终交付前必须移除
harness。

Vitest 命令逐一列出全部 15 个允许新增的 spec 文件，随后运行 `vue-tsc -b` 与隔离的
Vite build 输出目录。浏览器验收使用实际导入组件、mock API、固定 `4328` 端口、
`v027-ui` session 和专属 profile
`outputs/upstream-v027-ui/profile`；合同要求记录 Vite PID、desktop/390px 截图、关闭
session、检查 task-owned profile/cliDaemon 归属并只结束该 Vite PID。这满足仓库的浏览器
隔离和精确清理门禁，且不要求真实支付、认证或数据写入。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- browser_profile_and_cleanup_explicit: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra`)
- base_commit_confirmed: `yes` (`d6e6c34718e6eee6388391b346f40e1492e81d57`)
- topology_confirmed: `yes`

## Approval

允许独立 Terra Developer 和独立 Terra QA 在本合同 allowlist 内执行。浏览器验收失败或
无法确认进程归属时必须以 `BLOCKED` 报告，不能按进程名进行全局清理。

## Amendment Review 2026-09-22

### FAIL: upstream-v027-ui amendment

`0838e0e610` 标记为 `NOT APPLICABLE` 的事实成立：不可变基线
`d6e6c34718e6eee6388391b346f40e1492e81d57` 的
`frontend/src/components/admin/user/` 只包含 GroupReplace、UserAllowedGroups、
UserApiKeys、UserBalanceHistory、UserBalance、UserCreate、UserEdit、
UserTicketMessage 等 modal；没有 `UserPlatformQuotaModal.vue` 或对应 spec。当前
工作树同样不存在两文件，且针对 admin source 的 scoped `rg` 未找到
`platform_quotas` 或等价的平台配额编辑器。因此不应为该上游修复新建本地功能，删除
该 item 的生产/test allowlist 是正确范围收缩。

但合同中的显式 Vitest 命令仍含
`src/components/admin/user/__tests__/UserPlatformQuotaModal.spec.ts`。这会使命令在
缺失文件上失败，且与“remaining 14 listed specs must execute”矛盾。修订前
`acceptance_commands_executable: no`，本次 amendment 不通过。

Required correction: 从 Allowed Paths 和唯一的 `pnpm.cmd exec vitest run` 命令删除
该不存在的 spec（生产路径也应移除）；保留其余 14 个 spec，随后重新提交 amendment
review。其他既有 UI `PASS` gate 不受本条未完成修订影响。

### PASS: upstream-v027-ui amendment re-review

Required correction 已完成：`UserPlatformQuotaModal.vue` 和
`UserPlatformQuotaModal.spec.ts` 均已从 Allowed Paths 移除，唯一显式 Vitest 命令也不再
引用该不存在的 spec，恰好保留 14 个明确的测试文件。`0838e0e610` 的不适用记录仍保留，
因此 Developer 不会误把缺失的本地管理员平台配额编辑器引入为新功能。

当前工作树中多数 `Test` 文件尚未创建是实施进度，不改变合同可执行性：它们都是 allowlist
中明确要求由 Developer 新增的测试路径。该事实不得被解释为可以缩减命令或降低门禁；最终
QA 必须确认全部 14 个文件存在并实际执行该完整 Vitest 命令，失败则按合同的 baseline
证据规则处理。

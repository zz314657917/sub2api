---
task_id: user-proxy-smart-input-s113
status: contract-approved
owner: Codex
qa_mode: runtime
---

# Task Contract

## Goal

让用户侧“我的代理”弹窗和管理员批量代理导入识别带协议的冒号分隔凭据格式，例如
`socks5h://host:port:username:password`，同时保留标准
`scheme://username:password@host:port` 格式，并将结果提交为现有结构化代理字段。

## Success Criteria

- 冒号分隔格式能正确识别协议、主机、端口、用户名和密码，`socks5h` 原样保留。
- 标准 URL、无认证 URL、IPv6（带方括号）和凭据中的额外冒号不被错误拆分。
- 用户侧单条输入可填充表单并保存，多行输入可校验后批量创建；空名称时自动生成可用名称。
- 管理员批量导入复用同一解析规则，现有重复检测和无效计数行为不变。
- 解析器单元测试、前端 typecheck、生产构建和 diff 静态门禁通过。

## Context

- Repo: `F:/mcplugins/sub2api`
- Read first: `docs/workflow/status.md`, `docs/workflow/spec.md`
- Backend user/admin proxy handlers already accept `http`, `https`, `socks5`, and `socks5h`.

## Allowed Paths

- `frontend/src/utils/proxyInput.ts`
- `frontend/src/utils/__tests__/proxyInput.spec.ts`
- `frontend/src/views/user/__tests__/MyAccountsView.importFile.spec.ts`
- `frontend/src/views/user/MyAccountsView.vue`
- `frontend/src/views/admin/ProxiesView.vue`
- `frontend/src/i18n/locales/zh/myAccounts.ts`
- `frontend/src/i18n/locales/en/myAccounts.ts`
- `frontend/src/i18n/locales/zh/admin/proxies.ts`
- `frontend/src/i18n/locales/en/admin/proxies.ts`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/user-proxy-smart-input-s113.md`
- `docs/workflow/qa-reports/user-proxy-smart-input-s113-qa.md`
- `knowledge/tasks/current-task.md`

## Denied Paths

- `backend/**`
- `frontend/src/types/**`
- `knowledge/**` (except the approved current-task snapshot path above)
- deployment, database migrations, containers, and unrelated dirty files

## Constraints

- 保持现有结构化 API 和代理协议白名单不变。
- 不记录或输出代理凭据到日志、workflow 或知识库。
- 不回滚或覆盖已有用户改动。
- 解析失败必须显式反馈，不能静默直连或提交空字段。

## Acceptance Commands

```powershell
cd F:/mcplugins/sub2api/frontend
npm.cmd run test:run -- src/utils/__tests__/proxyInput.spec.ts src/views/user/__tests__/MyAccountsView.importFile.spec.ts src/__tests__/integration/proxy-data-import.spec.ts -- --pool=threads --maxWorkers=1 --minWorkers=1
npm.cmd run typecheck
npm.cmd run build
cd F:/mcplugins/sub2api
git diff --check
```

## Output

- 变更文件、测试命令及结果、未验证风险在最终回复中说明。

## Stop Rules

- 若需要修改后端协议、数据库或生产配置，停止并回 Codex 重新裁决。
- 若发现现有用户改动与允许路径冲突，不覆盖用户改动。

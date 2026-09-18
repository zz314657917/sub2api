---
type: contract-review
scope: task
status: approved
task_id: upstream-api-key-provider-filter
verdict: PASS
base_commit: 0ef64d868b35351c2a878cf022d2b93c447ea296
reviewer: evaluator
last_verified: 2026-09-18
---

### PASS: upstream-api-key-provider-filter

# Contract Review

## Verdict

`PASS` — 可派发独立 Terra Developer。此结论只批准隔离 worktree 内的实现，不代表 QA、合入、推送或部署批准。

## Findings

- 已对比上游 `55a95d4c674f0624fad1c31ee4afd0ec03e2a53f` 的 5 个改动路径。其 `keyGroupProviders.ts` 含 `minimax`、`composite`、`opencode_go`，而本地基线的 `GroupPlatform` 精确只有 `anthropic`、`openai`、`gemini`、`antigravity`、`grok`、`kimi`、`zhipu`、`deepseek`。合同已明确要求按本地 union 映射，故必须行为移植，不能整体 cherry-pick；运行时未知字符串仍须归入 `other` 并有测试覆盖。
- 已核对本地 `KeysView` 的多分组 owner：默认组下拉使用 `groupOptions`，路由行使用独立的 `getRouteGroupOptions`，且创建和编辑提交都会携带 `multi_group_routes`；该路径保存顺序、启用状态和 `first_response_timeout_seconds`。合同将 provider 过滤限制在“创建默认组”并要求路由行仍用完整授权组，足以避免把跨平台 failover 路由误过滤。
- 编辑入口从 `key.route_groups` 和 `key.multi_group_routes` 重建表单，关闭/重开会重置表单。合同已要求编辑保留原组、创建切换/延迟加载/重开时清除失效选择；测试必须通过真实组件的 create、edit、submit 断言 API 参数，证明 provider 状态不会进入 payload。
- 成功标准、精确 allowlist 与 denied paths 完整。新增的 `docs/workflow/agent-matrix.md` 仅恢复仓库已声明的矩阵，明确 Developer 与 QA 均为独立 `gpt-5.6-terra` 角色，属于 controller-only 流程证据，不扩展业务范围。
- 所列 Vitest、`vue-tsc`、ESLint、Vite build 和 diff 检查均有对应本地二进制入口；acceptance 明确禁止以“无测试可跑”代替通过，并包含误导性名称、未知平台、空类别、延迟加载、重开、跨平台路由和超时回归。

## Gate Checks

- success_criteria_testable: `yes`
- allowed_paths_explicit: `yes`
- denied_paths_explicit: `yes`
- acceptance_commands_executable: `yes`
- worker_model_confirmed: `yes` (`gpt-5.6-terra` for separate Developer and QA)
- base_commit_confirmed: `yes` (`0ef64d868b35351c2a878cf022d2b93c447ea296`)
- upstream_behavior_checked: `yes` (`55a95d4c6`)
- local_multi_group_non_interference_specified: `yes`

## Preconditions and Limits

- `pnpm` frozen 安装因 esbuild/vue-demi 的 build-script policy 退出；其自动生成的 `frontend/pnpm-workspace.yaml` 已移除，manifest/lock 无变化。现有本地 `.cmd` 二进制可执行 contract 所列测试、类型检查、lint 与 build；最终 QA 仍须记录实际命令与退出码，且不得暂存任何安装生成的 package/config 文件。
- 最终 QA 仍须实际执行全部验收命令并审查完整 diff；本评审没有运行业务测试，也未验证真实 provider、浏览器、部署或推送。

## Scope Amendment Review — locale owners

### PASS: upstream-api-key-provider-filter locale-path-amendment

- 本地 `frontend/src/i18n/locales/en.ts` 与 `zh.ts` 已分别直接 import 并以顶层 `keys` 装配 `./en/keys`、`./zh/keys`；`KeysView.vue` 使用的 `t('keys.*')` 因而只能由这两个文件提供。上游 `dashboard.ts` 是其不同仓库结构中的 owner，在本地修改它不会满足新增 provider 文案。
- contract 已将 allowlist 和 ESLint 精确替换为 `frontend/src/i18n/locales/en/keys.ts`、`frontend/src/i18n/locales/zh/keys.ts`，没有扩大文件数量、API、依赖或运行时行为范围；`en.ts`、`zh.ts` 装配文件保持 denied，且不需要改动。
- 此修订恢复可执行的本地 locale 路径，并保持此前所有 provider 分类、多分组路由非干扰、payload 隔离与独立 Terra QA 门禁。可将该修订后的 contract 重新标记为 `approved/PASS` 后派发 Developer。

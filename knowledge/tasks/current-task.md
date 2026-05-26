# 当前任务快照

最后更新：2026-05-27 01:56 +08:00

## 背景

- 项目主仓库：`F:/mcplugins/sub2api`。
- 当前任务：用户侧使用记录改成更适合运维扫描的表格，重点补齐“实际消费分组”可见性；同时整理新人 API Key 引导弹窗体验。
- 用户反馈：API 密钥支持多路由后，使用记录里看不到某条消耗到底归属哪个分组 token。
- 约束：不新增数据库字段；复用 `usage_logs.group_id`、`UsageLog.group` 和用户侧 `/groups/available`。

## 当前目标

- 已完成：用户侧 `/api/v1/usage` 和 `/api/v1/usage/stats` 支持按 `group_id`、`billing_mode` 等筛选，统计卡片与表格口径一致。
- 已完成：前端使用记录新增消费分组列、分组筛选、更多筛选、列设置、行详情和 CSV 新字段。
- 已完成：默认表格变为“常用列优先、低频列可展开”的运维视图。
- 已完成：当前改动已整理并推送到 `origin/main`；使用记录主线提交为 `a5e998cb4 feat(user): improve usage visibility and onboarding`，新人引导后续收口到 `e73c60474 fix(user): update onboarding model tags`。

## 本次已完成

- 后端：
  - `UsageHandler.List` 解析 `group_id` 和 `billing_mode`，写入 `usagestats.UsageLogFilters`。
  - `UsageHandler.Stats` 改为构造完整 `UsageLogFilters` 并调用 `GetStatsWithFilters`，支持 `group_id`、`model`、`request_type`、`stream`、`billing_type`、`billing_mode`。
  - 用户侧仍强制叠加当前登录用户 `UserID`，`api_key_id` 仍做归属校验。
- 前端：
  - `UsageQueryParams` 增加 `billing_mode?: string`。
  - `usageAPI.getStatsByDateRange` 改为接收筛选条件对象。
  - `UsageView.vue` 新增消费分组筛选，来源为 `userGroupsAPI.getAvailable()`。
  - 表格新增 `group` 列，使用 `GroupBadge` 展示 `row.group.name/platform`，倍率使用本行 `row.rate_multiplier`。
  - 默认显示列改为 API 密钥、消费分组、模型、类型、计费模式、Token、费用、首 Token、耗时、时间、详情。
  - 推理强度、端点、User-Agent 默认隐藏，可通过列设置恢复，持久化到 `localStorage` key `usage-visible-columns:v1`。
  - `User-Agent` 显示时截断，完整内容进入行详情。
  - 新增本地详情抽屉，展示路由信息、请求信息、Token 明细、费用明细、耗时和 User-Agent。
  - 顶部统计卡片下方新增当前口径摘要。
  - 成本卡主值改为紧凑显示，完整值放 `title`。
  - CSV 按当前筛选导出，并新增 `Group Name`、`Group ID`、`Request ID`、`User-Agent`。
- 测试：
  - 更新后端用户 usage handler 测试，覆盖 `group_id`、`billing_mode` 和 stats filter。
  - 重写/扩展前端 `UsageView.spec.ts`，覆盖分组列、筛选、更多筛选、列设置、行详情和 CSV 新字段。
- 新人引导：
  - Dashboard 加载福利概览后，把账户余额或新手试用额度传给 API Key 引导弹窗。
  - 引导弹窗新增头图、福利文案、联系支持按钮和对应测试。
  - 后续提交继续简化福利文案，展示 `GPT-5.5`、`Image2`、`Claude`、`OpenClaw` 能力标签，并同步中英文 i18n 与组件测试。

## 已确认事实

- `usage_logs.group_id` 已存在，不需要 migration。
- usage repository 的 `ListWithFilters` 会 hydrate `Group`，DTO 已包含 `group`。
- 多路由认证链路会把实际命中的 route group 写入 APIKey clone 的 `GroupID/Group`，usage log 记录的是实际消费分组。
- 历史缺失 `group_id` 的记录继续显示 `-`；如果有 `group_id` 但分组未 hydrate，则显示 `#id`。
- 管理员使用记录页本轮未重做，只保留现有分组列。

## 待验证点

- 动作：用真实用户登录态打开使用记录页，选择一个消费分组和计费模式。
  验证：表格列表与顶部统计卡片数值同时按当前分组/计费模式变化，CSV 导出只包含当前筛选数据。
- 动作：查看一条多路由 API Key 产生的请求记录。
  验证：消费分组列展示实际扣费分组，详情抽屉中倍率与该行扣费倍率一致。
- 动作：在窄屏或长 User-Agent 数据下查看表格。
  验证：默认行高不被 User-Agent 撑高；打开列设置恢复 User-Agent 后内容截断且详情中可看完整值。

## 当前结论

- 用户现在可以直接在使用记录表格里看到每条消耗归属哪个消费分组。
- 分组筛选和统计口径已统一，不会出现表格按分组但卡片仍按全部数据统计的错位。
- 详情抽屉承担低频深挖信息，默认表格更适合扫描。
- 当前 `main` 与 `origin/main` 已同步到 `e73c60474`，工作区干净。

## 最近归档

- 已写入 `knowledge/tasks/timeline.md`：`2026-05-27 01:56 +08:00 - 用户使用记录与新人引导收口复核`。
- 多智能体审查最近提交 `HEAD~1..HEAD` 后未发现运行时代码明确回归；发现两个测试/CI 侧风险：默认 CI 未覆盖 `UserApiKeyOnboardingDialog.spec.ts`，以及 `toContain('Claude')` 断言不够精确。

## 下一步

- 若继续优化真实运营体验：建议用浏览器登录态做一次截图级验收，重点看筛选栏在 1440px 和超宽屏下的按钮拥挤度。
- 若用户要扩展管理员页：可复用本轮列设置、详情抽屉和统计口径摘要，但管理员页要额外考虑用户、账号、IP、上游映射链等字段。

## 验证记录

- 已通过：
  - `go test ./internal/handler -run UsageHandler`，工作目录 `F:/mcplugins/sub2api/backend`
  - `corepack.cmd pnpm exec vitest run src/views/user/__tests__/UsageView.spec.ts src/components/user/dashboard/__tests__/UserApiKeyOnboardingDialog.spec.ts src/components/common/__tests__/DateRangePicker.spec.ts`，工作目录 `F:/mcplugins/sub2api/frontend`
  - `corepack.cmd pnpm run typecheck`，工作目录 `F:/mcplugins/sub2api/frontend`
  - `git diff --check`
- 未做：
  - 未启动浏览器做真实登录态视觉验收；当前任务已用单测和类型检查覆盖功能逻辑。

# 当前任务快照

最后更新：2026-07-23 19:14 +08:00

## 背景

- S109 在隔离 worktree `E:/codex-worktrees/sub2api-pricing-s109` 手工迁移上游模型定价与图片输入计费修复；主工作树的用户脏改保持未触碰。
- 本地与上游架构已明显分叉，本轮只迁移可独立验证的 pricing/image usage 语义，不整体 merge `upstream/main`。

## 当前目标

- 对齐 LiteLLM 和渠道图片输入价、Claude 模型名、GLM-5.2 官方价、OpenAI OAuth 图片 usage，以及 usage log/API/UI 对账。
- Contract：`docs/workflow/tasks/upstream-model-pricing-alignment-s109.md`。

## 本次已完成

- 图片输入 token/费用已贯通 billing、账号统计、repository、migration、DTO/API 和管理员/用户 usage 页面。
- hosted `tool_usage.image_gen` 保留独立图片 token；普通 usage 仍做有界解析和总量钳制。
- 长上下文图片 token 比例拆分使用无溢出乘除，总输入使用非负饱和加法。
- 可用渠道 fallback 和 token/image 模式均展示 LiteLLM `InputCostPerImageToken`。

## 已确认事实

- composite 别名计费缺少本地约 2,000 行平台/gateway 前置链，按 stop rule 留到独立 Sprint。
- 上游已刷新到 `cb24522dd` / `v0.1.164`，没有新的模型单价补丁。
- migration 193/194 连续；usage 单条 INSERT 为 53 个数据参数，timezone 为 `$54`。

## 待验证点

- 未执行生产数据库实迁、真实 OpenAI OAuth 图片请求或登录态浏览器 smoke。
- `go test -tags=unit ./internal/service` 仍被既有重复 `stringPtr`、旧 billing 测试签名和缺失 Grok helper 阻断；不得将聚合 unit suite 写为 PASS。

## 当前结论

- `PASS / source-only`：四项最终复审 finding 已修复，完整 S109 backend/frontend 门禁和三路审查覆盖均通过。
- 当前待精确提交并从隔离分支推送到远端 `main`；未部署、未更新容器。

## 下一步

1. 精确暂存 contract allowlist 文件和 workflow/handoff 产物；验证：cached path、cached diff、`git diff --cached --check`。
2. fetch 后确认 `origin/main` 未移动，提交并推送 `HEAD:main`；验证：`HEAD`、`origin/main`、`git ls-remote origin refs/heads/main` 一致。
3. 用 publication 提交记录远端 parity；不更新或清理 dirty 主工作树，不部署、不更新容器。

## 验证记录

- focused/default-tag service、宽范围 pricing/billing/image/record-usage、repository、migration integration、`/api/v1/usage` contract 全部 PASS。
- 前端 Vitest `3 files / 33 tests`、typecheck、production build（1089 modules）、scoped ESLint PASS。
- gofmt、56 路径 allowlist、diff、冲突标记和未合并索引门禁 PASS；三路智能体审查覆盖无阻断 finding。

## 历史快照

## S105 当前目标

- 在管理员账号管理中增加 OpenAI 套餐分类与筛选：`Plus / Pro / K12 /
  Team / Free / 其他 / 未识别`。
- 列表、共享账号列表、按筛选批量编辑、按筛选批量共享状态和按筛选导出使用
  同一个 `credentials.plan_type` 条件；`share_display_tier` 只影响展示。
- Contract：`docs/workflow/tasks/admin-account-plan-type-filter-s105.md`。

## S105 当前状态

- 已完成并通过 `PASS / source-only`：管理员账号列表可按 `Plus / Pro / K12 /
  Team / Free / 其他 / 未识别` 筛选，列表、共享账号、按筛选批量编辑、按筛选
  共享状态和按筛选导出使用同一个 `credentials.plan_type` 条件。
- 仓储集成、聚焦 service/handler、全量 handler/repository、Vitest `13/13`、
  typecheck、生产构建（1089 modules）、格式、diff、冲突、allowlist 和未合并
  索引门禁通过。
- 全量 service 仍只失败于既有 peak-rate 时区基线；测试启动日志为 UTC，
  与 S105 无关。未使用真实 K12 凭据或登录态浏览器 smoke。
- 不改数据库 schema/migration，不重写凭据，不增加手动套餐编辑，不改导入/OAuth
  enrichment、调度、网关、计费、部署或容器。
- 当前 `main` 比 `origin/main` 领先 S104 本地提交 `cb4e0b4aa`；S105 本轮作为
  本地提交保留，不推送。

## S104 当前目标

- 手工迁移上游 `d0b8760eb`：保留 token/JWT 已识别的 OpenAI/K12
  `plan_type`，并在 `accounts/check` fallback 中跳过停用或过期 workspace。
- Contract：
  `docs/workflow/tasks/upstream-openai-inactive-workspace-plan-s104.md`。

## S104 当前结论

- 已完成并通过 `PASS / source-only`：K12 等 token-derived plan 不再被
  无效 workspace 的内部 billing plan 覆盖；当前 plan 为空时仍可使用有效
  `accounts/check` 候选回退。
- 4 个聚焦测试连续 10 轮通过，更宽的 OpenAI OAuth/account-info 回归通过，
  gofmt、diff、冲突、allowlist 和未合并索引门禁通过。
- 未使用真实 K12 凭据，未部署、未更新容器；`at-*` PAT 和 K12
  `gpt-5.6-sol` 503 不属于本 Sprint。

## S103 当前目标

- 手工迁移上游 `3401a971a` 的适用前端硬编码文案本地化：AppHeader 的两个
  aria-label、CustomPageView 的 404 文案，并补齐中英文 common locale。
- Contract：`docs/workflow/tasks/upstream-frontend-hardcoded-i18n-s103.md`。

## S103 当前结论

- 已完成并通过 `PASS / published`：AppHeader 两个 aria-label、
  CustomPageView 404 文案和中英文 common locale 已本地化。
- 聚焦测试 `2 files / 9 tests`、typecheck、生产构建（1089 modules）、
  静态目标、diff/冲突/索引门禁均通过。未做登录态浏览器 smoke、部署或容器更新。
- 功能与 QA 提交 `f2720fdbb` 及 workflow publication 提交 `abd590761` 均已
  推送，并与 `origin/main`、远端 `main` 一致；本 Sprint 已全部收口。
- 支付自定义方式 placeholder 因本地没有对应 UI/字段而跳过，后续若补齐前置
  功能需另开 contract；本次不加入 orphan locale key。
- 只允许 4 个现有前端业务/locale 文件及 workflow 记录；支付自定义方式
  placeholder 因本地没有对应 UI/字段而跳过，不整体合并上游，不改 backend、
  API、计费、数据库、部署或容器。

## S102 当前目标

- 将渠道监控时间线柱子改为更易辨识的语义色：正常绿、降级橙、失败/错误红、
  未测试灰；不改状态判定、柱高、顺序、数量和响应式宽度。
- Contract：`docs/workflow/tasks/channel-monitor-timeline-status-colors-s102.md`。

## S102 当前结论

- 已完成并通过 `PASS / published`：`operational` 使用 `bg-emerald-500`，
  `degraded` 使用 `bg-orange-500`，`failed/error` 使用 `bg-red-500`。
- `MonitorTimeline` 专用测试 `2/2`、typecheck、生产构建（1089 modules）、
  diff/冲突/索引门禁和 1000px 浏览器截图均通过。
- 本地预览仍为 `http://127.0.0.1:62080/monitor`；浏览器使用合成 API 数据，
  本机没有 8080 后端，因此真实 API/auth 未验证。S101/S102 组合提交
  `fd8719b4c` 已推送；未部署、未更新容器。

## S101 当前结论

- S101 `channel-monitor-timeline-overflow-s101` 已完成：监控卡时间线不再被
  60 根柱条的固定最小宽度撑出卡片，桌面和移动布局均保持 60 根完整柱条。
- `MonitorTimeline.vue` 使用 `min-w-0 w-full`，柱条使用 `min-w-0`；新增
  `MonitorTimeline.spec.ts` 覆盖柱条数量、顺序、状态编码和收缩约束。
- 验证通过：专用测试 `2/2`，连同 `ChannelStatusView.capacityPools` 为
  `2 files / 9 tests`，`vue-tsc`，生产构建（1089 modules），diff/冲突/索引门禁，
  1706px 桌面和 390px 移动 Playwright 截图及 DOM 边界检查。
- 本地 Vite 预览：`http://127.0.0.1:62080/monitor`。视觉验证使用 Playwright
  拦截的合成认证与监控数据；本机没有 8080 后端监听，因此真实 API/auth 未验证。
- S101 与 S102 已在组合提交 `fd8719b4c` 中推送到 `origin/main`；没有部署或
  容器更新。

## 下一步

1. 用户单独授权后再推送 S104/S105；验证：`HEAD`、`origin/main` 和
   `git ls-remote` 一致。
2. 需要确认真实 K12 效果时，另做脱敏账号运行态 smoke，并单独授权容器更新。

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- S98/S99 组合功能提交 `34b1844ab` 和 workflow 收口提交 `be5d4114d` 已推送。
- 当前按已批准的上游差异计划推进独立小步迁移，不整体 merge `upstream/main`。

## 当前目标

- S100：移植上游 `d0fa8c63f`，将订阅剩余天数从向下截断改为正数向上取整。
- 保持 24 小时 duration 语义，不扩展成用户时区日历日，不触碰前端、计费、持久化或部署。

## 本次已完成

- `DaysRemaining()` 复用可确定测试的 `daysRemainingAt(now)`。
- 过期和恰好到期返回 `0`；不足一天返回 `1`；超过整天的余数向上取整。
- progress DTO 的剩余天数断言收敛为精确值。
- 上游新增测试移除 `unit` build tag，使其进入本地默认 service 测试集，避开既有 unit-tag 聚合编译漂移。

## 验证记录

- 测试发现同时列出 `TestCalculateProgress_BasicFields` 和 `TestUserSubscriptionDaysRemainingAt`。
- focused 和 broader progress service tests PASS。
- `gofmt -d`、`git diff --check`、冲突标记、精确路径和未合并索引检查 PASS。

## 当前结论

- `PASS / published`：S100 功能提交 `e567ecbb5` 已推送到 `origin/main`。
- 未部署、未更新容器；24 小时 duration 语义未改为日历日。

## 下一步

1. 进入下一独立 Sprint，手工迁移上游零散硬编码文案本地化。
2. Responses SSE 热路径优化随后单独处理，不引入上游整套 gateway 文件拆分。

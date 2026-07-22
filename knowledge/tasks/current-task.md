# 当前任务快照

最后更新：2026-07-22 21:16 +08:00

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

1. 按授权发布 S104 精确补丁；不要混入 PAT 或 Agent Identity 功能链。
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

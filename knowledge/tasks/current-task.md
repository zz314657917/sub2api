# 当前任务快照

最后更新：2026-05-28 11:58 +08:00

## 背景

- 项目主仓库：`F:/mcplugins/sub2api`。
- 用户要求：整理最近更新并提交。
- 当前工作区已有一组后端共享展示窗口重置/用量口径改动，以及一组前端控制台顶部公告轮播改动。
- 本轮目标是复核改动范围、跑相关验证、同步任务记录并创建提交。

## 当前目标

- 后端：账号配额重置时同步重置共享展示 5h/7d 伪装窗口基线，并让容量池展示用量可按每个窗口自己的起始时间统计。
- 前端：控制台顶部在已登录且有公告时展示最多 3 条公告轮播，点击复用公告铃详情弹窗。
- 提交前保持改动范围收敛，不混入无关文件。

## 本次已完成

- 后端新增 `GetAccountUsageCostsSinceByWindow`，支持按账号和窗口起点批量统计 `usage_logs` 成本。
- 容量池展示用量读取 `share_display_5h_start` / `share_display_7d_start`；缺失时回退到固定 5h / 7d 窗口。
- `ResetQuotaUsed` 清零本地配额时同步清零 `share_display_5h_used` / `share_display_7d_used` 并刷新对应 start 字段，同时保留 `codex_*` 真实上游快照。
- 账号创建、编辑和用户侧导入保存时，关闭共享展示会清理 `share_display_*_start` 残留字段。
- 前端新增 `HeaderAnnouncementCarousel.vue`，接入 `AppHeader.vue`，并通过 `AnnouncementBell` 暴露的 `openDetail` 打开公告详情。
- 已补后端 repository/service 测试和前端 header 公告轮播测试。

## 已确认事实

- 当前改动文件集中在 `backend/internal/repository/account_repo.go`、`backend/internal/service/user_account_service.go`、共享展示相关测试、账号弹窗/用户账号页 extra 清理、`AnnouncementBell`、`AppHeader` 和新增公告轮播组件。
- `knowledge/tasks/current-task.md` 与 `knowledge/tasks/timeline.md` 是 tracked 文件，本轮作为最近更新整理一起提交。
- `git diff --check` 已通过，没有发现空白错误。

## 待验证点

- 动作：如需要视觉验收，用真实登录态打开控制台宽屏页面。
  验证：顶部中间显示公告轮播，hover/focus 暂停，点击后打开公告详情，不挤压右侧操作区。
- 动作：如需要运行时数据验收，准备带共享展示窗口的账号，执行配额重置后刷新容量池。
  验证：展示窗口从重置时间重新累计，本地伪装用量归零，`codex_*` 上游快照不被清除。

## 当前结论

- 当前代码和文档整理已通过相关窄范围自动化验证。
- 尚未做真实浏览器视觉复核，也未连真实数据库做共享展示窗口重置后的运行时验收。

## 下一步

- 动作：暂存本轮源码、测试和任务记录并创建提交。
  验证：`git status --short --branch` 显示工作区干净或仅剩用户未纳入的无关改动。
- 动作：如继续收口，做一次真实登录态宽屏控制台截图和一次共享展示窗口重置数据验收。
  验证：公告轮播交互和容量池窗口口径与预期一致。

## 验证记录

- 已执行检查：
  - `gofmt -w backend/internal/repository/account_repo.go backend/internal/repository/account_repo_quota_reset_test.go backend/internal/service/user_account_service.go backend/internal/service/account_sharing_test.go`
  - `go test ./internal/repository -run "ResetQuotaUsed|GetAccountUsageCostsSinceByWindow" -count=1`，工作目录 `F:/mcplugins/sub2api/backend`，通过。
  - `go test ./internal/service -run "CapacityPools|ShareDisplay" -count=1`，工作目录 `F:/mcplugins/sub2api/backend`，通过。
  - `corepack.cmd pnpm exec vitest run src/components/layout/__tests__/HeaderAnnouncementCarousel.spec.ts src/components/layout/__tests__/AppHeader.spec.ts`，工作目录 `F:/mcplugins/sub2api/frontend`，2 个测试文件 6 个测试通过。
  - `npm.cmd run typecheck`，工作目录 `F:/mcplugins/sub2api/frontend`，通过。
  - `git diff --check`，通过。

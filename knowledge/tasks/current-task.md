# 当前任务快照

最后更新：2026-06-24 01:37 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/upstream-v0138-small-patches`。
- 本轮用户要求：排行榜增加“模型榜”，先放假数据看效果；模型榜图标改为模型商图标；右侧增加“增长”和“排名变化”；把模型榜排名后面的 Token 文本改为百分比；按截图在右侧再补 Token 指标卡；最后把“Token 消耗榜 / 模型榜”切换收进排行榜卡片，去掉中间大空白，并更新本地容器。
- 当前工作树是 mixed dirty tree，另有 S20 上游小补丁、Payment UI、knowledge 等无关脏改；提交时必须外科式 staging，不能 `git add .`。

## 本次已完成

- 后端排行榜响应新增 `model_ranking` / `total_models`。
- 后端新增模型榜聚合：按当前周期统计模型请求数、输入 Token、输出 Token、总 Token，并与上一同长度周期对比。
- 模型榜新增趋势字段：
  - `growth_percent`：相对上一周期 Token 增长百分比，保留 1 位小数。
  - `rank_change`：上一周期排名减当前排名；正数表示上升，负数表示下降。
- 空模型榜样例数据只在 `SUB2API_LEADERBOARD_SAMPLE_MODELS=true`，或 `SERVER_MODE=debug` 且 Gin debug 模式下启用。
- 前端排行榜增加 Token 榜 / 模型榜切换。
- 模型榜行内图标改为 `ModelIcon`，按模型名显示模型商图标。
- 模型榜右侧指标区已按截图改为三列指标卡：
  - Token：例如 `7.29B`、`194.35M`。
  - 增长：例如 `+28.4%`、`-77.7%`、`—`。
  - 排名变化：例如 `↑ 1`、`↓ 1`、`—`。
- 模型榜移动端已做响应式堆叠，避免右侧指标挤压主内容。
- 模型榜条形区域后面的值已从绝对 Token 改为当前可见模型榜 Token 占比：
  - 正常显示 1 位小数百分比，例如 `83.3%`、`16.7%`。
  - `0 < share < 0.1` 显示 `<0.1%`。
  - tooltip / aria-label 仍保留绝对 Token，并追加百分比。
- “Token 消耗榜 / 模型榜”切换按钮已从排行榜卡片外移入榜单卡片内部：
  - 空态、Token 榜、模型榜三种分支都在卡片顶部内嵌同一组切换按钮。
  - 左侧榜单外层去掉额外 `space-y-3` 间距，避免 tab 和榜单卡片之间出现割裂大空白。
- 已补中英文 i18n、前端类型、后端/前端定向测试。
- 已修复模型榜 SQL 歧义导致的排行榜接口 500：最终 SELECT / WHERE / ORDER BY 改为显式 `ranked.*` 别名，避免 `model` 在 `ranked` 与 `previous_ranked` join 后被 PostgreSQL 判定为 ambiguous。

## 本地容器

- 当前在线容器：`sub2api`，不是 `sub2api-dev`。
- 地址：`http://127.0.0.1:62080`。
- 当前镜像：`sub2api:codex-20260624-0132-leaderboard-tabs-card`。
- 当前镜像 ID：`sha256:08635c5210a05c1136fee41aeff556f486c01b7a85ff41b2192f592939aeb658`。
- `sub2api:local` 已指向同一新镜像。
- 当前容器状态：`running healthy`，端口 `127.0.0.1:62080->8080`。
- 本次 tab 收进卡片更新前容器备份：`sub2api-before-leaderboard-tabs-card-20260624-0132`，旧镜像 `sub2api:codex-20260624-0120-payment-layout-wide`。
- Payment 宽布局更新前容器备份：`sub2api-before-payment-layout-wide-20260624-0120`，旧镜像 `sub2api:codex-20260624-0115-leaderboard-token-card`。
- 本次 Token 卡更新前容器备份：`sub2api-before-leaderboard-token-card-20260624-0115`，旧镜像 `sub2api:codex-20260624-0051-payment-layout`。
- 百分比显示更新前容器备份：`sub2api-before-leaderboard-percent-20260624-0050`，旧镜像 `sub2api:codex-20260624-0036-leaderboard-sqlfix`。
- 更早 SQL 修复前容器备份：`sub2api-before-leaderboard-sqlfix-20260624-0036`，旧镜像 `sub2api:codex-20260624-0011-leaderboard-trend`。
- `sub2api-postgres` 和 `sub2api-redis` 未重建，数据容器未动。
- Docker 更新锁已获取并释放。

## 验证记录

- `go test ./internal/handler -run "TestUsageHandlerDashboardLeaderboard" -count=1` 通过。
- `go test ./internal/repository -run "TestUsageLogRepositoryGetLeaderboardModelRanking|TestUsageLogRepositoryGetUserLeaderboard" -count=1` 通过。
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/LeaderboardView.spec.ts src/__tests__/leaderboard-theme.spec.ts` 通过，2 files / 26 tests。
- `corepack.cmd pnpm --dir frontend run typecheck` 通过。
- `corepack.cmd pnpm --dir frontend run build` 通过，仅有既有 Browserslist、Node `DEP0190`、Vite dynamic/static import chunk warning 和大 chunk 提示。
- `git diff --check -- frontend/src/views/user/LeaderboardView.vue frontend/src/views/user/__tests__/LeaderboardView.spec.ts` 通过。
- `GET http://127.0.0.1:62080/health` 返回 200。
- `docker exec sub2api /app/sub2api --version` 输出 `Sub2API 0.1.126 (commit: codex-leaderboard-tabs-card, built: 2026-06-23T17:33:03Z)`。
- 当前容器日志无 `panic`、`pq:` 或 `ERROR`；`/leaderboard` 未登录访问返回 200 HTML。
- 已确认 `http://127.0.0.1:62080/assets/LeaderboardView-CJliFWp5.js` 包含 `leaderboard-ranking-card-toolbar` 和 `leaderboard-model-token`。
- 已确认 `http://127.0.0.1:62080/assets/LeaderboardView-ClHQr9PL.css` 包含 `leaderboard-ranking-card-toolbar` 和 `leaderboard-token-ranking-card`。
- 早前临时登录后请求 `GET /api/v1/usage/dashboard/leaderboard?period=day` 返回 200，`code=0`，`model_ranking_count=3`，`total_models=3`，随后已调用登出接口撤销本次临时 refresh token。

## 未验证点

- 内置浏览器打开 `http://127.0.0.1:62080/leaderboard` 后标题为 `Login - 落叶网络`，未持有登录态，因此未完成真实登录态视觉截图验收。
- 本轮收尾按精确 staging 分批提交；源码变更、知识记录和交接快照不混入无关未审文件。

## 提交边界

如用户要求提交，只 stage 本轮 leaderboard 相关文件：

- `backend/internal/handler/usage_handler.go`
- `backend/internal/handler/usage_handler_leaderboard_test.go`
- `backend/internal/pkg/usagestats/usage_log_types.go`
- `backend/internal/repository/usage_log_repo.go`
- `backend/internal/repository/usage_log_repo_request_type_test.go`
- `backend/internal/service/usage_service.go`
- `frontend/src/types/index.ts`
- `frontend/src/views/user/LeaderboardView.vue`
- `frontend/src/views/user/__tests__/LeaderboardView.spec.ts`
- `frontend/src/__tests__/leaderboard-theme.spec.ts`
- `frontend/src/i18n/locales/zh/leaderboard.ts`
- `frontend/src/i18n/locales/en/leaderboard.ts`

不要混入 Payment、S20、`knowledge/05-current-focus.md`、`knowledge/studio-bridge-luoye.md` 或其它无关改动，除非用户明确要求。

## 下一步

- 用户浏览器刷新 `http://127.0.0.1:62080/leaderboard` 后，切到“模型榜”，应看到“Token 消耗榜 / 模型榜”切换按钮被包进排行榜卡片顶部，下面紧跟 `模型 Top 10 / 总榜` 标题和榜单内容；条形区域后面是百分比，右侧指标卡依次为 `Token / 增长 / 排名变化`。
- 如需视觉复核，先登录本地页面，再确认模型商图标、右侧三列指标、tab 内嵌卡片和移动端布局。
- 合并/推送前继续按 cached diff 审核 staged 文件，并运行 `git diff --cached --check`。

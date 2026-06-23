# 当前任务快照

最后更新：2026-06-24 00:20 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前分支：`codex/upstream-v0138-small-patches`。
- 本轮用户要求：排行榜增加“模型榜”，先放假数据看效果；模型榜图标改为模型商图标；右侧增加“增长”和“排名变化”；最后更新本地容器。
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
- 模型榜右侧增加两个指标卡：
  - 增长：例如 `+28.4%`、`-77.7%`、`—`。
  - 排名变化：例如 `↑ 1`、`↓ 1`、`—`。
- 模型榜移动端已做响应式堆叠，避免右侧指标挤压主内容。
- 已补中英文 i18n、前端类型、后端/前端定向测试。

## 本地容器

- 当前在线容器：`sub2api`，不是 `sub2api-dev`。
- 地址：`http://127.0.0.1:62080`。
- 当前镜像：`sub2api:codex-20260624-0011-leaderboard-trend`。
- 当前镜像 ID：`sha256:a123f00a67aa9073dcdbe00bbd6e4355b9a14d1f29d5914d079f0ba3bc3ef87b`。
- `sub2api:local` 已指向同一新镜像。
- 当前容器状态：`running healthy`，端口 `127.0.0.1:62080->8080`。
- 旧容器备份：`sub2api-before-leaderboard-trend-20260624-0011`，旧镜像 `sub2api:codex-20260623-2336-model-icons`。
- `sub2api-postgres` 和 `sub2api-redis` 未重建，数据容器未动。
- Docker 更新锁已获取并释放。

## 验证记录

- `go test ./internal/handler -run "TestUsageHandlerDashboardLeaderboard" -count=1` 通过。
- `go test ./internal/repository -run "TestUsageLogRepositoryGetLeaderboardModelRanking|TestUsageLogRepositoryGetUserLeaderboard" -count=1` 通过。
- `corepack.cmd pnpm --dir frontend exec vitest run src/views/user/__tests__/LeaderboardView.spec.ts src/__tests__/leaderboard-theme.spec.ts` 通过。
- `corepack.cmd pnpm --dir frontend run typecheck` 通过。
- `corepack.cmd pnpm --dir frontend run build` 通过，仅有既有 Browserslist、Node `DEP0190`、Vite chunk warning 和大 chunk 提示。
- `git diff --check` 通过，仅提示部分 Markdown/knowledge 文件下次 Git 触碰时 LF 会替换为 CRLF。
- `GET http://127.0.0.1:62080/health` 返回 200。
- `docker exec sub2api /app/sub2api --version` 输出 `Sub2API 0.1.126 (commit: codex-leaderboard-trend, built: 2026-06-23T16:14:00Z)`。

## 未验证点

- 内置浏览器访问 `/leaderboard` 被重定向到 `/login?redirect=/leaderboard`，当前未持有登录态，因此未完成真实登录态视觉截图验收。
- 未 stage/commit。

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

- 若用户提供登录态或自行登录后，可打开 `http://127.0.0.1:62080/leaderboard`，切到“模型榜”，人工确认模型商图标、右侧“增长/排名变化”和移动端布局。
- 若要提交，先用 `git diff --cached --name-only` 审核 staged 文件，再跑 `git diff --cached --check`。

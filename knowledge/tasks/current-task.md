# 当前任务快照

最后更新：2026-07-12 20:58 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 发布工作树：`E:/codex-worktrees/sub2api/release-upstream-v0151-followups-s71-s73`，分支 `codex/release-upstream-v0151-followups-s71-s73`。
- 发布 merge：`ccac358e4`，父提交为最新主线 `f6ee836d4` 与 S71-S73 集成/流程收口 `d101ac2d2`。
- 主工作树仅有用户自己的 `knowledge/05-current-focus.md` 修改，SHA256 为 `99B6CEF620315114851B1BD6CAEB8CDB0BE851AF895A14E4B8C78F9E5DCDE882`；本轮不得暂存、覆盖或回滚。
- 本轮不部署、不更新本地容器。

## 当前目标

- 精确提交 post-merge 发布记录，刷新并确认 `origin/main` 未移动后推送。
- 验证远端 HEAD，再让本地主工作树 `main` 仅以 fast-forward 跟进，并清理已证明合并的 integration/release worktree 和本地分支。
- 支付并发补丁 `fc66a30ff` 继续保持独立高风险审计边界。

## 本次已完成

- Upstream S71：用户级 Fast/Flex 策略只读取可信 `ctxkey.APIKeyUserID`，用户规则组优先于全局规则；managed/API-key/OAuth HTTP 与 parsed/passthrough WS 均有 capture 覆盖。
- Upstream S72：裸 `gpt-5.6` alias/catalog 严格归一到 Sol；未知 suffix 不误映射，Terra/Luna 不坍缩；前后端目录与 OpenCode `max` 已补齐。
- Upstream S73：UserBreakdown 支持 aliased legacy `request_type` fallback，并保持非零类型权威、RequestType+Stream AND、七列 scan、排序/LIMIT 与排行榜边界。
- S74：管理端工单只返回最小用户摘要，详情复用只读用户信息弹窗，最近使用扩为 30 条滚动记录。
- S75：UsageView 仅在列设置菜单打开时把筛选卡片提升到 `z-[221]`。
- 三个 workflow 冲突已按时间与 ownership 合并；较晚主线任务从原本重号的 S71/S72 重编号为 S74/S75，`total_sprints` 为 75。

## 已确认事实

- 最新主线与 S71-S73 集成分支的业务路径交集为 0；merge 审计逐文件确认主线 15 个业务 blob 与 upstream 24 个业务 blob 均完整保留。
- `ccac358e4` 恰有两个正确父提交，且两者均为其祖先；无冲突标记、无 unmerged index 项。
- S71-S73 的 3 份 contract、3 份 worker result、3 份独立 QA 与 1 份组合 QA 共 10 个 artifact 均存在且引用有效。
- release worktree 无残留 `frontend/node_modules` junction，`backend/internal/web/dist` 无生成改动。
- 当前快照尚未推送；`origin/main` 仍需在发布前再次 fetch 核对。

## 待验证点

- 未调用真实 OpenAI/Codex 上游；HTTP/WS 行为由 in-process server 与本地真实 relay capture 证明。
- S73 未连接真实 PostgreSQL；SQL 行为由 SQLMock 与精确 SQL 断言覆盖。
- S74/S75 未在真实管理员会话中做浏览器 smoke；Docker、PostgreSQL、Redis 当前不可用。
- 未执行 race、生产部署或容器更新。

## 当前结论

- `PASS`：代码、merge 结构、Sprint 编号和合并态回归均已通过发布门禁；可以进入精确提交与远端发布。
- 此快照不把“准备推送”描述为“已发布”；必须以 `git ls-remote` 的远端 HEAD 复核作为发布完成证据。

## 下一步

1. 仅暂存 `docs/workflow/status.md`、`docs/workflow/main-log.md`、`knowledge/tasks/current-task.md`、`knowledge/tasks/timeline.md`，验证：检查 cached 文件列表、`git diff --cached --check` 与旧 task ID 扫描。
2. `git fetch origin` 后确认 `origin/main` 仍为 `f6ee836d4`，再把 release HEAD 推送到 `origin/main`，验证：`git ls-remote origin refs/heads/main` 等于本地 HEAD。
3. 主工作树执行 `git merge --ff-only origin/main`，验证：`knowledge/05-current-focus.md` SHA256 保持不变。
4. 仅删除已证明为 `main` 祖先或 patch-equivalent 的 integration/release worktree 和分支，验证：最终仅保留主工作树 `main`。

## 验证记录

- Post-merge backend：S71 exact discovery/tests `4/4`、S72 `6/6`、S73 `2/2`；Fast/Flex HTTP/WS/DTO/service、legacy request-type、leaderboard、S74 ticket 与 repository compile-only 回归均 PASS。
- Post-merge frontend：`6 files / 59 tests` PASS；`vue-tsc --noEmit` PASS；production build PASS。
- Release audit：merge parent/ancestor、39 个业务 blob、10 个 workflow artifact、S74/S75 编号、冲突标记与 diff check 全部 PASS。
- 已知非阻断警告：router-link test stub、stale `caniuse-lite`、Vite dynamic/static import 与 chunk-size、Node `DEP0190`。

# 当前任务快照

最后更新：2026-07-12 20:46 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 最新发布基线：`main` / `origin/main` 均为 `f6ee836d4`。
- 发布工作树：`E:/codex-worktrees/sub2api/release-upstream-v0151-followups-s71-s73`，分支 `codex/release-upstream-v0151-followups-s71-s73`。
- 主工作树仅保留用户自己的 `knowledge/05-current-focus.md` 修改；发布工作树不得暂存或覆盖该文件。
- 本轮不部署、不更新本地容器。

## 当前目标

- 把已完成并通过独立 QA 的 upstream S71-S73 合入最新主线，执行合并态回归后推送 `origin/main`。
- 合并前主线已包含工单用户上下文和 UsageView 列菜单层级两项后续工作；为避免与 upstream S71-S73 重号，发布整理中将其文档编号顺延为 S74/S75。
- 支付并发补丁 `fc66a30ff` 继续保持独立高风险审计边界。

## 已完成范围

- Upstream S71：用户级 Fast/Flex 策略只读取可信 `ctxkey.APIKeyUserID`，用户规则组优先于全局规则；HTTP managed/API-key/OAuth passthrough 与 parsed/passthrough WS capture 均已覆盖。
- Upstream S72：裸 `gpt-5.6` alias/catalog 归一到 Sol；未知 suffix 不误映射，Terra/Luna 不坍缩；前后端目录和 OpenCode `max` 已补齐。
- Upstream S73：UserBreakdown 支持 aliased legacy `request_type` fallback，并保持非零类型权威、RequestType+Stream AND、七列 scan、排序/LIMIT 和排行榜边界。
- S71-S73 三个 Generator、独立语义审查和三个 fresh integration QA 均完成；S71 首轮 i18n namespace finding 已修复并复测 PASS。
- S71-S73 组合验证：后端精确 discovery `4/6/2`，前端 `3 files / 46 tests`、typecheck、路径与冲突标记审计 PASS。
- S74（原主线本地编号 S71）：管理端工单返回最小用户摘要，详情复用只读用户信息弹窗，最近使用扩为 30 条滚动记录；定向 Go/frontend/typecheck/build PASS。
- S75（原主线本地编号 S72）：UsageView 仅在列设置菜单打开时把筛选卡片提升到 `z-[221]`；定向 Vitest、typecheck、build 和 CSS 检查 PASS。

## 合并整理

- 最新主线与 upstream 集成分支的业务路径交集为 0。
- 冲突仅发生在 `docs/workflow/main-log.md`、`docs/workflow/status.md`、`knowledge/tasks/current-task.md`。
- `main-log` 已按时间合并两组记录；完整 upstream S71-S73 artifacts 保留，后完成且缺少 tracked artifact 的主线工单/列菜单记录改为 S74/S75，并由 `docs/workflow/spec.md` 持久化范围。
- 当前 merge 尚未提交；必须先确认冲突标记清零、运行合并态回归并检查 staged 边界。

## 待验证点

- 合并态尚未重新运行 S71-S73 后端、前端 46 项和 typecheck。
- 未调用真实 OpenAI/Codex 上游；HTTP/WS 行为来自 in-process server 与本地真实 relay capture。
- S73 未连接真实 PostgreSQL；SQL 行为由 SQLMock 和精确片段断言覆盖。
- S74/S75 尚未在真实管理员会话中做浏览器 smoke；Docker、PostgreSQL、Redis 当前不可用。
- 未执行 race、生产部署或容器更新。

## 当前结论

- 业务实现可以合并；发布仍处于 `post-merge regression pending`，尚不能描述为已推送。

## 下一步

1. 完成冲突解析并创建 `--no-ff` merge commit。
2. 运行 S71-S73 合并态后端/前端/typecheck 与最终 diff/path 审计。
3. 更新 workflow 为真实发布状态，提交并推送 `origin/main`。
4. 验证远端 HEAD 后，让本地 `main` 快进并清理已证明合并的 release/integration 分支。

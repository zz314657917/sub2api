# 当前任务快照

最后更新：2026-07-30 18:21 +08:00

## 背景

- 用户要求检查上游 `v0.1.168` 与本地差异，并仅选择适合的兼容性行为合入本地。
- S128 在隔离 worktree 完成 source-level QA 后，用户授权本地提交与普通 merge；推送、
  部署、容器更新和外部运行态验证仍不在范围内。
- 随后近期提交复核确认 OpenAI 容量重试、团购退款、Grok 请求清洗和排行榜首屏守卫的
  五项回归；S130 已完成最小修复与 source-level QA。

## 当前目标

- 保留 S130 的已验证本地修复，等待用户决定是否整理提交；没有推送、部署、容器或真实
  支付渠道操作授权。

## 本次已完成

- 未提交改动已分为四条独立提交：`1ca5b0e89` Opus 5 支持、`295176d11` Grok
  Responses 清理、`5cc202465` group-buy 弹窗修复和 `3007bbc2e` workflow/knowledge 更新。
- `origin/main` 合并为 `6bd90ba1b`；其 S121 审计安全、S122 渠道定价目录和 S123
  异步图片任务均已进入本地 `main`。
- S110 group-buy 生命周期/退款硬化合并为 `a397d5966`；S120 排行榜账号时长门槛
  合并为 `0c74fa192`。
- S114、S115、S125、S126 的代码补丁已通过 `git cherry` 证明与 `main` 等价，缺失的
  workflow contract/QA/result 文件已补入主线。
- 已删除 9 条冗余本地 branch ref，并将含冲突的旧 S121 worktree 压缩备份后删除；当前
  保留主工作区及 S127、S128 两个隔离 worktree。没有删除远端分支、备份分支或 stash。
- `main` 已成功推送到 `origin/main`，远端从 `811a6d3c0` 前进到 `86845f9be`。
- S128 的选择性 `v0.1.168` 兼容修复提交为 `85439ff50`，并通过普通 merge
  `fbf4ea10e` 合入本地 `main`；未产生冲突。
- S129 已校正 workflow 与 handoff 记录，使其区分本地合入与远端发布状态；未修改业务
  代码、测试、部署或容器配置。
- S130 补齐 11 个 OpenAI failover 构造器的精确容量同账号重试限额；部分退款转人工复核，
  超时释放后的迟到支付转 `refund_pending` 并写入退款排队事件；Grok 清理
  `tools: null`/空数组下的 `tool_choice`；排行榜仅在数值门槛已加载时执行年龄拦截。

## 已确认事实

- S128 的本地 merge 为 `main@fbf4ea10e`，其上游跟踪基线仍为
  `origin/main@49af8e1bb`；发布前必须重新测量精确分叉，尚未推送。
- Redis 实例可通过 `127.0.0.1:6380` 访问；临时键 PING、TTL、GET、DEL 和 EXISTS
  清理均通过。默认 `127.0.0.1:6379` 未映射。
- `outputs/` 未跟踪且未纳入任何提交；两条历史 stash 未查看、未应用、未删除。
- S130 隔离 Go 回归、全仓生产构建/compile probe、前端路由 44/44 Vitest、前端 typecheck、
  格式、diff、冲突标记与允许路径检查通过。

## 待验证点

- 动作：修复仓库既有服务/unit 测试漂移后重跑完整 Go 回归 -> 验证：`stringPtr`、billing
  测试签名和 Grok runtime-block 测试不再阻断当前包。
- 动作：获得测试环境和认证授权后验证部分退款、团购超时后迟到支付、容量重试和排行榜首屏 ->
  验证：支付记录、退款队列、重试日志和路由行为与 S130 一致。

## 当前结论

- `PASS / local-only`：S128 的定向 Go/Antigravity、前端 Vitest/typecheck、全仓 Go
  编译、格式与 Git 完整性检查已通过；`85439ff50` 已由 `fbf4ea10e` 合入本地 `main`。
- `outputs/` 未跟踪且未纳入任何提交。没有推送、部署、容器更新或真实上游 OAuth、
  Anthropic、Gemini、Antigravity、数据库或认证态浏览器验证。
- `PASS / source-level`：S130 的变更和回归证据已通过；正常服务/`unit` 测试二进制仍受
  本轮未修改的陈旧测试源码阻断，真实外部运行态未验证。

## 下一步

1. 用户如需提交，先审阅 S130 允许路径 diff -> 验证：仅包含本轮服务、测试、前端与
   workflow/handoff 记录。
2. 用户如需运行态交付，提供测试环境和认证授权 -> 验证：支付、退款、容量重试和首屏
   路由的日志与持久化状态。

## 验证记录

- `go mod verify`、`go test ./... -run '^$'`、S121/S123 定向 Go 回归和前端 typecheck：PASS。
- S110 服务/退款、管理员 handler、Wire cleanup、group-buy Vitest 和 typecheck：PASS。
- S120 服务/handler、前端路由/侧栏/设置 85 项 Vitest 和 typecheck：PASS。
- `git diff --check`、暂存区检查和未合并索引检查：PASS。
- 最终 `go mod verify`、`go test ./... -run '^$'`、前端 typecheck 和 production build
  （1101 modules）：PASS。
- `git diff --check origin/main..HEAD`、`git diff --cached --check` 和
  `git push origin main`：PASS；远端更新为 `86845f9be`。
- S128 merge 后：`go test ./... -run '^$'`、`AccountStatusIndicator.spec.ts`（6/6）和
  前端 typecheck：PASS；`git diff --check`、`git show --check` 与未合并索引检查：PASS。
- S130：隔离 Go 回归、`go build ./...`、`go test ./... -run '^$'`、路由 Vitest（44/44）、
  `vue-tsc --noEmit`、格式、11 个 failover 构造器静态审计、diff 和允许路径检查：PASS。

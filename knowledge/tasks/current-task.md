# 当前任务快照

最后更新：2026-08-31 22:18 +08:00

## 背景

- 用户要求检查 `v0.1.184` 上游差异并选择可安全合入本地的行为。
- 只读审计已拒绝整体 merge/cherry-pick，并把近期候选拆为 S276-S279。
- 当前先执行 S276：四项互不扩张的后端兼容修复。

## 当前目标

- 审核并执行 `upstream-v0184-compat-fixes-s276` contract。
- 按本地拓扑移植 Anthropic-to-Responses 流式生命周期、Anthropic 工具参数、SMTP 测试 TLS fallback 和版本后缀比较修复。

## 当前状态

- Workflow phase: `done`。
- Contract: `docs/workflow/tasks/upstream-v0184-compat-fixes-s276.md`。
- Base commit: `a4844cc6b43a3fdf23e91a35e8202bb5eb1d2165`。
- Contract review: `PASS`；本地 owner、测试命令、allowlist、denied paths 和 base commit 均已确认。
- `pge-doctor --strict`: 20 checks，0 issues。
- 外部 `gpt-5.6-terra` 与用户授权的 S276-only `gpt-5.6-sol` 调度均在处理 prompt 前返回 HTTP 403；两次调用均为 0 token/cost。
- 用户随后明确授权协作子智能体作为 S276 Developer 替代执行路径；实现和两轮 Controller finding 修正均已完成，worker report 首行为 `DONE`。
- Controller review: `PASS`；7 类 stream regressions、5 条 gateway tool-argument 路径、2 条 native bridge、SMTP 三态、版本后缀和普通版本比较均 x10 通过；默认 tag 的 3 个受影响包和 server compile 通过。
- `go test -tags unit ./internal/service` 仍被 contract 外的旧测试编译基线阻断，具体为重复 helper、旧函数签名和已删除字段引用；不计为 S276 回归。
- 独立 QA 已按用户授权由新的协作子智能体完成，只读审查并写入 QA report 首行为 `PASS`；全局 Agent Matrix 不修改。

## 保护边界

- 不整体 merge、rebase 或 cherry-pick `v0.1.184`。
- 保留 API-key route breaker/auth、`backend/internal/service/admin_service.go`、Pixel Cafe 管理页的现有修改和 `outputs/**`。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## 下一步

1. S276 已由 Codex 最终裁决并精确提交本地；不 push。
2. 如继续上游迁移，进入 S277 Planner：前端本地时间解析、Usage tooltip 和 Claude attribution。

## 后续队列

- S277：前端本地时间解析、Usage tooltip、Claude attribution。
- S278：带后缀模型的渠道定价归一化。
- S279：分组部分更新限额；必须额外保护当前脏的 `admin_service.go`。

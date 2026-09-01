# 当前任务快照

最后更新：2026-09-01 11:52 +08:00

## 背景

- 用户要求继续检查并选择 `v0.1.179`--`v0.1.194` 中可安全合入本地的上游行为。
- 上游最新实际 tag 是 `v0.1.184`；`v0.1.185`--`v0.1.194` 尚未发布。只读审计拒绝整体 merge/cherry-pick，并把剩余候选拆为 S277-S279。
- S276、S277 已精确提交本地但未 push；当前启动 S278 的渠道定价归一化修复。

## 当前目标

- 审核并执行 `upstream-v0184-channel-pricing-s278` contract。
- 按本地拓扑让带后缀的 OpenAI/Codex 模型在渠道字面查找失败后回退到已知归一化模型名，避免官方兜底价覆盖渠道价。

## 当前状态

- Workflow phase: `done`。
- Contract: `docs/workflow/tasks/upstream-v0184-channel-pricing-s278.md`；contract review: `PASS`。
- Base commit: `f81bb2a55`（S278 产品提交 `43d109581` 的实际父提交；并行 17 文件明确排除）。
- 上游来源 `eb4237a2b` 已按本地 `ModelPricingResolver` 拓扑适配；没有整体 merge/cherry-pick。
- S277 Developer report 首行为 `DONE`；独立 Terra QA report 首行为 `PASS`，并已提交为 `e5ff9b299`。
- S278 产品提交 `43d109581` 的实际父提交是并行提交 `f81bb2a55`；后者 17 个文件不属于 S278。S278 只允许 resolver、定向计费回归、worker/QA 证据；当前 backend 其他脏改、锁文件、Pixel Cafe、knowledge、outputs 仍按基线保护。
- Terra Developer 两次在模型接口层失败（HTTP 524、HTTP 503）；用户授权 S278 Developer 与独立 QA 改用 `gpt-5.6-sol`。Sol worker report 为 `DONE`，独立 QA 为 `PASS`：8 个定向用例 x10、完整 service 非缓存复跑、server 编译、格式、diff/conflict 和保护摘要均通过。

## 保护边界

- 不整体 merge、rebase 或 cherry-pick `v0.1.184`。
- 保留所有 `backend/**` 未提交 edits（包括不在 S277 的 apicompat）、`frontend/pnpm-lock.yaml`、API-key route breaker/auth、`backend/internal/service/admin_service.go`、Pixel Cafe 管理页和 `outputs/**`。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## 下一步

1. S278 已本地集成并独立 QA PASS；不 push。
2. 如继续上游迁移，为 S279 新建 contract，评估分组部分更新限额，同时保护当前脏的 `admin_service.go` 与 API-key cache/auth 路径。

## 后续队列

- S279：分组部分更新限额；必须额外保护当前脏的 `admin_service.go`。

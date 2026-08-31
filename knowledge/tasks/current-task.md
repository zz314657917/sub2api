# 当前任务快照

最后更新：2026-08-31 23:12 +08:00

## 背景

- 用户要求继续检查并选择 `v0.1.179`--`v0.1.194` 中可安全合入本地的上游行为。
- 上游最新实际 tag 是 `v0.1.184`；`v0.1.185`--`v0.1.194` 尚未发布。只读审计拒绝整体 merge/cherry-pick，并把剩余候选拆为 S277-S279。
- S276 已精确提交本地但未 push；当前执行 S277 的三项前端兼容修复。

## 当前目标

- 审核并执行 `upstream-v0184-frontend-compat-s277` contract。
- 按本地拓扑严格解析 `datetime-local`、让兑换码批量过期时间使用该解析器，并保留 Claude attribution headers。

## 当前状态

- Workflow phase: `done`。
- Contract: `docs/workflow/tasks/upstream-v0184-frontend-compat-s277.md`；contract review: `PASS`。
- Base commit: `53484808e7b1cab0049c2066d1a53816848e8b3c`。
- 上游来源 `81e461f65`、`b7aca87fd`、`5778739cd` 和 `c03776604` 均已映射到本地 frontend owner；原始 patch 在分叉拓扑中无法 `git apply --check`，必须手工保留行为。
- S277 Developer report 首行为 `DONE`；独立 Terra QA report 首行为 `PASS`。定向 Vitest 31/31、typecheck、build、diff/conflict 和保护路径摘要哈希均通过。
- 只有六个 S277 前端产品/测试文件和 workflow 证据会被精确提交；backend、锁文件、Pixel Cafe、knowledge、outputs 仍按基线保护。

## 保护边界

- 不整体 merge、rebase 或 cherry-pick `v0.1.184`。
- 保留所有 `backend/**` 未提交 edits（包括不在 S277 的 apicompat）、`frontend/pnpm-lock.yaml`、API-key route breaker/auth、`backend/internal/service/admin_service.go`、Pixel Cafe 管理页和 `outputs/**`。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## 下一步

1. S277 已经通过独立 QA，精确提交前端六文件和 workflow 证据；不 push。
2. 如继续上游迁移，进入 S278 Planner：带后缀模型的渠道定价归一化。

## 后续队列

- S278：带后缀模型的渠道定价归一化。
- S279：分组部分更新限额；必须额外保护当前脏的 `admin_service.go`。

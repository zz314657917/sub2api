# 当前任务快照

最后更新：2026-09-01 12:28 +08:00

## 背景

- 用户要求继续检查并选择 `v0.1.179`--`v0.1.194` 中可安全合入本地的上游行为。
- 上游最新实际 tag 是 `v0.1.184`；`v0.1.185`--`v0.1.194` 尚未发布。只读审计拒绝整体 merge/cherry-pick，并把剩余候选拆为 S277-S279。
- S276--S278 已精确提交本地但未 push；当前启动 S279 的分组限额部分更新修复。

## 当前目标

- 审核并执行 `upstream-v0184-group-limit-partial-s279` contract。
- 让管理员部分更新分组时保留省略的日/周/月限额，同时保留本地 room-managed 强制无限的约束。

## 当前状态

- Workflow phase: `done`。
- Contract: `docs/workflow/tasks/upstream-v0184-group-limit-partial-s279.md`；contract review: `PASS`。
- Base commit: `408916129`。
- 上游来源 `9f1effd71` 属于 `v0.1.184`；上游 service owner 在本地已合并进 `admin_service.go`，必须手工适配。
- S277 Developer report 首行为 `DONE`；独立 Terra QA report 首行为 `PASS`，并已提交为 `e5ff9b299`。
- S278 已由产品提交 `43d109581` 与 closeout `5ad3d5e73` 完成，独立 QA PASS。
- `admin_service.go` 当前有 Pixel Cafe 配额重置脏改；控制器已保存仓库外 SHA-256 基线。S279 只允许修改 `UpdateGroup` 限额块，不能吸收该脏改。
- Terra Developer report 为 `DONE`；handler/service 定向 x10、完整受影响包、server 编译、格式与外部基线检查通过。Controller 复核确认 Cafe hunk 未变。
- 独立 Terra QA report 为 `PASS`；合同完整 handler/service 命令、server 编译、格式、diff/conflict 与两份外部基线检查通过。

## 保护边界

- 不整体 merge、rebase 或 cherry-pick `v0.1.184`。
- 保留所有 `backend/**` 未提交 edits（包括不在 S277 的 apicompat）、`frontend/pnpm-lock.yaml`、API-key route breaker/auth、`backend/internal/service/admin_service.go`、Pixel Cafe 管理页和 `outputs/**`。
- 不 push、不部署、不更新容器，不操作数据库、共享数据或真实 provider。

## 下一步

1. 精确暂存并本地提交 S279；`admin_service.go` 只提交 UpdateGroup 限额 hunk，Pixel Cafe hunk继续留在工作区。
2. 不 push；`v0.1.179`--`v0.1.184` 当前已批准的 S276--S279 队列完成，后续候选需新开 Planner 审计。

## 后续队列

- S279：独立 QA PASS，精确本地集成收口中。

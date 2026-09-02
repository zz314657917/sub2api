---
type: task-contract
scope: project
status: approved
review_verdict: PASS
task_id: upstream-v0185-quota-reset-cooldown-s281
worker_model: gpt-5.6-terra
base_commit: 5d4810801
spec_ref: docs/workflow/spec.md
openspec_change: none
last_verified: 2026-09-01
---

# Upstream v0.1.185 Quota Reset Cooldown S281

## Task ID

`upstream-v0185-quota-reset-cooldown-s281`

## Role

你是 P/G/E 流程里的 Developer Worker。只执行本 contract，不做架构裁决，不扩大范围。

## Goal

按本地拓扑适配上游 `897faea33`：管理员手工重置账号 quota 时，在同一条 SQL 更新中清除账号级 `rate_limited_at` 与 `rate_limit_reset_at`，并在目标账号不存在时返回 `ErrAccountNotFound` 且不写 scheduler outbox。保留本地 quota、share-display、model-rate-limit、overload 与 temporary-unschedulable 语义。

## Success Criteria

- `accountRepository.ResetQuotaUsed` 保持现有接口名和调用方不变；单条 `UPDATE accounts` 同时清零本地 quota 并将 `rate_limited_at = NULL, rate_limit_reset_at = NULL`。
- 成功更新后才检查/写入 scheduler outbox，并调用已有 scheduler snapshot sync；零行更新返回 `service.ErrAccountNotFound`，不写 outbox。
- SQL 继续保留本地 monthly quota、share-display reset、固定重置字段删除和 `model_rate_limits` 不受影响；不清除 `overload_until`、`temp_unschedulable_until/reason`。
- 现有 quota reset 测试继续通过，并新增零行与 SQL cooldown 断言；不修改 AccountRepository 接口、Ent、迁移或其他 test doubles。
- 定向 repository 测试 ×10、完整 repository（若集成环境缺失需据实记录）、完整 service、server 编译、gofmt、diff/conflict 和受保护脏文件哈希门禁通过。

## Context

- Repo: `F:\mcplugins\sub2api`
- Upstream provenance: `897faea3332148d4dbe522c6198c1861ac02e076`。
- Local owner: `backend/internal/repository/account_repo.go:ResetQuotaUsed`；local admin service already calls this method, and many test doubles implement it.
- Existing scheduler helpers: `enqueueSchedulerOutbox` and `syncSchedulerAccountSnapshot` in the same repository package.
- Base commit: `5d4810801` (S280 local integration).

## Allowed Paths

- `backend/internal/repository/account_repo.go`（仅 `ResetQuotaUsed` 函数及其紧邻注释）
- `backend/internal/repository/account_repo_quota_reset_test.go`
- `docs/workflow/worker-results/upstream-v0185-quota-reset-cooldown-s281-result.md`
- `docs/workflow/qa-reports/upstream-v0185-quota-reset-cooldown-s281-qa.md`

## Denied Paths

- `backend/internal/service/account_service.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/account_service_delete_test.go`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/repository/account_repo_integration_test.go`
- `backend/internal/repository/**`（除上述 `account_repo.go` 与 `account_repo_quota_reset_test.go`）
- `backend/internal/service/**`
- `backend/internal/server/**`
- `backend/internal/pkg/apicompat/**`
- `frontend/**`
- `frontend/pnpm-lock.yaml`
- `VERSION`
- `docker-compose*.yml`
- `deploy/**`
- `knowledge/**`
- `outputs/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/**`
- `docs/workflow/contract-reviews/**`
- `C:/Users/Administrator/.codex/memories/**`
- 未在 Allowed Paths 中列出的任何源码、测试、工作流状态、配置、容器、部署或数据文件

## Constraints

- 不照搬上游接口重命名 `ResetQuotaUsedAndClearRateLimitCooldown`；保持本地已有接口和所有实现方兼容。
- 使用本地已有 `sqlExecutor`/Ent context 模式；SQL 更新与两个 cooldown 清除必须在一个数据库语句中完成。
- 只有 `RowsAffected > 0` 才允许 enqueue scheduler outbox 和 snapshot sync；outbox 失败沿用本地“记录日志、不回滚主更新”的既有语义。
- 不清理 overload、temporary-unschedulable、model_rate_limits、fixed-reset 配置或 share-display 之外的字段。
- 不 commit、push、调用 provider、操作真实数据库、容器、部署或共享数据。

## Acceptance Commands

From `F:\mcplugins\sub2api\backend`:

```powershell
go test ./internal/repository -run '^TestAccountRepositoryResetQuotaUsed' -count=10
go test ./internal/repository
go test ./internal/service
go test ./cmd/server -run '^$'
gofmt -w internal/repository/account_repo.go internal/repository/account_repo_quota_reset_test.go
```

From repo root:

```powershell
git diff --check -- backend/internal/repository/account_repo.go backend/internal/repository/account_repo_quota_reset_test.go
git diff --name-only --diff-filter=U
git status --short
```

Also inspect: target function diff is limited to cooldown assignment, result/row-count handling, existing outbox flow and snapshot sync; no interface/test-double churn; protected dirty paths and `outputs/**` remain unchanged.

## Output

- Write `docs/workflow/worker-results/upstream-v0185-quota-reset-cooldown-s281-result.md` using the worker-result template.
- First line must be `### DONE: upstream-v0185-quota-reset-cooldown-s281`, `### BLOCKED: upstream-v0185-quota-reset-cooldown-s281` or `### FAILED: upstream-v0185-quota-reset-cooldown-s281`.
- Independent QA writes only `docs/workflow/qa-reports/upstream-v0185-quota-reset-cooldown-s281-qa.md`.

## Stop Rules

- If implementation requires interface renames, Ent/migration/service/frontend changes, stop with `BLOCKED`.
- If the target function has unrelated concurrent edits or any protected baseline changes, stop and report; do not overwrite.
- Record integration-environment failures truthfully; do not weaken row-count or preservation assertions.

## Budget

- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

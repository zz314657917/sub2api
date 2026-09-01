### PASS: upstream-v0185-quota-reset-cooldown-s281

# QA Report

## Task ID

`upstream-v0185-quota-reset-cooldown-s281`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-quota-reset-cooldown-s281.md`
- `docs/workflow/contract-reviews/upstream-v0185-quota-reset-cooldown-s281-review.md`
- `docs/workflow/worker-results/upstream-v0185-quota-reset-cooldown-s281-result.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- target diff is limited to `accountRepository.ResetQuotaUsed` plus its dedicated quota-reset test file; no interface, test-double, Ent, migration, service or frontend churn
- protected dirty files and outputs remained unchanged
- aggregate protected dirty diff hash: `0e467987fd7aec5fc451983bdb8f8216f97ba69c`

## Executed Checks

```text
cd backend && go test ./internal/repository -run '^TestAccountRepositoryResetQuotaUsed' -count=10 -> PASS
cd backend && go test ./internal/repository -> FAIL (pre-existing fixture drift in account_repo_upstream_billing_probe_update_test.go: expected 32 columns, actual 34; outside S281 paths)
cd backend && go test ./internal/service -> PASS (cached)
cd backend && go test ./cmd/server -run '^$' -count=1 -> PASS
cd backend && gofmt -d internal/repository/account_repo.go internal/repository/account_repo_quota_reset_test.go -> PASS (no diff)
git diff --check -- backend/internal/repository/account_repo.go backend/internal/repository/account_repo_quota_reset_test.go -> PASS
git diff --name-only --diff-filter=U -> PASS (no unmerged paths)
git status --short -> PASS (only the two S281 target files plus pre-existing dirty paths/outputs)
Get-FileHash protected dirty files -> PASS (all contract baselines unchanged)
git diff --no-ext-diff --binary -- <six protected files> | git hash-object --stdin -> 0e467987fd7aec5fc451983bdb8f8216f97ba69c (PASS)
```

## Manual Checks

```text
Single SQL UPDATE clears quota counters and sets rate_limited_at = NULL, rate_limit_reset_at = NULL -> PASS
RowsAffected == 0 returns service.ErrAccountNotFound -> PASS
Zero-row path performs no scheduler_outbox INSERT -> PASS
Successful path enqueues scheduler outbox only after the UPDATE and then calls existing syncSchedulerAccountSnapshot -> PASS
Existing monthly quota, share-display reset, fixed reset field removal and model-rate-limit isolation remain intact; overload/temp-unschedulable fields are not touched -> PASS
```

## Findings

未发现 S281 实现问题。SQL cooldown 清除与本地 quota reset 在同一 UPDATE 中完成；零行更新正确返回 `ErrAccountNotFound` 且不写 outbox；成功路径保持既有 outbox 与 snapshot sync 顺序。

## Unverified Risks

- 完整 `./internal/repository` 被既有 `updatedAccountRows` fixture 漂移阻塞：`account_repo_upstream_billing_probe_update_test.go:559` 提供 34 个值而测试列数为 32。该测试和文件均不在 S281 allowlist，故归因于基线而非本次实现。
- 未执行真实 PostgreSQL、provider、scheduler 多账号、容器、部署或浏览器 smoke；这些不在合同范围内。

## Recommendation

`PASS`。可进入控制器的精确暂存/本地集成步骤；保留完整 repository 基线 fixture 失败记录，不修改其测试或其它 denied paths，不 commit/push。

## Bug Owner Recommendation

`none`（repository fixture 漂移属于既有基线，需另行处理）

## Root Cause

`none`（S281）；完整 repository 失败根因为 `environment-blocked` / baseline fixture drift。

## Retest Scope

- 若后续修改 `ResetQuotaUsed` 或其测试，重跑 quota reset 定向 x10、完整 service、server 编译、gofmt、diff/conflict、保护哈希和 aggregate dirty diff hash；repository fixture 修复后再重跑完整 repository。

## Knowledge Promotion

`none`

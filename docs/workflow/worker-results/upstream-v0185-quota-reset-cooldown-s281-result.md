### DONE: upstream-v0185-quota-reset-cooldown-s281

# Worker Result

## Task ID

`upstream-v0185-quota-reset-cooldown-s281`

## Status

`done`

## Summary

- 按 approved contract 在本地保留 `ResetQuotaUsed` 接口名，手工适配上游 `897faea33` 的 quota reset cooldown 行为。
- 单条 `UPDATE accounts` 同时清零本地总/日/周/月 quota、保留现有 share-display reset 逻辑，并将 `rate_limited_at` 与 `rate_limit_reset_at` 原子置空。
- 读取 `RowsAffected`；目标账号不存在时返回 `service.ErrAccountNotFound`，不会写 scheduler outbox；成功更新后沿用既有 outbox 与 scheduler snapshot sync。
- 新增零行更新测试，并扩展 SQL 断言确认两个 cooldown 字段被清除。

## Changed Files

- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_quota_reset_test.go`
- `docs/workflow/worker-results/upstream-v0185-quota-reset-cooldown-s281-result.md`

## Commands Run

```text
gofmt -w backend/internal/repository/account_repo.go backend/internal/repository/account_repo_quota_reset_test.go -> PASS
go test ./internal/repository -run '^TestAccountRepositoryResetQuotaUsed' -count=10 -> PASS
go test ./internal/repository -> FAIL (out-of-scope baseline fixture drift: updatedAccountRows expects 32 columns but supplies 34 at account_repo_upstream_billing_probe_update_test.go:559)
go test ./internal/service -> PASS
go test ./cmd/server -run '^$' -> PASS
git diff --check -- backend/internal/repository/account_repo.go backend/internal/repository/account_repo_quota_reset_test.go -> PASS
git diff --name-only --diff-filter=U -> empty
```

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/repository 5.458s (focused quota reset tests x10)
ok github.com/Wei-Shaw/sub2api/internal/service (cached)
ok github.com/Wei-Shaw/sub2api/cmd/server (cached) [no tests to run]
FAIL internal/repository: TestUpdateWithAccountBillingSettingsRollsBackWhenOutboxFails
panic: Expected number of values to match number of columns: expected 32, actual 34
at backend/internal/repository/account_repo_upstream_billing_probe_update_test.go:559
```

## Risks

- 完整 `internal/repository` 套件仍受既有 `updatedAccountRows` fixture 漂移阻塞；该测试不触及 S281 文件或行为，本轮未修改它。
- 未执行真实 PostgreSQL、provider、scheduler 多账号、容器、部署或浏览器 smoke；这些不在本 contract 范围内。
- 独立 QA 尚未执行；本报告不是最终 PASS。

## Knowledge Candidates

- 本地 quota reset 已扩展 monthly/share-display 语义时，应在同一条 SQL 中叠加 cooldown 清除，并保持其它调度阻断字段不变。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`（完整 repository 的独立 baseline fixture 失败除外）
- stop_rules_triggered: `no`

## Blocked Reason

- none

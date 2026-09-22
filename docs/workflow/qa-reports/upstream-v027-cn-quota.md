### PASS: upstream-v027-cn-quota

# QA Report

## Task ID
upstream-v027-cn-quota

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/upstream-v027-cn-quota.md`
- `docs/workflow/contract-reviews/upstream-v027-cn-quota-review.md`

## Initial Finding And Fix
- Initial QA found P1: `handleCNProviderQuotaExhausted403` persisted a recognized quota 403 but did not call `notifyAccountSchedulingBlocked`, leaving an attached in-process `AccountRuntimeBlocker` unaware of the durable pause. This differed from upstream `db8692d67` and its runtime test.
- Fix reviewed: notification now follows a successful `SetRateLimited` and follows a successful `SetTempUnschedulable`; neither notification runs when its corresponding write fails. The new `TestV027CNQuotaRuntimeBlockerFollowsSuccessfulPersistence` drives `HandleUpstreamError` and passes for future snapshot, missing snapshot, and both-writes-fail cases.

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched by this task: `no`
- commands run:
```text
backend: go test ./internal/service -run '^TestV027CNQuota' -count=1 -v -> PASS
backend: go test ./internal/service -run 'CN.*403|403.*CN|Concurrency' -count=1 -v -> PASS
backend: go build ./... -> PASS
backend: gofmt -d internal/service/ratelimit_cn_providers.go internal/service/ratelimit_service.go internal/service/cn_quota_v027_test.go -> no output
root: git diff HEAD --check -> PASS
root: git diff --name-only --diff-filter=U -> empty
```
- manual checks:
```text
403 priority -> PASS: exact Kimi concurrency 403 returns before quota recognition.
quota boundary -> PASS: only CN Coding Plan accounts match text signals or error.type=access_terminated_error; non-Coding Plan accounts retain generic 403 behavior.
future snapshot -> PASS: earliest future 5h/weekly reset persists SetRateLimited and then blocks runtime scheduling with cn_quota_exhausted.
stale or missing snapshot -> PASS: uses bounded 10-minute SetTempUnschedulable and then blocks runtime scheduling.
SetRateLimited failure then temporary-write success -> PASS by reviewed control flow: failed rate-limit write makes no notification; successful fallback SetTempUnschedulable immediately calls notifyAccountSchedulingBlocked with the temporary cooldown and cn_quota_exhausted reason.
both writes fail -> PASS: test verifies one failed rate-limit write, one failed temporary write, no SetError, and no runtime blocker notification.
non-target and concurrency regressions -> PASS: non-Coding Plan quota signal follows existing generic 403 path; exact Kimi concurrency message remains on its dedicated temporary-cooldown path.
```

## Findings
- 初始 P1 已修复，未发现当前实现的明确问题。
- 测试覆盖限制：运行时 blocker 测试直接覆盖未来快照、缺失快照和双写失败；“SetRateLimited 失败、SetTempUnschedulable 成功”的通知由同一成功后通知分支静态复核确认，未单列测试子用例。该缺口不改变本次 PASS 结论。

## Bug Owner Recommendation
`none`

## Root Cause
- `none`

## Retest Scope
- 已完成合同定向测试、既有 CN 403/并发回归和 backend 全量构建。

## Knowledge Promotion
- `none`

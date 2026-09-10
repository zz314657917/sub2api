---
type: task-contract
scope: repository
status: approved
review_verdict: PASS
task_id: upstream-v024-payment-fulfillment-isolation-s300
base_commit: 41282aa1383d452ef1d1f5b8b0f7791f8f8085ba
last_verified: 2026-09-11
---

# Task Contract: Payment Fulfillment Isolation S300

## Goal

按上游 `7a70de401` 的行为修复支付与兑换码公共限流边界：支付履约和管理员履约不得被用户公开兑换失败计数阻断，也不得增加该计数；公共 `Redeem` 保持现有安全限流。支付履约重试必须对已有兑换码做归属、类型、金额和状态校验，并对非“未找到”查询错误 fail-closed。

## Allowed Paths

- `backend/internal/service/redeem_service.go`
- `backend/internal/service/payment_fulfillment.go`
- `backend/internal/handler/admin/redeem_handler.go`
- `backend/internal/service/payment_fulfillment_test.go`
- `backend/internal/service/redeem_admin_fulfillment_test.go`
- `docs/workflow/tasks/upstream-v024-payment-fulfillment-isolation-s300.md`
- `docs/workflow/contract-reviews/upstream-v024-payment-fulfillment-isolation-s300-review.md`
- `docs/workflow/worker-results/upstream-v024-payment-fulfillment-isolation-s300-result.md`
- `docs/workflow/qa-reports/upstream-v024-payment-fulfillment-isolation-s300-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths and Constraints

- 不修改迁移、Ent、wire、数据库、支付 provider、前端、锁文件、容器、部署或任何现有并发脏文件。
- 不整体 merge、rebase 或 cherry-pick 上游提交。
- 不执行真实支付、生产兑换、数据库迁移、部署或推送。
- 保持公共兑换限流、兑换事务、affiliate 语义和管理员 API 兼容。

## Acceptance

- `go test ./internal/service -run 'Test(ResolveRedeemAction|ValidatePaymentRedeemCode|PublicRedeem|PaymentRedeem|ExecuteBalanceFulfillment|AdminFulfillment)' -count=1`
- `go test ./internal/handler/admin -run 'Test.*Redeem' -count=1`
- `go build ./...`
- `gofmt -d` on four code/test paths
- exact allowlist `git diff --check`
- unmerged-index check

## Required Behaviors

- 公共 `Redeem` 超过失败阈值仍返回 `ErrRedeemRateLimited`，无效码仍增加计数。
- payment/admin fulfillment bypasses only the public failure counter, retains distributed code lock and all validation/transactional updates; admin retains redeem-level affiliate rebate.
- payment existing-code lookup treats `ErrRedeemCodeNotFound` as create, but propagates other lookup errors.
- Existing payment code must match order code, balance type, amount, and user/status invariants; mismatches fail before `Use` or credit.

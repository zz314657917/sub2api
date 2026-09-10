### PASS: upstream-v024-payment-fulfillment-isolation-s300

# QA Report

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v024-payment-fulfillment-isolation-s300.md`

## Findings

- 未发现阻断问题。
- Public `Redeem` 仍先执行失败限流，错误码、过期码和已用码仍增加失败计数。
- Payment/admin fulfillment 仅绕过公开失败计数，仍获取兑换码锁并复用事务和 `Use` 路径。
- Payment lookup 仅将 `ErrRedeemCodeNotFound` 视为首次创建；其他查询错误 fail-closed。
- 已存在支付兑换码在 `Use` 前校验 code、balance type、amount、status 和 `UsedBy` 用户归属。
- Admin fulfillment 保留 redeem-level affiliate rebate；payment fulfillment 继续使用订单级 affiliate rebate。
- allowlist、冲突索引和并发 denied-path 脏改均符合合同。

## Executed Checks

```text
go test ./internal/service -run 'Test(PublicRedeemStillEnforcesFailureLimit|PaymentAndAdminFulfillmentBypassFailureLimit|ValidatePaymentRedeemCodeRejectsMismatches)' -count=1 -> PASS
go test ./internal/handler/admin -run 'Test.*Redeem' -count=1 -> PASS
go build ./... -> PASS
gofmt -d on five allowed code/test paths -> PASS
git diff --check on exact allowlist -> PASS
git diff --name-only --diff-filter=U -> PASS
```

## Unverified Risks

- 未执行真实支付 provider、PostgreSQL 并发履约、迁移、容器、部署或推送。
- 带 `unit` 标签的全量历史测试仍有既有 fixture/signature 漂移；本批新增 focused tests 在普通测试路径通过。

## Recommendation

可进入本地集成提交；继续排除现有用户并发脏改。

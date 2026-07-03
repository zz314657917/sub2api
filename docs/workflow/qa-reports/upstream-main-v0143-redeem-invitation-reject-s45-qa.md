### PASS: upstream-main-v0143-redeem-invitation-reject-s45

## Findings
- 未发现 S45 diff 的明确问题。
- PASS: 普通兑换接口现在只允许 `balance`、`concurrency` 和有效 `subscription` 类型继续进入用户查询与事务。
- PASS: `invitation` 类型会返回 `REDEEM_CODE_UNSUPPORTED_TYPE`，并提示只能在注册流程使用。
- PASS: 本地额外的内部标记类型，例如 `monthly_recharge`，也会在事务前被拒绝，不会被普通兑换接口烧码。
- PASS: 管理员 create-and-redeem 的普通余额码和订阅参数校验路径未被本轮改动破坏。
- PASS: 公开 handler 邀请码注册相关定向测试通过，说明普通兑换接口的拒绝没有改注册入口行为。

## Executed Checks
- `go test ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestFulfillPaidOrder.*Redeem|TestPaymentRechargePackage|TestFirstRechargeBonus|TestMonthlyRecharge" -count=1`
  - Result: PASS.
- `go test ./internal/handler -run "TestEmailOAuthCallbackRequiresPendingRegistrationWhenInvitationEnabled|TestEmailOAuthCallbackExistingEmailLogsInWhenInvitationEnabled|TestCompleteWeChatOAuthRegistrationAfterInvitationPendingSessionReturnsPendingSession" -count=1`
  - Result: PASS.
- `go test ./internal/handler/admin -run "TestCreateAndRedeem_TypeDefaultsToBalance|TestCreateAndRedeem_SubscriptionRequiresGroupID|TestCreateAndRedeem_SubscriptionValidParamsPassValidation" -count=1`
  - Result: PASS.
- `git diff --check`
  - Result: PASS.
- Diff review:
  - `backend/internal/service/redeem_service.go` only adds pre-transaction unsupported type validation and reuses the same error helper in the defensive switch default.
  - `backend/internal/service/redeem_service_redeem_test.go` uses a panic-on-unexpected-call repository stub, proving unsupported paths do not call `Use`.

## Unverified Risks
- `go test -tags=unit ./internal/service ...` could not run because existing unrelated unit tests fail to compile:
  - `proxyRepoStub` / `proxyRepoStubForAdminList` lack `CountUserOwnedAccountsByProxyID`.
  - older billing tests still reference `ImageOutputPriceExplicit` and an outdated `computeTokenBreakdown` signature.
- Full backend suite was not run.

## Recommendation
- PASS S45 for scoped commit after staged denied-path audit.
- Keep Codex session import identity fix and ops realtime stats performance as later independent S46/S47 candidates; do not mix them into this redeem-code commit.

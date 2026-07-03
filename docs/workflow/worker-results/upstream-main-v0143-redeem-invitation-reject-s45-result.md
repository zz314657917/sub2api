### DONE: upstream-main-v0143-redeem-invitation-reject-s45

## Summary
- Ported upstream `372436323` / `修复邀请码普通兑换错误` into the local fork.
- Added `unsupportedRedeemTypeError` so invitation codes receive the registration-flow specific `REDEEM_CODE_UNSUPPORTED_TYPE` message.
- Moved unsupported redeem type validation before user lookup, transaction start, and `redeemRepo.Use`.
- Kept normal `balance`, `concurrency`, and valid `subscription` redeem behavior unchanged.
- Added focused tests proving invitation and local unsupported marker codes are rejected before `Use` can mark them used.

## Changed Files
- `backend/internal/service/redeem_service.go`
- `backend/internal/service/redeem_service_redeem_test.go`
- `docs/workflow/tasks/upstream-main-v0143-redeem-invitation-reject-s45.md`
- `docs/workflow/worker-results/upstream-main-v0143-redeem-invitation-reject-s45-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-redeem-invitation-reject-s45-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `go test ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestFulfillPaidOrder.*Redeem|TestPaymentRechargePackage|TestFirstRechargeBonus|TestMonthlyRecharge" -count=1`
  - Result: PASS.
- `go test ./internal/handler -run "TestEmailOAuthCallbackRequiresPendingRegistrationWhenInvitationEnabled|TestEmailOAuthCallbackExistingEmailLogsInWhenInvitationEnabled|TestCompleteWeChatOAuthRegistrationAfterInvitationPendingSessionReturnsPendingSession" -count=1`
  - Result: PASS.
- `go test ./internal/handler/admin -run "TestCreateAndRedeem_TypeDefaultsToBalance|TestCreateAndRedeem_SubscriptionRequiresGroupID|TestCreateAndRedeem_SubscriptionValidParamsPassValidation" -count=1`
  - Result: PASS.
- `git diff --check`
  - Result: PASS.
- `go test -tags=unit ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestResolveRedeemAction|TestMonthlyRechargeBonusClaimReleasedWhenBalanceRedeemFails|TestRegisterOAuthEmailAccount.*|TestRollbackOAuthEmailAccountCreationRestoresInvitationUsage" -count=1`
  - Result: BLOCKED by existing unrelated unit-test compile errors in `proxyRepoStub` / `ProxyRepository` and older billing test expectations.

## Contract Compliance
- Did not merge all upstream `v0.1.143` or `upstream/main`.
- Did not touch registration auth service code, payment fulfillment code, handler logic, repository logic, Ent, migrations, frontend, deploy, `.github`, README, or knowledge paths.
- Kept implementation in the isolated worktree and used focused tests around the changed service behavior.

## Risks
- The full `-tags=unit` service test package is currently blocked by unrelated existing test-stub drift, so unit-tag registration service tests were not executable in this worktree.
- Full repository test suite was not run.

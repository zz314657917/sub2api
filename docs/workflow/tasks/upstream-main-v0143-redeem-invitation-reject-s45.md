---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 13:24 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-redeem-invitation-reject-s45

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `372436323` / `修复邀请码普通兑换错误` so the normal redeem endpoint rejects invitation and other non-user-redeemable code types before transaction side effects.

## Success Criteria
- `RedeemService.Redeem` still allows only normal `balance`, `concurrency`, and valid `subscription` codes.
- `RedeemTypeInvitation` returns `REDEEM_CODE_UNSUPPORTED_TYPE` with a registration-flow message.
- Other unsupported local redeem types, including welfare/monthly bonus markers, return `REDEEM_CODE_UNSUPPORTED_TYPE`.
- Unsupported types are rejected before user lookup, transaction start, and `redeemRepo.Use`, so the code is not marked used.
- Existing subscription `group_id` validation remains unchanged.

## Allowed Paths
- `backend/internal/service/redeem_service.go`
- `backend/internal/service/redeem_service_redeem_test.go`
- `docs/workflow/tasks/upstream-main-v0143-redeem-invitation-reject-s45.md`
- `docs/workflow/worker-results/upstream-main-v0143-redeem-invitation-reject-s45-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-redeem-invitation-reject-s45-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `backend/internal/handler/**`
- `backend/internal/repository/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of upstream `v0.1.143` or `upstream/main`.
- Do not change registration invitation-code handling in auth flows.
- Do not change payment fulfillment recharge-code behavior beyond preserving normal balance-code redemption.
- Do not introduce migrations, Ent generation, frontend changes, or wire changes.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44
go test ./internal/service -run "TestRedeemRejects.*BeforeTransaction|TestFulfillPaidOrder.*Redeem|TestPaymentRechargePackage|TestFirstRechargeBonus|TestMonthlyRecharge" -count=1
git diff --check
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|backend/internal/handler/|backend/internal/repository/|frontend/|deploy/|knowledge/|\\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-redeem-invitation-reject-s45`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-redeem-invitation-reject-s45-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-redeem-invitation-reject-s45-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if staged diff includes denied paths.
- Stop if invitation registration flow tests show changed behavior.
- Stop if balance-code payment fulfillment redemption no longer works.
- Stop if unsupported codes are marked used before the error is returned.

## Review Result
- Reviewed at: 2026-07-03 13:24 +08:00.
- Verdict: approved.
- Reason: the upstream fix is narrow, backend-only, and directly prevents burning invitation or internal marker codes through the normal redeem endpoint.

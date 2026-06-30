### PASS: upstream-main-v0141-payment-validity-units-s33

# QA Report

## Task ID
upstream-main-v0141-payment-validity-units-s33

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-v0141-payment-validity-units-s33.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
go test ./internal/service -run "TestComputeValidityDaysSupportsSingularAndPluralUnits" -count=1 -> pass
git diff --check -- backend/internal/service/payment_service.go backend/internal/service/payment_order_result_test.go -> pass
```
- manual checks:
```text
Reviewed upstream 147c1879d against local payment_service.go -> pass
Confirmed frontend plan editor can emit weeks/months and backend now recognizes both plural values -> pass
Confirmed unknown-unit/days fallback still returns the raw validity_days value -> pass
Confirmed no frontend, Ent, migration, proxy/account, Ops, or knowledge paths were changed -> pass
```

## Findings
- 未发现 S33 范围内明确问题。

## Bug Owner Recommendation
codex-planner

## Root Cause
none

## Retest Scope
- If subscription plan unit options change again, rerun `TestComputeValidityDaysSupportsSingularAndPluralUnits`.
- If payment order creation is later refactored, add an order-level test asserting `subscription_days` for plural units.

## Knowledge Promotion
none

### DONE: upstream-main-v0141-payment-validity-units-s33

# Worker Result

## Task ID
upstream-main-v0141-payment-validity-units-s33

## Status
done

## Summary
- Ported upstream `147c1879d95f520b6db3af0291a52921322de421`.
- Added plural validity unit support to `psComputeValidityDays`.
- `weeks` now converts to `days * 7`.
- `months` now converts to `days * 30`.
- Added a regression test covering `days`, `week`, `weeks`, `month`, and `months`.

## Changed Files
- `backend/internal/service/payment_service.go`
- `backend/internal/service/payment_order_result_test.go`
- `docs/workflow/tasks/upstream-main-v0141-payment-validity-units-s33.md`

## Commands Run
```text
gofmt -w backend/internal/service/payment_service.go backend/internal/service/payment_order_result_test.go -> pass
go test ./internal/service -run "TestComputeValidityDaysSupportsSingularAndPluralUnits" -count=1 -> pass
git diff --check -- backend/internal/service/payment_service.go backend/internal/service/payment_order_result_test.go -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/service	5.761s
```

## Risks
- Verification is focused on the conversion helper. No live payment order was created.
- This does not change subscription stacking rules; it only fixes the day conversion before fulfillment receives `ValidityDays`.
- The wider main worktree still has unrelated user-owned-proxy dirty files, but they are outside S33.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

### DONE: upstream-v024-payment-fulfillment-isolation-s300

## Summary

- `ported`: payment/admin fulfillment bypasses the public redeem failure counter while retaining locks, validation and transactions.
- `ported`: payment idempotency distinguishes not-found from lookup failures and validates existing code ownership, type, amount and status.
- `ported`: admin CreateAndRedeem uses the trusted fulfillment entry point and keeps redeem-level affiliate behavior.
- `preserved`: public Redeem rate-limit and invalid-code failure accounting.

## Changed Files

- `backend/internal/service/redeem_service.go`
- `backend/internal/service/payment_fulfillment.go`
- `backend/internal/handler/admin/redeem_handler.go`
- `backend/internal/service/payment_fulfillment_test.go`
- `backend/internal/service/redeem_admin_fulfillment_test.go`

## Checks

- focused service tests: PASS
- admin Redeem tests: PASS
- `go build ./...`: PASS
- `gofmt -d`: PASS
- no database, provider, deployment or push action

## Risks

- Real payment provider, PostgreSQL concurrency and production fulfillment remain unverified.

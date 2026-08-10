### DONE: upstream-api-key-validation-s209

## Summary

- Added handler-side finite/non-negative validation for API Key quota and all
  rate-limit fields, plus positive-only Create `expires_in_days` validation.
- Placed handler validation before Create's idempotency execution and before
  Update's service call.
- Added matching service-side validation with structured
  `API_KEY_LIMIT_INVALID` and `API_KEY_EXPIRY_INVALID` HTTP `400` errors so
  internal callers cannot bypass the HTTP boundary.
- Added default-tag handler and service regressions for negative, `NaN`,
  positive/negative infinity, zero/unlimited, large finite values, invalid
  expiry, and pre-side-effect rejection.

## Changed Files

- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/api_key_handler_validation_s209_test.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/service/api_key_service_validation_s209_test.go`
- S209 workflow contract, status, spec, main log, and this result.

## Commands Run

- `gofmt -w <four changed Go files>`: PASS.
- `git diff --check`: PASS.
- `go test ./internal/handler -run '^TestS209' -count=10`: PASS.
- `go test ./internal/service -run '^TestS209' -count=10`: PASS.
- `go test ./internal/handler ./internal/service`: PASS.
- `go test ./cmd/server -run '^$' -count=0`: PASS.

## Contract Compliance

- Product changes are limited to the four allowed handler/service paths.
- No repository, schema/migration, generated code, dependency, frontend,
  configuration, container, deployment, primary-worktree S208, or `outputs/`
  path was changed.
- No cherry-pick, provider call, shared resource operation, push, deployment,
  or production validation was performed.

## Risks

- This is source-level and local-regression evidence only; no deployed HTTP
  runtime or production traffic was exercised.
- S209 cannot be fast-forwarded onto local `main` while the separate S208
  working-tree changes remain uncommitted and absent from this branch.

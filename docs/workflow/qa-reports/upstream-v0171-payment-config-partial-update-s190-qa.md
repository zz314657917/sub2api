### PASS: upstream-v0171-payment-config-partial-update-s190

## Findings

- Upstream `3deb2f17d` exposed a local PATCH-semantics defect: `UpdatePaymentConfig` populated every setting key, so omitted request fields overwrote stored values with empty strings. This could clear enabled payment types and visible-method routing while updating an unrelated setting.
- The update map now receives a key only when the request explicitly supplied that field. Explicit `false`, empty string, and empty payment-type slice still write their intended values. Existing recharge-package and FAQ presence handling is unchanged.

## Executed Checks

- `gofmt -w backend/internal/service/payment_config_service.go backend/internal/service/payment_config_service_test.go`: passed.
- `go test ./internal/service -run '^Test(UpdatePaymentConfig|GetPaymentConfig|ApplyVisibleMethod)' -count=1` from `backend/`: passed. Covers stored visible-method routing, omitted-field preservation, explicit empty/false values, and existing configuration reads.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0` from `backend/`: passed.
- `git diff --check`: passed; conflict-marker scan and `git ls-files -u` were empty.

## Unverified Risks

- No payment request, provider, database, historical setting, container, or deployment operation was performed. The conclusion is limited to the in-process payment-settings write boundary.

## Recommendation

Commit to the isolated branch `codex/upstream-v0171-integration-s183`; do not merge the primary worktree, push, or deploy.

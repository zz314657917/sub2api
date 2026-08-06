### PASS: upstream-v0171-subscription-quota-windows-s192

## Findings

- Upstream `7b6111f2f` fixes quota windows that were anchored to the calendar day rather than the subscription term. The local service now keeps the effective activation, renewal, and administrator-reset timestamps, advances automatic windows from their persisted anchor, and prevents a new automatic window at or beyond subscription expiry.
- The concrete Ent repository conditionally activates or resets a window. A stale conditional write is inert when the subscription still exists, while a missing subscription preserves `SUBSCRIPTION_NOT_FOUND`. This capability is optional so existing repository test doubles retain the original core interface.
- Legacy subscriptions whose initial window was stored at their start-date midnight are realigned to `StartsAt`; later manual anchors remain authoritative. One-time daily quotas remain non-resettable.

## Executed Checks

- `gofmt -w internal/service/subscription_service.go internal/service/user_subscription.go internal/service/user_subscription_port.go internal/service/user_subscription_daily_quota_test.go internal/service/subscription_quota_window_test.go internal/repository/user_subscription_repo.go` from `backend/`: passed.
- `go test ./internal/service -run 'Test(SubscriptionQuotaWindows|CheckAndResetWindows|ValidateAndCheckLimits|AssignOrExtendSubscription)' -count=1` from `backend/`: passed. Covers exact delayed activation and manual reset timestamps, renewed term anchors, legacy initial midnight anchors, partial final windows, explicit manual anchors, one-time daily quota behavior, and exact-expiry rejection.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0` from `backend/`: passed.
- `git diff --check`: passed. Conflict-marker scan and `git ls-files -u` were empty.

## Unverified Risks

- No PostgreSQL integration/concurrency run was performed for the Ent conditional-update queries. The repository behavior is source-reviewed and compile-covered; its concurrent stale-write behavior still needs an isolated database contract if runtime evidence is required.
- No subscription API request, production database, payment flow, container, deployment, push, or primary-worktree modification occurred.

## Recommendation

Commit to the isolated branch `codex/upstream-v0171-integration-s183`; do not merge the primary worktree, push, or deploy.

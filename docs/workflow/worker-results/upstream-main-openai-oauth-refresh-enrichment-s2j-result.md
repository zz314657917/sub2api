### DONE: upstream-main-openai-oauth-refresh-enrichment-s2j

## Task ID
upstream-main-openai-oauth-refresh-enrichment-s2j

## Status
done

## Summary
- Ported the bounded service behavior from upstream `eba204632 fix: enrich OpenAI OAuth token refresh`.
- Existing-access-token OpenAI OAuth refresh now preserves stored `subscription_expires_at` and runs best-effort enrichment without calling the OAuth refresh endpoint.
- Added ChatGPT `/backend-api/subscriptions` fallback for `subscription_expires_at` when accounts/check does not provide entitlement expiry.
- Added a small `ProvideOpenAIOAuthService` provider so production `OpenAIOAuthService` receives `PrivacyClientFactory`.
- Kept the patch limited to OpenAI OAuth/privacy service code, tests, and minimal wire provider/generated wiring.

## Changed Files
- `backend/internal/service/openai_oauth_service.go`
- `backend/internal/service/openai_privacy_service.go`
- `backend/internal/service/openai_oauth_service_refresh_test.go`
- `backend/internal/service/openai_subscription_test.go`
- `backend/internal/service/wire.go`
- `backend/cmd/server/wire_gen.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-openai-oauth-refresh-enrichment-s2j.md`
- `docs/workflow/worker-results/upstream-main-openai-oauth-refresh-enrichment-s2j-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-oauth-refresh-enrichment-s2j-qa.md`
- `knowledge/tasks/current-task.md`

## Commands Run
```text
gofmt -w backend/internal/service/openai_oauth_service.go backend/internal/service/openai_privacy_service.go backend/internal/service/openai_oauth_service_refresh_test.go backend/internal/service/openai_subscription_test.go backend/internal/service/wire.go backend/cmd/server/wire_gen.go -> pass
git diff --check -> pass
go test ./internal/service -run "OpenAIOAuthService_RefreshAccountToken_NoRefreshTokenUsesExistingAccessToken|FetchChatGPTSubscriptionExpiresAt|OpenAI.*Refresh|OpenAI.*Privacy|OpenAI.*Subscription|BuildAccountCredentials|RefreshIfNeeded" -count=1 -> pass
go test ./cmd/server -run TestNonExistent -count=1 -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 5.669s
ok github.com/Wei-Shaw/sub2api/cmd/server 5.657s [no tests to run]
```

## Implementation Notes
- `RefreshAccountToken` now resolves proxy URL before either refresh path so existing-token enrichment uses the same proxy as token refresh.
- The no-refresh-token branch now carries forward `subscription_expires_at`, then calls `enrichTokenInfo`.
- `enrichTokenInfo` first uses existing accounts/check enrichment, then falls back to the subscription endpoint only when expiry is still blank.
- `fetchChatGPTSubscriptionExpiresAt` validates `active_until` as RFC3339 and returns empty on client/request/status/body errors.
- `wire_gen.go` was manually kept in sync with `wire.go`: it now creates `privacyClientFactory` before `OpenAIOAuthService` and calls `service.ProvideOpenAIOAuthService`.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

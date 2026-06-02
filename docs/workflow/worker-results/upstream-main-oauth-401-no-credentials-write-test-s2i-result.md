### DONE: upstream-main-oauth-401-no-credentials-write-test-s2i

## Task ID
upstream-main-oauth-401-no-credentials-write-test-s2i

## Status
done

## Summary
- Completed a test-only semantic port of upstream `be3613593 test(oauth): update OAuth 401 tests to match new no-write behavior`.
- Production `ratelimit_service.go` was not changed because local implementation already matches upstream `6aec50501`: OAuth 401 invalidates token cache and sets temp-unschedulable state without persisting the request-start credentials snapshot.
- Updated unit tests now assert `updateCredentialsCalls == 0` for OAuth 401 no-write behavior.
- No production code, handler, schema, migration, config, frontend, OpenAI WS, or bridge files were changed.

## Changed Files
- `backend/internal/service/ratelimit_service_401_test.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-oauth-401-no-credentials-write-test-s2i.md`
- `docs/workflow/worker-results/upstream-main-oauth-401-no-credentials-write-test-s2i-result.md`
- `docs/workflow/qa-reports/upstream-main-oauth-401-no-credentials-write-test-s2i-qa.md`
- `knowledge/tasks/current-task.md`

## Commands Run
```text
git status --short --branch -> on codex/upstream-main-oauth-401-no-credentials-write-test-s2i
gofmt -w backend/internal/service/ratelimit_service_401_test.go -> pass
git diff --check -> pass
go test -tags unit ./internal/service -run "OAuth401InvalidatorError|OAuth401DoesNotOverwriteCredentials|OAuth401NoRefreshToken" -count=1 -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service
```

## Implementation Notes
- Updated `TestRateLimitService_HandleUpstreamError_OAuth401InvalidatorError` to assert no credentials write-back when token cache invalidation fails.
- Renamed and inverted the old `OAuth401UsesCredentialsUpdater` test into `OAuth401DoesNotOverwriteCredentials`.
- Added assertions that the OAuth 401 handler still sets temp-unschedulable cooldown and persists no credentials snapshot.

## Candidate Review Notes
- Equivalent locally, not re-ported: `32ea9cfe` API Key Responses SSE body fallback. Local code and `TestHandleNonStreamingResponse_APIKeyFallsBackToSSEBodyWhenContentTypeIsWrong` already exist.
- Equivalent locally, not re-ported: `b9509e823` and `ed2aac25a` long-context multipliers for cache_read/cache_creation. Local billing code and corresponding tests already exist.
- Already handled in migration Sprint: `f597c1581` group custom `/v1/models` list.
- Deferred: `eba204632` OpenAI OAuth refresh enrichment because it changes service wiring and generated wire paths.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

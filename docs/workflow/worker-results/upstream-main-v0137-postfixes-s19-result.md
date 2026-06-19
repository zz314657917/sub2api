### DONE: upstream-main-v0137-postfixes-s19

## Changed Files

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_images.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_failover_cached_body_test.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/ratelimit_service_anthropic_window_limit_test.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_test.go`
- `docs/workflow/tasks/upstream-main-v0137-postfixes-s19.md`
- `docs/workflow/spec.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Upstream Mapping

- `46bd7968a fix: reuse OpenAI failover error body`: equivalent. Local `OpenAIGatewayService.handleFailoverSideEffects` now receives cached `respBody` instead of rereading `resp.Body`; OpenAI chat/images failover callers pass their already-read body. Added a panic-on-read regression test.
- `f6e0ebc6b fix: preserve Anthropic window cooldowns`: equivalent. Local `RateLimitService.HandleUpstreamError` now persists official Anthropic 5h/7d window limits before local temp-unsched rules can shorten them. Reused existing Anthropic header parser and session-window update path.
- `8b698ff4c fix account list parameter limit`: equivalent. Local account repository now deduplicates and batches proxy, account-group, and group ID lookups at `50000` IDs to stay below PostgreSQL parameter limits. Added a large active account set regression test.
- `acaffe29e fix(account-repo): refresh candidates SQL excluded healthy accounts; fix CI build`: skipped/not applicable. Local `AccountRepository` currently has no `ListOAuthRefreshCandidates` method or SQL path, so the upstream SQL three-valued-logic fix has no target without pulling the broader token refresh retry amplification chain, which S19 explicitly excludes.

## Commands Run

```powershell
go test ./internal/service -run "Test.*Failover.*Body|Test.*Cached.*Body|Test.*Anthropic.*Window|Test.*Cooldown|TestOpenAI.*Images" -count=1
go test ./internal/repository -run "Test.*Account.*List|Test.*Refresh.*Candidate|Test.*Temp.*Unscheduled|TestAccountsToService" -count=1
go test ./internal/server -run "Test.*APIContract" -count=1
go test -tags=unit ./internal/service -run "TestOpenAIGatewayService_HandleFailoverSideEffects_DoesNotRereadResponseBody|TestOpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|TestHandleUpstreamError_AnthropicWindowLimitPreemptsTempUnschedRule" -count=1
go test -tags=unit ./internal/repository -run "TestAccountsToService_LargeActiveAccountSetDoesNotExceedPostgresParameterLimit" -count=1
git diff --check
git diff --name-only | rg "^(backend/ent/|backend/migrations/|backend/cmd/server/VERSION|backend/cmd/server/wire_gen.go|backend/internal/service/wire.go|frontend/|backend/internal/service/studio_bridge|backend/internal/repository/studio_bridge|frontend/src/views/public/|frontend/src/views/payment/|frontend/src/views/canvas/|frontend/src/components/studio/|frontend/src/views/admin/ModelMarket|frontend/src/views/admin/Payment)" || echo NO_DENIED_PATHS
```

## Test Output

- `go test ./internal/service ...`: PASS.
- `go test ./internal/repository ...`: PASS.
- `go test ./internal/server ...`: PASS; package compiled, no matching tests.
- `go test -tags=unit ./internal/service ...`: PASS.
- `go test -tags=unit ./internal/repository ...`: PASS.
- `git diff --check`: PASS with existing LF/CRLF working-copy warnings only.
- Denied-path audit: `NO_DENIED_PATHS`.

## Risks

- `acaffe29e` remains skipped because the local repository lacks the upstream `ListOAuthRefreshCandidates` API. Pulling it would require the broader token refresh retry amplification chain, outside S19.
- S19 intentionally does not port OpenAI image server-error failover (`da30c5992`), scheduler outbox changes, OAuth promo signup, cyber policy, channel monitor jitter, or Claude OAuth system prompt blocks.
- `docs/workflow/tasks/upstream-main-v0137-postfixes-s19.md` is ignored by `.gitignore:133 docs/*`; submitting it requires explicit `git add -f`.

## Knowledge Candidates

- For future upstream merges: `acaffe29e` only applies after introducing a local `ListOAuthRefreshCandidates` path; do not try to port it in isolation.

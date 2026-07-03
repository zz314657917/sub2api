### DONE: upstream-main-v0143-claude-oauth-payload-s40

## Summary
- Ported upstream `5bd9368ab fix claude oauth token exchange payload`.
- Claude OAuth setup-token exchange no longer sends `expires_in` in the outbound token exchange request body.
- Updated `ClaudeOAuthServiceSuite.TestExchangeCodeForToken` so the setup-token case asserts `expires_in` is omitted.
- Broader `v0.1.143` feature chains and dirty-overlap fixes remain deferred.

## Changed Files
- `backend/internal/repository/claude_oauth_service.go`
- `backend/internal/repository/claude_oauth_service_test.go`
- `docs/workflow/tasks/upstream-main-v0143-claude-oauth-payload-s40.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `gofmt -w backend/internal/repository/claude_oauth_service.go backend/internal/repository/claude_oauth_service_test.go`
- `go test ./internal/repository -run "TestClaudeOAuthServiceSuite/TestExchangeCodeForToken" -count=1`
  - Result: PASS.

## Contract Compliance
- Changed only allowed Claude OAuth repository files and workflow artifacts.
- Did not edit gateway, billing, usage, frontend, Ent, migrations, deploy, knowledge, dependencies, or generated files.
- Did not merge/rebase `v0.1.143` or cherry-pick broader release content.

## Risks
- `v0.1.143` still contains several deferred features/fixes, including OpenAI WS http_bridge, peak-rate groups, IP geolocation, subscription restore, count_tokens latest scope handling, Codex compact image bridge skip, Claude Code keepalive, and Anthropic Bearer auth.

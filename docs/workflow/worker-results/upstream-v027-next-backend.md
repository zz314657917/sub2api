### PASS: upstream-v027-next-backend

## Changed Files

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/deepseek_reasoning_v027_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_gateway_deepseek_reasoning_v027_test.go`
- `backend/internal/service/token_refresh_service.go`
- `backend/internal/service/token_refresh_service_test.go`
- `backend/internal/service/token_refresh_paused_v027_test.go`

## Behavior Evidence

- The default Responses-to-Chat converter remains unchanged. A named opt-in
  converter alone consumes plaintext `summary_text` / `reasoning_text` and
  attaches it to the following assistant turn. Encrypted content is ignored.
- Only the raw Responses fallback chooses that converter for a DeepSeek
  platform account or exact `api.deepseek.com` hostname. It fills only empty
  assistant reasoning with a single space.
- Opt-in normalization retains a reasoning-only assistant when its tool calls
  have no replies, clears those orphan tool calls, and does not change default
  normalization.
- JSON string input items are user boundaries in the opt-in path too: they
  clear pending reasoning before appending their user message, preventing a
  following assistant turn from receiving prior-turn plaintext.
- Active administrator-paused OAuth accounts enter the refresh loop. Refresh
  persists credentials without `SetSchedulable`, so the stored pause remains.

## Commands

- PASS: `go test ./internal/pkg/apicompat -count=1` (package suite passed).
- PASS: `go test ./internal/service -run 'TestV027(DeepSeekReasoning|PausedRefresh)' -count=1 -v` (4 top-level tests passed; DeepSeek host subcases cover exact, userinfo/suffix lookalikes, and platform).
- PASS: `go build ./...`.
- PASS: `git diff --check`; no unmerged paths.
- Follow-up PASS after the string-user-boundary fix: `go test
  ./internal/pkg/apicompat -count=1`, `go test ./internal/service -run
  'TestV027(DeepSeekReasoning|PausedRefresh)' -count=1`, and `go build ./...`.
- BLOCKED baseline: `go test -tags unit ./internal/service -run 'TestForwardResponses_ForceChatCompletions|TestTokenRefreshService|TestV027' -count=1 -v` cannot compile unrelated existing tagged tests (`stringPtr` duplicate, stale billing signatures, stale proxy fields, and count-token arity). The test binary was not produced, so this is not recorded as a passing unit execution.
- Follow-up tagged compile used `-gcflags=all=-e` after correcting the
  task-local embedded test-stub initialization. It reports no task-local
  compile error; the ordinary tagged selector still stops at the same
  unrelated baseline failures above.

## Scope And Limits

No commit, push, deployment, repository/migration/dependency change, real
OAuth refresh, or real DeepSeek request was performed. The worktree retained
the pre-existing `docs/workflow/*` edits and four temporary untracked patch
files untouched.

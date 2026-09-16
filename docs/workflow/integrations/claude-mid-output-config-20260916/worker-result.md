### DONE: claude-mid-output-config

## Changed Paths

- `backend/internal/pkg/claude/constants.go`
- `backend/internal/service/gateway_request.go`
- `backend/internal/service/gateway_mid_conversation_output_config_test.go`

## Implementation

- Added `BetaMidConversationOutputConfig` to the OAuth Claude Code mimic beta list.
- Preserved the existing context/thinking/fallback sanitizer and appended a bounded
  message-level `output_config` sanitizer. It is a byte no-op when the exact beta
  token is present or no message-level field exists; malformed complete bodies are
  rejected conservatively without rewriting.
- Empty system control messages with that field are removed only when content is
  absent, null, empty, or exclusively empty string text blocks. Unknown blocks and
  non-string text payloads are retained conservatively.
- Added untagged `MidConversation` tests for all four approved request builders,
  their default/drop or client-header present/absent cases, and an enabled-CCH
  upstream-builder case that recomputes the signature from the final sanitized body.

## Executed Commands

- `go test ./internal/service -list 'Test.*MidConversation'` — PASS; selected tests:
  `TestSanitizeAnthropicBodyForBetaTokens_MidConversationOutputConfig`,
  `TestBuildUpstreamRequestOAuthMimic_MidConversationOutputConfig`,
  `TestBuildUpstreamRequestAnthropicAPIKeyPassthrough_MidConversationOutputConfig`,
  `TestBuildCountTokensRequestOAuthMimic_MidConversationOutputConfig`,
  `TestBuildCountTokensRequestAnthropicAPIKeyPassthrough_MidConversationOutputConfig`,
  `TestBuildUpstreamRequestOAuthMimicEnabledCCH_MidConversationOutputConfig`.
- `go test ./internal/service -run 'Test.*MidConversation' -count=1` — PASS.
- `go test ./internal/service -count=1` — FAIL after 66.148s at existing
  `TestClaudeCodeValidator_MessagesWithoutProbeStillNeedStrictValidation`
  (`claude_code_validator_test.go:52`, expected false); the focused rerun fails
  identically. This task did not modify that file or validator path.
- `go build ./...` — PASS.
- `git diff --check` — PASS (only pre-existing controller CRLF warnings for
  `docs/workflow/main-log.md` and `docs/workflow/status.md`).
- `git ls-files -u` — empty; `gofmt -l` on the three changed Go paths plus the
  optional contract path — empty.
- QA follow-up: `go test ./internal/service -run 'Test.*MidConversation' -count=1`
  — PASS; `go build ./...` — PASS after the CCH cache-isolation fix.

## Scope and Risks

- Follow-up after independent QA: the enabled-CCH test now snapshots and restores
  the process-global forwarding cache with `t.Cleanup`; an initially unset cache
  is restored as an explicitly expired entry, so no `cchSigning=true` value leaks
  to later package tests.
- Task-owned business changes are exactly the three paths above. The modified
  `docs/workflow/main-log.md` and `docs/workflow/status.md` pre-existed and belong
  to the controller; they were not changed or reverted.
- No provider, database, container, deployment, commit, or push was performed.
- Full service-suite acceptance remains blocked by the independently reproducible
  pre-existing validator failure; fresh independent QA must decide final status.

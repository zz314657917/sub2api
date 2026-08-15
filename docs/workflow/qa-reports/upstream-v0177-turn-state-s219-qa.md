### PASS: upstream-v0177-turn-state-s219

## Findings

- None. Independent review of `8884ee10c6978b4823d7e1419d56aed1708bbfac..5eb5ffb168280e1c858ea9fabc049de37596cacf` found only the six contract-allowed paths.
- `x-codex-turn-state` remains outside the generic response-header allowlist. The normal and passthrough paths relay/clear it explicitly; neither native HTTP path injects stored turn state.
- Provenance requires a positive API-key ID and the original inbound `session-id` (falling back to `session_id`), records only after downstream commit, is guarded by the existing sticky TTL, and sweeps malformed or expired entries opportunistically.
- Normal and passthrough request builders run the stripping-only cross-account guard after client-header whitelisting and before account-side overrides. Same-account, unknown, expired, malformed, and untracked values remain unchanged.
- Streaming code stages/relays the header and records exactly once after the first successful downstream flush. All normal/passthrough JSON and SSE-to-JSON paths call provenance recording only after their downstream writer is committed.
- No fingerprint behavior, frontend, migrations, providers, containers, deployment, push, database, or WebSocket/Claude bridge changes were found.

## Executed Checks

- Worktree: `E:/codex-worktrees/sub2api/s219-turn-state`; `git rev-parse HEAD` returned `5eb5ffb168280e1c858ea9fabc049de37596cacf`; `git merge-base HEAD main` returned `8884ee10c6978b4823d7e1419d56aed1708bbfac`.
- Read the task contract, Developer result, and `F:/mcplugins/sub2api/docs/workflow/agent-matrix.md`; Developer result first line is `### DONE: upstream-v0177-turn-state-s219`.
- Default-tag discovery: all seven contract service tests were listed by `go test ./internal/service -list "^<test>$"`.
- Focused repeat: `go test ./internal/service -run '^(TestOpenAICodexTurnStateSeedRequiresAPIKeyAndSession|TestOpenAICodexTurnStateRelayGuardAndExpiry|TestOpenAIHTTPBuildersGuardCrossAccountTurnState|TestOpenAIStreamingTurnStateRecordsOnlyOnCommit|TestOpenAINonStreamingTurnStateRelaysJSONAndSSE|TestOpenAIPassthroughTurnStateRelayAndGuard|TestWriteOpenAIPassthroughResponseHeadersTurnState)$' -count=10` passed.
- Compatibility: `go test ./internal/service -run '^(TestOpenAIStreamingReadErrorBeforeOutputReturnsFailover|TestOpenAIStreamingPreambleOnlyMissingTerminalReturnsFailover|TestOpenAIStreamingPassthroughResponseFailedBeforeOutputReturnsFailover|TestForwardAsAnthropic_ReusesOAuthCodexTurnState|TestForwardAsAnthropic_OAuthDigestFallbackReusesTurnStateWithoutExplicitKey|TestOpenAIGatewayService_Forward_WSv2_TurnStateAndMetadataReplayOnReconnect|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_PassthroughHeadersUsePromptCacheAndTurnState)$' -count=1` passed (`ok`, 1.706s).
- Complete runtime/compile chain passed in order: `go test ./internal/service -count=1`; `go test ./internal/handler -count=1`; `go test ./internal/server -count=1`; `go test ./cmd/server -run '^$' -count=0`.
- Static gates passed: `git diff --check 8884ee10c..HEAD`; `gofmt -d` on all five changed Go files; `git diff --name-only --diff-filter=U` and `git ls-files -u` empty; exact allowlist audit `changed=6 outside=0`; denied-path audit `denied=0`.
- Provenance gates passed: `git merge-base --is-ancestor 8219dcfc8 upstream/main`, `4d9fedee2`, and `fce41e318` all succeeded.
- Main-worktree scope check: `git -C F:/mcplugins/sub2api status --short -- frontend/src/components/account/EditAccountModal.vue frontend/src/components/account/__tests__/EditAccountModal.spec.ts outputs` showed only pre-existing user changes (`M`, `M`, `??`); S219 commit range did not modify or stage them.

## Unverified Risks

- Runtime verification used only in-process fakes and loopback `httptest` fixtures. No real provider, shared database/Redis, container, deployment, or browser session was used, by contract.

## Recommendation

- PASS. The S219 implementation meets the approved contract and is ready for the Final Evaluator's integration decision.

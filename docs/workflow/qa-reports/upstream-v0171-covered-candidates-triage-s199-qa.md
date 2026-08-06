### PASS: upstream-v0171-covered-candidates-triage-s199

## Findings

- `272735b0a` (`fix(openai): preserve Codex namespace tools on OAuth Responses forwarding`): the local OAuth upstream builder already fixes both normal and passthrough traffic to `https://chatgpt.com/backend-api/codex/responses`; API-key `base_url` is not used for OAuth. The non-passthrough Codex transform retains namespace declarations and copies `input[].namespace` on tool-call items, while the passthrough normalizer only removes unsupported top-level fields and normalizes `store`/`stream`. The upstream per-account flatten fallback and its compact/transport restoration machinery have no small equivalent here; they are not a safe duplicate port.
- `7d3bf86e5` (`fix(openai): retry streamed capacity errors in pool mode`): local `newOpenAIStreamFailoverError` already applies the broader `openAISameAccountRetryPolicy` to stream failures. Capacity and overload events receive explicit retry budgets even outside pool mode, so the upstream pool-only retry assignment is subsumed.
- `da49ce3f2` (`fix(openai): fail open proxy stream circuit and collapse burst disconnects`): local code already has the disabled circuit setting, three-second collapse interval, active quarantine count, context-scoped scheduler bypass, and a fail-open scheduler retry when quarantine is the only capacity blocker.

## Executed Checks

- `go test ./internal/service -run '^(TestApplyCodexOAuthTransform_PreservesFunctionCallInputName|TestOpenAIGatewayService_OAuthPassthrough_StreamKeepsToolNameAndBodyNormalized|TestOpenAIGatewayService_OAuthPassthrough_CompactUsesJSONAndKeepsNonStreaming|TestS92NormalizeOpenAIResponsesCustomToolNamespaces)$' -count=1`: passed.
- `go test ./internal/service -run '^(TestOpenAIStreamingPassthroughResponseFailedBeforeOutput(CapacityErrorRetriesSameAccount|ServerOverloadedRetriesSameAccount)|TestOpenAIGatewayService_SelectAccountWithScheduler_ProxyStreamCircuit|TestOpenAIProxyStreamCircuit(CollapsesBurstFailures|Disabled|ActiveBlockCount)|TestOpenAIProxyStreamQuarantineBypassContext)$' -count=1`: passed.
- Source-level route/transform audit verified OAuth endpoint selection, namespace preservation on normal and passthrough `/responses`, and the local proxy-circuit topology.

## Unverified Risks

- No live OpenAI OAuth request, namespace relay, compact upstream response, proxy outage, database state, container, deployment, primary-worktree modification, push, or merge to `main` was performed.
- `272735b0a` compact-path namespace compatibility and a user-configurable fallback for external OAuth relays remain a separate architectural change, not evidence that the normal `/responses` behavior is missing.

## Recommendation

Do not cherry-pick or reimplement these three upstream commits. Keep the current local behavior and revisit the deferred compact/relay namespace concern only with a separate cross-transport contract.

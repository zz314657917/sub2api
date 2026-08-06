### PASS: upstream-v0171-encrypted-compaction-recovery-s194

## Findings

- The existing `invalid_encrypted_content` retry correctly handled encrypted reasoning items but retained encrypted `compaction` and `compaction_summary` input items, allowing stale account-bound data to survive its recovery retry.
- The shared retry sanitizer now removes an encrypted compaction item as a whole. Its behavior for reasoning items, unencrypted compaction, and unrelated input is unchanged.
- HTTP and WS initial requests retain the encrypted compaction; only their existing single recovery retry uses the sanitized payload.

## Executed Checks

- `go test ./internal/service -run 'Test(TrimOpenAIEncryptedReasoningItems|OpenAIGatewayService_Forward_HTTPIngressRetriesInvalidEncryptedContentOnce|OpenAIGatewayService_Forward_WSv2InvalidEncryptedContentRecoversOnce)' -count=1`: passed.
- `go test -v ./internal/service -run '^TestTrimOpenAIEncryptedReasoningItemsDropsOnlyEncryptedCompaction$' -count=1`: passed for encrypted `compaction`, encrypted `compaction_summary`, and unencrypted compaction cases.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `gofmt -w internal/service/openai_gateway_service.go internal/service/openai_ws_protocol_forward_test.go`: passed.
- `git diff --check`, conflict-marker scan, unmerged-index check, and S194 allowlist audit: passed.

## Unverified Risks

- No real OpenAI request, account credential, database, container, deployment, primary-worktree modification, push, or merge to `main` was performed.

## Recommendation

Commit this scoped recovery correction to the isolated integration branch. It is source-level HTTP/WS regression evidence, not live provider certification.

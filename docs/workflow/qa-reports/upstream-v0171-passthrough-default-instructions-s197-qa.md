### PASS: upstream-v0171-passthrough-default-instructions-s197

## Findings

- The local OAuth Codex passthrough branch executed before the standard OAuth request-body defaulting path, so an omitted `instructions` field was rejected locally even though the same model has an existing model-aware default instruction helper.
- S197 injects that default only for an omitted field, immediately before the established passthrough reject predicate. Explicit empty and non-string values remain rejected; API-key and non-Codex paths are unchanged.
- Both stream and compact request regressions reached the local upstream recorder with the expected default instruction and their existing normalized stream shape.

## Executed Checks

- `go test ./internal/service -run 'Test(OpenAIGatewayService_OAuthPassthrough_CodexMissingInstructionsGetsDefault|DetectOpenAIPassthroughInstructionsRejectReason|ForwardAsChatCompletions_OAuthDoesNotInjectDefaultInstructions)' -count=1`: passed.
- `go test ./internal/service -run '^TestOpenAIGatewayService_OAuthPassthrough' -count=1`: passed.
- `go test ./internal/service -run 'Test(OpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|OpenAIGatewayService_OAuthPassthrough_UpstreamRequestIgnoresClientCancel)' -count=1`: passed.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `gofmt -w internal/service/openai_gateway_service.go internal/service/openai_oauth_passthrough_test.go internal/service/openai_passthrough_normalization_test.go`, `git diff --check`, scoped conflict-marker scan, unmerged-index check, and staged allowlist audit: passed; exactly the six contract-allowed files were staged.

## Unverified Risks

- No real OpenAI OAuth request, account credential, database, container, deployment, primary-worktree modification, push, or merge to `main` was performed.

## Recommendation

Commit this scoped passthrough compatibility correction to the isolated integration branch. It has source-level stream/compact evidence, not live OpenAI certification.

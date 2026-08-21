### DONE: upstream-cn-account-test-routing-s237-a

## Controller Note

The native Terra Developer started but failed twice with `429 Too Many
Requests` before producing a complete candidate. Controller implementation was
therefore used after stopping the Worker loop, per the P/G/E stop rule. The
partial dispatcher diff was completed and validated in this isolated worktree.

## Changed Files

- `backend/internal/service/account_test_service.go`
- `backend/internal/service/account_test_service_cn_protocol.go`
- `backend/internal/service/account_test_service_cn_protocol_test.go`

## Behavior

- Fixed CN-provider `chat_completions` accounts use the existing provider-aware
  OpenAI base URL/key and Chat Completions stream parser.
- Fixed CN-provider `anthropic` accounts use the provider-native Anthropic base
  URL, `/v1/messages` without the generic `?beta=true` suffix, the shared API
  key auth helper, and no Anthropic-host fallback.
- DeepSeek fixed `responses` accounts use the platform-aware Responses URL and
  Bearer key regardless of stale `openai_responses_supported` metadata.
- OpenAI-shaped Anthropic base URLs are rejected before outbound traffic.

## Commands / Evidence

- `go test ./internal/service -run 'TestAccountTestService_(CN|DeepSeek)' -count=10` -> PASS, 4 tests.
- `go test ./internal/service -count=1` -> PASS, 64.563s.
- `go test ./cmd/server -run '^$' -count=1` -> PASS.
- `gofmt -l internal/service/account_test_service.go internal/service/account_test_service_cn_protocol.go internal/service/account_test_service_cn_protocol_test.go` -> no output.
- `git diff --check` -> PASS.
- Conflict-marker scan over all three changed product/test files -> no matches.

## Contract Compliance / Risks

- Product/test scope is limited to the dispatcher, one fixed-protocol owner, and
  focused tests. No Adaptive, gateway, schema, migration, frontend, dependency,
  provider, database, deployment, or user dirty path was changed.
- Local `4e59289ec` is the refreshed frozen base; affected account-test owners
  are unchanged from the earlier `e191ebc5d` draft base.
- Tests use an in-memory `HTTPUpstream` recorder and do not contact real
  providers.
- Independent QA is still required before main integration.

## knowledge_candidates

- None.

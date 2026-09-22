### PASS: upstream-v027-main-integration-backend

## Integration Baseline And Source Parity

- Reviewed in `E:/codex-worktrees/sub2api/upstream-v027-main-integration` at
  `796003b8d0563c7d64971b6bb2fe0f8f685c6589`.
- All 21 dirty backend V027 integration files match their counterparts in
  `E:/codex-worktrees/sub2api/upstream-v027-port` byte-for-byte by SHA-256.
  This includes the OpenAI, Gemini, CN 403, DeepSeek media, response media,
  and Anthropic schema-normalization portions of the five behavior batches plus
  the earlier Anthropic port.
- `git diff --check` passed and `git diff --name-only --diff-filter=U` was
  empty.

## Focused Integration Review

- OpenAI strict Chat role conversion remains bounded to API-key CN accounts or
  exact strict-provider hostnames, retains OAuth and retry-body behavior, and
  preserves raw unknown fields and numeric literals.
- Codex manifest validation keeps the lowercase, case-sensitive `models` key
  and a single enclosing JSON decode. HTTP response affinity retains values
  after client cancellation while applying the existing timeout.
- CN 403 quota handling, Gemini thinking-variant and SSE behavior, custom
  Gemini model lists, forced Antigravity fallback, DeepSeek media lifting, and
  Responses-to-Anthropic tool-schema handling compile and pass their selected
  regression coverage together on the latest main baseline.

## Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v027-main-integration/backend
go test ./internal/pkg/apicompat -count=1
go test ./internal/service ./internal/handler -run 'TestV027|TestFetchCodexModelsManifest|TestOpenAIGatewayService_BindHTTPResponseAccount|CN.*403|403.*CN' -count=1
go test -tags unit ./internal/handler -run '^(TestGeminiV1BetaListModels_CustomGroupListUsesNativeResponse|TestGeminiV1BetaListModels_ForcedAntigravityIgnoresCustomGroupList|TestCustomGeminiModelsList_DisabledKeepsExistingFlow)$' -count=1
go build ./...
```

All commands passed.

## Limits

This is source-parity, focused local-test, and build evidence only. It does
not establish real provider, Redis, database, container, deployment, commit,
or push behavior.

### DONE: upstream-openai-client-tools-s242

- Worktree/branch: `E:/codex-worktrees/sub2api/upstream-openai-client-tools-s242`, `pge/upstream-openai-client-tools-s242`.
- Base: local `main` at `baa6541acb1ef909b85d6d3cdb4817d2da0564c9`; workflow-only S242 contract commits remain outside the business scope.
- Provenance: reviewed upstream `44ef88f65`, `7e579cb28`, and `b94e484e2`; each is an ancestor of `upstream/main@d45135d87df16d48637f04ccd245727bc955ba54`.
- Scope: API-key OpenAI Responses custom declarations are lowered to function tools and restored for JSON/SSE; WS-HTTP bridge stores lowered tools and inherits custom mapping on tool-omitting follow-ups, while explicit `tools` replaces inherited state. Ordinary function and namespace-only tools are no-ops. Mapping is cleared/re-established per passthrough HTTP attempt.
- Tests:
  - `go test ./internal/pkg/apicompat -run "TestResponsesClientTool|TestAdaptResponsesClientTools" -count=10`
- `go test ./internal/service -run "TestOpenAIPassthroughAPIKey|TestOpenAIWSHTTPBridge.*ClientTool|TestAdaptOpenAIResponsesClientTools|TestClearOpenAIResponsesClientToolMapping" -count=10`
  - `go test ./internal/pkg/apicompat`
  - `go test ./internal/service`
  - `go test ./cmd/server -run '^$' -count=1`
  - `gofmt` and `git diff --check` passed.
- No real provider, Redis/PostgreSQL, container, deployment, or push operation was performed.
- Risk: full integration with external function-only providers was not exercised; validation is local fake-upstream and package/runtime tests only.

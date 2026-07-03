### DONE: upstream-main-v0143-antigravity-reasoning-params-s41

## Summary
- Ported upstream `f5b296127 fix: Handle invalid arguments correctly for Gemini reasoning models`.
- Marked Gemini reasoning models in the Antigravity model table and added `IsGeminiReasoningModel`.
- Antigravity now omits forced empty `toolConfig` for Gemini reasoning models when no tools are present.
- Antigravity now omits `stopSequences`, `temperature`, `topP`, and `topK` for Gemini reasoning models.
- Added focused tests for reasoning model parameter filtering and non-reasoning behavior preservation.

## Changed Files
- `backend/internal/pkg/antigravity/claude_types.go`
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run
- `gofmt -w backend/internal/pkg/antigravity/claude_types.go backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go`
- `go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_ReasoningModelOmitsInvalidArgs|TestBuildGenerationConfig_ReasoningModelOmitsUnsupportedParams|TestTransformClaudeToGeminiWithOptions_PreservesWebSearchAlongsideFunctions|TestTransformClaudeToGeminiWithOptions_MessageRoles" -count=1`
  - Result: PASS.
- `git diff --check -- backend/internal/pkg/antigravity/claude_types.go backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go docs/workflow/tasks/upstream-main-v0143-antigravity-reasoning-params-s41.md docs/workflow/worker-results/upstream-main-v0143-antigravity-reasoning-params-s41-result.md docs/workflow/qa-reports/upstream-main-v0143-antigravity-reasoning-params-s41-qa.md docs/workflow/status.md docs/workflow/main-log.md`
  - Result: PASS, with LF/CRLF warnings for workflow docs only.
- staged denied-path audit
  - Result: `NO_DENIED_PATHS` because no files were staged.

## Contract Compliance
- Changed only allowed Antigravity package files and workflow artifacts.
- Did not edit gateway, billing, usage, frontend, Ent, migrations, deploy, knowledge, dependencies, or generated files.
- Did not merge/rebase `v0.1.143` or cherry-pick broader release content.

## Risks
- The broader `v0.1.143` and post-release `upstream/main` queue still contains deferred candidates: user model stats, OpenAI subscription state, Codex session import identity, compact skip, keepalive/Bearer auth, and count_tokens latest scope.
- Full backend test suite was not run because S41 is a narrow Antigravity transformer fix and the worktree contains unrelated dirty files.

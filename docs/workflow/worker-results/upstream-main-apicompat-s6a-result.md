### DONE: upstream-main-apicompat-s6a

## Summary
- Worktree: `E:/codex-worktrees/sub2api/upstream-main-apicompat-s6a`
- Branch: `codex/upstream-main-apicompat-s6a`
- Contract: `docs/workflow/tasks/upstream-main-apicompat-s6a.md`
- Result: all candidate fixes are already present as equivalent tree content on this branch; direct `git cherry-pick -x` attempts were empty after conflict-resolution checks, so no empty candidate commits were created.

## Candidate Results
- `348a48773` APPLIED_EQUIVALENT: `backend/internal/service/openai_codex_transform.go` already preserves underscore when rewriting `call_*` ids.
- `a729752de` APPLIED_EQUIVALENT: `backend/internal/service/openai_codex_transform_test.go` already has the aligned `fc_` assertions.
- `e9a25e7b9` APPLIED_EQUIVALENT: `backend/internal/pkg/apicompat/types.go` and related tests already preserve empty streaming thinking blocks.
- `df82a3bc6` APPLIED_EQUIVALENT: null-content fallback and regression tests are already present; test insertion conflict resolved to the existing equivalent tree, then cherry-pick was empty.
- `276b5c775` APPLIED_EQUIVALENT: reasoning-model `temperature` / `top_p` stripping and tests are already present; test insertion conflict resolved to the existing equivalent tree, then cherry-pick was empty.
- `c4d7edba0` APPLIED_EQUIVALENT: developer-role-to-system mapping and bridge test are already present.

## Deferred
- None.

## Acceptance Commands
- `git status --short --branch` PASS
- `git diff --check` PASS
- `git diff --name-status f78566b5d..HEAD` PASS
- `git diff --name-only f78566b5d..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/main-log\\.md|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"` PASS: no denied-path matches; `rg` returned 1 because no files matched.
- `go test ./internal/pkg/apicompat -run "ChatCompletions|Responses|Anthropic|Thinking|Reasoning|Developer|Tool|Temperature|TopP" -count=1` PASS from `backend/`
- `go test ./internal/service -run "CodexTransform|OpenAICodex" -count=1` PASS from `backend/`

## Changed Files
- `docs/workflow/worker-results/upstream-main-apicompat-s6a-result.md`
- `docs/workflow/qa-reports/upstream-main-apicompat-s6a-qa.md`

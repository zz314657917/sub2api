# Task Contract

## Task ID
upstream-main-claude-count-tokens-s2g

## Role
Codex acts as Planner, Generator, and Final Evaluator for this small backend validation Sprint. Implement only the Claude Code `count_tokens` validator exception selected here.

## Goal
Port the safe subset of upstream `bf3787de1 fix(gateway): allow Claude Code count_tokens` onto the current upstream-sync branch. Claude Code's `/v1/messages/count_tokens` helper request should pass validator checks when the User-Agent matches Claude Code, because it may not include the full `/v1/messages` system prompt and headers.

## Success Criteria
- `ClaudeCodeValidator.Validate` still rejects non-Claude-Code User-Agent values.
- Non-`messages` paths keep the existing User-Agent-only behavior.
- `/v1/messages/count_tokens` keeps User-Agent validation but skips strict `/v1/messages` system prompt/header/metadata checks.
- Normal `/v1/messages` requests still require strict validation unless they are the existing max_tokens=1 haiku probe bypass.
- No gateway routing, schema, migration, config, public API, OpenAI WS/Responses bridge, or frontend behavior changes are introduced.

## Context
- Repo/worktree: `E:/codex-worktrees/sub2api/upstream-main-safe-patches-s1`
- Base branch for this Sprint: `codex/upstream-main-admin-account-redaction-s2f`
- Work branch: `codex/upstream-main-claude-count-tokens-s2g`
- Upstream source commit: `bf3787de1`
- Main worktree `F:/mcplugins/sub2api` must not be modified.

## Allowed Paths
- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `docs/workflow/tasks/upstream-main-claude-count-tokens-s2g.md`
- `docs/workflow/worker-results/upstream-main-claude-count-tokens-s2g-result.md`
- `docs/workflow/qa-reports/upstream-main-claude-count-tokens-s2g-qa.md`
- `docs/workflow/main-log.md`
- `knowledge/tasks/current-task.md`

## Denied Paths
- `backend/ent/**`
- `backend/migrations/**`
- `backend/cmd/server/**`
- `backend/internal/handler/**`
- `backend/internal/server/**`
- `frontend/**`
- `deploy/**`
- `README*`
- `.github/**`
- `assets/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- Payment, subscription notify, redeem expiry, channel monitor API mode, DingTalk, `user_platform_quotas`, OpenAI WS/Responses bridge redesign, OpenAI gateway routing redesign, or unrelated local UI/public page changes.

## Constraints
- Do not direct-merge `upstream/main`.
- Do not cherry-pick if it attempts to widen the patch beyond allowed paths.
- Keep the implementation limited to validator path classification and tests.
- Do not add live upstream smoke tests or require credentials.
- If the selected patch requires touching handlers, schema, frontend, config, or bridge code, stop and split a new Sprint.

## Candidate Commit
- Primary: `bf3787de1 fix(gateway): allow Claude Code count_tokens`

## Explicitly Deferred
- `a39163519` OpenAI key config model defaults, because it is an external product configuration policy update.
- `08e19bb15`, `d7bed40dd`, `08061717b` OpenAI WS bridge/failover changes, because they are bridge-sized.
- `f10bca815` Codex Responses bridge redesign.
- `003b2786d` apicompat wire test lint, because its target test file belongs to the deferred bridge test chain.
- Pricing resource updates such as `5fd9a3509`, unless the matching resource data is also intentionally synced in a separate Sprint.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
go test ./internal/service -run ClaudeCodeValidator -count=1
go test ./internal/service ./internal/handler -run "ClaudeCode|CountTokens" -count=1
```

## Output
- Write `docs/workflow/worker-results/upstream-main-claude-count-tokens-s2g-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-claude-count-tokens-s2g-qa.md`.
- Update `docs/workflow/main-log.md` with contract approval and QA events.
- Update `knowledge/tasks/current-task.md` with the current handoff snapshot after QA.

## Stop Rules
- Stop if implementation requires touching denied paths.
- Stop if tests fail for reasons requiring gateway handler, schema, config, frontend, or bridge changes.
- Stop if the validator change would allow non-Claude-Code clients through this path.

## Budget
- worker_mode: `codex-direct`
- qa_mode: `runtime`
- worktree_root: `E:/codex-worktrees`

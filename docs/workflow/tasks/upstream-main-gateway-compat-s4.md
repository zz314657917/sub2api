# Task Contract

## Task ID
upstream-main-gateway-compat-s4

## Role
Codex acts as Generator and Final Evaluator for this Sprint. Implement only the approved gateway/apicompat compatibility subset of upstream fixes.

## Goal
Port selected OpenAI Images, Chat Completions failed response, upstream `response.failed`, DeepSeek reasoning-only, and Responses-to-Anthropic tool pairing fixes from `upstream/main` onto a dedicated isolated branch without directly merging `upstream/main`. Preserve local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow changes.

## Success Criteria
- Selected upstream fixes are applied by cherry-pick or equivalent minimal porting.
- No Ent schema, SQL migration, frontend UI, public API field, production config, README/logo/deploy-only sync, DingTalk, notification email, user-platform quota, Channel Monitor API mode, or broad gateway refactor is introduced.
- OpenAI Images upstream errors preserve the real upstream error body and semantics where possible instead of being generalized to a misleading `502`.
- Chat/Responses compatibility paths preserve failed/error events and do not convert them into misleading success responses or generic errors.
- DeepSeek reasoning-only output remains visible through Responses/Chat compatibility conversion.
- Responses-to-Anthropic conversion keeps `tool_use` and `tool_result` blocks legally paired.
- Skipped or deferred commits are documented with a reason.
- Target checks and feasible regression commands are executed and recorded.

## Context
- Repo: `F:/mcplugins/sub2api`
- Isolated worktree: `E:/codex-worktrees/sub2api/upstream-main-gateway-compat-s4`
- Base branch: `main`
- Work branch: `codex/upstream-main-gateway-compat-s4`
- Upstream source: `upstream/main`
- Baseline local commit: `34d02457b`

## Allowed Paths
- `backend/internal/service/openai_images*`
- `backend/internal/service/openai_gateway_chat_completions*`
- `backend/internal/service/openai_gateway_responses_chat_fallback_test.go`
- `backend/internal/handler/openai_*`
- `backend/internal/handler/openai_gateway_handler*`
- `backend/internal/pkg/apicompat/**`
- `docs/workflow/tasks/upstream-main-gateway-compat-s4.md`
- `docs/workflow/worker-results/upstream-main-gateway-compat-s4-result.md`
- `docs/workflow/qa-reports/upstream-main-gateway-compat-s4-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `frontend/**`
- `backend/ent/**`
- `backend/migrations/**`
- `deploy/**`
- `knowledge/**`
- `docs/workflow/status.md`
- `docs/workflow/spec.md`
- `.github/**`
- `assets/**`
- `README*`

## Candidate Commits
- `381d1d6d6` fix OpenAI Images upstream real error propagation.
- `2e212d18e` handle failed responses in Chat Completions compatibility.
- `5bd3d9043` preserve upstream `response.failed` errors.
- `9b99f6c1f` surface DeepSeek reasoning-only replies in apicompat.
- `60867022b` repair `tool_use`/`tool_result` pairing on the Responses-to-Anthropic path.

## Explicitly Deferred
- Broad gateway architecture refactor chain beyond the five listed commits.
- DingTalk, notification emails, user-platform quota, Channel Monitor API mode, frontend UI, Ent/migration, version/sponsors, and CI-only changes.
- Upstream deletion or restructuring of local model market, APIMart billing, tickets, Canvas, Chat/Image Studio, OpenWebUI, knowledge, and workflow features.

## Constraints
- Do not direct-merge `upstream/main`.
- Work only inside the isolated worktree for this Sprint.
- Prefer `git cherry-pick -x`; if conflicts would touch denied paths or broaden scope, stop that commit and document it as deferred.
- Keep local behavior when local code already contains an equivalent fix.
- Do not run code generation that rewrites denied generated code.
- Do not include generated frontend build output, Docker artifacts, `node_modules`, or unrelated temp files.

## Acceptance Commands
```powershell
git status --short --branch
git diff --check
git diff --name-status main..HEAD
go test ./internal/pkg/apicompat -run "Responses|ChatCompletions|Anthropic|Tool|DeepSeek|Reasoning" -count=1
go test ./internal/service -run "OpenAIImages|ChatCompletions|Responses|Failed|Tool|DeepSeek" -count=1
go test ./internal/handler -run "OpenAI|Gateway|Images|Failed" -count=1
go test ./internal/pkg/apicompat ./internal/service ./internal/handler -count=1
```

Run `go test ./internal/server/routes ./cmd/server -count=1` only if this Sprint touches route/server wiring.

## Output
- Write `docs/workflow/worker-results/upstream-main-gateway-compat-s4-result.md`.
- Write `docs/workflow/qa-reports/upstream-main-gateway-compat-s4-qa.md`.
- Update `docs/workflow/main-log.md` with contract, implementation, and QA events.

## Stop Rules
- Stop a candidate commit if it requires denied paths.
- Stop a candidate commit if conflict resolution requires new schema, migration, frontend UI, public API fields, production config, or broad gateway architecture changes.
- Stop Sprint implementation if the working tree cannot be returned to a clean state between candidate commits.

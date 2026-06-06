### DONE: upstream-main-openai-ops-hardening-s5

## Summary
- Created isolated worktree `E:/codex-worktrees/sub2api/upstream-main-openai-ops-hardening-s5` and branch `codex/upstream-main-openai-ops-hardening-s5` from baseline `b708d0552`.
- Added Sprint contract at `docs/workflow/tasks/upstream-main-openai-ops-hardening-s5.md`.
- Ported all five approved OpenAI/Ops/admin backend hardening fixes without directly merging `upstream/main`.
- No candidate commit was deferred.

## Commits
- `5ce3cdb1b` docs: add openai ops hardening sprint contract
- `e7ec84b60` cherry-pick of `8e27ff20a`: handle missing messages stream terminal.
- `35e30ec28` cherry-pick of `86d9b6bff`: self-heal stale Codex used-percent snapshots and lock semantics.
- `bd0a267d9` cherry-pick of `32ef47110`: treat allowed proxy quality statuses as pass.
- `d1d0377be` cherry-pick of `bc7ce1857`: persist cleared group descriptions.
- `3e8368c6d` cherry-pick of `d626ccce1`: recognize Claude Code clients via billing block.

## Notes
- `8e27ff20a` conflicted in `backend/internal/handler/openai_gateway_handler.go`; resolved by keeping the upstream writer-size guard and deferred account-slot release while preserving local failover handling.
- `86d9b6bff` conflicted in Codex usage snapshot tests and `openai_gateway_service.go`; resolved by adding the stale snapshot guard while keeping local quota helpers and avoiding duplicate test names/account IDs.
- `d626ccce1` conflicted only in `claude_code_validator_test.go`; resolved by keeping the new billing-block tests plus existing local validator tests.
- `32ef47110` and `bc7ce1857` cherry-picked cleanly.

## Changed Files
- `backend/internal/handler/admin/group_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_group_test.go`
- `backend/internal/service/admin_service_proxy_quality_test.go`
- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `backend/internal/service/openai_account_scheduler_test.go`
- `backend/internal/service/openai_compat_model_test.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_snapshot_test.go`
- `backend/internal/service/testdata/security_monitor_system_prompt.txt`
- `docs/workflow/tasks/upstream-main-openai-ops-hardening-s5.md`

## Verification
- `git status --short --branch`
- `git diff --check`
- `git diff --name-status b708d0552..HEAD`
- `git diff --name-only b708d0552..HEAD | rg "^(frontend/|backend/ent/|backend/migrations/|deploy/|knowledge/|docs/workflow/status\\.md|docs/workflow/spec\\.md|\\.github/|assets/|README)"`
- `go test ./internal/service -run "OpenAI|Codex|Proxy|Group|Claude|Terminal|Snapshot|Quality" -count=1`
- `go test ./internal/handler ./internal/service -run "OpenAI|Gateway|Group|Proxy|Claude|Terminal" -count=1`
- `go test ./internal/service ./internal/handler -count=1`

## Integration Verification
- Clean integration worktree `E:/codex-worktrees/sub2api/upstream-main-openai-ops-hardening-s5-integration` was created from current `main@b708d0552`.
- `codex/upstream-main-openai-ops-hardening-s5` merged without conflicts into integration commit `a121e6389`.
- Integration reran the same path audit and Go target/regression tests; all passed.

# Task Contract: upstream-runtime-hotfixes-s66a

## Task ID

`upstream-runtime-hotfixes-s66a`

## Status

`approved`

## Role

You are the Generator worker for the narrow runtime-hotfix lane. Implement only this contract; do not make architecture decisions or broaden scope.

## Goal

Selectively adapt upstream setup-token refresh, released `opsCaptureWriter` nil guards, and Windows WebSocket reset recognition to the local tree without importing unrelated upstream refactors.

## Success Criteria

- Anthropic `setup-token` accounts with a non-empty refresh token are included in automatic refresh candidates and accepted by `ClaudeTokenRefresher`.
- Existing OAuth refresh behavior and refresh-window/distributed-lock behavior remain unchanged.
- Every delegated `gin.ResponseWriter` method on a released `opsCaptureWriter` is nil-safe and covered by regression tests.
- Windows reset messages such as `forcibly closed by the remote host` are classified as disconnects without weakening other error handling.
- Targeted repository, service, handler, and WebSocket tests pass.

## Context

- Repo: `F:/mcplugins/sub2api`
- Baseline: current branch commit containing S65 latency health.
- Upstream references: `99da30819`, `a495d5e30`, `89a551b96`, `bc3cb2902`, `0a5f34a2e`.
- Preserve the local monolithic/file layout; direct cherry-pick is not required.

## Allowed Paths

- `backend/internal/repository/account_repo.go`
- `backend/internal/repository/account_repo_test.go`
- `backend/internal/repository/account_repo_integration_test.go`
- `backend/internal/service/token_refresher.go`
- `backend/internal/service/token_refresher_test.go`
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/ops_error_logger_test.go`
- `backend/internal/handler/ops_capture_writer_nil_test.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay_test.go`
- `backend/internal/service/openai_ws_v2/passthrough_relay_internal_test.go`
- `docs/workflow/worker-results/upstream-runtime-hotfixes-s66a-result.md`

## Denied Paths

- All paths not listed above.
- `frontend/**`, `backend/migrations/**`, payment, billing, deployment, and production configuration.
- `knowledge/**` and global memories.

## Constraints

- Keep the patch minimal; do not split or reformat surrounding files.
- Do not change refresh cadence, retry count, token credential formats, or account status semantics.
- Do not change normal response-writer behavior while the inner writer is present.
- Do not modify another worker's files.

## Acceptance Commands

```powershell
Push-Location backend
go test ./internal/repository -run "Test.*OAuthRefreshCandidates|Test.*RefreshCandidate" -count=1
go test ./internal/service -run "TestClaudeTokenRefresher|Test.*DisconnectError" -count=1
go test ./internal/handler -run "TestOpsCaptureWriter" -count=1
go test ./internal/service/openai_ws_v2 -run "Test.*Disconnect" -count=1
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-runtime-hotfixes-s66a-result.md` using the worker-result format.
- First line must be `### DONE: upstream-runtime-hotfixes-s66a`, `### BLOCKED: ...`, or `### FAILED: ...`.
- Commit the implementation and report on the assigned worktree branch.

## Stop Rules

- Stop if the fix requires schema migration, credential-format changes, payment/billing changes, or files outside Allowed Paths.
- Stop if an existing local behavior conflicts with the upstream assumption and cannot be resolved without architecture judgment.
- Do not revert other worktree or user changes.

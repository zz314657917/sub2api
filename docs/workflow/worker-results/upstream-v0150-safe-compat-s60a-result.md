### DONE: upstream-v0150-safe-compat-s60a

- task_id: `upstream-v0150-safe-compat-s60a`
- branch: `pge/upstream-v0150-safe-compat-s60a`
- baseline: `3332c6883e7480f030fcffbccb6dc7ee0a3f69ca`
- contract: `docs/workflow/tasks/upstream-v0150-safe-compat-s60a.md`

Changed files:
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/handler/admin/dashboard_handler.go`
- `backend/internal/handler/admin/dashboard_handler_user_breakdown_test.go`
- `frontend/src/api/admin/dashboard.ts`
- `docs/workflow/worker-results/upstream-v0150-safe-compat-s60a-result.md`

Implementation summary:
- Added `ChatCompletionsRequest.ParallelToolCalls *bool` and preserved pointer semantics across Chat Completions -> Responses and Responses -> Chat Completions conversion.
- Ensured serialized payloads retain explicit `"parallel_tool_calls": false` when the source pointer is non-nil false.
- Updated admin `GetUserBreakdown` to parse `request_type` through `service.ParseUsageRequestType`, returning HTTP 400 for invalid values.
- Updated frontend `UserBreakdownParams.request_type` to use `UsageRequestType` instead of `number`.
- Added focused backend tests for both compatibility behaviors and the admin filter parser.

Commands run:
- `git status --short`
- `git rev-parse --verify HEAD && git branch --show-current`
- `git show --stat --oneline ad8afc8a2`
- `git show --stat --oneline dda8f7873`
- `git show --stat --oneline ea9f40b63`
- `go test ./internal/pkg/apicompat -run 'ParallelToolCalls|ChatCompletionsToResponses|ResponsesToChatCompletionsRequest' -count=1`
- `go test ./internal/handler/admin -run 'TestGetUserBreakdown' -count=1`
- `npm.cmd run typecheck`
- `gofmt -w <changed-go-files>`
- `git diff --check`
- `rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend frontend`
- fixed-baseline denied-path audit script from contract

Key output:
- `go test ./internal/pkg/apicompat ...` -> `ok`
- `go test ./internal/handler/admin ...` -> `ok`
- `npm.cmd run typecheck` -> passed after restoring missing frontend dependency installation
- denied-path audit -> `DENIED_PATH_AUDIT_PASS`
- conflict-marker scan -> no matches
- `git diff --check` -> clean

Notes:
- Initial `npm.cmd run typecheck` failed because `frontend/node_modules/.bin/vue-tsc` was missing in the worktree environment. I restored frontend dependencies locally with `npm.cmd install`, reran typecheck successfully, and removed the generated out-of-contract `frontend/package-lock.json` before final denied-path audit.

Risks:
- Frontend dependency installation was environment recovery, not a source change. No dependency manifests or lockfiles are included in the final diff.
- No API/schema/generated-code/deployment/payment/image paths were touched.

Contract compliance:
- Only Allowed Paths were modified.
- Denied Paths were not modified.
- Stop Rules were not triggered.
- Implementation stayed within the upstream behavior described by `ad8afc8a2`, `dda8f7873`, and `ea9f40b63`.

knowledge_candidates:
- none

### PASS: upstream-v027-openai

## Changed Paths

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_chat_roles.go`
- `backend/internal/service/openai_codex_models_service.go`
- `backend/internal/service/openai_compat_v027_test.go`
- `docs/workflow/worker-results/upstream-v027-openai.md`

## Upstream Mapping

- `18bfa4bf2`: strict Chat upstreams translate only `developer` to `system` after account and target URL selection. CN platform API-key accounts are strict; OpenAI-compatible API-key accounts require one of the exact official hostnames. The raw-message representation preserves unknown fields and numeric literals, and does not mutate the retry input.
- `2f16e0984` + `31f3003ff`: manifest validation keeps the exact, case-sensitive `models` key and array-token validation, relying on the enclosing JSON decode for malformed-array rejection instead of parsing the same value twice.
- `d7ee1ab6b`: HTTP response affinity uses `context.WithoutCancel` with the existing `openAIWSStateStoreRedisTimeout`, retaining values while limiting durable write time. No response-owner subsystem was added.

## Commands

```powershell
cd E:/codex-worktrees/sub2api/upstream-v027-port/backend
gofmt -w internal/service/openai_gateway_service.go internal/service/openai_gateway_chat_completions_raw.go internal/service/openai_chat_roles.go internal/service/openai_codex_models_service.go internal/service/openai_compat_v027_test.go
go test ./internal/service -run '^TestV027OpenAI' -count=1
go build ./...
```

All commands passed. The runtime test reaches a task-owned `httptest` localhost upstream through the raw Chat forwarding path. It verifies a CN API-key strict target sends `system`, retains unknown numeric input, and leaves the original retry body as `developer`. Focused tests also cover exact-host scope, OAuth/non-strict preservation, invalid strict payloads, manifest null/wrong type/malformed/case variants, cancellation, value propagation, deadline bounds, and nil guards.

## Final Checks

- `git diff HEAD --check`: passed.
- `git diff --name-only --diff-filter=U`: empty.
- No staging, commit, push, deployment, or external provider call was performed.

## Limits

The evidence uses local fake HTTP transport only. It does not establish real strict-provider acceptance, Redis availability, or deployed runtime behavior. Other dirty paths in the shared worktree belong to concurrent tasks and were not modified.

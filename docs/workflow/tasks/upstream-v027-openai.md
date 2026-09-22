---
status: approved
review_verdict: PASS
task_id: upstream-v027-openai
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---

## Task ID
upstream-v027-openai
## Role
Independent Terra Developer, separately dispatched Terra QA. Controller owns contract and final decision.
## Goal
Port 18bfa4bf2 strict Chat developer roles, final combination 2f16e0984 +31f3003ff manifest validation, d7ee1ab6b cancel-safe affinity.
## Success Criteria
In monolith only bindHTTPResponseAccount may change. Preserve all guards and existing group/cache key semantics. Bounded context.WithoutCancel with existing timeout constant, nil context handled, no new response owner subsystem. Test canceled request, value propagation, deadline enforcement and nil/invalid inputs. Role normalization only strict CN platform/APIKey or exact official hostname, preserve other roles, OAuth, unknown JSON fields and numeric precision; localhost fake upstream payload test. Manifest map key stays case-sensitive; remove only redundant parsing, reject null, Models, malformed/wrong-type arrays, accept empty array. TestV027OpenAI* names.
## Allowed Paths
backend/internal/service/openai_gateway_service.go
backend/internal/service/openai_gateway_chat_completions_raw.go
backend/internal/service/openai_chat_roles.go
backend/internal/service/openai_codex_models_service.go
backend/internal/service/openai_compat_v027_test.go
docs/workflow/worker-results/upstream-v027-openai.md
docs/workflow/qa-reports/upstream-v027-openai.md
## Denied Paths
Everything else, especially main checkout, migrations, dependencies, databases/containers, secrets and memory.
## Constraints
Behavior port only. No whole-file upstream overwrite, cherry-pick, commit, push, deployment or real paid providers. Preserve earlier Anthropic port. Shared monolith editing is SERIAL between DeepSeek and OpenAI agents. No other agent may edit assigned paths concurrently. Preserve unrelated dirty files. Deterministic httptest runtime evidence required; no external service calls.
## Acceptance Commands
Run in backend:
```powershell
go test ./internal/service -run '^TestV027OpenAI' -count=1
go build ./...
```
From root: git diff HEAD --check; git diff --name-only --diff-filter=U (must empty). Format only listed Go files with gofmt. Include new-file diff inspection and changed path allowlist. Baseline d6e6c3471 apicompat and full build freshly PASS; any alleged unrelated failure must be reproduced using git show d6e6c3471:<path> in separate task-owned temporary source with exact original command, never changing current files to fake a pass.
## Output
Report first line ### PASS/FAIL/BLOCKED: upstream-v027-openai, changed paths, commands, runtime evidence, upstream mapping, remaining limits. QA reruns independently.
## Stop Rules
Missing dependency or need for out-of-allowlist edits => notify controller before edit. Terra unavailable => BLOCKED, no model fallback. No claiming real provider/deployment success from mock tests.

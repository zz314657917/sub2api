---
status: approved
review_verdict: PASS
task_id: upstream-v027-cn-quota
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---

## Task ID
upstream-v027-cn-quota
## Role
Independent Terra Developer, separately dispatched Terra QA. Controller owns contract and final decision.
## Goal
Port db8692d67 CN Coding Plan quota-exhausted 403 temporary cooldown.
## Success Criteria
Use existing cnAccountIsCodingPlan (read-only cn_provider_quota_service.go). Exact structured/text exhaustion only CN coding plan, preserve ordinary 403, concurrency-specific 403 and all existing breakers. Future reset snapshot follows existing semantics; no snapshot uses bounded temp pause. SetRateLimited failure must fallback, SetTempUnschedulable failure observable; never SetError for recognized exhaustion. TestV027CNQuota* table tests include target/non-target, stale/future/missing snapshots, persistence failures; exercise HandleUpstreamError public entry and fake repository state.
## Allowed Paths
backend/internal/service/ratelimit_cn_providers.go
backend/internal/service/ratelimit_service.go
backend/internal/service/cn_quota_v027_test.go
docs/workflow/worker-results/upstream-v027-cn-quota.md
docs/workflow/qa-reports/upstream-v027-cn-quota.md
## Denied Paths
Everything else, especially main checkout, migrations, dependencies, databases/containers, secrets and memory.
## Constraints
Behavior port only. No whole-file upstream overwrite, cherry-pick, commit, push, deployment or real paid providers. Preserve earlier Anthropic port. Shared monolith editing is SERIAL between DeepSeek and OpenAI agents. No other agent may edit assigned paths concurrently. Preserve unrelated dirty files. Deterministic httptest runtime evidence required; no external service calls.
## Acceptance Commands
Run in backend:
```powershell
go test ./internal/service -run '^TestV027CNQuota' -count=1
go build ./...
```
From root: git diff HEAD --check; git diff --name-only --diff-filter=U (must empty). Format only listed Go files with gofmt. Include new-file diff inspection and changed path allowlist. Baseline d6e6c3471 apicompat and full build freshly PASS; any alleged unrelated failure must be reproduced using git show d6e6c3471:<path> in separate task-owned temporary source with exact original command, never changing current files to fake a pass.
## Output
Report first line ### PASS/FAIL/BLOCKED: upstream-v027-cn-quota, changed paths, commands, runtime evidence, upstream mapping, remaining limits. QA reruns independently.
## Stop Rules
Missing dependency or need for out-of-allowlist edits => notify controller before edit. Terra unavailable => BLOCKED, no model fallback. No claiming real provider/deployment success from mock tests.

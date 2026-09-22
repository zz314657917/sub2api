---
status: approved
review_verdict: PASS
task_id: upstream-v027-deepseek
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---

## Task ID
upstream-v027-deepseek
## Role
Independent Terra Developer, separately dispatched Terra QA. Controller owns contract and final decision.
## Goal
Port 881ab1b0c + acc05620c: native DeepSeek Responses media lifting and contiguous parallel tool results.
## Success Criteria
Only normalizeDeepSeekResponsesRequestBody in monolithic service may change. Keep store=false and previous_response_id deletion. Decode UseNumber; retain large integer precision, call IDs, text, order across user/assistant turn boundaries. Lift images into following user message after complete output batch, developer/system notices after batch. Preserve plain output/no-op and all non-DeepSeek/non-Responses accounts. Tests TestV027DeepSeek* must include localhost fake upstream actual forwarding payload with two calls, images+text, notices, numeric precision and non-target accounts.
## Allowed Paths
backend/internal/pkg/apicompat/responses_tool_output_media.go
backend/internal/pkg/apicompat/responses_tool_output_media_test.go
backend/internal/service/openai_gateway_service.go
backend/internal/service/deepseek_media_v027_test.go
docs/workflow/worker-results/upstream-v027-deepseek.md
docs/workflow/qa-reports/upstream-v027-deepseek.md
## Denied Paths
Everything else, especially main checkout, migrations, dependencies, databases/containers, secrets and memory.
## Constraints
Behavior port only. No whole-file upstream overwrite, cherry-pick, commit, push, deployment or real paid providers. Preserve earlier Anthropic port. Shared monolith editing is SERIAL between DeepSeek and OpenAI agents. No other agent may edit assigned paths concurrently. Preserve unrelated dirty files. Deterministic httptest runtime evidence required; no external service calls.
## Acceptance Commands
Run in backend:
```powershell
go test ./internal/pkg/apicompat -count=1
go test ./internal/service -run '^TestV027DeepSeek' -count=1
go build ./...
```
From root: git diff HEAD --check; git diff --name-only --diff-filter=U (must empty). Format only listed Go files with gofmt. Include new-file diff inspection and changed path allowlist. Baseline d6e6c3471 apicompat and full build freshly PASS; any alleged unrelated failure must be reproduced using git show d6e6c3471:<path> in separate task-owned temporary source with exact original command, never changing current files to fake a pass.
## Output
Report first line ### PASS/FAIL/BLOCKED: upstream-v027-deepseek, changed paths, commands, runtime evidence, upstream mapping, remaining limits. QA reruns independently.
## Stop Rules
Missing dependency or need for out-of-allowlist edits => notify controller before edit. Terra unavailable => BLOCKED, no model fallback. No claiming real provider/deployment success from mock tests.

---
status: approved
review_verdict: PASS
task_id: upstream-v027-gemini
worker_model: gpt-5.6-terra
base_commit: d6e6c34718e6eee6388391b346f40e1492e81d57
spec_ref: docs/workflow/tasks/upstream-v027-remaining.md
---

## Task ID
upstream-v027-gemini
## Role
Independent Terra Developer, separately dispatched Terra QA. Controller owns contract and final decision.
## Goal
Port 0f4d8acaa/f79b8bf96/da74bf13a Gemini thinking mapping, SDK-compatible heartbeats, account-mapped model listing.
## Success Criteria
Local ForwardGemini and native streaming in antigravity_gateway_service.go own mapping/heartbeat (not upstream split files). Only bare supported model gets thinkingConfig-based variant; explicit suffix and account custom mapping priority preserved. go-genai/python-genai clients omit comment heartbeat but other clients retain existing heartbeat; preserve cancellation/error/usage semantics. GeminiV1BetaListModels and compat account listing must retain custom group list precedence, forced platform behavior, mixed opt-in, group visibility/allowlist and upstream metadata/pagination/errors. TestV027Gemini* exercise fake HTTP/SSE traffic and real handler, supported/unsupported models and non-target clients. If exact streaming owner differs, report before edit.
## Allowed Paths
backend/internal/service/antigravity_gateway_service.go
backend/internal/service/antigravity_gemini_thinking_variant.go
backend/internal/service/gemini_sse_comment_compat.go
backend/internal/service/gemini_messages_compat_service.go
backend/internal/handler/gemini_v1beta_handler.go
backend/internal/service/gemini_compat_v027_test.go
backend/internal/handler/gemini_models_v027_test.go
docs/workflow/worker-results/upstream-v027-gemini.md
docs/workflow/qa-reports/upstream-v027-gemini.md
## Denied Paths
Everything else, especially main checkout, migrations, dependencies, databases/containers, secrets and memory.
## Constraints
Behavior port only. No whole-file upstream overwrite, cherry-pick, commit, push, deployment or real paid providers. Preserve earlier Anthropic port. Shared monolith editing is SERIAL between DeepSeek and OpenAI agents. No other agent may edit assigned paths concurrently. Preserve unrelated dirty files. Deterministic httptest runtime evidence required; no external service calls.
## Acceptance Commands
Run in backend:
```powershell
go test ./internal/service ./internal/handler -run '^TestV027Gemini' -count=1
go build ./...
```
From root: git diff HEAD --check; git diff --name-only --diff-filter=U (must empty). Format only listed Go files with gofmt. Include new-file diff inspection and changed path allowlist. Baseline d6e6c3471 apicompat and full build freshly PASS; any alleged unrelated failure must be reproduced using git show d6e6c3471:<path> in separate task-owned temporary source with exact original command, never changing current files to fake a pass.
## Output
Report first line ### PASS/FAIL/BLOCKED: upstream-v027-gemini, changed paths, commands, runtime evidence, upstream mapping, remaining limits. QA reruns independently.
## Stop Rules
Missing dependency or need for out-of-allowlist edits => notify controller before edit. Terra unavailable => BLOCKED, no model fallback. No claiming real provider/deployment success from mock tests.

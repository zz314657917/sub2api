---
status: approved
review_verdict: PASS
task_id: upstream-v0215-protocol-20261009
worker_model: current-session-default
base_commit: ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b
spec_ref: docs/workflow/tasks/upstream-v0215-20261009-spec.md
---
# Task Contract
## Task ID
upstream-v0215-protocol-20261009
## Role
Developer; implement only after independent review PASS. Other Developers share this isolated checkout with disjoint paths.
## Goal
Adapt v0.2.15 protocol fixes: developer role, named tool_choice, deterministic legacy function pairing, refusals, thinking signature serialization, negative buffered block indexes, Anthropic ping as SSE comment.
## Success Criteria
- Preserve developer/system role and content. Flatten only named Chat function tool_choice; preserve other choices.
- Pair repeated legacy function names in order with unique collision-free deterministic call IDs; keep modern tool calls unchanged.
- Preserve refusals in streamed/nonstreamed/buffered Responses-to-Chat conversion without duplicating text.
- Thinking blocks always include signature, including empty strings, without leaking irrelevant fields into other block types.
- Reject negative block indexing in every existing compatible buffered path. Stream pings remain valid SSE comments rather than Anthropic events.
- New named regression tests prove each case and existing package tests pass or exact unrelated baseline blockers are documented. Production backend build is required by QA.
## Context
Upstream commits: 9a8ab4f88, f3d752693, 7a767d5ce, 775c1a3c1, 412718edf, 15f469c86, 9f84204c2. Read using git show; adapt behavior, do not cherry-pick whole commits.
## Allowed Paths
- `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- `backend/internal/pkg/apicompat/responses_to_chatcompletions.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/pkg/apicompat/*v0215*test.go`
- `backend/internal/pkg/apicompat/anthropic_responses_test.go`
- `backend/internal/pkg/antigravity/claude_types.go`
- `backend/internal/pkg/antigravity/stream_transformer.go`
- `backend/internal/pkg/antigravity/*v0215*test.go`
- `backend/internal/service/gateway_forward_as_chat_completions.go`
- `backend/internal/service/gateway_forward_as_responses.go`
- `backend/internal/service/openai_gateway_chat_completions_anthropic_native.go`
- `backend/internal/service/openai_gateway_responses_anthropic_native.go`
- `backend/internal/service/*v0215_protocol*test.go`
- `docs/workflow/worker-results/upstream-v0215-protocol-20261009-result.md`
## Denied Paths
- All paths outside this allowlist, especially other worker files, knowledge, migrations, dependencies, production config, shared original repository.
## Constraints
- Preserve copied pre-existing dirty changes. Minimal hunks only; no whole-file replacement from upstream, no commit/push/deploy.
- No edits to baseline fixtures to hide failures. External temporary compile overlays may repair independently proven pre-existing signature drift; log original error and overlay contents.
- Narrow relevant-test exception: update only TestStreamingReasoning's content_block SSE expectation to require the newly mandatory empty signature field; preserve the copied unrelated ThinkingDisabled test changes. This assertion update is part of the requested wire-contract change, not suppression of a baseline failure.
## Acceptance Commands
From backend: `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/antigravity`; `go test -tags unit ./internal/service -run V0215Protocol -count=1`; `go build ./cmd/server`; `git diff --check`.
## Output
Write result with first line `### DONE: upstream-v0215-protocol-20261009` or FAILED/BLOCKED. Include changed files, commands and exit codes, test counts, risks, baseline blockers, knowledge_candidates. Write only own report.
## Stop Rules
Missing review PASS or approved frontmatter; model unavailable; out-of-scope architectural need; conflicts with other owners: stop and report, no silent workaround.

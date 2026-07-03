---
status: approved
owner: codex
qa_mode: runtime
created_at: 2026-07-03 15:44 +08:00
---

# Task Contract

## Task ID
upstream-main-v0143-claude-code-stream-keepalive-s50

## Role
Codex acts as Planner, Generator, and Final Evaluator in the isolated worktree `E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44`.

## Goal
Port upstream `a5781fe31` so Anthropic/Claude streaming keepalive is based on downstream idle time and affected Claude Code versions receive safe no-op content deltas during open content blocks instead of `event: ping`.

## Success Criteria
- Standard Anthropic streaming keepalive uses an idle timer that resets after downstream output, avoiding periodic pings immediately after normal stream data.
- Anthropic API-key passthrough streaming uses the same idle-timer behavior.
- Claude Code user agents at or above `claude-cli/2.1.193` receive empty `content_block_delta` keepalives during open text/tool/thinking blocks.
- Older Claude Code clients and non-Claude clients continue to receive regular `event: ping` keepalives.
- Usage parsing, model replacement, cache usage patching, stream timeout handling, and client-disconnect draining remain unchanged.
- No frontend, i18n, Anthropic API key auth scheme UI, Ent, migrations, deploy, README, `.github`, or knowledge files are modified.

## Allowed Paths
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_service_streaming_test.go`
- `docs/workflow/tasks/upstream-main-v0143-claude-code-stream-keepalive-s50.md`
- `docs/workflow/worker-results/upstream-main-v0143-claude-code-stream-keepalive-s50-result.md`
- `docs/workflow/qa-reports/upstream-main-v0143-claude-code-stream-keepalive-s50-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `backend/cmd/server/wire_gen.go`
- `backend/ent/**`
- `backend/migrations/**`
- `frontend/**`
- `deploy/**`
- `knowledge/**`
- `.github/**`
- `README*`
- Any unrelated dirty file from the main worktree.

## Constraints
- Do not merge all of `upstream/main` or the full release.
- Do not include upstream `7869b7fe3` Anthropic API Key Bearer auth in this Sprint.
- Do not change billing, account selection, tool-name reversal, cache TTL override, or SSE usage extraction semantics except where required for keepalive timing.
- Upstream patch does not apply cleanly to local `gateway_service.go`; hand-port only the keepalive helpers, timer reset behavior, and targeted tests.
- Do not use `git add .`.

## Acceptance Commands
```powershell
cd E:/codex-worktrees/sub2api/upstream-main-v0143-group-peak-rate-impl-s44/backend
go test ./internal/service -run "TestGatewayService_StreamingKeepalive|TestGatewayService_StreamingReusesScannerBufferAndStillParsesUsage|TestDetachUpstreamContextIgnoresClientCancel" -count=1
cd ..
git diff --check -- backend/internal/service/gateway_service.go backend/internal/service/gateway_service_streaming_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0143-claude-code-stream-keepalive-s50.md docs/workflow/worker-results/upstream-main-v0143-claude-code-stream-keepalive-s50-result.md docs/workflow/qa-reports/upstream-main-v0143-claude-code-stream-keepalive-s50-qa.md
git diff --cached --name-only | rg "^(backend/cmd/server/wire_gen.go|backend/ent/|backend/migrations/|frontend/|deploy/|knowledge/|\.github/|README)" || echo NO_DENIED_PATHS
```

## Output
- Final implementation commit on `codex/upstream-main-v0143-claude-code-stream-keepalive-s50`.
- Worker result: `docs/workflow/worker-results/upstream-main-v0143-claude-code-stream-keepalive-s50-result.md`.
- QA report: `docs/workflow/qa-reports/upstream-main-v0143-claude-code-stream-keepalive-s50-qa.md`.
- Updated `docs/workflow/status.md` and `docs/workflow/main-log.md`.

## Stop Rules
- Stop if the fix requires changing frontend account settings, Anthropic auth scheme persistence, billing, migrations, or unrelated streaming transformations.
- Stop if keepalive tests require long sleeps beyond the existing upstream-style short idle interval pattern.
- Stop if staged diff includes denied paths.

## Review Result
- Reviewed at: 2026-07-03 15:44 +08:00.
- Verdict: approved.
- Reason: the upstream patch is backend-only, addresses a concrete Claude Code streaming stall, and can be verified with local targeted SSE tests.

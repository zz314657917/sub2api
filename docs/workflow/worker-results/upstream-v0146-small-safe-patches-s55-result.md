### PASS: upstream-v0146-small-safe-patches-s55

Changed files:
- OpenAI compact usage handling: `openai_gateway_service.go`, `openai_gateway_service_test.go`.
- Codex function call id filtering: `openai_codex_transform.go`, `openai_codex_function_call_id_test.go`.
- Antigravity forward base default: `antigravity_gateway_service.go`.
- Usage log overflow handling: config, repository, gateway recording, and worker-pool files.
- Concurrency slot cleanup: `concurrency_cache.go`, `concurrency_service.go`, related handler/service/repository tests, and `testutil` stubs.

Implementation notes:
- Cherry-picked all five selected upstream commits with `-x` and no conflicts.
- Kept the branch based on `origin/main` in the isolated worktree `E:/codex-worktrees/sub2api/upstream-s55-small-safe-patches`.
- The concurrency cleanup commit builds on the S54 API Key concurrency display work without changing frontend key-list columns.
- No Ent, migration, deploy, README, container, or unrelated frontend files were changed.

Commands run:
- `go test ./internal/service -run "Test.*(Compact|SSE|Codex|FunctionCall|Antigravity|Usage|Queue|Concurrency|Slot)" -count=1` PASS.
- `go test ./internal/repository -run "Test.*(UsageLog|Concurrency)" -count=1` PASS.
- `go test ./internal/config -count=1` PASS.
- `go test ./internal/handler -run "Test.*(Gateway|Concurrency|Warmup|Fastpath|Hotpath)" -count=1` PASS.
- `git diff --check origin/main..HEAD` PASS.

Risks / follow-up:
- This S55 batch was validated at targeted Go test level only; no live Redis/runtime smoke or local container rebuild was run.
- Conflicting or broad upstream candidates remain deferred, including websearch history filtering, request-body parse logging, image namespace, Grok video, messages fallback, and batch-image feature series.

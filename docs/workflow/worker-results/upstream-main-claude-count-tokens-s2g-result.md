### DONE: upstream-main-claude-count-tokens-s2g

## Task ID
upstream-main-claude-count-tokens-s2g

## Status
done

## Summary
- Completed a minimal semantic port of upstream `bf3787de1 fix(gateway): allow Claude Code count_tokens`.
- Claude Code `/v1/messages/count_tokens` now passes `ClaudeCodeValidator` with a valid Claude Code User-Agent even when the request lacks the full `/v1/messages` system prompt and headers.
- Non-Claude-Code User-Agent values are still rejected, and normal `/v1/messages` strict validation is unchanged.
- No handler, schema, migration, config, frontend, OpenAI WS, or bridge files were changed.

## Changed Files
- `backend/internal/service/claude_code_validator.go`
- `backend/internal/service/claude_code_validator_test.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-claude-count-tokens-s2g.md`
- `docs/workflow/worker-results/upstream-main-claude-count-tokens-s2g-result.md`
- `docs/workflow/qa-reports/upstream-main-claude-count-tokens-s2g-qa.md`
- `knowledge/tasks/current-task.md`

## Commands Run
```text
git status --short --branch -> on codex/upstream-main-claude-count-tokens-s2g
gofmt -w backend/internal/service/claude_code_validator.go backend/internal/service/claude_code_validator_test.go -> pass
git diff --check -> pass
go test ./internal/service -run ClaudeCodeValidator -count=1 -> pass
go test ./internal/service ./internal/handler -run "ClaudeCode|CountTokens" -count=1 -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/handler
```

## Implementation Notes
- Added `isMessagesCountTokensPath(path)` with a narrow suffix check for `/messages/count_tokens`.
- Inserted the count_tokens exemption only after the Claude Code User-Agent check and before normal `/v1/messages` strict validation.
- Added tests for:
  - valid Claude Code User-Agent on `/v1/messages/count_tokens` passes with a minimal body;
  - non-Claude-Code User-Agent on `/v1/messages/count_tokens` remains rejected.

## Candidate Review Notes
- Equivalent locally, not re-ported: `a6117429`, `26ca73a`, `2c14efeaa`, `6acb46c11`, `1d47fd630`, `b15375dfb`, `56e96fdd8`, `f1cc83e0e`, `a66f771cb`, `0cfabaa82`, `0a521f09f`, `20f534078`, `89dffdd2e`, `6010c3cca`, `1e6d0b602`, `888cd8092`, `d3d5843b9`.
- Deferred: `a39163519` because it changes Codex/OpenAI generated config defaults to `gpt-5.5`, which is a product/config policy decision outside this small stable-fix Sprint.
- Deferred: `003b2786d` because its target file belongs to the deferred apicompat bridge test chain.
- Deferred: `08e19bb15`, `d7bed40dd`, `08061717b` because they are OpenAI WS bridge/failover-sized changes.
- Deferred: `5fd9a3509` because the current local pricing resource still matches the old `codex-auto-review` assertion; porting only the test would make it fail.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

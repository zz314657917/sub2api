### DONE: upstream-main-v0141-antigravity-system-role-s34

# Worker Result

## Task ID
upstream-main-v0141-antigravity-system-role-s34

## Status
done

## Summary
- Ported upstream `65559ac58993c5eb42eb14d9f889ec76f2f44c8e`.
- `buildContents` now separates message-level `system` role parts from normal Gemini contents.
- `TransformClaudeToGeminiWithOptions` appends those parts to `systemInstruction`, after top-level system instructions.
- Added regression coverage for message-level system role handling, merge order, assistant role mapping, and ordinary user/assistant stability.

## Changed Files
- `backend/internal/pkg/antigravity/request_transformer.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `docs/workflow/tasks/upstream-main-v0141-antigravity-system-role-s34.md`

## Commands Run
```text
gofmt -w backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go -> pass
go test ./internal/pkg/antigravity -run "TestTransformClaudeToGeminiWithOptions_MessageRoles|TestTransformClaudeToGeminiWithOptions_PreservesBillingHeaderSystemBlock" -count=1 -> pass
git diff --check -- backend/internal/pkg/antigravity/request_transformer.go backend/internal/pkg/antigravity/request_transformer_test.go -> pass
```

## Test Output
```text
ok  	github.com/Wei-Shaw/sub2api/internal/pkg/antigravity	2.133s
```

## Risks
- Verification is focused on the transformer behavior. No live Antigravity upstream request was sent.
- This does not change routing, account selection, tools, thinking, identity patch, MCP XML, or web-search fallback behavior.
- The wider main worktree still has unrelated user-owned-proxy dirty files, but they are outside S34.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

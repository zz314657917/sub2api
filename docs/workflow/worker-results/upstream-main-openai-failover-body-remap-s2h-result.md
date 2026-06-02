### DONE: upstream-main-openai-failover-body-remap-s2h

## Task ID
upstream-main-openai-failover-body-remap-s2h

## Status
done

## Summary
- Completed a minimal semantic port for upstream `c8cd91e3c test(openai): 覆盖 failover 请求体重映射`.
- `getOpenAIRequestBodyMap` now parses the supplied request body every time and no longer reads or writes `OpenAIParsedRequestBodyKey`.
- OpenAI failover retries can re-apply per-account `model_mapping` from the original client model instead of a previously rewritten cached model.
- Handler-side `OpenAIParsedRequestBodyKey` usage is preserved for validation and Claude Code helper paths.
- No schema, migration, config, frontend, handler, public API, OpenAI WS bridge, or Responses bridge files were changed.

## Changed Files
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_hotpath_test.go`
- `backend/internal/service/openai_failover_cached_body_test.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-openai-failover-body-remap-s2h.md`
- `docs/workflow/worker-results/upstream-main-openai-failover-body-remap-s2h-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-failover-body-remap-s2h-qa.md`
- `knowledge/tasks/current-task.md`

## Commands Run
```text
git status --short --branch -> on codex/upstream-main-openai-failover-body-remap-s2h
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_service_hotpath_test.go backend/internal/service/openai_failover_cached_body_test.go -> pass
git diff --check -> pass
go test ./internal/service -run "FailoverReparsesCachedBody|GetOpenAIRequestBodyMap" -count=1 -> pass
go test ./internal/service -run "OpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|GetOpenAIRequestBodyMap" -count=1 -> pass
go test ./internal/service ./internal/handler -run "OpenAIRequestBodyMap|ClaudeCodeBodyMap|FunctionCallOutput" -count=1 -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/handler
```

## Implementation Notes
- Changed the service helper signature body to ignore `*gin.Context` and always `json.Unmarshal(body, &reqBody)`.
- Updated the nearby hotpath tests to assert parse error behavior and no context-cache writeback.
- Added upstream-style failover regression coverage for:
  - both accounts with different mappings;
  - first account mapped and second account unmapped;
  - first account unmapped and second account mapped;
  - legacy context cache ignored between failover attempts.
- Added a local no-op account repo stub in the new test file so current-branch 429 failover side effects can run without nil panics.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

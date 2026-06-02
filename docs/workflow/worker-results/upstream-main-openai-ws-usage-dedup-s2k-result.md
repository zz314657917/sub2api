### DONE: upstream-main-openai-ws-usage-dedup-s2k

## Task ID
upstream-main-openai-ws-usage-dedup-s2k

## Status
done

## Summary
- Ported upstream `1e2193c3d fix: avoid websocket usage dedup conflicts`.
- `OpenAIGatewayService.RecordUsage` now uses the upstream/response request id for OpenAI WS mode when present, even if the connection context has a stable `client_request_id`.
- Non-WS OpenAI usage still prefers `ctxkey.ClientRequestID` over upstream request id.
- Local `RequestIDOverride` remains the final override after WS/non-WS request-id resolution.

## Changed Files
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-openai-ws-usage-dedup-s2k.md`
- `docs/workflow/worker-results/upstream-main-openai-ws-usage-dedup-s2k-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-ws-usage-dedup-s2k-qa.md`
- `knowledge/tasks/current-task.md`

## Commands Run
```text
gofmt -w backend/internal/service/openai_gateway_service.go backend/internal/service/openai_gateway_record_usage_test.go -> pass
git diff --check -> pass
go test ./internal/service -run "OpenAIGatewayServiceRecordUsage_(PrefersClientRequestIDOverUpstreamRequestID|WSModePrefersUpstreamRequestIDOverClientRequestID|GeneratesRequestIDWhenAllSourcesMissing)" -count=1 -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service 8.389s
```

## Implementation Notes
- Added a WS-mode request-id override only when `result.OpenAIWSMode` is true and `result.RequestID` is non-empty.
- Added a regression test proving WS mode does not deduplicate every turn under the connection-level client request id.
- Did not modify OpenAI WS forwarder, WS bridge, handlers, schema, config, frontend, gateway routing, or billing repositories.

## Candidate Review Notes
- Equivalent locally, not re-ported: `8a999f438 fix(ws): exclude terminal events from first-token detection`.
- Equivalent locally, not re-ported: `2bd3125d Preserve usage request context`.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

### PASS: upstream-main-openai-ws-usage-dedup-s2k

# upstream-main-openai-ws-usage-dedup-s2k QA Report

## Task ID
upstream-main-openai-ws-usage-dedup-s2k

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-openai-ws-usage-dedup-s2k.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run from `backend/` where applicable:
```text
git diff --check -> pass
go test ./internal/service -run "OpenAIGatewayServiceRecordUsage_(PrefersClientRequestIDOverUpstreamRequestID|WSModePrefersUpstreamRequestIDOverClientRequestID|GeneratesRequestIDWhenAllSourcesMissing)" -count=1 -> pass
```
- manual checks:
```text
Only OpenAIGatewayService.RecordUsage and its record-usage regression tests changed among business files.
OpenAI WS forwarder/bridge files were not touched.
RequestIDOverride still executes after WS/non-WS request-id selection.
Non-WS client_request_id preference remains covered by the target test.
```

## Findings
- 未发现本 Sprint 引入的阻断问题。
- OpenAI WS usage billing/logging will no longer reuse the same connection-level client request id for different upstream response turns.

## Bug Owner Recommendation
none

## Root Cause
- Usage request-id resolution preferred stable `ctxkey.ClientRequestID` globally. That is correct for HTTP requests, but OpenAI WS can carry multiple upstream turns under one connection context, so deduplication needs the upstream response id in WS mode.

## Retest Scope
- Re-run the target service tests if future patches touch `OpenAIGatewayService.RecordUsage`, `OpenAIForwardResult.OpenAIWSMode`, request-id resolution, or usage billing dedup semantics.

## Knowledge Promotion
- none

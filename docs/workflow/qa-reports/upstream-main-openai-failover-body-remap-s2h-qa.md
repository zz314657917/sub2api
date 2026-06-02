### PASS: upstream-main-openai-failover-body-remap-s2h

# upstream-main-openai-failover-body-remap-s2h QA Report

## Task ID
upstream-main-openai-failover-body-remap-s2h

## Verdict
PASS

## Contract Checked
- `docs/workflow/tasks/upstream-main-openai-failover-body-remap-s2h.md`

## Evidence
- diff reviewed: yes
- allowed paths checked: yes
- denied paths touched: no
- commands run:
```text
git diff --check -> pass
go test ./internal/service -run "FailoverReparsesCachedBody|GetOpenAIRequestBodyMap" -count=1 -> pass
go test ./internal/service -run "OpenAIGatewayService_Forward_FailoverReparsesCachedBodyForNextAccount|GetOpenAIRequestBodyMap" -count=1 -> pass
go test ./internal/service ./internal/handler -run "OpenAIRequestBodyMap|ClaudeCodeBodyMap|FunctionCallOutput" -count=1 -> pass
```
- manual checks:
```text
getOpenAIRequestBodyMap no longer reads or writes OpenAIParsedRequestBodyKey.
OpenAIParsedRequestBodyKey constant and handler-side uses remain present.
New failover regression confirms second account mapping is applied from original body, not legacy context cache.
No handler, schema, migration, frontend, config, WS bridge, Responses bridge, or routing redesign files changed.
```

## Findings
- 未发现本 Sprint 引入的阻断问题。
- Service 层 request map helper 已从 context-cache 语义切到 per-call body parse 语义。
- Handler 侧预校验/Claude Code helper 的 context-cache 行为通过关联测试保留。

## Bug Owner Recommendation
none

## Root Cause
- Previous service helper behavior could reuse `openai_parsed_request_body` after the first account had rewritten `model`, causing a later failover attempt to skip or misapply its own `model_mapping`.

## Retest Scope
- Re-run the service OpenAI failover/body-map tests if future patches touch `OpenAIGatewayService.Forward`, model mapping, failover retries, or handler request-body parsing cache.

## Knowledge Promotion
- none

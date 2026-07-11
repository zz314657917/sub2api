### DONE: upstream-codex-mcp-tool-bridge-s67b

# Worker Result

## Task ID

`upstream-codex-mcp-tool-bridge-s67b`

## Summary

- 按顺序适配 `75fb3c41c`、`27e29f056`、`794233832`、`f1082bb78`、`a2cdaa641`、`e2b68d1f9`、`90e9d03de`，未引入上游 service 拆文件或共享 pipeline 重构。
- Responses custom/freeform 工具降级为代理 function，并在非流式、流式和历史消息回程恢复为 `custom_tool_call`。
- `tool_search` 降级为客户端代理 function，回程恢复 `tool_search_call`；namespace 子工具使用稳定摊平名并恢复 `namespace + name`。
- 同名工具、摊平名碰撞显式报错；`tool_choice` 只引用实际发出的工具，强制 `tool_search` 指向代理 function。
- Responses 和本地未拆分的 Messages fallback 保留 mapped/requested model 候选的 effort 元数据；独立分支通过兼容 wrapper 调用现有单模型 extractor。

## Changed Files

- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_bridge_custom_tools_test.go`
- `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- `backend/internal/pkg/apicompat/responses_stream_event_wire.go`
- `backend/internal/pkg/apicompat/responses_stream_event_wire_test.go`
- `backend/internal/pkg/apicompat/types.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_messages_usage_test.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `docs/workflow/worker-results/upstream-codex-mcp-tool-bridge-s67b-result.md`

## Commands Run

```text
go test ./internal/pkg/apicompat -run "Test.*Custom.*Tool|Test.*ToolSearch|Test.*Namespace|Test.*ToolChoice|Test.*Responses.*Chat" -count=1 -> PASS
go test ./internal/service -run "TestForwardResponsesChatCompletionsFallback|Test.*Messages.*Fallback|Test.*Messages.*Usage" -count=1 -> PASS
go test ./internal/pkg/apicompat -count=1 -> PASS
git diff --check -> PASS
```

## Risks

- 未对真实 Codex MCP server 或真实 Chat-only 上游做端到端联调；协议行为由定向/完整 apicompat 测试和 fallback service 测试覆盖。
- S67a 合入后共享 effort extractor 会改为 variadic；集成时可把本分支兼容 wrapper 收敛为直接候选调用，当前实现保持独立可编译和等价后缀回退。

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Knowledge Candidates

- None.

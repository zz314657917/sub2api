# cf3577a3c 完整行为映射（实施完成）

41 个上游路径已逐项映射：39 项 implemented/equivalent，2 项 not-applicable。测试路径映射表示对应行为由本地定向回归覆盖，不代表原上游测试文件逐字导入；最终独立验收以 qa-reports/cf3577a3c-behavior-port-qa.md 为准。

| 上游路径 | 本地 owner 与验收测试 | 状态 |
| --- | --- | --- |
| backend/internal/handler/openai_chat_completions.go | handler/openai_chat_completions.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/openai_gateway_credential_failover_test.go | handler/openai_gateway_handler.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/openai_gateway_handler.go | handler/openai_gateway_handler.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/ops_error_logger.go | handler/ops_error_logger.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/ops_error_logger_test.go | handler/ops_error_logger.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/request_body_read_log.go | handler/request_body_read_log.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/handler/request_body_read_log_test.go | handler/request_body_read_log.go + handler/cfport_observability_test.go | implemented/equivalent |
| backend/internal/service/image_generation_intent.go | service/image_generation_intent.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_alpha_search.go | not-applicable：本地无 alpha/search 网关；禁止导入前置功能 | not-applicable |
| backend/internal/service/openai_alpha_search_test.go | not-applicable：本地无 alpha/search 网关；禁止导入前置功能 | not-applicable |
| backend/internal/service/openai_codex_function_call_id_test.go | service/openai_codex_transform.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_codex_tool_names.go | service/openai_codex_tool_names.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_codex_tool_names_test.go | service/openai_codex_tool_names.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_codex_transform.go | service/openai_codex_transform.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_codex_transform_test.go | service/openai_codex_transform.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_codex_turn_state_test.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_compact_body_signal.go | service/openai_compact_body_signal.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_compact_body_signal_test.go | service/openai_compact_body_signal.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_apikey_item_id_test.go | service/openai_responses_item_id.go + cfport_request_test.go (基线已有核心修复) | implemented/equivalent |
| backend/internal/service/openai_gateway_chat_completions.go | service/openai_gateway_chat_completions.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_forward.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_messages.go | service/openai_gateway_messages.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_passthrough.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_request_body.go | service/openai_responses_compatibility.go + openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_response_handling.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_service_codex_cli_only_test.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_service_test.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_gateway_upstream_errors.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_passthrough_normalization_test.go | service/openai_gateway_service.go + cfport_gateway_test.go | implemented/equivalent |
| backend/internal/service/openai_responses_input_compat.go | service/openai_responses_input_compat.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_responses_input_compat_test.go | service/openai_responses_input_compat.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_responses_item_id.go | service/openai_responses_item_id.go + cfport_request_test.go (基线已有核心修复) | implemented/equivalent |
| backend/internal/service/openai_responses_rejected_field_retry.go | service/openai_responses_rejected_field_retry.go + cfport_request_test.go | equivalent（基线已有 helper，新增实际透传接线与负向测试） |
| backend/internal/service/openai_responses_rejected_field_retry_test.go | service/openai_responses_rejected_field_retry.go + cfport_request_test.go | equivalent（基线已有 helper，新增实际透传接线与负向测试） |
| backend/internal/service/openai_responses_tool_schema.go | service/openai_responses_schema_compat.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_responses_tool_schema_test.go | service/openai_responses_schema_compat.go + cfport_request_test.go | implemented/equivalent |
| backend/internal/service/openai_ws_forwarder_ingress.go | service/openai_ws_forwarder.go + openai_ws_v2_passthrough_adapter.go + cfport_websocket_test.go | implemented/equivalent |
| backend/internal/service/openai_ws_forwarder_v2.go | service/openai_ws_forwarder.go + openai_ws_v2_passthrough_adapter.go + cfport_websocket_test.go | implemented/equivalent |
| backend/internal/service/openai_ws_v2_passthrough_adapter.go | service/openai_ws_forwarder.go + openai_ws_v2_passthrough_adapter.go + cfport_websocket_test.go | implemented/equivalent |
| backend/internal/util/responseheaders/responseheaders.go | util/responseheaders/responseheaders.go + cfport_headers_test.go | implemented/equivalent |
| backend/internal/util/responseheaders/responseheaders_test.go | util/responseheaders/responseheaders.go + cfport_headers_test.go | implemented/equivalent |

明确允许修改的产品/测试路径（相对仓库）：
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/request_body_read_log.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_tool_names.go`
- `backend/internal/service/image_generation_intent.go`
- `backend/internal/service/openai_responses_input_compat.go`
- `backend/internal/service/openai_responses_item_id.go`
- `backend/internal/service/openai_responses_rejected_field_retry.go`
- `backend/internal/service/openai_responses_compatibility.go`
- `backend/internal/service/openai_responses_schema_compat.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_compact_body_signal.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/util/responseheaders/responseheaders.go`
- `backend/internal/service/cfport_request_test.go`
- `backend/internal/service/cfport_gateway_test.go`
- `backend/internal/service/cfport_websocket_test.go`
- `backend/internal/handler/cfport_observability_test.go`
- `backend/internal/util/responseheaders/cfport_headers_test.go`

禁止新建 upstream split gateway 文件（forward/passthrough/request_body/response_handling/upstream_errors）、split ws ingress/v2、alpha_search。主控独占 openai_gateway_service.go；worker 分别独占 request helpers/transform、WS/bridges、handler/headers/compact。

保护案例必须覆盖：allowed_tools 对象与子级工具选择保留/别名一致；gpt-6-astra 的 reasoning.mode=pro 保留，不作旧模型 max 映射；原生 ctc/tsc 与 function fc 配对；API Key multiplier/账户控制与 frontend 的 git diff 必须为空。

精确验收命令（PowerShell，backend cwd）：
```powershell
go build ./...
$names = (go list -f '{{join .GoFiles " "}}' ./internal/service) -split ' '
$paths = @($names | ForEach-Object { 'internal/service/' + $_ })
go test @paths internal/service/cfport_request_test.go internal/service/cfport_gateway_test.go internal/service/cfport_websocket_test.go -count=1 -v
$names = (go list -f '{{join .GoFiles " "}}' ./internal/handler) -split ' '
$paths = @($names | ForEach-Object { 'internal/handler/' + $_ })
go test @paths internal/handler/cfport_observability_test.go -count=1 -v
go test ./internal/util/responseheaders -count=1
git diff --check
```

Go 测试包含真实生产源文件和 fake/httptest HTTP、WS 运行验证，不使用复制实现 stub。服务整体单测已知存在测试 fixture 编译漂移，单独报告，不允许据此跳过 mock runtime 验收。

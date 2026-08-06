### PASS: upstream-v0171-codex-originator-normalization-s187

## Findings

- 上游 `e1b76e224` 将观测到落入降载桶的 `codex-tui` identity 归一化为官方 CLI，避免上游 `server_is_overloaded` 被本地当作瞬时故障而冷却账号。
- 本地所有 OAuth HTTP、WebSocket、Live、探针出站路径已收口到 `enforceCodexIdentityHeaders`；本次只在该共享边界归一化已配对的 identity，保留版本、OS、架构和终端指纹，并仅裁剪尾部官方客户端身份组。
- `gateway.disable_codex_originator_normalization` 为启动期负向开关，默认 `false`。本地不存在上游动态 gateway-settings 刷新层，因此不虚构热更新语义；修改该项后需重启服务生效。

## Executed Checks

- `gofmt -w internal/config/config.go internal/pkg/openai/request.go internal/pkg/openai/request_load_shed_test.go internal/service/openai_codex_identity.go internal/service/openai_codex_identity_test.go internal/service/openai_gateway_service.go internal/service/openai_gateway_service_test.go internal/service/openai_ws_forwarder_success_test.go`: passed.
- `go test ./internal/pkg/openai -run 'Test(IsCodexLoadShedOriginator|NormalizeCodexClientIdentityToCLI)' -count=1`: passed.
- `go test ./internal/service -run 'Test(EnforceCodexIdentityHeaders|CodexOriginatorNormalization|OpenAIBuildUpstreamRequestOAuthOfficialClientOriginatorCompatibility|OpenAIGatewayService_Forward_WSv2_OAuthOriginatorCompatibility)' -count=1`: passed.
- `git diff --check`: passed; scope review found only contract-allowed paths. Conflict-marker search and `git ls-files -u` both returned empty.

## Unverified Risks

- 未对真实 ChatGPT/Codex 上游发起请求，结论限于本地 header 归一化、HTTP 和 WS 出站回归；上游分桶策略变更时可将 `disable_codex_originator_normalization` 设为 `true` 并重启回退。
- 未验证部署、运行中配置热更新、账号冷却恢复、数据库、真实代理或生产流量。

## Recommendation

可将 S187 提交到隔离分支 `codex/upstream-v0171-integration-s183`；不合并主工作树、不推送、不部署。

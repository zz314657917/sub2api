### PASS: upstream-v0185-anthropic-fallback-s287

## Findings

- 未发现 S287 实现的明确功能问题。`sanitizeAnthropicBodyForBetaTokens` 在最终 beta header 确定后清洗客户端 fallback 字段：`fallbacks` 仅接受 `server-side-fallback-2026-07-01`，`fallback_credit_token` 接受 server-side、2026-07-01 或 2026-06-01 credit token；默认 mimic beta 不被注入。
- Bedrock 清洗无条件移除 `fallbacks` 与 `fallback_credit_token`，符合其不支持 Anthropic server-side fallback beta 的边界。
- S287 代码/测试变更仅位于 contract allowlist；现有保护性脏改与 `outputs/**` 保持不变。

## Executed Checks

- `go test ./internal/service -run 'Test(SanitizeAnthropicBodyForBetaTokens|SanitizeBedrockFieldsForBetaTokens)' -count=10` -> PASS。
- `go test ./internal/service` -> PASS（约 66.5s）。
- `go test ./cmd/server -run '^$' -count=1` -> PASS。
- `gofmt -d backend/internal/pkg/claude/constants.go backend/internal/service/gateway_request.go backend/internal/service/bedrock_request.go backend/internal/service/gateway_fallbacks_sanitize_test.go` -> PASS（无输出）。
- `git diff --check -- backend/internal/pkg/claude/constants.go backend/internal/service/gateway_request.go backend/internal/service/bedrock_request.go backend/internal/service/gateway_fallbacks_sanitize_test.go` -> PASS。
- `git diff --name-only --diff-filter=U` -> PASS（无冲突路径）。
- S287 allowlist audit -> PASS：目标源码/测试均在四个允许路径；其余工作区变更为既有保护性 dirty paths、workflow/user state 与 `outputs/`。
- 保护性脏改 aggregate diff hash -> `0e467987fd7aec5fc451983bdb8f8216f97ba69c`，与既定基线一致。

## Unverified Risks

- 未执行真实 Anthropic/Bedrock provider、数据库、容器、部署或浏览器 smoke，符合 contract 排除范围。
- 未执行带 `unit` tag 的完整测试；该套件存在既有 fixture/API 漂移风险，未作为 S287 通过证据。

## Recommendation

S287 可通过独立 QA，进入下一门禁或集成阶段。保持当前 fallback beta 白名单和 Bedrock 无条件清洗边界；真实 provider 验证应另行安排，不应在本 Sprint 扩大范围。

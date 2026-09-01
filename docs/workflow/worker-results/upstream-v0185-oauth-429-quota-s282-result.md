### DONE: upstream-v0185-oauth-429-quota-s282

## Changed Files

- `backend/internal/service/openai_account_runtime_block_fastpath.go`
- `backend/internal/service/openai_oauth_429_runtime_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_account_model_transient.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/openai_gateway_grok.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_gateway_chat_completions.go`
- `backend/internal/service/openai_gateway_chat_completions_raw.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/openai_gateway_responses_chat_fallback.go`
- `backend/internal/service/openai_embeddings.go`
- `backend/internal/service/openai_images.go`

## Implementation

- 按上游 `f1aadd48d` 的行为目标，在本地统一 OpenAI service owner 中增加 OAuth 429 分类：5h、7d、明确 reset quota 与普通瞬时 429。
- 增加并发安全的账号级运行时调度阻断，保留更长 cooldown；`ClearRateLimit` 同步清除运行时阻断，并将既有 Grok 临时暂停镜像到该状态。
- 普通 OpenAI OAuth 429 在 2 分钟同账号重试窗口内跳过持久化 cooldown，并为 HTTP failover 注入受限的同账号重试 metadata；quota/reset 429 立即 runtime block 并继续既有 snapshot、`SetRateLimited`、temporary-unschedulable 与 body reset 逻辑。
- 保持 API key、Spark shadow、非 OpenAI 账号、pool-mode、model transient 与语义 WS header 隔离的既有边界。
- 新增默认 tag 定向测试，覆盖分类、transient 短路、5h/7d/body reset、API key/shadow 边界、最长阻断与 clear、failover metadata。

## Commands Run

- `go test ./internal/service -run 'Test(OpenAIOAuth429|OpenAIAccountRuntime|HandleGrokAccountUpstreamError|OpenAIModelTransient)' -count=10`：PASS
- `go test ./internal/service`：PASS
- `go test ./cmd/server -run '^$'`：PASS
- `gofmt -d`（全部 S282 service 文件）：PASS
- `git diff --check -- backend/internal/service`：PASS
- `git diff --name-only --diff-filter=U`：PASS（无未合并路径）
- 保护文件及 aggregate dirty diff hash：PASS，aggregate=`0e467987fd7aec5fc451983bdb8f8216f97ba69c`

## Contract Compliance

- 仅修改 S282 allowlist 中的 service/test 与本 worker-result、QA report 证据文件。
- 未修改 denied paths，未执行 cherry-pick、merge、rebase、push、部署、容器更新、数据库或真实 provider/WS 操作。
- 语义 WS header 隔离按 contract 留给后续 S284；setup-token 不扩大纳入 OAuth fast path。

## QA

- 独立 Terra QA：`docs/workflow/qa-reports/upstream-v0185-oauth-429-quota-s282-qa.md`，结论 `PASS`。
- QA 记录 `go test -tags unit ./internal/service` 为既有 test/API fixture 漂移，与本 S282 allowlist 无关；默认 tag 验证通过。

## Risks

- 未执行真实 OAuth provider、真实 WebSocket、数据库、多进程 scheduler、容器或部署 smoke；均不在本 contract 范围内。

### DONE: upstream-v0185-spark-429-model-s283

## Changed Files

- `backend/internal/service/model_rate_limit.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/openai_images_responses.go`
- `backend/internal/service/openai_ws_forwarder.go`
- `backend/internal/service/openai_ws_http_bridge.go`
- `backend/internal/service/openai_ws_v2_passthrough_adapter.go`
- `backend/internal/service/openai_spark_quota_model_test.go`

## Implementation

- 按上游 `5d9c7abed` 在本地统一错误入口增加 Spark OAuth 429 模型级短路：请求模型或账号映射后的模型命中 Codex Spark 时，使用归一化 key 写入既有 `SetModelRateLimit`，不进入账号级 runtime block 或 `SetRateLimited`。
- Spark 5h/7d、body reset 与无 reset fallback 均复用 S282 的分类/解析；Spark shadow 的 global Codex header 只保留模型级状态，不污染母账号或账号级 cooldown。
- 扩展模型限流读取对 Spark effort/date alias 做同样归一化，保证 canonical `model_rate_limits` key 能阻断所有 Spark 变体而不影响其他模型。
- WS 标准 forwarder、ingress acquire/error、prewarm、HTTP bridge、v2 passthrough 握手/error event 透传 canonical/mapped model；单体 HTTP/SSE passthrough 沿用 S282 已有 requested-model 传递，不改其 owner。
- 保持本地四参数 `AccountRepository.SetModelRateLimit` 接口，不引入上游 variadic reason、repository/Ent/migration 变更。

## Commands Run

- `go test ./internal/service -run '^TestS283' -count=10`：PASS
- `go test ./internal/service -run 'Test(S283|OpenAI.*Spark|Spark.*Quota|OpenAIWS.*RateLimit)' -count=10`：PASS
- `go test ./internal/service`：PASS
- `go test ./cmd/server -run '^$' -count=1`：PASS
- `gofmt -d`（全部 S283 service/test 文件）：PASS
- `git diff --check -- backend/internal/service`：PASS
- `git diff --name-only --diff-filter=U`：PASS（无未合并路径）
- 保护文件 aggregate dirty diff hash：`0e467987fd7aec5fc451983bdb8f8216f97ba69c`

## Contract Compliance

- 仅修改 S283 allowlist 中的 service/test 与证据文件；未修改 denied paths、`outputs/**`、repository、Ent、migration、handler、frontend、容器或部署。
- 未执行 cherry-pick、merge、rebase、push、真实 provider/数据库/WebSocket 操作。
- S282 普通 OAuth retry window、API key/非 OAuth、Grok、pool-mode 与语义 WS header 隔离边界保持不变。

## Risks

- 尚未执行真实 OAuth provider、真实 WebSocket、数据库、多进程 scheduler、容器或部署 smoke；均不在 contract 范围内。
- `go test -tags unit ./internal/service` 若失败，按 QA 记录为既有 test/API fixture 漂移，不调整 S283 测试以规避。

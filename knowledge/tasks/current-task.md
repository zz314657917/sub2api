# 当前任务快照

最后更新：2026-06-02 15:30 +08:00

## 背景

- 仓库：`F:/mcplugins/sub2api`。
- 当前主线从“模型感知智能分组路由”延伸到 OpenAI-compatible 视频模型接入与异步任务计费。
- 工作区仍有多条并行未提交改动，包括 ticket、公共模型广场、前端类型等；本轮收尾只处理后端视频异步任务、余额预扣和计费链路，不回滚其它改动。

## 当前目标

- 用户余额不足时，`POST /v1/videos/generations` 必须在上游调用前拦截，避免平台垫付视频生成成本。
- 钱包分组提交视频任务前按模型价格预估并预扣余额；上游提交失败、缺少 task_id、任务失败/取消时退款；任务成功时写 usage/billing，但不重复扣余额。
- 订阅分组不预扣钱包余额，任务成功后仍走订阅额度计费。
- `simple` 运行模式不做真实预扣或计费。

## 本次已完成

- 新增 `openai_video_tasks` 任务表迁移，记录异步视频任务、模型映射、提交/状态响应、计费状态。
- 增加 `estimated_cost`、`reserved_cost`、`refunded_cost`，区分估算价、钱包预扣额和退款额。
- `OpenAIGatewayHandler.Videos` 在钱包模式下先预估价格并原子预扣余额，余额不足直接返回 `billing_error`，不调用上游。
- 上游 failover、普通失败、响应缺少 `task_id`、任务记录失败时只对已预扣任务执行退款。
- 订阅分组只检查视频价格配置，不写 `reserved_cost`，避免钱包和订阅双扣。
- `SettleOpenAIVideoTaskIfTerminal`：
  - 失败状态退款、刷新用户余额缓存并标记 `failed_no_charge`。
  - 成功状态对已预扣任务使用 `PrepaidBalanceCost` 和 `CostOverride`，写使用记录但不再次扣钱包余额。
- 通用 usage billing 增加 `PrepaidBalanceCost`，钱包余额扣减、缓存扣减和低余额通知都只针对剩余未预付成本。
- 预扣/退款后失效用户余额缓存，降低前端余额和后续 eligibility 判断滞后。
- 补跑真实 PostgreSQL 容器迁移集成测试，确认 `166_openai_video_tasks.sql` 可随完整迁移链执行。

## 已确认事实

- 钱包模式：`reserved_cost > 0` 才会被当成预付余额；单纯 `estimated_cost` 不会触发预付抵扣。
- 订阅模式：不预扣钱包，结算时由已有订阅计费逻辑处理。
- `simple` 模式：跳过视频价格估算、预扣和真实扣费。
- 视频模型价格必须在渠道模型定价里配置为 `per_request` 或 `image`，否则非 simple 模式会拒绝视频任务提交。

## 验证记录

- `go test ./internal/service -run OpenAIGatewayServiceForwardVideos -count=1`：通过。
- `go test ./internal/service -run OpenAIGatewayServiceEstimateOpenAIVideoCost -count=1`：通过。
- `go test ./internal/service -run OpenAIGatewayServiceVideoTaskSettlement -count=1`：通过，覆盖失败任务退款后余额缓存失效。
- `go test ./internal/handler ./internal/repository -count=1`：通过。
- `go test ./internal/service ./internal/handler ./internal/repository -count=1`：通过。
- `go test ./internal/server/routes ./cmd/server -count=1`：通过。
- `go test -tags=integration ./internal/repository -run TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate -count=1`：通过，覆盖 PostgreSQL/Redis testcontainers 与完整迁移链。
- `git diff --check`：通过；仅有 `.dockerignore` 和 `knowledge/tasks/current-task.md` 的 LF/CRLF warning。

## 待验证点

- 用真实上游视频账号 smoke：
  - 余额不足提交视频任务应直接失败，且上游无任务产生。
  - 余额充足提交成功后余额立即减少预扣额。
  - 上游任务失败后余额退回。
  - 上游任务成功后 usage log 生成，余额不二次减少。
- 前端/管理台如需展示视频任务或预扣状态，后续再补 UI。

## 下一步

- 若继续开发：真实上游视频 smoke 可能消耗上游额度/余额，需用户确认使用哪个本地环境和测试账号后再跑；随后再考虑前端展示异步任务状态和管理员定价提示。

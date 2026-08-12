### DONE: standard-group-time-rate-s211

# Worker Result

## Task ID

`standard-group-time-rate-s211`

## Status

`done`

## Summary

- 扩展现有 `peak_rate_*` 分时因子，使标准分组与订阅分组都能按服务器时区的同日左闭右开窗口计费。
- 最终 Token 倍率为原有效倍率乘以分时因子；窗口外或关闭功能时因子为 `1.0`，不会禁用分组，也不改变 API Key 路由、可用状态或绑定关系。
- 三类用量输入携带 `RequestStartedAt`，请求、重试、故障切换、WebSocket 和异步落账均按首次请求开始时间判定窗口。
- 标准分组启用时倍率必须大于 `0`，订阅分组保留倍率 `0` 兼容行为；图片和视频按次计费不叠加分时因子。
- 管理端创建/编辑页面支持两类分组，保留类型切换时的合法配置，并展示服务器时区、叠加公式和类型化校验提示。

## Changed Files

- `backend/internal/service/group.go`
- `backend/internal/service/group_peak_rate_test.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/gateway_record_usage_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_record_usage_test.go`
- `backend/internal/service/openai_videos.go`
- `backend/internal/service/admin_service.go`
- `backend/internal/service/admin_service_peak_rate_test.go`
- `backend/internal/service/admin_service_group_test.go`
- `backend/internal/handler/gateway_handler.go`
- `backend/internal/handler/gateway_handler_chat_completions.go`
- `backend/internal/handler/gateway_handler_responses.go`
- `backend/internal/handler/gemini_v1beta_handler.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/handler/openai_images.go`
- `backend/internal/handler/openai_embeddings.go`
- `backend/internal/handler/openai_chat_completions.go`
- `backend/internal/handler/gateway_key_billing_test.go`
- `frontend/src/views/admin/GroupsView.vue`
- `frontend/src/views/admin/__tests__/GroupsView.peakRate.spec.ts`
- `frontend/src/i18n/locales/zh/admin/groups.ts`
- `frontend/src/i18n/locales/en/admin/groups.ts`
- `docs/workflow/spec.md`

## Commands Run

```text
go test ./internal/service -run '^(TestPeakMultiplier|TestValidatePeakRateConfig|TestNormalizePeakRateConfig|TestAdminService_.*PeakRate|TestGatewayServiceRecordUsage_.*PeakRate|TestOpenAIGatewayServiceRecordUsage_.*PeakRate)' -count=10 -> PASS
go test ./internal/service -count=1 -> PASS (63.588s)
go test ./internal/handler ./internal/handler/admin -count=1 -> PASS
go test ./cmd/server -run '^$' -count=0 -> PASS
npm.cmd run test:run -- src/views/admin/__tests__/GroupsView.peakRate.spec.ts -> PASS (3/3)
npm.cmd run lint:check -> PASS
npm.cmd run typecheck -> PASS
npm.cmd run build -> PASS
gofmt -d <S211 Go allowlist> -> PASS (no output)
git diff --check -> PASS
git ls-files -u -> PASS (no unmerged entries)
```

## Test Output

```text
基础倍率 1.5、分时因子 0.7 时 Token 最终倍率为 1.05；窗口结束后恢复为 1.5。
请求跨越结束时间仍按 RequestStartedAt 计费。
视频回归：基础费用 0.48 使用原有效倍率 1.5 计为 0.72，不错误套用 0.7 因子；图片/视频按次费用不叠加分时因子。
浏览器桌面与 390px 窄屏验收通过；截图位于 E:/codex-artifacts/sub2api/standard-group-time-rate-s211，两个视口均 overflow=false。
```

## Risks

- 初始 `deepseek-v4-pro` CLI 独立 QA 因模型不可用而阻塞；用户授权的
  `gpt-5.6-terra` agent 随后发现非有限倍率边界，修复和完整独立复测已通过，
  详见 S211 QA report。
- 未执行真实 provider、共享 PostgreSQL/Redis、容器、部署、远端 push 或生产流量验证；这些均在 contract 外。

## Knowledge Candidates

- `none`

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## QA Resolution

- `gpt-5.6-terra` independent QA PASS: focused/full Go、frontend、视频回归、
  浏览器工件和 Git gate 均通过；其首轮发现的 `NaN`/infinity 漏洞已在
  `ValidatePeakRateConfig`、`NormalizePeakRateConfig` 和运行时安全降级中收口。
- 提交前第二轮计费复核 PASS：generic Gateway 非图片 `per_request` 和图片/视频
  按次路径均使用原有效倍率，不叠加分时因子；`UsageLog.RateMultiplier` 与实际
  扣费倍率一致。所有直接构造三类用量输入的 handler 优先复用
  `ctxkey.RequestStartedAt`，仅在缺失或零值时兼容回退到当前时间。聚焦高重复、
  完整 Go、server compile、格式、diff 和索引门禁均通过。

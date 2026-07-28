### PASS: upstream-v0165-chatgpt-live-s116

## Task ID

`upstream-v0165-chatgpt-live-s116`

## Verdict

`PASS / source-only`

## Contract Checked

- `docs/workflow/tasks/upstream-v0165-chatgpt-live-s116.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes` for the Live slice; existing S114/group-buy/knowledge/output dirt remains separate
- denied paths touched: `no` attributable to S116
- commands run:

```text
go generate ./ent -> PASS
go test ./internal/service ./internal/repository ./internal/handler/... -run "Test(ExtractClientSessionID|UsageLog.*Session|.*SessionID|Live|OpenAI.*Live)" -count=1 -> PASS
go test ./internal/repository -run "TestUsageLogRepository|Test.*UsageLog.*RequestType|Test.*Session" -count=1 -> PASS
go test ./internal/server/routes -run "TestGatewayRoutes|PromptAudit" -count=1 -> PASS
go test ./... -run '^$' -> PASS (all packages compile)
go test -tags=integration ./internal/repository -run 'TestGatewayCacheSuite/TestLiveCallRoundTripPreservesAttestationCiphertext' -> NOT RUN (Docker/Redis integration environment unavailable)
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS (1091 modules transformed)
gofmt -d <Live Go files> -> PASS (no output)
git diff --check -> PASS
```

- manual checks:

```text
allow_live is persisted with default false and enforced for OpenAI groups only -> PASS
/v1/live and /backend-api/codex/realtime/calls route to SDP creation; sideband aliases resolve the stored call -> PASS
Redis call mapping, controller handoff, lease refresh/release, expiry finalization, and request_type=live usage are covered by focused tests -> PASS
Live account selection rejects API-key, PAT, and Agent Identity accounts and reuses local OAuth transport/auth headers -> PASS
Live creation passes through the local OpenAI content-moderation gate before billing/account selection -> PASS
Windows does not fabricate DeviceCheck attestation and returns an explicit unavailable error -> PASS (fail-closed behavior)
```

## Findings

- 未发现明确的源码级阻断问题。Live 是 opt-in，未打开 group 开关时不会暴露会话能力；租约丢失会终止会话并释放并发槽，避免同账号重试被永久阻断。
- `session_id` 是 usage 记录的显式关联字段；ChatGPT Live 是独立的实时语音/媒体会话网关，不是普通 `/v1/responses` 请求。
- Redis call 映射已包含加密 DeviceCheck attestation；新增 integration-only round-trip 回归覆盖该字段，避免 sideband 从 Redis 恢复后丢失认证材料。

## Bug Owner Recommendation

`original-worker`

## Root Cause

`none`

## Retest Scope

- 无待修复项；在 macOS Apple Silicon、官方 ChatGPT app 和可用 Redis 上应重跑真实 SDP、sideband、断线重连、租约过期和 usage smoke。

## Unverified Risks

- 当前 Windows 环境无法生成官方 macOS DeviceCheck attestation，因此真实 Live 请求会 fail-closed，未执行 ChatGPT 上游请求。
- 未启动 Redis 做真实 call 状态/租约生命周期验证；完整路由套件在 Redis 指向 `127.0.0.1:1` 的测试环境不可作为运行态证据。
- `go test ./internal/server/routes -count=1` 复跑仍受同一 Redis 地址和三个既有路由断言失败影响；Live 定向路由测试通过，失败文件未由本次 Live 改动触碰。
- 未执行认证态浏览器、真实 OAuth 账号、部署或容器刷新。

## Knowledge Promotion

- `none`

## Recommendation

`PASS / source-only`。代码和接口可合入本地工作树；在 macOS/Redis/真实 OAuth 环境完成运行态 smoke 前，不建议宣称 ChatGPT Live 已上线。

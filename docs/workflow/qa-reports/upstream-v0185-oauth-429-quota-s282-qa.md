### PASS: upstream-v0185-oauth-429-quota-s282

# QA Report

## Task ID

`upstream-v0185-oauth-429-quota-s282`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-oauth-429-quota-s282.md`
- `docs/workflow/contract-reviews/upstream-v0185-oauth-429-quota-s282-review.md`
- `docs/workflow/worker-results/upstream-v0185-oauth-429-quota-s282-result.md` was not present at QA start or completion

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no` by S282; all S282 source/test changes are within the contract OpenAI service allowlist
- no denied path is staged
- protected dirty-file SHA-256 manifest: `PASS`
  - `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`: `3EF121902FA6707467F9449F4B7F35EB30DBFC9EA9144B2E70EA7D89C46D6488`
  - `backend/internal/pkg/apicompat/chatcompletions_responses_bridge.go`: `ECC2B47B60BA2D0BFA39631867164CFDB95EB8ED4033502401D367348B6C37F7`
  - `backend/internal/pkg/apicompat/types.go`: `A0BAD7ABE0F6DCF0F6D5F515E1AA394A1FBC6495D6A608D62C8ADF1D64B3BDDC`
  - `backend/internal/service/admin_service.go`: `858C0CFCEC7CB0AC27ED03694B3536C3AF5E2C695E5644A281A2955745CB3841`
  - `frontend/pnpm-lock.yaml`: `8B545157E34CC0DDC1866A43B7147326B91549879EE6C3360F094DB300CE135E`
  - `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`: `4999C1582056C3BC5E1B15EECC1EEB1DCA945EB5BC00CB5D97D17A6248D8EBB7`
- aggregate protected dirty diff hash: `0e467987fd7aec5fc451983bdb8f8216f97ba69c` (`PASS`)

## Executed Checks

```text
cd backend && go test ./internal/service -run 'Test(S282|OpenAI.*429|HandleGrokAccountUpstreamError|OpenAIModelTransient)' -count=10 -> PASS
cd backend && go test ./internal/service -> completed naturally; immediate retry -> PASS (cached)
cd backend && go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -d <all S282 formatted service paths> -> PASS (no diff)
git diff --check -- backend/internal/service -> PASS
git diff --name-only --diff-filter=U -> PASS (no unmerged paths)
git status --short -> PASS (S282 paths plus preserved pre-existing dirty paths and outputs)
Get-FileHash protected dirty files -> PASS
git diff --no-ext-diff --binary -- <six protected files> | git hash-object --stdin -> 0e467987fd7aec5fc451983bdb8f8216f97ba69c
go test -tags unit ./internal/service -> FAIL (pre-existing test/API fixture drift; see Unverified Risks)
```

## Findings

未发现明确 S282 实现问题。

- HTTP/WS 429 入口均先调用 OAuth runtime 分类，再由既有 `RateLimitService` 处理持久化；普通 OpenAI OAuth 429 在短重试窗口内跳过 durable cooldown。
- 5h、7d 和 body-reset quota 信号会建立运行时账号阻断；既有 quota snapshot、`SetRateLimited`/temporary-unschedulable 和 reset fallback 分支仍保留。阻断更新不会缩短已有更长截止时间，`ClearRateLimit` 会清除阻断。
- runtime scheduler 判定合并账号级阻断与既有 account-model transient 阻断；Grok 的既有 temporary pause 同步到该账号级 blocker。
- 分类仅面向 OpenAI OAuth 且排除 shadow；API key 不取得 OAuth same-account retry metadata，Spark shadow 不会把全局 Codex quota 信号写为账号状态。
- 所有直接 429 failover 构造仅附加 transient OAuth 的同账号 retry metadata；failover 分支在返回前不向客户端写入错误响应。定向测试覆盖 5h/7d/body-reset、API key/shadow 边界、最长阻断/clear、Grok 与既有 model transient。

## Unverified Risks

- Developer worker result `docs/workflow/worker-results/upstream-v0185-oauth-429-quota-s282-result.md` 缺失；本 QA 不采信 Developer 自述，结论基于当前 diff、独立测试和保护门禁。该缺口应在本地集成前补齐流程证据。
- 默认 tag 完整 service 已 PASS；首次完整运行在终端桥 30 秒截断后自然结束，随后相同默认 tag 命令返回 cached PASS。
- `go test -tags unit ./internal/service` 为既有基线漂移：`stringPtr` redeclared、billing/context/proxy 测试调用过期 API/字段等，与 S282 allowlist 无关。
- 未执行真实 OAuth provider、真实 WebSocket、数据库、scheduler 多进程、容器或部署 smoke；均在合同范围外。

## Recommendation

`PASS`。可进入控制器的精确暂存/本地集成步骤，但应补写缺失的 S282 worker result 作为流程证据。仅纳入 S282 allowlist 与 workflow 证据；不得修改保护脏路径、`outputs/**` 或 unit-tag 基线测试，不 commit/push。

## Bug Owner Recommendation

`none`

## Root Cause

`none`（S282）；unit-tag 失败为既有基线测试/API 漂移。

## Retest Scope

- 若修改 S282 runtime blocker、OAuth 429 入口或 RateLimit hook，重跑定向 service x10、完整默认 tag service、server 编译、gofmt、diff/conflict 和保护 aggregate hash。

## Knowledge Promotion

`none`

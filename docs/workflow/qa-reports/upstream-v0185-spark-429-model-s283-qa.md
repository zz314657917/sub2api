### PASS: upstream-v0185-spark-429-model-s283

# QA Report

## Task ID

`upstream-v0185-spark-429-model-s283`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-v0185-spark-429-model-s283.md`
- `docs/workflow/contract-reviews/upstream-v0185-spark-429-model-s283-review.md`
- `docs/workflow/worker-results/upstream-v0185-spark-429-model-s283-result.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes` for S283 changes; current S283 diff contains only the seven allowlisted service/test paths (plus preserved unrelated dirty paths)
- denied paths touched: `no` by S283
- protected dirty diff aggregate hash: `0e467987fd7aec5fc451983bdb8f8216f97ba69c`
- no unmerged paths; no denied path staged

## Executed Checks

```text
cd backend && go test ./internal/service -run 'Test(S283|OpenAI.*Spark|Spark.*Quota|OpenAIWS.*RateLimit)' -count=10 -> PASS
cd backend && go test ./internal/service -> PASS (cached; first run completed naturally)
cd backend && go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -d <all S283 allowlisted Go files> -> PASS (no diff)
git diff --check -- backend/internal/service -> PASS
git diff --name-only --diff-filter=U -> PASS (no conflicts)
git status --short -> PASS for scope audit; unrelated dirty paths/outputs preserved
protected SHA-256 manifest and aggregate dirty diff hash -> PASS; aggregate `0e467987fd7aec5fc451983bdb8f8216f97ba69c`
go test -tags unit ./internal/service -> FAIL (pre-existing API/test drift: duplicate stringPtr, stale billing signatures, buildCountTokensRequest arity, proxy fields)
```

## Findings

未发现明确实现问题。新增 `TestS283SparkOAuth429BodyResetUsesModelCooldown` 已覆盖 `usage_limit_reached` body reset 的模型级写入，并断言不触发账号级 `SetRateLimited` 或 runtime block。Spark 429 在账号级 OAuth 分类前调用现有 `SetModelRateLimit`，模型 key 经过映射/规范化；WS bridge、forwarder、v2 passthrough 均透传模型；Spark shadow 不建立账号级 runtime block。

## Unverified Risks

- 未执行真实 OpenAI OAuth provider、WebSocket、数据库、scheduler 多进程、容器或部署 smoke；均在合同范围外。
- 默认 tag 完整 service 为 PASS；`-tags unit` 失败属于既有测试/API 漂移，未修改。

## Recommendation

`PASS`。可进入控制器的精确暂存/本地集成步骤；仅纳入 S283 allowlist 与 workflow 证据，不触碰保护脏路径、`outputs/**` 或其它 denied paths，不 commit/push。

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- 如后续修改 Spark 429 处理，重跑定向 service x10、完整默认 tag service、server compile、gofmt/diff/conflict、allowlist 和保护 hash/aggregate hash。

## Knowledge Promotion

`none`

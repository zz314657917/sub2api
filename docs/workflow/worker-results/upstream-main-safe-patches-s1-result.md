### DONE: upstream-main-safe-patches-s1

## Task ID
upstream-main-safe-patches-s1

## Status
done

## Summary
- 完成 Sprint 1 的上游低风险补丁分批同步，未直接 merge `upstream/main`。
- 已合入 5 个上游补丁提交，并追加 1 个本地收尾修复，覆盖 ops 归因、OpenAI 5h 用量语义、局部策略拦截和相关测试同步。
- 保持了 Sprint 1 边界：未引入 Ent schema/codegen/migrations，未触碰 deploy/README/assets/knowledge，未做大规模 gateway/bridge 重构。

## Changed Files
- `backend/internal/handler/ops_error_logger.go`
- `backend/internal/handler/ops_error_logger_test.go`
- `backend/internal/server/middleware/api_key_auth_test.go`
- `backend/internal/service/account_test_service_openai_test.go`
- `backend/internal/service/account_usage_service_test.go`
- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_service_codex_snapshot_test.go`
- `backend/internal/service/openai_gateway_service_test.go`
- `backend/internal/service/ops_metrics_collector_test.go`
- `backend/internal/service/ops_upstream_context.go`
- `backend/internal/service/ratelimit_service_openai_test.go`
- `frontend/src/views/admin/ops/components/__tests__/OpsSettingsDialog.spec.ts`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-safe-patches-s1.md`
- `docs/workflow/worker-results/upstream-main-safe-patches-s1-result.md`
- `docs/workflow/qa-reports/upstream-main-safe-patches-s1-qa.md`

## Commands Run
```text
git status --short --branch -> clean on codex/upstream-main-safe-patches-s1
git cherry-pick --continue (after resolving b65dde634) -> success, commit 0daf0e613
go test ./internal/service -run "TestWriteOpenAIFastPolicyBlockedResponseMarksBusinessLimited|TestBuildCodexUsageProgressFromExtra_UsesCanonicalUsedPercent|TestAccountUsageService_PersistOpenAICodexProbeSnapshotBlocksOpenAIOAuth7dExhausted" -count=1 -> pass
go test ./internal/service -run "OpenAI|Codex|Ops|APIKey|Gateway|Billing|AccountUsage|Metrics|Ratelimit" -count=1 -> pass
go test ./internal/service ./internal/handler ./internal/repository -count=1 -> pass
go test ./internal/server/routes ./cmd/server -count=1 -> pass
corepack.cmd pnpm install --frozen-lockfile (frontend/) -> pass
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/views/admin/ops/components/__tests__/OpsSettingsDialog.spec.ts -> pass
corepack.cmd pnpm --dir frontend run build -> pass
docker build -f tmp/Dockerfile.runtime-prebuilt -t sub2api:s1-runtime-smoke . -> pass
Docker runtime smoke with temporary Postgres/Redis/app containers -> /health returned {"status":"ok"}
git diff --check -> pass
```

## Test Output
```text
ok github.com/Wei-Shaw/sub2api/internal/service
ok github.com/Wei-Shaw/sub2api/internal/handler
ok github.com/Wei-Shaw/sub2api/internal/repository
ok github.com/Wei-Shaw/sub2api/internal/server/routes
ok github.com/Wei-Shaw/sub2api/cmd/server
frontend typecheck: vue-tsc --noEmit -> pass
frontend lint:check -> pass
frontend OpsSettingsDialog Vitest -> 1 test passed
frontend build -> built successfully; emitted existing Vite/Browserslist warnings only
Docker runtime smoke -> health=pass body={"status":"ok"}
temporary Docker containers/network -> cleaned up
```

## Skipped Upstream Candidates
- `f597c1581` group custom `/v1/models` list: touches `backend/ent/**` and `backend/migrations/**`; Sprint 2 schema/API evaluation.
- `33ac8eb27` OpenAI http2 timeout: touches `deploy/**`; outside Sprint 1.
- `5e5c2062b`, `cff2f291b` response.failed handling: conflicts in core stream/handler bridge flow; Sprint 3 bridge/rewrite evaluation.
- `27600b1d2` count_tokens generation fields: touches gateway core behavior and local branch already has partial handling; deferred.
- `2bd3125d0` preserve usage request context: conflicts across gateway/OpenAI/Gemini handlers; too broad for Sprint 1.
- `eba204632` OpenAI OAuth refresh enrichment: requires `backend/cmd/server/wire_gen.go` and wider wiring paths not authorized by Sprint 1 contract.
- `bf24b6113`, `b60d8bb4c`, `57d9e15e0`: usage/admin/sync upstream models theme is migration-sized; Sprint 2.
- Empty/equivalent locally: `6aec50501`, `ed2aac25a`, `b9509e823`, `00eb3abbe`, `32ea9cfe`, `888cd8092`, `b34cc71be`, `fc66cd704`, `8a999f438`, `89dffdd2`, `0a521f09f`, `164e2f610`, `20f534078`.

## Risks
- A large part of the upstream delta was intentionally deferred to Sprint 2/3 because it requires schema, deployment, or bridge-level changes.

## Knowledge Candidates
- None.

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

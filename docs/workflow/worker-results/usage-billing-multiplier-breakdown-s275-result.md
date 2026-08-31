### DONE: usage-billing-multiplier-breakdown-s275

# Worker Result

## Task ID
usage-billing-multiplier-breakdown-s275

## Status
done

## Summary
- 保持 APIMart 图片的既有扣费、余额、`total_cost`、`actual_cost` 和持久化 `rate_multiplier` 不变。
- 新增只读 DTO 投影 `pricing_rate_multiplier` 与 `balance_conversion_multiplier`。图片记录复用既有 `apimartImageUsageMultiplierForModels` 判定，在命中 APIMart 账号或官方模型候选时将综合倍率按 `8.4` 拆分；其他记录保留既有倍率与 `1`。
- 用户与管理端 tooltip/详情、分组徽章和 CSV/XLSX 导出均显示计价倍率，并在余额换算不是 `1x` 时单独展示；旧响应继续回退 `rate_multiplier`。

## Changed Files
- `backend/internal/service/usage_log.go`
- `backend/internal/service/usage_log_multiplier_breakdown_test.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/handler/dto/mappers.go`
- `backend/internal/handler/dto/mappers_usage_test.go`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/zh/usage.ts`
- `frontend/src/i18n/locales/en/usage.ts`
- `frontend/src/views/user/UsageView.vue`
- `frontend/src/views/user/__tests__/UsageView.spec.ts`
- `frontend/src/components/admin/usage/UsageTable.vue`
- `frontend/src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `frontend/src/views/admin/UsageView.vue`
- `docs/workflow/worker-results/usage-billing-multiplier-breakdown-s275-result.md`

## Commands Run
```text
backend: go test ./internal/service -run 'TestUsageRateMultiplierBreakdown' -count=10 -> PASS
backend: go test ./internal/handler/dto -run 'TestUsageLogFromService.*MultiplierBreakdown' -count=10 -> PASS
backend: Controller bounded fix 后重跑上述两项 x10 -> PASS
backend: go test ./internal/service ./internal/handler/dto -> PASS
backend: go test ./cmd/server -run '^$' -> PASS
backend: gofmt -w internal/service/usage_log.go internal/service/usage_log_multiplier_breakdown_test.go internal/handler/dto/types.go internal/handler/dto/mappers.go internal/handler/dto/mappers_usage_test.go -> PASS
frontend: cmd.exe /d /s /c "corepack.cmd pnpm exec vitest run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts" -> PASS (43 tests)
frontend: cmd.exe /d /s /c "corepack.cmd pnpm run typecheck" -> PASS
root: git diff --check -- <allowed paths> -> PASS
root: git diff --name-only --diff-filter=U -> PASS (no output)
```

## Test Output
```text
TestUsageRateMultiplierBreakdown x10: PASS
TestUsageLogFromService.*MultiplierBreakdown x10: PASS
Controller bounded fix 后：service 5.545s，DTO 0.889s，均 PASS
UsageTable.spec.ts + UsageView.spec.ts: 43 passed
vue-tsc --noEmit: PASS
```

## Risks
- APIMart 账号触发使用查询时关联的当前账号配置；账号后续变更无法还原为不可变历史快照，已在 service 注释中明确。展示拆分直接复用当前计费倍率判定，避免两处触发边界漂移。
- 未执行真实 provider、数据库、容器或浏览器人工 smoke；均不在本 contract 授权范围。

## Knowledge Candidates
- 使用日志展示可在不修改持久化综合倍率的情况下，通过 read-only DTO 投影拆分业务计价倍率和余额换算倍率。

## Contract Compliance
- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- N/A

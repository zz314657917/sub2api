### PASS: usage-billing-multiplier-breakdown-s275

## Findings

- 无阻断问题。`UsageRateMultiplierBreakdown` 仅为只读投影：保留 `rate_multiplier` 原值，未改动 `total_cost`、`actual_cost` 或扣费写入路径；拆分复用了既有 `apimartImageUsageMultiplierForModels` 触发 helper。
- 覆盖核对：官方图片 1x -> 8.4x、2x -> 16.8x、映射后的上游 official、APIMart OpenAI API Key 任意图片、普通图片、非图片 official、无账号关联的历史 official 均有服务层回归；每例断言 `rate_multiplier == pricing_rate_multiplier * balance_conversion_multiplier`。
- 用户与管理员 DTO 均投影附加字段，且普通记录为现有 `rate_multiplier` 和 `1`。前端使用 `??` 回退旧响应的 `rate_multiplier` / `1`。
- 用户端分组徽章、tooltip、详情和 CSV，以及管理端 tooltip 和 XLSX 导出，均单列“计价倍率”“余额换算”“综合倍率”；管理端导出按代码审查确认。
- 当前工作区另有受保护的并行修改。其四个指定 SHA-256 均与合同值一致：`api_key_auth.go`、`api_key_auth_route_breaker_test.go`、`admin_service.go`、`AdminCafeRoomsView.vue`。S275 产品差异及新增测试均在 allowlist；没有冲突文件。

## Executed Checks

- `backend`: `go test ./internal/service -run 'TestUsageRateMultiplierBreakdown' -count=10` -> PASS.
- `backend`: `go test ./internal/handler/dto -run 'TestUsageLogFromService.*MultiplierBreakdown' -count=10` -> PASS.
- `backend`: `go test ./internal/service ./internal/handler/dto` -> PASS.
- `backend`: `go test ./cmd/server -run '^$'` -> PASS.
- `frontend`: `cmd.exe /d /s /c "corepack.cmd pnpm exec vitest run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts"` -> PASS, 2 files / 43 tests.
- `frontend`: `cmd.exe /d /s /c "corepack.cmd pnpm run typecheck"` -> PASS.
- Root: scoped `git diff --check` -> PASS; `git diff --name-only --diff-filter=U` -> empty.
- Root: `gofmt -d` over the five S275 Go paths -> no output. QA did not run the contract's `gofmt -w`, because this role is prohibited from modifying source/test files.
- Code review: `ListWithFilters` hydrates associated Accounts before user/admin DTO mapping, so account-triggered image rows retain the APIMart account evidence used by the projection.

## Unverified Risks

- 未执行真实数据库、provider、网络、容器、部署或浏览器 E2E，均不在合同授权范围内。
- `GetByID` 的 repository 路径未批量 hydrate Account；单条详情 API 的 APIMart“任意图片账号”拆分取决于模型 official 触发。现有用户详情抽屉由已 hydrate 的列表行驱动，合同所列 UI 路径不受影响；若未来该单条 API 成为独立展示来源，应补充账号关联投影。
- 页面控制台的 Browserslist 数据过期提示不影响本次 Vitest/typecheck 结果。

## Recommendation

接受 S275 的只读 usage 倍率拆分实现，可交 Final Evaluator；保持扣费、余额、schema、历史账本和受保护并行修改不变。

# Task Contract

## Task ID
studio-bridge-model-scoped-apimart-billing

## Role
Codex 作为 Generator 实现，Evaluator 独立审查 contract、diff 和验收证据；本任务不调用 worker。

## Goal
修复 Studio Bridge 将普通 `gpt-image-2` 的 APIMart 上游成本错误换算并补扣的问题。固定价 `gpt-image-2` 不允许使用 `amount_unit=apimart_cost`；`gpt-image-2-official` 以及其他原本按 APIMart 成本结算的模型保持现有行为。

## Success Criteria
- Sub2API 对固定价 `gpt-image-2` 携带 `apimart_cost` 返回明确错误，不把原始上游成本当余额扣除。
- `gpt-image-2-official`、Midjourney、Grok 等已有成本型模型继续换算 `apimart_cost`。
- official 模型现有 `7 * 1.2` 换算、reserve/commit 幂等和 usage log 行为不变。
- `chatgpt2api` 普通 `gpt-image-2` 即使收到任务状态成本，也只 reserve/commit 固定预扣，不产生 refund 或 surcharge。

## Context
- Primary ledger repo: `F:/mcplugins/sub2api`
- Executor repo: `F:/java/chatgpt2api`
- Related contract: `F:/java/chatgpt2api/docs/workflow/tasks/studio-bridge-model-scoped-apimart-billing.md`
- Deployment order: executor policy first, ledger validation second.

## Allowed Paths
- `backend/internal/service/studio_bridge.go`
- `backend/internal/service/studio_bridge_test.go`
- `docs/workflow/tasks/studio-bridge-model-scoped-apimart-billing.md`
- `docs/workflow/qa-reports/studio-bridge-model-scoped-apimart-billing-qa.md`
- `docs/workflow/main-log.md`

## Denied Paths
- `docs/workflow/status.md` and existing dirty frontend/knowledge files.
- Database schema, migrations, repository SQL, payment callbacks, production config and deploy files.
- Any path in `F:/java/chatgpt2api` except that repo's mirrored contract.

## Constraints
- Keep the change scoped to the invalid fixed-price `gpt-image-2 + apimart_cost` combination.
- Do not change fixed prices, the `7 * 1.2` official multiplier, fingerprints or charge idempotency.
- Do not repair historical charges automatically; report historical and in-flight settlement risk separately.
- Preserve unrelated working-tree changes.

## Acceptance Commands
```powershell
go test ./internal/service -run 'TestStudioBridge(APIMartImageChargeUsesSub2APIMultiplierForOfficialModel|APIMartImageChargeRejectsFixedPriceGPTImage2Model|APIMartImageChargeKeepsOtherCostBasedModels|ImageChargeWithoutAPIMartAmountUnitKeepsAmount)$' -count=1
go test ./internal/service -run 'TestStudioBridge' -count=1
go test ./internal/service -run 'TestPeakMultiplier' -count=1
git diff --check -- backend/internal/service/studio_bridge.go backend/internal/service/studio_bridge_test.go
```

## Output
- QA evidence: `docs/workflow/qa-reports/studio-bridge-model-scoped-apimart-billing-qa.md`.
- Final verdict must be `PASS`, `FAIL` or `BLOCKED` with cross-repo evidence.

## Stop Rules
- Stop if the fix requires database changes, historical balance mutation, production deployment or unrelated dirty files.
- Stop if normal-model fixed settlement cannot be enforced without changing the bridge contract.

## Contract Review
- Verdict: approved.
- The allowed paths are sufficient for a fixed-price-model validation guard and focused tests.
- The cross-repo executor change is required before this fail-closed ledger guard is deployed.
- QA amendment: the full service package has unrelated global-timezone test interference; all Studio Bridge tests plus the isolated peak-rate baseline are the required scoped gate.

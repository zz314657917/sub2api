### PASS: upstream-model-pricing-alignment-s109

# QA Report

## Task ID

`upstream-model-pricing-alignment-s109`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-model-pricing-alignment-s109.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run <S109 focused/default-tag regressions> -count=1 -> PASS
go test ./internal/service -run "Test.*Pricing|Test.*Billing|Test.*ImageInput|Test.*OpenAIImages.*Usage|Test.*RecordUsage" -count=1 -> PASS
go test ./internal/service -run <account-stats/available-channel/hosted-usage/max-int regressions> -count=1 -> PASS
go test ./internal/repository -run "TestUsageLogRepository|TestPrepareUsageLogInsert|TestScanUsageLog" -count=1 -> PASS
go test -tags=integration ./internal/repository -run "TestMigrationsRunner_IsIdempotent_AndSchemaIsUpToDate" -count=1 -> PASS
go test -tags=unit ./internal/server -run "TestAPIContracts/GET_/api/v1/usage" -count=1 -> PASS
corepack.cmd pnpm exec vitest run <S109 usage tests> -> PASS (3 files / 33 tests)
corepack.cmd pnpm run typecheck -> PASS
corepack.cmd pnpm run build -> PASS (1089 modules)
corepack.cmd pnpm exec eslint <S109 frontend paths> -> PASS
gofmt -d <origin/main..S109 Go paths> -> PASS (no output)
git diff origin/main --check -> PASS
git diff --name-only --diff-filter=U -> PASS (no output)
```

- manual checks:

```text
GLM-5.2 default-tag fallback regression executes and matches GLM-5.1 USD rates -> PASS
long-context image input split preserves pre-multiplier TotalCost and applies 2x only to ActualCost -> PASS
negative/overflow/malformed OpenAI usage values are bounded; generic hosted tool usage preserves independent image tokens -> PASS
explicit billing_mode=token with image_count>0 remains token billing in admin/user views -> PASS
legacy rows without billing_mode still derive image billing from image_count -> PASS
migrations 193/194, usage insert/scan order, DTO and /usage fields align -> PASS
account stats and available-channel fallback preserve image-input pricing -> PASS
origin/main..worktree paths match the approved S109 allowlist; composite and account-import paths are absent -> PASS
```

## Findings

- 最终复审未发现剩余阻断问题。18:55 PASS 撤销后发现的账号统计图片输入价遗漏、hosted tool usage 钳制、长上下文整数溢出和可用渠道图片输入价遗漏均已修复。
- 四项修复的默认标签回归、完整 S109 backend/frontend 门禁和最新三智能体只读复审均已通过。
- QA 期间修复了四项问题：显式 token 图片记录被前端误判、GLM-5.2 unit 断言仍为旧输出价、普通 OpenAI usage 缺少有界解析/图片 token 钳制、长上下文测试把倍率错误计入 `TotalCost`。
- `go test -tags=unit ./internal/service ...` 仍在编译阶段被仓库既有漂移阻断：`stringPtr` 重复、billing helper 旧签名、Grok runtime-block helper 缺失。S109 的 GLM 与长上下文关键语义已增加默认 build-tag 测试并实际通过，未将 unit 聚合伪报为 PASS。
- 完整 `TestAPIContracts` 仍有与 S109 无关的 settings 快照漂移；S109 `/api/v1/usage` 子契约独立通过。

## Bug Owner Recommendation

`original-worker`

## Root Cause

`none`

## Retest Scope

- 无待修复项；可以进入精确 staging、commit、merge、push 和远端 parity 验证。

## Unverified Risks

- 未向真实 OpenAI OAuth 图片上游发请求。
- 未做管理员/用户登录态浏览器 smoke；`per_request` 分支仅通过代码枚举和既有按图用例覆盖。
- 未部署、未更新容器，也未运行生产数据迁移。
- composite 别名计费因本地缺少完整平台/gateway 前置链，按 contract stop rule 留到独立 Sprint。

## Knowledge Promotion

- `candidate`: monolithic usage insert 固定为 53 个数据参数与第 54 个 timezone；OAuth split 聚合必须累计 `ImageInputTokens`。

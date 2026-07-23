### PASS: upstream-v0164-small-fixes-s111

# QA Report

## Task ID

`upstream-v0164-small-fixes-s111`

## Verdict

`PASS / published`

## Contract Checked

- `docs/workflow/tasks/upstream-v0164-small-fixes-s111.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/service -run TestHandleGrokAccountUpstreamErrorPaymentRequiredPausesAccount -count=10 -> PASS
go test ./internal/pkg/openai -run TestDefaultModels -count=1 -> PASS
go test ./internal/handler/admin -run TestAccountHandlerGetAvailableModels_OpenAIAPIKeyDefaultsToConcreteGPT56Sol -count=1 -> PASS
go test ./internal/pkg/openai ./internal/handler/admin -count=1 -> PASS
corepack.cmd pnpm --dir frontend exec vitest run src/utils/__tests__/ccswitchImport.spec.ts src/components/account/__tests__/AccountStatusIndicator.spec.ts -> PASS (2 files / 17 tests)
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS (1090 modules)
corepack.cmd pnpm --dir frontend exec eslint <four S111 frontend paths> -> PASS
gofmt -d <six S111 Go paths> -> PASS (no output)
git diff --check -> PASS
git diff --name-only --diff-filter=U -> PASS (no output)
git diff --cached --check -> PASS
git fetch origin main -> PASS (remote baseline remained 7e2013fdd)
git push origin main -> PASS (7e2013fdd..15496ed12)
git rev-list --left-right --count HEAD...origin/main -> PASS (0 0)
git ls-remote --heads origin main -> PASS (15496ed12)
```

- manual checks:

```text
local diff compared with upstream a3a1575e9, ca0d3314c, 48d58d72f, and dd5956be5 -> PASS
Grok CC Switch keeps normalized homepage and emits grokbuild + one /v1 + grok-4.5 -> PASS
Grok 402 persists a 30-minute pause, blocks scheduling during cooldown, and recovers after expiry -> PASS
401, 403, 429, and 5xx Grok policies remain unchanged -> PASS
model countdown uses shared day-aware formatter and tooltip uses complete local date/time -> PASS
console tooltip surface/arrow classes remain present -> PASS
gpt-5.6-sol is first; bare gpt-5.6 alias remains exactly once -> PASS
ten business/test paths and workflow artifacts match the approved allowlist -> PASS
separate group-buy S110 worktree remains outside this task -> PASS
feature commit contains exactly the 17 approved S111 paths -> PASS
HEAD, origin/main, and remote main match at 15496ed12 -> PASS
```

## Findings

- 未发现明确的 S111 阻断问题；四项补丁逐项与上游源提交对照，并按本地拓扑完成适配。
- 默认标签 Grok 402 回归真实执行了生产 handler、repository persistence 调用、调度阻断和冷却过期恢复，避免只依赖当前不可编译的 `unit` 聚合测试。
- `go test -tags=unit ./internal/service` 仍在编译阶段被仓库既有漂移阻断，包括 `stringPtr` 重复、旧 billing helper 签名和既有 Grok runtime-block 测试引用；未把该聚合套件伪报为通过。
- 前端生产构建只有既有 Browserslist 数据过期、chunk 切分、large chunk 和 Node `DEP0190` 告警。
- 功能提交 `15496ed12` 已推送到 `origin/main`；本地、remote-tracking 和远端引用一致，分歧为 `0/0`。

## Bug Owner Recommendation

`original-worker`

## Root Cause

`none`

## Retest Scope

- 无待修复项；如后续与其他分支集成，重跑 focused Go/Vitest、typecheck、build、ESLint 和静态门禁即可。

## Unverified Risks

- 未使用真实 CC Switch 客户端打开 Grok Build deeplink。
- 未向真实 Grok 上游制造 HTTP 402。
- 未做管理员登录态浏览器视觉 smoke。
- 未部署、未更新容器，也未触碰独立 S110 工作树。
- 完整 `unit` tag service 聚合仍受上述既有编译漂移阻断。

## Knowledge Promotion

- `none`

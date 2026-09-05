### DONE: upstream-v0200-claude-fable-5-1-s294

# Worker Result

## Task ID
upstream-v0200-claude-fable-5-1-s294

## Status
`done`

## Summary

- 按批准合同完成上游 Fable 5.1 的行为级适配，没有 merge、rebase 或 cherry-pick 上游历史。
- 补齐 Claude、Antigravity、Bedrock 模型目录和映射、OpenCode 配置、Fable fallback pricing、OAuth system prompt 兼容、Anthropic `7d_oi` 家族级限流、被动用量采样及前端用量展示。
- 追加 Antigravity Claude 汇总对 `claude-fable-5-1` 的覆盖，避免 5.1 配额被漏算。

## Changed Files

- 合同允许的 backend domain、Claude/Antigravity model owners、billing、gateway test、model rate-limit、Anthropic rate-limit、account usage owners。
- `frontend/src/types/index.ts`
- `frontend/src/components/account/AccountUsageCell.vue`
- `frontend/src/components/account/__tests__/AccountUsageCell.spec.ts`
- `frontend/src/composables/useModelWhitelist.ts`
- `frontend/src/composables/__tests__/useModelWhitelist.spec.ts`
- `frontend/src/components/keys/UseKeyModal.vue`
- `frontend/src/components/keys/__tests__/UseKeyModal.spec.ts`

## Commands Run

```text
go test ./internal/pkg/claude ./internal/pkg/antigravity ./internal/service -run 'Test(Default|.*Fable|.*Anthropic.*Window|.*ModelRateLimit|.*ClaudeOAuth|.*Usage)' -count=1 -> PASS
go test ./internal/service -> PASS
go build ./... -> PASS (rechecked after the independent Prompt Audit remediation)
npm.cmd run test -- --run src/composables/__tests__/useModelWhitelist.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts -> PASS, 47/47
npm.cmd run typecheck -> PASS (rechecked after the independent Prompt Audit remediation)
npm.cmd run build -> PASS (rechecked after the independent Prompt Audit remediation)
go test -tags unit ./... -> BLOCKED by the known repository fixture/test drift
gofmt -d <allowed Fable Go files> -> PASS
git diff --check -> PASS
git diff --name-only --diff-filter=U -> PASS, empty
```

## Test Output

```text
backend targeted packages: PASS
backend internal/service: PASS
frontend focused files: 3 passed, 47 tests passed
```

## Risks

- 未验证真实 Anthropic OAuth、Antigravity、Bedrock provider、代理网络、数据库、容器或部署运行态。
- 全仓 backend build 与 frontend typecheck/build 已通过；全量 unit-tag suite 仍受已知 repository fixture/test drift 阻断。
- `go test -tags unit ./...` 仍被仓库既有测试漂移阻断，包括 `stringPtr` 重复定义、旧函数签名、Proxy 字段缺失和 repository fixture `32`/`34` 列不一致。
- 上游 `AccountStatusIndicator.vue` 的 Fable 5.1 scope 别名未纳入本 Sprint allowlist，保留为后续低风险 UI 补充。

## Knowledge Candidates

- none

## Contract Compliance

- allowed_paths_only: `yes` for the S294 delta
- denied_paths_touched: `no` by S294; pre-existing denied-path dirty changes remain untouched
- success_criteria_met: `partial` pending blocked full-suite gates and provider smoke
- stop_rules_triggered: `yes: full unit-tag suite retains known fixture drift and provider smoke is deferred`

## Blocked Reason

- 最新重测已解除 `mustJSON`、`owasp_tags`、backend build 和前端 typecheck/build 阻断；剩余阻断为已知 unit-tag repository fixture/test drift，另有真实 provider/runtime smoke 未执行。

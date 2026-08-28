### PASS: group-model-match-legacy-compat-s273

# QA Report

## Task ID
group-model-match-legacy-compat-s273

## Verdict
`PASS`

## Contract Checked
- `docs/workflow/tasks/group-model-match-legacy-compat-s273.md`

## Evidence
- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:
```text
go test ./internal/service -run 'TestS273|TestS91' -count=10 -> PASS
go test ./internal/server/middleware -run 'TestS272|TestS273' -count=10 -> PASS
go test ./internal/service -count=1 -> PASS (65.352s)
go test ./internal/server/middleware -count=1 -> PASS
go test ./internal/handler -count=1 -> PASS (32.900s)
go test ./internal/server/routes -run 'TestGatewayRoutes' -count=1 -> PASS
go test ./cmd/server -run '^$' -count=1 -> PASS
gofmt -w <changed Go files> -> PASS
git diff --check -> PASS
git ls-files -u -> PASS (no unmerged entries)
```
- manual checks:
```text
Legacy single-group + empty rules -> resolves the original key.
Legacy single-group + configured rule -> accepts matching model and rejects mismatch, including image intent.
Legacy Grok key on /v1/responses -> configured Grok rule is evaluated without endpoint platform rejection.
Legacy disabled-image key + empty rules -> reaches handler-owned permission_error boundary.
Multi-group or pinned key + empty rules -> remains nil/fail-closed.
Incomplete default group + valid multi-group route -> middleware still resolves the valid route.
Group.MatchesModel implementation -> unchanged.
Changed product paths -> approved service routing and middleware compatibility guard plus S273 tests only.
outputs/ and protected worktrees -> preserved; outputs/ remains untracked.
```

## Findings
- 未发现明确问题。兼容例外只存在于无多组路由且未 pinned 的单组默认回退；配置
  了模型规则后仍调用 `Group.MatchesModel`，不会把空规则推广为通配符。
- middleware 仅对不完整的单组非 pinned 快照保留 handler fixture 兼容；存在多组
  路由或 pinned 账号时仍进入 resolver，不会绕过路由解析。

## Bug Owner Recommendation
`original-worker`

## Root Cause
`none`

## Retest Scope
- None; final focused and package-level checks are green.

## Knowledge Promotion
- `none`

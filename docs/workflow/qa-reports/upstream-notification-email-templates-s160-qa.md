### PASS: upstream-notification-email-templates-s160

# QA Report

## Task ID

`upstream-notification-email-templates-s160`

## Verdict

`PASS`

## Contract Checked

- `docs/workflow/tasks/upstream-notification-email-templates-s160.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go generate ./cmd/server -> PASS; generated Wire output is stable.
go test ./internal/service -run 'TestNotificationEmail|Test.*(Balance|Subscription|Payment|Email).*' -count=1 -> PASS.
go test ./internal/handler ./internal/handler/admin ./internal/server/routes -run 'Test.*(Notification|Unsubscribe|EmailTemplate|Auth|Payment).*' -count=1 -> handler/admin PASS; routes package is blocked by the existing TestAuthRoutesRateLimitFailCloseWhenRedisUnavailable nil-audit-middleware panic.
go test ./... -run '^$' -count=1 -> PASS; all packages compile, including routes.
npm.cmd run typecheck -> PASS.
npx.cmd vitest run src/views/admin/settings/__tests__/EmailTemplateEditor.spec.ts src/views/admin/__tests__/SettingsView.spec.ts -> PASS (2 files, 25 tests).
npm.cmd run build -> PASS; vue-tsc and Vite production build complete.
npx.cmd eslint <changed frontend files> -> PASS.
gofmt -d <changed Go files> -> PASS.
git diff --check -> PASS.
git ls-files -u -> PASS; no unmerged index entries.
allowlist, denied-path, and conflict-marker audits -> PASS.
```

- manual checks:

```text
Fallback classification -> template/config errors permit exactly one legacy-email fallback; delivery errors do not trigger a second send.
Unsubscribe boundary -> only optional balance/subscription reminders can be opted out; verification and password-reset events remain transactional.
Rendering boundary -> placeholder allowlists, HTML escaping, subject header sanitization, URL validation, and sandboxed admin preview are present.
Locale/dedup -> Accept-Language memory resolves zh/en templates; v2 hashed preference/delivery keys retain legacy-key compatibility.
Isolation -> no migration, configured database, SMTP, OAuth, payment provider, container, deployment, main-worktree, merge, or push action was performed.
```

## Findings

- 未发现明确的 S160 实现问题。
- `TestAuthRoutesRateLimitFailCloseWhenRedisUnavailable` 在合同的 routes 筛选命令中因测试传入 `nil` 审计中间件而在既有 `BackendModeAuthGuard(...).Next()` 调度时 panic；测试、guard 和该 auth-group 注册均未被 S160 修改。新增退订路由位于独立的 public settings group，完整路由包编译通过，因此该基线测试不阻断本 Sprint。
- 前端构建保留仓库既有的动态导入/大 chunk 警告；聚焦测试通过。

## Bug Owner Recommendation

`integration-owner`（仅针对既有 routes 测试的 nil middleware fixture，不属于 S160）

## Root Cause

`none`

## Retest Scope

- 无 S160 修复待重测。若单独修复 routes 测试 fixture，重跑 `TestAuthRoutesRateLimitFailCloseWhenRedisUnavailable` 及 auth routes 套件。

## Knowledge Promotion

- `none`

## Unverified Risks

- 未执行真实 SMTP、配置数据库、OAuth、支付提供商、管理员浏览器会话、部署或生产环境验证。
- 未验证并发同一事件下的持久化去重竞争；本 Sprint 仅覆盖现有 SettingRepository 的顺序去重与旧键兼容。

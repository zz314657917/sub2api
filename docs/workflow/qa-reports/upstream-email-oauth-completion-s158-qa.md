### PASS: upstream-email-oauth-completion-s158

# Upstream Email OAuth Completion S158 QA

## Task ID

`upstream-email-oauth-completion-s158`

## Verdict

`PASS` for the isolated source-level port. This is not a provider, database, deployment, or production validation.

## Contract Checked

- `docs/workflow/tasks/upstream-email-oauth-completion-s158.md`

## Evidence

- diff reviewed: `yes`
- allowed paths checked: `yes`
- denied paths touched: `no`
- commands run:

```text
go test ./internal/handler -run '^TestLinuxDoOAuthCallbackEmailVerificationCompletesWithBoundEmail$' -count=1 -> PASS
go test -tags=unit ./internal/server/middleware -run '^TestBackendModeAuthGuard$' -count=1 -> PASS
go test ./internal/handler -run '^TestLinuxDoOAuthCallback' -count=1 -> PASS
go test ./internal/handler ./internal/server/middleware -run '^$' -count=1 -> PASS (compile-only; no non-unit tests selected)
npm.cmd run test:run -- src/views/auth/__tests__/OAuthCallbackView.spec.ts src/api/__tests__/client.spec.ts -> PASS, 2 files / 17 tests
npm.cmd run typecheck -> PASS
gofmt -d <four changed Go files> -> PASS (no output)
git diff --check -> PASS
git ls-files -u -> PASS (no output)
scoped conflict-marker scan -> PASS (none found)
```

- manual checks:

```text
LinuxDO pending response -> verification-required signup returns create_account_required, clears synthetic/resolved email, and creates the verified email account plus LinuxDO identity through the existing pending create-account endpoint
backend mode -> GitHub and Google complete-registration are allowed; their OAuth start endpoints remain covered as forbidden by the table test
API URL -> a relative api/v1 environment value becomes /api/v1 and buildApiUrl emits one API prefix; the OAuth callback view now uses that shared helper
upstream topology -> local checkout lacks upstream's unrelated direct-login/DingTalk context, so only the three required behaviors were manually adapted
```

## Findings

- 未发现本次允许范围内的明确问题。
- Initial frontend test launch found no `node_modules` in the isolated worktree. A temporary junction to the primary worktree dependency directory allowed the tests and typecheck to run. Cleanup of that junction was requested after validation but blocked by the local command policy; it remains ignored and is not staged.

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

- Not applicable for this PASS verdict.

## Knowledge Promotion

`none`

## Unverified Risks

- No real OAuth provider callback, email delivery, authenticated browser session, configured database, deployment, or production smoke was run.
- The temporary ignored `frontend/node_modules` junction remains in the isolated worktree because the local command policy blocked its removal. It points at `F:/mcplugins/sub2api/frontend/node_modules`; the target was not altered.

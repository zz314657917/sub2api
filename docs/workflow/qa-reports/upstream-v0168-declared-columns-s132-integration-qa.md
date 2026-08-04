### PASS: upstream-v0168-declared-columns-s132-integration

## Scope

- Adapted only S132's declared-column user/API-key persistence behavior onto
  `main@9099db0c4` in an isolated worktree.
- Excluded Passkey, Kimi K3, Model Plaza, Codex manifest, migrations,
  frontend, Docker, deployment, and production state.

## Executed Checks

- `gofmt -w` on every changed Go file.
- `go test ./internal/repository ./internal/service ./internal/handler ./internal/server/... -run "Test(User|APIKey|UpdateQuotaUsed|UpdateProfile|ChangePassword|UpdateStatus)" -count=1`
- `go test ./... -run "^$"`
- `go build ./...`
- `git diff --cached --check`, `git diff --check`, and `git ls-files -u`
- Allowlist audit of every changed path and a search for legacy two-argument
  `userRepo.Update` / `apiKeyRepo.Update` calls.

## Findings

- User and API-key repositories now write only declared fields, preventing a
  stale profile/editor snapshot from replaying unrelated balances, quotas,
  rate-limit counters, groups, routing, or key lifecycle fields.
- Administrative balance mutation uses the new repository balance primitives;
  promo-code administration no longer writes back the atomic redemption
  counter.
- A current-main Cafe test stub used the old API-key repository method
  signature. Its mechanical signature update was added to the approved scope;
  focused tests, full package compilation, and production build then passed.

## Unverified Risks

- The new PostgreSQL integration-tag lost-update tests were not run: this
  contract denies database and Docker/container use. They remain source-level
  coverage only.
- No authenticated browser flow, deployment, container refresh, provider call,
  or production validation was performed.

## Recommendation

The integration branch is ready for a local mainline fast-forward after the
primary worktree's concurrent `main-log.md` change has been committed or
otherwise cleared. Do not push or modify remote branches in this task.

### PASS: upstream-fingerprint-user-agent-validation-s225

# Independent QA Report

## Task ID

`upstream-fingerprint-user-agent-validation-s225`

## Verdict

`PASS`

## Contract Checked

- Authoritative contract: `F:/mcplugins/sub2api/docs/workflow/tasks/upstream-fingerprint-user-agent-validation-s225.md`.
- Frozen business base: `06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec`.
- Candidate business commit: `569080eb0bf337d1311c9e3777b2f481200ebfac`.
- Developer report commit: `74c2b2b9e592abac7e3e42c3964ee895325b7950`.

## Evidence

- Diff reviewed: yes. The business candidate changes exactly `backend/internal/service/identity_service.go` and `backend/internal/service/identity_service_user_agent_validation_test.go`.
- Allowed paths checked: yes. The Developer report is the contract-required workflow artifact and is within the HEAD allowlist.
- Denied paths touched: no. `backend/internal/pkg/claude/constants.go`, cache interfaces, TTL implementation, and all other denied paths are unchanged.
- Formatting, whitespace, conflict, and index checks: pass (`gofmt -d` empty; `git diff --check` clean; no unresolved paths; `git ls-files -u` empty).
- Upstream provenance: pass. `fe2c265c91f58c68426495acb875ff9bd1b0440c` is an ancestor of `upstream/main` at `e330c243a8f142f8963d784916da0093ab7084ee`.

## Commands Run

```text
go test -list ^<each of eleven S225 names>$ ./internal/service -> PASS, 11/11 discoverable
go test ./internal/service -run ^(<eleven S225 names>)$ -count=10 -> PASS (0.077s)
go test ./internal/service -count=1 -> PASS (60.243s)
go test ./internal/server -run '^$' -count=1 -> PASS (0.631s, no tests to run)
gofmt -d backend/internal/service/identity_service.go backend/internal/service/identity_service_user_agent_validation_test.go -> PASS, empty output
git diff --check 06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec...569080eb0 -> PASS
git diff --name-only 06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec...569080eb0 -> PASS, exactly the two business paths
git diff --name-only --diff-filter=U; git ls-files -u -> PASS, empty output
git merge-base --is-ancestor fe2c265c91f58c68426495acb875ff9bd1b0440c upstream/main -> PASS
```

## Manual Checks

- Creation and cached-version upgrade use `isAcceptableFingerprintUserAgent`; malformed, sentinel, local/dev/build-suffixed, overlong, two-segment, and leading-junk forms are rejected before persistence.
- `claude.CLICurrentVersion` is read only to derive the Claude CLI major-version ceiling. Valid non-Claude products do not receive that ceiling.
- Exact local defaults remain unchanged: `claude-cli/2.1.92 (external, cli)`, `js`, `0.70.0`, `Linux`, `arm64`, `node`, and `v24.13.0`.
- Both poisoned-cache recovery paths write a valid request User-Agent or the local default while preserving `ClientID`; the merge helper never changes `ClientID`.
- Healthy cache and valid newer-version behavior remain on the existing merge path. The unchanged lazy refresh branch writes only when no repair/upgrade occurred and `UpdatedAt` is older than 24 hours; seven-day cache storage ownership remains outside this change.

## Findings

- No implementation defect found.

## Unverified Risks

- No real Redis or upstream provider integration was run: the contract denies cache implementation changes, provider calls, containers, and deployment. Unit-cache coverage plus complete service regression exercised the permitted behavior.

## Bug Owner Recommendation

`none`

## Root Cause

`none`

## Retest Scope

`none`

## Knowledge Promotion

`none`

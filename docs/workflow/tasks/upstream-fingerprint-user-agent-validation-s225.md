# Task Contract: upstream-fingerprint-user-agent-validation-s225

## Task ID

`upstream-fingerprint-user-agent-validation-s225`

## Status

`contract-approved`

## Role

Planner, Terra Developer Worker, independent Terra QA Worker, and Final Evaluator.
The Developer may start only after the Final Evaluator approves this contract.

## Goal

Behaviorally port upstream `fe2c265c91f58c68426495acb875ff9bd1b0440c`:
validate User-Agent values before they become account-level persistent
fingerprints, and heal already-poisoned cached User-Agent values without changing
the account's existing `ClientID`.

## Success Criteria

- First creation and cached-version upgrade share the same acceptance rule.
- Reject empty, malformed, overlong, two-segment, leading-junk, and local/dev/
  build-suffixed User-Agent values. Reject implausible Claude CLI major versions.
- Accept syntactically valid non-Claude products without applying the
  Claude-specific major-version ceiling.
- A poisoned cache heals to a valid current request User-Agent when available,
  otherwise to the existing local default; both paths preserve `ClientID`.
- Healthy cached fingerprints and valid upgrades retain existing behavior,
  including the 24-hour lazy TTL refresh path.
- The exact local defaults remain unchanged: `claude-cli/2.1.92 (external, cli)`,
  Stainless language `js`, package `0.70.0`, OS `Linux`, arch `arm64`, runtime
  `node`, and runtime version `v24.13.0`.

## Context

- Repo: `F:/mcplugins/sub2api`
- Frozen implementation base: `06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec`
- Worktree: `E:/codex-worktrees/sub2api/upstream-fingerprint-user-agent-validation-s225`
- Branch: `pge/upstream-fingerprint-user-agent-validation-s225`
- Upstream head at review: `e330c243a8f142f8963d784916da0093ab7084ee`
- Direct `git apply --check` of upstream fails at `identity_service.go` because
  local fingerprint defaults and imports diverged. Adapt behavior manually.
- Read first: this contract, `docs/workflow/spec.md`, and `docs/workflow/status.md`
  from the main checkout.

## Allowed Paths

Business implementation:

- `backend/internal/service/identity_service.go`
- `backend/internal/service/identity_service_user_agent_validation_test.go`

Developer evidence:

- `docs/workflow/worker-results/upstream-fingerprint-user-agent-validation-s225-result.md`

Independent QA evidence, owned only by the QA Worker after controller approval:

- `docs/workflow/qa-reports/upstream-fingerprint-user-agent-validation-s225-qa.md`

## Denied Paths

- Every path outside the allowlist, especially cache interfaces/implementations,
  Redis keys or TTLs, gateway/request code, frontend, `knowledge/**`, `outputs/**`,
  dependencies, lockfiles, migrations, Docker/deployment files, and VERSION.
- `backend/internal/pkg/claude/constants.go`: import and read
  `claude.CLICurrentVersion`; do not edit it.
- User-owned dirty files: `frontend/src/components/account/EditAccountModal.vue`,
  its test, `knowledge/00-start-here.md`, `knowledge/05-current-focus.md`, and
  `outputs/`.
- Provider calls, database execution, containers, deployment, remote push,
  release tagging, wholesale upstream merge, or direct cherry-pick.

## Constraints

- Preserve `defaultFingerprint` values exactly. Do not replace the local
  User-Agent or Stainless defaults with upstream values.
- Use `claude.CLICurrentVersion` only to derive the Claude CLI current major
  version for a bounded upper limit. Do not couple minor or patch versions.
- Keep validation in the common create/upgrade flow. Sanitizing only
  `isNewerVersion` or only first creation is incomplete.
- Healing mutates only allowed fingerprint header fields and timestamps; it must
  preserve `ClientID`. Do not add cache deletion/reset operations.
- Keep the current cache API, seven-day storage TTL, 24-hour lazy refresh,
  logging framework, public API, and caller behavior unchanged.
- Adapt to local code; do not replace the file wholesale with upstream.

## Acceptance Commands

```powershell
Push-Location backend
$tests = @(
  'TestIsAcceptableFingerprintUserAgent',
  'TestGetOrCreateFingerprintRejectsMalformedUserAgentOnCreate',
  'TestGetOrCreateFingerprintRejectsSentinelVersionOnUpgrade',
  'TestGetOrCreateFingerprintStillUpgradesOnValidNewerVersion',
  'TestGetOrCreateFingerprintAcceptsValidUserAgentOnCreate',
  'TestDefaultFingerprintUserAgentIsAcceptable',
  'TestGetOrCreateFingerprintHealsPoisonedCacheUsingValidClientUA',
  'TestGetOrCreateFingerprintHealsPoisonedCacheWithoutValidClientUA',
  'TestGetOrCreateFingerprintDoesNotRewriteHealthyCache',
  'TestGetOrCreateFingerprintMissingUserAgentKeepsDefault',
  'TestDefaultFingerprintRetainsLocalDefaults'
)
foreach ($test in $tests) {
  $listed = go test -list "^$test$" ./internal/service
  if ($LASTEXITCODE -ne 0 -or (($listed -join "`n") -notmatch "(?m)^$test$")) { throw "S225 undiscoverable: $test" }
}
$pattern = '^(' + ($tests -join '|') + ')$'
go test ./internal/service -run $pattern -count=10
if ($LASTEXITCODE -ne 0) { throw 'S225 focused regressions failed' }
go test ./internal/service -count=1
if ($LASTEXITCODE -ne 0) { throw 'S225 service regression failed' }
go test ./internal/server -run '^$' -count=1
if ($LASTEXITCODE -ne 0) { throw 'S225 server compile failed' }
Pop-Location

$base = '06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec'
$allowed = @(
  'backend/internal/service/identity_service.go',
  'backend/internal/service/identity_service_user_agent_validation_test.go',
  'docs/workflow/worker-results/upstream-fingerprint-user-agent-validation-s225-result.md'
)
$formatPaths = @(
  'backend/internal/service/identity_service.go',
  'backend/internal/service/identity_service_user_agent_validation_test.go'
)
$formatDiff = gofmt -d $formatPaths
if ($LASTEXITCODE -ne 0 -or $formatDiff) { throw 'S225 formatting check failed' }
git diff --check $base...HEAD
if ($LASTEXITCODE -ne 0) { throw 'S225 diff check failed' }
$unexpected = @(git diff --name-only $base...HEAD | Where-Object { $_ -notin $allowed })
if ($unexpected.Count -gt 0) { throw "S225 changed paths outside allowlist: $unexpected" }
if (git diff --name-only --diff-filter=U) { throw 'S225 has unresolved conflicts' }
if (git ls-files -u) { throw 'S225 index has unresolved entries' }
git merge-base --is-ancestor fe2c265c91f58c68426495acb875ff9bd1b0440c upstream/main
if ($LASTEXITCODE -ne 0) { throw 'S225 upstream provenance failed' }
```

## Output

- Developer report:
  `docs/workflow/worker-results/upstream-fingerprint-user-agent-validation-s225-result.md`.
- Independent QA report:
  `docs/workflow/qa-reports/upstream-fingerprint-user-agent-validation-s225-qa.md`.
- Developer report first line must be
  `### DONE: upstream-fingerprint-user-agent-validation-s225`,
  `### BLOCKED: upstream-fingerprint-user-agent-validation-s225`, or
  `### FAILED: upstream-fingerprint-user-agent-validation-s225`.
- Report changed files, commands, test outputs, risks, contract compliance, and
  `knowledge_candidates`. Do not paste unrelated long logs.
- No push or deployment. Integrate only after controller review and independent
  QA PASS.

## Stop Rules

- Stop if a denied path is required, a local default changes, either create or
  upgrade remains unvalidated, healing changes `ClientID`, or a valid normal
  upgrade/healthy-cache path regresses.
- Stop if the worktree is not based exactly on `06e0e6ea5`, the focused tests
  are undiscoverable, the service suite cannot run, or Terra is unavailable.

## Budget

- developer_worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- worktree_root: `E:/codex-worktrees/sub2api`

## Contract Review

`PASS / contract-approved`: local ownership for creation, upgrade, healing, and
default fallback is confined to `identity_service.go`; the existing Claude
constant supplies a stable major-version reference without changing local
defaults; a default-tag focused test can cover all security and compatibility
branches. No cache, Redis, gateway, frontend, dependency, or schema path is
required.

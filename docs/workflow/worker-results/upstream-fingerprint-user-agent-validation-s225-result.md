### DONE: upstream-fingerprint-user-agent-validation-s225

# Worker Result

## Task ID

`upstream-fingerprint-user-agent-validation-s225`

## Status

`done`

## Summary

- Ported upstream User-Agent validation into the local identity service without
  changing local fingerprint defaults or the cache interface.
- Creation and cached-version upgrade share syntax and Claude CLI major
  validation. Poisoned cached values heal from a valid request UA or the local
  default while retaining `ClientID`.
- Healthy cache behavior, valid upgrades, lazy refresh, non-Claude syntax
  acceptance, malformed/local-build rejection, and exact local defaults are
  covered by default-tag tests.

## Changed Files

- `backend/internal/service/identity_service.go`
- `backend/internal/service/identity_service_user_agent_validation_test.go`
- `docs/workflow/worker-results/upstream-fingerprint-user-agent-validation-s225-result.md`

## Commands Run

```text
go test ./internal/service -run <11 S225 tests> -count=10 -> PASS (0.082s)
go test ./internal/service -count=1 -> PASS (60.547s, exit 0)
go test ./internal/server -run '^$' -count=1 -> PASS (0.099s)
gofmt -w <two business paths> -> PASS
gofmt -d <two business paths> -> PASS
git diff --check 06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec...HEAD -> PASS
git diff --name-only 06e0e6ea5aff41cedc3d79819e4ab3fb692d61ec...HEAD -> exactly two business paths before this report
git diff --name-only --diff-filter=U -> PASS (empty)
git ls-files -u -> PASS (empty)
git merge-base --is-ancestor fe2c265c91f58c68426495acb875ff9bd1b0440c upstream/main -> PASS
```

The contract's per-test `go test -list` discovery was also run for all eleven
names; every name was discoverable. The final complete service run includes
the corrected overlong valid-form test case.

## Test Output

```text
ok github.com/Wei-Shaw/sub2api/internal/service 0.082s
ok github.com/Wei-Shaw/sub2api/internal/service 60.547s
ok github.com/Wei-Shaw/sub2api/internal/server 0.099s [no tests to run]
```

## Risks

- The Claude CLI upper bound uses only the major component of
  `claude.CLICurrentVersion` plus the approved two-major skew; minor and patch
  versions do not constrain future valid clients.
- Existing cache values with surrounding whitespace are evaluated after trim,
  matching upstream acceptance behavior; valid request headers retain existing
  merge semantics.

## Knowledge Candidates

- None.

## Contract Compliance

- allowed_paths_only: `yes`
- denied_paths_touched: `no`
- success_criteria_met: `yes`
- stop_rules_triggered: `no`

## Blocked Reason

- None.

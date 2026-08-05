---
task_id: upstream-v0169-release-security-integration
status: contract-approved
role: Generator
qa_mode: runtime
source_commits:
  - a53b0bf6f
  - 98c0b779b
base_commit: 580ecea3c
---

# Task Contract

## Goal

Port the coupled S169 release hardening: Compose containers run with
`no-new-privileges`, release images contain required runtime resources, and CI
checks those two static invariants.

## Success Criteria

- Every supported Compose file has the intended `no-new-privileges:true` security option.
- GoReleaser and both release Dockerfiles copy the runtime resource directory needed by packaged binaries.
- The two repository static checks pass without building images, changing containers, or deployment.

## Allowed Paths

- `.github/workflows/backend-ci.yml`
- `.goreleaser.simple.yaml`
- `.goreleaser.yaml`
- `Dockerfile.goreleaser`
- `deploy/Dockerfile`
- `deploy/docker-compose.yml`
- `deploy/docker-compose.dev.yml`
- `deploy/docker-compose.local.yml`
- `deploy/docker-compose.standalone.yml`
- `deploy/tests/docker-compose-security-test.sh`
- `deploy/tests/docker-runtime-resources-test.sh`
- `docs/workflow/tasks/upstream-v0169-release-security-integration.md`
- `docs/workflow/qa-reports/upstream-v0169-release-security-integration-qa.md`

## Denied Paths

- Product Go/frontend code, migrations, configuration values, secrets, actual Docker builds,
  Compose up/down, container removal, database, remote, push, and primary-worktree changes.

## Constraints

- Apply `a53b0bf6f` and `98c0b779b` as one indivisible release-safety slice.
- Do not add privileges, change image tags, alter services, or execute runtime deployment operations.
- The shell checks must remain static repository checks.

## Acceptance Commands

```powershell
D:/Git/bin/bash.exe -lc 'cd /e/codex-worktrees/sub2api-v0169-release-security-integration-20260805 && bash deploy/tests/docker-compose-security-test.sh'
D:/Git/bin/bash.exe -lc 'cd /e/codex-worktrees/sub2api-v0169-release-security-integration-20260805 && bash deploy/tests/docker-runtime-resources-test.sh'
git diff --check
git ls-files -u
```

## Contract Review

`PASS / contract-approved`: the cumulative `a53b0bf6f^..98c0b779b` release-safety
patch applies cleanly to `main@580ecea3c`. It has no application runtime dependency
and its two checks inspect tracked configuration only.

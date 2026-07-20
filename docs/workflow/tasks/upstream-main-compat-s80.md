# Task Contract: upstream-main-compat-s80

## Task ID

`upstream-main-compat-s80`

## Status

`approved`

## Role

Direct Codex implementation of one configuration-only upstream compatibility
fix. The upstream behavior is extended to the two identical local deployment
topologies so the repository's recommended deployment entry remains correct.

## Goal

Port upstream `be74deae73f19e35919b49eae63dc0a5b503a415` from live snapshot
`d4b9797ff72024960a035cf22fdd8f213e149169` onto local baseline
`366e590b3e61ef8c5596bb43445aa8952953497b`.

Make each built-in Redis service invoke `redis-server`, `--save 60 1`,
`--appendonly yes`, `--appendfsync everysec`, and the optional `--requirepass`
fragment as one continued shell command. Upstream fixed only the main Compose;
S80 applies the same behavior to local/dev because the documented one-click
deployment consumes `docker-compose.local.yml` and both files contain the same
defect.

## Success Criteria

- In all three built-in Redis Compose files, `redis-server` and each fixed
  option line end with a shell continuation (`\`) so the inner `sh -c` sees
  one command rather than five newline-separated commands.
- With an empty controlled `REDIS_PASSWORD`, rendered `redis.command` contains
  all persistence flags and no `--requirepass` argument.
- With a non-empty controlled hex password, rendered `redis.command` contains
  the same persistence flags plus `--requirepass` in the same continued shell
  command.
- Redis image, container name, restart policy, ulimits, volume, environment,
  network, healthcheck, dependency, and port-exposure behavior remain unchanged.
- `docker-compose.standalone.yml` remains unchanged because it uses external
  Redis and has no Redis service.
- Docker Compose static rendering, exact path audit, conflict scan, and diff
  checks pass without starting, rebuilding, restarting, or removing a container.

## Context

- Repo: `F:/mcplugins/sub2api`
- Worktree: `F:/mcplugins/sub2api/.tmp/codex-worktrees/upstream-main-compat-s80`
- Branch: `codex/upstream-main-compat-s80`
- Baseline: `366e590b3e61ef8c5596bb43445aa8952953497b`
- Upstream snapshot: `d4b9797ff72024960a035cf22fdd8f213e149169`
- Latest tag: `v0.1.161` (`19149ca196eeae4a4482e5299dc6fa4ba0b06c8c`)
- `d4b9797ff..upstream/main`: zero commits after the 2026-07-20 refresh.
- Protected primary-checkout files and SHA-256 baselines:
  - `knowledge/00-start-here.md`: `2BEB6CA5625A89E872BC8CA2A9A707EE172F3A492CDE691F629E3F6C978C93DB`
  - `knowledge/05-current-focus.md`: `C6C0EAF7851F016D06645914A12A4BF50950011EA0925E8CB9CB5747DEBF57FF`
  - `knowledge/tasks/current-task.md`: `DC719B584F0866D32CE539955EFDB70EFB68D0D6AAF7744A81EDD13F04603295`

## Allowed Paths

- `deploy/docker-compose.yml`
- `deploy/docker-compose.local.yml`
- `deploy/docker-compose.dev.yml`
- `docs/workflow/spec.md`
- `docs/workflow/tasks/upstream-main-compat-s80.md`
- `docs/workflow/worker-results/upstream-main-compat-s80-result.md`
- `docs/workflow/qa-reports/upstream-main-compat-s80-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Denied Paths

- `deploy/docker-compose.standalone.yml`, `deploy/.env.example`,
  `deploy/docker-deploy.sh`, `deploy/README.md`, all other deploy documentation,
  Dockerfiles, and entrypoints.
- Any backend, frontend, Ent, migration, Wire, dependency manifest, lockfile,
  VERSION, or product source file.
- `knowledge/**`, global memories, handoff/timeline files, Redis data directories,
  Docker volumes, and ignored build/runtime artifacts.
- Upstream subscription renewal (`1db10dc55`), WS documentation (`8b75dd557`),
  Docker cross-build changes, and every behavior outside `be74deae7`.

## Constraints

- Work only in the isolated S80 worktree; preserve the dirty primary checkout.
- Hand-port the behavior. Do not cherry-pick because the local/dev equivalents
  are required and upstream changed only the main Compose file.
- Change only the four Redis command lines in each business file plus a concise
  explanatory comment. Do not alter image tags, service topology, auth env,
  healthchecks, volumes, networks, ports, or restart policy.
- Preserve `${REDIS_PASSWORD:+--requirepass "$REDIS_PASSWORD"}` exactly. Shell
  hardening for arbitrary special-character passwords is a separate security
  review; S80 validation uses controlled alphanumeric/hex values only.
- Do not run `docker compose up/down/restart/build/pull`, `docker run`, or any
  command that contacts, mutates, replaces, or removes a container or volume.
- Docker daemon availability is not required. Runtime `CONFIG GET` validation
  is explicitly deferred until a disposable environment is separately authorized.
- New contract/result/QA files match the repository `docs/*` ignore rule and
  must be force-added by exact path; never force-add a directory.
- Do not push, deploy, update containers, or merge S80 automatically.

## Acceptance Commands

```powershell
$files = @(
  'deploy/docker-compose.yml',
  'deploy/docker-compose.local.yml',
  'deploy/docker-compose.dev.yml'
)
$env:POSTGRES_PASSWORD = 's80-postgres-test'
foreach ($file in $files) {
  foreach ($password in @('', 'a1b2c3d4e5f6')) {
    $env:REDIS_PASSWORD = $password
    docker compose -f $file config --quiet
    if ($LASTEXITCODE -ne 0) { throw "Compose validation failed: $file" }
    $json = docker compose -f $file config --format json
    if ($LASTEXITCODE -ne 0) { throw "Compose JSON rendering failed: $file" }
    $config = $json | ConvertFrom-Json
    $command = [string]$config.services.redis.command
    foreach ($line in @(
      'redis-server \',
      '--save 60 1 \',
      '--appendonly yes \',
      '--appendfsync everysec \'
    )) {
      if (($command -split "`r?`n" | ForEach-Object { $_.Trim() }) -notcontains $line) {
        throw "Missing continued Redis command line in ${file}: $line"
      }
    }
    if ($password -eq '' -and $command.Contains('--requirepass')) {
      throw "Empty password unexpectedly rendered requirepass: $file"
    }
    if ($password -ne '' -and -not $command.Contains("--requirepass `"$password`"")) {
      throw "Non-empty password did not render requirepass: $file"
    }
  }
}
Remove-Item Env:POSTGRES_PASSWORD -ErrorAction SilentlyContinue
Remove-Item Env:REDIS_PASSWORD -ErrorAction SilentlyContinue

git diff --check
if ($LASTEXITCODE -ne 0) { throw 'S80 diff check failed' }
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S80 has unmerged index entries'
}
```

Evaluator additionally verifies that only the three Redis command blocks differ,
the rendered image/healthcheck/environment/volume/network values match baseline,
all changed paths are allowlisted, no real conflict marker exists, and all three
primary-checkout hashes remain unchanged.

### Pre-commit Tracking Gate

```powershell
git add -u -- deploy/docker-compose.yml deploy/docker-compose.local.yml deploy/docker-compose.dev.yml docs/workflow/spec.md docs/workflow/status.md docs/workflow/main-log.md
git add -f -- docs/workflow/tasks/upstream-main-compat-s80.md docs/workflow/worker-results/upstream-main-compat-s80-result.md docs/workflow/qa-reports/upstream-main-compat-s80-qa.md
git ls-files --error-unmatch docs/workflow/tasks/upstream-main-compat-s80.md docs/workflow/worker-results/upstream-main-compat-s80-result.md docs/workflow/qa-reports/upstream-main-compat-s80-qa.md
if ($LASTEXITCODE -ne 0) { throw 'S80 workflow evidence is not tracked' }
$expected = @(
  'deploy/docker-compose.yml',
  'deploy/docker-compose.local.yml',
  'deploy/docker-compose.dev.yml',
  'docs/workflow/spec.md',
  'docs/workflow/status.md',
  'docs/workflow/main-log.md',
  'docs/workflow/tasks/upstream-main-compat-s80.md',
  'docs/workflow/worker-results/upstream-main-compat-s80-result.md',
  'docs/workflow/qa-reports/upstream-main-compat-s80-qa.md'
)
$actual = @(git diff --cached --name-only)
$pathDelta = @(Compare-Object ($expected | Sort-Object) ($actual | Sort-Object))
if ($pathDelta.Count -ne 0) {
  throw "S80 staged path set differs from allowlist: $($pathDelta | Out-String)"
}
$unmerged = @(git ls-files -u)
if ($LASTEXITCODE -ne 0 -or $unmerged.Count -ne 0) {
  throw 'S80 has unmerged index entries'
}
$conflictMarkers = @(git grep --cached -n -E '^(<<<<<<< .+|=======|>>>>>>> .+)$' -- $expected)
if ($LASTEXITCODE -eq 0 -or $conflictMarkers.Count -ne 0) {
  throw "S80 contains conflict markers: $($conflictMarkers -join ', ')"
}
if ($LASTEXITCODE -ne 1) { throw 'S80 conflict-marker scan failed' }
git diff --cached --check
if ($LASTEXITCODE -ne 0) { throw 'S80 cached diff check failed' }
```

The complete staged path list must contain exactly the nine Allowed Paths and
must exclude ignored `dist`, Redis data, `.env`, and Docker runtime artifacts.

## Output

- Worker result: `docs/workflow/worker-results/upstream-main-compat-s80-result.md`
- QA report: `docs/workflow/qa-reports/upstream-main-compat-s80-qa.md`
- Workflow status/log entries for contract review, implementation, QA, and verdict.

## Stop Rules

- Stop if any Compose file cannot render with controlled empty and non-empty
  passwords, or if command continuation cannot be proven from rendered JSON.
- Stop if the fix requires changing Redis auth design, healthcheck, image,
  volumes, networks, ports, data, or any denied path.
- Stop if validation would need an existing/shared container or volume. Runtime
  proof remains deferred rather than broadening authority.
- Stop if any protected primary-checkout hash changes.
- Stop if a workflow evidence file is absent or untracked at pre-commit time.
- Stop if the index has an unmerged entry, any real conflict marker exists, or
  the staged path set is missing an expected path or contains a non-allowlisted path.
- Stop after two failed attempts on the same behavior and return to Planner.

## Budget

- worker_mode: `Codex direct implementation for a three-file config-only batch`
- qa_mode: `independent static Compose evaluation; runtime deferred by scope`
- worktree_root: `F:/mcplugins/sub2api/.tmp/codex-worktrees`

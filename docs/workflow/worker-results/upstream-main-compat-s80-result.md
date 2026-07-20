### PASS: upstream-main-compat-s80-generator

## Changed Files

- `deploy/docker-compose.yml`
- `deploy/docker-compose.local.yml`
- `deploy/docker-compose.dev.yml`

Each Redis command block now uses a trailing shell continuation after
`redis-server`, `--save 60 1`, `--appendonly yes`, and
`--appendfsync everysec`. The optional `${REDIS_PASSWORD:+--requirepass ...}`
fragment remains the final portion of that one command.

## Contract Compliance

- Hand-ported upstream `be74deae7` behavior to the main Compose and the two
  locally equivalent built-in Redis topologies.
- Did not alter image, container name, restart policy, volume, environment,
  network, healthcheck, dependency, port, external Redis, auth design, or any
  backend/frontend/product file.
- Did not start, stop, build, pull, recreate, inspect, or remove a container or
  volume. Docker daemon runtime validation remains deferred by contract.

## Commands Run

- Docker Compose `v2.29.2-desktop.2`.
- For all three files with controlled empty and non-empty hex Redis passwords:
  - `docker compose -f <file> config --quiet`: PASS.
  - `docker compose -f <file> config --format json`: PASS.
  - Rendered command assertions for all four continuations: PASS (6/6 cases).
  - Empty password excludes `--requirepass`: PASS (3/3).
  - Non-empty password includes the controlled value: PASS (3/3).
- Baseline/current Compose JSON comparison with only `redis.command`
  normalized: PASS for all three files, proving all other rendered topology is
  unchanged. Both sides used `--project-directory deploy` so bind-volume
  resolution was comparable.
- `git diff --check`: PASS; only existing line-ending notices for workflow
  documents were emitted.

## Risks / Deferred Checks

- Empty password is unset by PowerShell 7, so Compose prints its expected
  "variable is not set" warning while returning exit code 0.
- No live `redis:8-alpine` `CONFIG GET appendonly/appendfsync/save` smoke was
  run because container operations are explicitly outside S80 authority.
- Enabling the previously ignored AOF/snapshot flags after a future deployment
  may increase Redis disk I/O and storage use.
- Arbitrary shell-special-character passwords remain an existing separate
  hardening concern; S80 preserves the upstream expression and tests only
  controlled alphanumeric/hex values.

# Task Contract: upstream-x-text-security-s107

- Task ID: `upstream-x-text-security-s107`
- Role: Planner / Generator / Evaluator
- Goal: Adapt upstream `c5971a6fc` by upgrading `golang.org/x/text` to
  `v0.39.0` and its compatible sibling `golang.org/x/*` modules to remove
  GO-2026-5970 without changing application code.
- Success Criteria:
  - `x/text` resolves to `v0.39.0` and GO-2026-5970 is no longer reported.
  - `x/crypto`, `x/mod`, `x/net`, `x/sync`, `x/sys`, `x/term`, and `x/tools`
    resolve to the exact versions selected by the upstream security commit.
  - Existing direct/indirect dependency classification is preserved unless
    the Go toolchain proves a current source-level requirement.
  - Backend production code compiles and dependency verification succeeds.
  - No source, generated, schema, frontend, deployment, or container file is
    changed by this Sprint.
- Allowed Paths:
  - `backend/go.mod`
  - `backend/go.sum`
  - `docs/workflow/tasks/upstream-x-text-security-s107.md`
  - `docs/workflow/qa-reports/upstream-x-text-security-s107-qa.md`
  - `docs/workflow/spec.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: All Go source and generated files, schema/migrations,
  frontend, deployment, containers, VERSION, and unrelated dependencies.
- Constraints:
  - Manually adapt only the eight `golang.org/x/*` version changes from
    `c5971a6fc`; do not copy its parent branch's Go version or dependency
    topology.
  - Preserve concurrent user changes in gateway route files and every S106
    source change.
  - Do not run `go mod tidy` unless module verification demonstrates it is
    required; avoid unrelated checksum churn.
  - Do not commit, push, deploy, or update containers without separate
    authorization.
- Acceptance Commands:
  - `go mod download` for the eight upgraded modules.
  - `go mod verify`.
  - `go list -m` for the eight upgraded modules and exact-version audit.
  - `go build ./cmd/server`.
  - `go test ./internal/service -run '^$' -count=1`.
  - Relevant focused backend tests plus a broad `go test ./...` attempt, with
    unrelated existing failures recorded rather than hidden.
  - `govulncheck ./...` or the repository-equivalent security scan proving
    GO-2026-5970 is absent.
  - Exact two-file dependency diff, `git diff --check`, conflict-marker scan,
    and unmerged-index check.
- Output: Scoped dependency update, security and compile evidence, QA report,
  and final `PASS`, `FAIL`, or `BLOCKED`. No automatic commit or push.
- Stop Rules: Stop if the required versions demand a Go toolchain upgrade,
  application source change, unrelated dependency change, schema/deployment
  work, or any business path outside the approved allowlist.

## Contract Review

`PASS`: The local module graph contains the same eight older `golang.org/x/*`
versions and uses Go 1.26.3, while upstream applies the security set on the
same module family. The local `x/mod` entry is indirect rather than direct, so
the patch must be adapted by version rather than applied as a raw diff.

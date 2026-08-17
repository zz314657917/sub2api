### DONE: upstream-cn-providers-s226-b

## Scope

- Worktree: `E:/codex-worktrees/sub2api/upstream-cn-providers-s226-b`
- Branch: `pge/upstream-cn-providers-s226-b`
- Approved base: `3ed89c9952f09e03861e197d59aad456f3b19b29`
- Implementation commit: `316fa46c62f4e1158a953773f6b287d5c7d42444`
- Upstream provenance checked: `901a0439f1575a45c29150e04d2ccc3ed87f4948`, `4b667ccd45747162ac58545cdbb6f6d88737bf04`, and `e7285453829930e2889432dcd18d7a0c2ba18481`.

## Delivered

- Added Kimi/Zhipu Coding Plan 5h/weekly probes, provider-prefixed snapshots
  including reset timestamps, and failure preservation.
- Added Kimi/DeepSeek payg balance probes with CNY/USD details. Any healthy
  currency prevents a balance pause; recovery only clears the B-owned
  `cn-provider-balance:` temporary-pause reason.
- Reused account proxy resolution, bounded each upstream request to 15 seconds,
  and sized periodic probe deadlines from quota batches and payg workload.
- Validated each final quota/balance endpoint before creating or sending its
  credentialed request. Focused counting transports prove policy-rejected
  targets make zero `HTTPUpstream.Do` calls.
- Body read failures and malformed or incomplete provider payloads are probe
  failures. They keep `Success=false`, skip `UpdateExtra`, and cannot trigger
  a balance-check pause.
- Added admin reads at `/admin/cn-providers/accounts/:id/quota` and
  `/admin/cn-providers/accounts/:id/balance`, with generated Wire wiring and
  cleanup lifecycle integration.
- Limited upstream billing-probe eligibility to known platforms and short-circuited
  official Kimi/Zhipu/DeepSeek domains without a probe request.

## Verification

- `go test -list` discovered all 17 required S226-B focused tests.
- `go test ./internal/service -run <17-test-pattern> -count=10`: PASS (`0.079s`).
- `go test ./internal/service -run <read-failure-and-invalid-payload-pattern> -count=10`: PASS (`5.482s`).
- `go test ./internal/config ./internal/service ./internal/server/routes ./cmd/server -count=1`: PASS.
  - config `0.703s`; service `60.367s`; routes `6.909s`; cmd/server `4.995s`.
- `gofmt -d` on all changed Go files: no output.
- `git diff --cached --check`: PASS before implementation commit.
- `git ls-files -u`: no unmerged entries before implementation commit.

## Contract Compliance

- Changed only S226-B allowlist source/test/Wire paths and this required report.
- Did not modify account foundation, gateway/rate-limit C ownership, migrations,
  dependencies, frontend, database, containers, deployment, remotes, or main.
- No push, real provider request, database operation, or container operation ran.

## Risks And Knowledge Candidates

- Provider response parsing rejects truncated, malformed, and incomplete
  fixture payloads; no real credentials were used, so future provider payload
  changes remain an integration risk for independent QA.
- The periodic job is configuration-disabled by default and needs runtime QA with
  a non-production configuration to observe scheduling lifecycle behavior.
- Knowledge candidate: final-probe URL validation must occur before credentialed
  request construction, with a counting transport regression to prove zero I/O.

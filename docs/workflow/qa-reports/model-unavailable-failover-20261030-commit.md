# Model-unavailable failover: local commit acceptance

Date: 2026-10-30. Verdict: PASS for the scoped local commit.

The user subsequently requested “整理下提交”, authorizing a local commit of the
completed adaptation. Earlier no-commit statements in the immutable contract
and adaptation report describe their original phase. Push, container updates
and deployment remain outside the authorized scope.

## Scope

- Eight frozen Go files: the shared missing-model classifier/cooldown helper,
  three HTTP error owners, pre-top-K scheduler filtering and three new test files.
- Six dedicated documents: original contract and scheduler amendment, their
  review records, adaptation acceptance report and this commit acceptance.
- Shared workflow/status/task snapshots, dependencies, CI, Docker and unrelated
  frontend/backend changes remain outside the commit.

Source freeze SHA256:
`028480614bbced62ba62389a37ae5271e93d4d3fcdd2ab9630f3c1ff19fb40c0`.
Independent adaptation QA seal SHA256:
`f2ed28d1b2ff605d6c49023b0b526e2590e52613032102a81e3efcbb0127fb36`.

## Executed checks

An isolated snapshot starts from parent
`424b96a6e9602d2922a1b8f524644ca2a427fa5b` and overlays only the approved eight
source files and dedicated documents. All source hashes match the accepted
freeze. The four modified production files have no unrelated uncommitted
baseline differences. Committed go.mod/go.sum are preserved exactly, including
the original Go1.26.3 directive; the verification toolchain is Go1.26.10 with
GOTOOLCHAIN=local, readonly modules and offline dependencies.

| Check | Result |
| --- | --- |
| Focused service, default | 108 PASS, zero failures/skips |
| Focused service, unit | 117 PASS, zero failures/skips |
| Model-limit/sticky/scheduler/image cross-unit | 225 PASS, zero failures/skips |
| Complete backend build | PASS |

Counts include Go parent/subtest pass events. The prior adaptation report also
records handler and loopback verification. The existing
`TestOpenAIResponsesWebSocket_FailoverOnUpstreamUsageLimitEvent` failure remains
documented there; this publication check does not claim all backend tests pass.

Raw commands, logs, exact snapshot, source/dependency bindings and protected
worktree hashes are retained under:
`E:/codex-runtime/sub2api/model-unavailable-failover-20261030/publication-20261030`.
`verification.json` binds the passing snapshot; `commit-result.json` records
the resulting commit/tree and the final worktree preservation check.

The first archive preparation stopped before testing or staging because Git
checkout line-ending conversion changed go.mod/go.sum byte hashes. Its original
snapshot and diagnostic remain preserved. The passing snapshot uses
`git -c core.autocrlf=false archive` and verifies the exact committed dependency
blobs. No dependency files in the user's worktree were changed.

## Limits

No real-provider, application-database, deployed-container or production-account
verification. Normal cooldown persistence policy and explicit pool/custom
opt-outs remain as documented in the adaptation acceptance report.

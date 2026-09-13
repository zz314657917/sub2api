---
type: contract-review
scope: repository
status: approved
task_id: upstream-main-ws-cap-s302
verdict: PASS
base_commit: cc5951896
reviewer: final-evaluator
last_verified: 2026-09-13
---

### PASS: upstream-main-ws-cap-s302

The contract is bounded and reviewable. It identifies the local pool owner and
the upstream behavior difference, keeps the global hard cap and non-positive
concurrency semantics explicit, and requires retained-session regression
coverage. Allowed paths are limited to the pool implementation, tests, optional
matching comments, and workflow evidence. Schema, frontend, provider, runtime
data, deployment, merge operations, and protected artifacts are denied. The
acceptance commands are executable for the local Go topology and include build,
formatting, diff, and unmerged-index checks. Real provider/runtime validation
is correctly outside this code-level contract.


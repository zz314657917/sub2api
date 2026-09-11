---
status: approved
review_verdict: PASS
task_id: cf3577a3c-behavior-port
worker_model: gpt-5.6-terra
base_commit: c1833f66c5ca5777ae0d11c0645d5e433ba1502e
spec_ref: docs/workflow/tasks/cf3577a3c-behavior-port.md
---

# Task ID: cf3577a3c-behavior-port

Role: Generator; independent contract and final QA by Evaluator.

Goal: Port every behavior introduced by upstream cf3577a3c to the local monolithic gateway, preserving newer local fixes and documenting equivalent implementations. Never import upstream split gateway files wholesale.

Success Criteria:
- Inventory all 41 upstream paths with local owner, implemented/equivalent result and evidence.
- Request normalization: OAuth unsupported/prompt/commands/reasoning fields, JSON schema compatibility, tool schema lookarounds, image input_fidelity, orphan outputs and Unicode text limit.
- Reserved tool aliases and restoration in HTTP stream/JSON, bridge and WebSocket flows; deterministic native call IDs and unambiguous item references; preserve newer allowed_tools and Astra semantics.
- Explicit status/null-content/cache field rejection retry, bounded and non-mutating on unrelated errors.
- Compaction ordering, nonterminal usage fallback, structured-error classification, request-body diagnostics, Ops stream failure recording and backend response headers.
- Backend go build ./... passes. Focused tests compile actual production sources and run new upstream-derived cases; report baseline test-suite drift separately. Mock transport HTTP/WS tests where feasible; never claim real-provider/database/container evidence.

Allowed Paths: exact product/test list in docs/workflow/tasks/cf3577a3c-coverage.md (part of this contract), and cf3577a3c-prefixed contract/review/result/QA artifacts under docs/workflow. No broad source glob. Main controller may update workflow status.

Denied Paths: main workspace uncommitted changes, frontend/**, migrations/**, Ent/database/schema, dependencies, credentials/config, production, containers, memory/knowledge, git history rewriting and worker commit/push.

Constraints: Work only E:/codex-worktrees/sub2api/cf3577a3c-port. Minimal hunks on existing owners. No checkout-theirs/cherry-pick of target. Do not revert later local functionality. Controller coordinates shared files. Upstream bugs may be corrected only with regression evidence, recorded explicitly.

Acceptance Commands: exact commands and 41-path owner map in cf3577a3c-coverage.md; gofmt and duplicate declaration/conflict marker review. alpha/search store removal is not applicable because the local route does not exist, not authorization to import it. Preserve Astra reasoning.mode=pro and verify this explicit newer-model compatibility exception; allowed_tools contents must survive the port and aliasing.

Output: changed paths, behavior map, executed checks, unverified risks, report under docs/workflow/qa-reports/ or worker-results/ prefixed cf3577a3c.

Stop Rules: report unexpected production/schema/credential dependency rather than importing unrelated code; no silent model fallback; QA must not report overall PASS if required code-path coverage is missing.

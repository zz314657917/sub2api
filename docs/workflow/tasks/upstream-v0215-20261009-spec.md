# Upstream v0.2.15 selective integration

- User approved parallel implementation on 2026-10-09, after the read-only release audit.
- Source tag: `v0.2.15`, commit `86a80c13dcba86f52f9ca815b5a471cecc236227`.
- Base HEAD: `ec2a9a2c90d3c8002fe9e9f01d660a6451c03f3b`; dirty business baseline was copied without secrets to this isolated checkout.
- Snapshot and immutable comparison files: `E:/codex-runtime/upstream-v0215-20261009/`.
- Three disjoint Developer contracts, independent contract review and independent QA.
- Authorized model exception: after the configured gpt-6.1-sol CLI route returned HTTP 403, the controller asked to use the current-session default model for multiple agents and the user replied "继续". Native agents inherit the current-session model; no global/repository model policy is changed.
- Integration means minimal validated changes copied into the original working tree after source-hash conflict checks. No commit, push, Docker operation, migration, real-provider call or production change.
- Scope: protocol conversion/serialization robustness; independent frontend lifecycle/numeric fixes; WS image usage/current group billing and OAuth history/reasoning recovery.
- Deferred: Cline, Command Code, platform/profile refactor, toolchain/dependency upgrade, payments, TPS features, monitor V2, namespace performance refactor.
- Preserve `/models` public model plaza, current privacy policies, GroupID-only channel pricing, account/user multipliers, long-context pricing, image/cached-image pricing and idempotent ledger/trial bookkeeping.
- Existing shared `status.md` belongs to another task and must not be overwritten. Only the isolated status tracks this batch.
- Runtime contract: local fake HTTP/WS upstreams where feasible, no paid providers or shared database. Vue interactions via Vitest mounted components; browser visual acceptance is not claimed.

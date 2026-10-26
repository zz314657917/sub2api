### DONE: kimi31-pro-cny-20261026

- Developer isolated candidate: `E:/codex-worktrees/sub2api/kimi31-pro-cny-20261026`.
- Full report, six-file SHA256 freeze, declaration inventory, red/green logs and scope audit: `E:/codex-runtime/kimi31-pro-cny-20261026/developer/`.
- Implemented canonical `kimi-k3.1-pro` with finite aliases and known provider prefixes. Per-million site-credit rates: input13/output68/read1.3/write5m13/write1h26; CNY convention remains 1:1.
- Default dynamic, catalog and fallback pricing agree. Unknown reserved variants cannot inherit the standard default. Group/channel overrides, explicit zero, intervals and request/image modes remain available.
- Catalog246→247: exactly one canonical addition; all existing entries semantically unchanged. No dependency changes are included.
- Full default selection: 33 top-level,309 including subtests passed. Unit:79 top-level,373 including subtests;372 passed and one independently reproduced original `long-image-model` failure.
- Six new test declarations each tag repeated3:18 top-level,369 including subtests passed in each tag, zero missing/skipped. Complete backend build, task gofmt and diff checks passed.
- Final assertions run against baseline source using external Go overlays fail on the new pricing/settlement behavior in both tags; the same assertions pass on the candidate. Existing tests were not changed.
- Two recorder paths execute actual in-memory settlement and assert component costs, debit, idempotency, multipliers, custom overrides and final account-model cost ownership.
- Limitations: no real provider/database/deployment verification. OpenAI usage has no1h write split. Existing Gateway account-cost projection forwards aggregate writes only; its TTL-specific account statistics are outside this change. Partial cache detail behavior follows the existing normalization algorithm.
- Developer stopped editing the frozen six paths. Final independent QA and protected integration are separate gates.

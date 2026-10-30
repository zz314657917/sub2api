### PASS: model-unavailable-failover-20261030

- Independent Contract Reviewer: native agent 01a18aee-2aa2-76f2-a9e6-20155e178188.
- Reviewed draft SHA256: 826f640c7b3d8777ecb79548d9050b6fba57995aaf73745ea72449459e24f55e.
- Full review: E:/codex-runtime/sub2api/model-unavailable-failover-20261030/contract-review-r1.md.
- Immutable reviewed draft: the same evidence directory, contract-reviewed-draft.md.
- Controller approved without changing scope: only `status: approved` and
  `review_verdict: PASS` replace the corresponding draft metadata.
- Independent metadata rebind PASS: contract-approval-binding.md in the same
  evidence directory confirms all other bytes unchanged. Approved SHA256:
  da964c460e0b768d12b136ab5821c547bd037532b0ec181a771eba1b9199a6f3.
- Review confirms allowlist covers Chat-to-Responses pre-error endpoint fallback
  and automatic passthrough; authoritative fields/conflicting codes, account
  scope, pool/custom policy, image guards, mapping-key consistency, bounded
  handler retries and post-output replay barriers remain mandatory.
- Baseline focused service/default and unit plus build passed; handler/default
  and unit each retain the original WebSocket usage-limit test failure.
- This is contract approval only. Independent implementation/QA and protected
  integration remain required. No commit, push or deployment authorization.

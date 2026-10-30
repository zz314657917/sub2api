### PASS: model-unavailable-failover-20261030

Independent Reviewer: 01a18aee-2aa2-76f2-a9e6-20155e178188.

- Scheduler amendment SHA256:
  dc13191e86bf9b0cd9c20f4b9e3b412234597110c14654db14701cdbeef4d611.
- Reviewed protect-r3 SHA256:
  9c4bdbe1d5bab7b68d43829a33178c3e2de951239e7b0264c7884b97180ca959.
- Full independent review:
  E:/codex-runtime/sub2api/model-unavailable-failover-20261030/scheduler-addendum-review.md.
- Authorizes only the eighth scheduler source path and additions to the three
  task-owned new tests. The four R2 production files retain exact R2 hashes.
- Requires unchanged valid top-K=1 next-request sticky bypass and default
  top-K=7/eight-account coverage. Enlarging test top-K cannot close the defect.
- Original approved contract and R1/R2 sources/failures remain immutable.

Protection reconciliation was independently reviewed separately:

- R1 FAIL: a concurrent Dashboard commit changed HEAD/index during review;
  original observed four frontend file hashes did not change.
- R2 PASS binds observed-r2 SHA256
  04962d291950e4f7f2acfbc29bbf8a52daea7a402faf41cf96de4b2743d58d59
  and protect-r5 SHA256
  bba82a09d4b11fdee8604befae4b3290b8fa3aafb6cf1fe792d1d8117f648d4a.
- Full report: the same evidence root, reconciliation/review-r2.md.
- Current expected HEAD f88b54e0c66ff66d732359c5cc80905e70d16988 and exact
  index/four frontend hashes are preserved. Any new drift fails closed.
- Fresh full-tree protection passed after review. Eight-file source freeze,
  independent QA and fresh main verification remain separate gates.

No commit, push, container, database or deployment authorization.

### Final protection reconciliation — PASS, 2026-10-30

The earlier protect-r5 approval above is historical. Final independent review
reconciliation/review-r4.md approves protect-r8 SHA256
2d4e84cfd92041a4d4ef3501ff61b5011c3b22fd60c6551625745aebd53dc24f
against observed-r3 SHA256
3e4f4eae5e538293f9db8439d73f670481fe15bbb5f3ba5e8ffad6bc36336b9b.
R6/R7 review failures remain preserved.

- Eight frontend files are bound to exact observed bytes. Every accepted
  concurrent commit, path and blob is checked; merges, intermediate unobserved
  content, backend edits followed by restoration and nonempty staging fail.
- Eleven synthetic Git controls exercise the actual R8 identity block and PASS,
  including metadata/same-byte acceptance and deletion/staging rejection.
  Evidence: reconciliation/identity-probe-r3-results.json.
- Protected integration and final audit PASS for12250 original paths with the
  four explicit workflow exclusions. Concurrent HEAD
  424b96a6e9602d2922a1b8f524644ca2a427fa5b, exact index and frontend bytes remain;
  staged tree is empty. This task did not write HEAD/index.
- Independent QA seal/source8 and fresh main verification remain unchanged.
  controller-final-result.json records final PASS at13:45 CST; the QA report
  states the identical pre-existing WS failure and real-provider/DB limitations.

This closes local adaptation only; publishing and deployment remain outside scope.

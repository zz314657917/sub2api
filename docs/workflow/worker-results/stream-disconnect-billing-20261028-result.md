### DONE: stream-disconnect-billing-20261028 implementation

Four independent scoped Developers delivered Gemini, Anthropic, OpenAI and
handler/worker changes in the isolated candidate. This records implementation
delivery only; independent final QA and exact main integration remain required.

- Candidate: `E:/codex-worktrees/sub2api/stream-disconnect-billing-20261028`.
- Evidence: `E:/codex-runtime/sub2api/stream-disconnect-billing-20261028`.
- Freeze24 SHA256: `ab0580e9e15be703d5861c5d130b0cbb0eb66e590f449e71802fdc0caa8cf46f`.
- All four Developers stopped editing. Reports: `gemini/report.md`,
  `anthropic/report.md`, `anthropic/r2/report.md`, `anthropic/r3/report.md`,
  `openai/report-r2.md`, `handlers/report.md`.
- R2 corrected Anthropic disconnected in-band errors/EOF usage tails and OpenAI
  hidden Committed fallback. Original failed runs and source hashes remain.
- Independently approved R3 migrates only three pre-existing ping test cases.
  All prior23 source bytes remain identical; the extra test is the sole24th
  path. Original freeze23, canonical FAIL and old QA checkout are preserved.
- Developer focused checks and deterministic stress passed within their
  respective scopes; controller's52 new declarations ran three times per tag,
  and prior A1-A7/B2/protocol/first-response cross checks passed.

No real provider, live database, container, commit, push or deployment is included.
Build and broader/final runtime outcomes must be read from the final QA report;
this implementation report does not waive independently reproduced old failures.

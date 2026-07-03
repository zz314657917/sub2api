### DONE: upstream-main-v0141-codex-reasoning-preserve-s48

## Summary

- Ported upstream `73de2ea7` into the isolated S48 worktree.
- Changed Codex input filtering so `reasoning` items are preserved instead of dropped.
- Stripped replay-unsafe `rs_*` ids from reasoning items and backfilled missing `summary` as an empty array.
- Added regression coverage for encrypted reasoning preservation, bare reasoning id stripping, summary backfill, non-empty summary/content preservation, and mixed input tool-call pairing.

## Changed Files

- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/openai_codex_transform_test.go`
- `docs/workflow/tasks/upstream-main-v0141-codex-reasoning-preserve-s48.md`
- `docs/workflow/worker-results/upstream-main-v0141-codex-reasoning-preserve-s48-result.md`
- `docs/workflow/qa-reports/upstream-main-v0141-codex-reasoning-preserve-s48-qa.md`
- `docs/workflow/status.md`
- `docs/workflow/main-log.md`

## Commands Run

```powershell
gofmt -w backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go
go test ./internal/service -run "TestFilterCodexInput|TestApplyCodexOAuthTransform" -count=1
git diff --check -- backend/internal/service/openai_codex_transform.go backend/internal/service/openai_codex_transform_test.go docs/workflow/status.md docs/workflow/main-log.md docs/workflow/tasks/upstream-main-v0141-codex-reasoning-preserve-s48.md
```

## Test Output

- `internal/service`: PASS.
- `git diff --check`: PASS for code and workflow files; only existing workflow doc line-ending warnings were emitted.

## Risks

- This preserves reasoning items that were previously dropped. The regression tests verify `rs_*` ids are still removed, but live chatgpt.com Codex OAuth behavior was not re-tested in this local Sprint.
- The change intentionally leaves non-reasoning item reference and tool-call id normalization behavior unchanged.

## Knowledge Candidates

- None.

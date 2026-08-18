### PASS: upstream-gemini-tool-config-s232

## Independent QA

QA ran from a separate clean worktree at business commit `6a5d4d95a`, without
using the Controller result report.

## Evidence

- Focused mixed/function-only/web-search-only tests, `-count=10`: PASS.
- Complete `go test ./internal/pkg/antigravity -count=1`: PASS.
- Complete `go test ./internal/service -count=1`: PASS (70.722s).
- `go test ./cmd/server -run '^$' -count=1`: PASS (10.580s).
- `gofmt -l` on the three allowed Go files: no output.
- `git diff --check 91e7b4f820..HEAD`: PASS.
- Changed product files are exactly the three contract allowlisted paths.
- No later upstream owner edits after `1ba92449c` were found.
- QA worktree has no uncommitted or unmerged entries.

## Findings

No implementation findings. No live Gemini/provider request was performed.

## Contract Compliance

PASS. The optional flag is emitted only for mixed built-in/function tool
declarations; function-only and web-search-only requests remain unchanged.

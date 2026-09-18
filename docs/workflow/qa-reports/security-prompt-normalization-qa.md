### PASS: security-prompt-normalization

# Independent QA Report

## Scope and boundary

- Reviewed the frozen implementation in `backend/internal/securityaudit/prompt_snapshot.go` and `backend/internal/securityaudit/prompt_snapshot_test.go` against `docs/workflow/tasks/security-prompt-normalization.md`, its independent contract review, and the Developer result.
- The feature diff changes only the approved production/test files. Existing uncommitted `docs/workflow/main-log.md` and `docs/workflow/status.md` changes were present outside this feature scope and were neither modified nor treated as feature evidence.
- No session-risk or Composite code is included. No provider, production data, browser, network provider traffic, commit, push, deploy, or configuration write was performed.

## Review findings

- `buildPrioritizedScanText` applies NFKC, strips U+200B/U+200C/U+200D/U+2060/U+FEFF, converts NUL/backspace/vertical-tab/form-feed/U+000E/U+000F to spaces, preserves readable internal newlines, trims outer whitespace, and omits normalized-empty segments.
- Existing role selection and latest-user priority assembly execute before the normalization loop. The focused and package tests preserve role coverage and priority behavior.
- `Request.Body` remains JSON-unmarshaled only; the six-protocol normalization test verifies the caller byte slice is unchanged. Redaction still derives from normalized metadata and the existing `Redacted`/storage gates still clear `ScanText`.
- The all-empty normalized case keeps the reviewed source behavior: a non-empty extracted segment list returns an empty scan/full/preview representation, SHA-256 of the empty metadata string, length zero, and its original message count.
- Production paths were traced: async enqueue calls `ExtractPromptSnapshot` then persists `snapshot.ScanText` as the worker payload; blocking evaluation calls `ExtractBlockingPromptSnapshot` then `GuardEvaluator.Evaluate`, which splits and scans `snapshot.ScanText`. The fake-scanner test proves normalized text reaches the Guard scanner without a provider call.

## Executed verification

From `backend`:

- `go test ./internal/securityaudit -count=1` -- PASS (`0.784s`).
- `go test -tags unit ./internal/securityaudit -count=1` -- PASS (`0.788s`).
- `go build ./...` -- PASS.
- `go test ./internal/securityaudit -run 'TestPromptSnapshot(NormalizesTextProtocolsWithoutMutatingRequestBody|NormalizationRetainsPriorityAndOmitsEmptySegments|NormalizationKeepsAllEmptySnapshotSemantics|NormalizedScanTextReachesGuardAndRedaction)$' -count=1 -v` -- PASS. This covers Chat, Responses, Anthropic, Gemini, Images, and Responses WebSocket; full-width/zero-width/control input; Unicode/emoji; body preservation; priority; empty segments; all-empty semantics; redaction; and scanner input.
- `gofmt -d internal/securityaudit/prompt_snapshot.go internal/securityaudit/prompt_snapshot_test.go` -- no output.
- `git diff --check` and `git diff --cached --check` -- PASS; no unmerged paths.
- Baseline `c30bb5350006490eaf0dac314e6dc66746818930` is an ancestor of the checked HEAD.

## Unverified limits

- Real Guard/provider behavior, PostgreSQL/Redis persistence, container/deployment, and production traffic are outside this local QA scope. The scanner assertion uses `PromptScannerFunc` and does not contact a provider.

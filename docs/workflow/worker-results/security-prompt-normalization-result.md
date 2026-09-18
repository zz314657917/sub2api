### PASS: security-prompt-normalization

# Developer Result

## Scope

- Modified only `backend/internal/securityaudit/prompt_snapshot.go` and `backend/internal/securityaudit/prompt_snapshot_test.go` for the feature.
- Added this result record. Pre-existing controller changes in `docs/workflow/main-log.md` and `docs/workflow/status.md` were preserved and not edited.

## Implemented behavior

- Ported the `bef505942` prompt-normalization hunk at prioritized scan-text assembly: NFKC normalization, removal of U+200B/U+200C/U+200D/U+2060/U+FEFF, replacement of NUL/backspace/vertical-tab/form-feed/U+000E/U+000F with spaces, outer trimming, and omission of empty normalized segments.
- Metadata, hash, preview, full prompt, and scanner input now derive from the normalized assembly. Request body bytes are only unmarshaled and are not modified.
- Kept the reviewed all-empty behavior: extracted non-empty segments that normalize away produce an empty snapshot payload with SHA-256 of the empty string, length zero, and the original pre-normalization message count.
- Did not change protocol extraction, role selection, guard policy/actions, persistence/redaction gates, provider payloads, session-risk, or Composite ownership.

## Test evidence

From `backend`:

- `go test ./internal/securityaudit -run 'TestPromptSnapshot(NormalizesTextProtocolsWithoutMutatingRequestBody|NormalizationRetainsPriorityAndOmitsEmptySegments|NormalizationKeepsAllEmptySnapshotSemantics|NormalizedScanTextReachesGuardAndRedaction)$' -count=1 -v` — PASS.
  - Covers Chat, Responses, Anthropic, Gemini, Images, and Responses WebSocket extraction; full-width/zero-width/control normalization; Unicode and emoji; request-body byte preservation; priority and empty segment semantics; all-empty reviewed semantics; redaction; and a fake Guard scanner receiving the normalized input.
- `go test ./internal/securityaudit -count=1` — PASS.
- `go test -tags unit ./internal/securityaudit -count=1` — PASS.
- `go build ./...` — PASS.
- `gofmt -d internal/securityaudit/prompt_snapshot.go internal/securityaudit/prompt_snapshot_test.go` — no output.
- `git diff --check` — PASS.

## Boundary and remaining risk

- No network, paid provider, browser, deploy, commit, or push was used. Scanner-path coverage uses `PromptScannerFunc` only.
- Composite remains outside this batch, as do the source commit's session-risk changes.

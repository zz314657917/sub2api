### PASS: group-model-match-auth-enforcement-s272

## Findings

- No blocking findings. `ResolveAPIKeyForModelRequest` now keeps only the existing nil guards before calling `APIKeyService.ResolveForModelRequest`; the former `len(MultiGroupRoutes) == 0 && PinnedAccountID <= 0` authorization bypass is removed.
- The existing resolver rejects an incompatible effective default group through `Group.MatchesModel`, and middleware returns the stable HTTP 403 `NO_MATCHING_GROUP_ROUTE` error. The S272 mismatch test verifies this for group 41 excluding `gpt-5.6-luna`; the wildcard-match control remains allowed.
- Production diff is one line in `backend/internal/server/middleware/api_key_auth.go`; the only other product artifact is the new focused S272 test file. S271 task/result/report paths are clean. `outputs/` remains an untracked protected directory and was not part of the S272 product patch.

## Executed Checks

- `go test ./internal/server/middleware -list '^TestS272'`: both required S272 test names found.
- `go test ./internal/server/middleware -run '^TestS272' -count=10`: PASS, exit 0.
- `go test ./internal/service -count=1`: PASS, exit 0, 65.381s.
- `go test ./internal/server/middleware -count=1`: PASS, exit 0, 0.060s.
- `go test ./internal/service -run '^TestS(88|91)' -count=1`: PASS, exit 0.
- Pinned-account no-fallback and bound-group focused suite: PASS, exit 0, including `TestGatewayService_PinnedAccountRejectsFallbackAndUsesOnlyBoundAccount`, `TestGeminiMessagesCompatService_PinnedAccountRejectsStickyAndFallback`, and `TestAPIKeyService_PinnedSnapshotRoundTripAndRouteStaysOnBoundGroup`.
- `go test ./cmd/server -run '^$' -count=1`: PASS, exit 0.
- `git diff --check`: PASS (only existing CRLF conversion warnings for workflow files).
- `git ls-files -u`: PASS, no unmerged entries.
- `git diff --name-only`, staged-path review, and targeted S271/`outputs/` status review: S272 production scope is limited to the approved middleware path; no staged changes or S271-path modifications observed.

## Unverified Risks

- No live provider request, deployment, container update, or production API probe was authorized or performed.
- `outputs/` is intentionally untracked, so Git cannot prove its historical ownership; this QA run did not create or modify it, and it remains outside S272 scope.

## Recommendation

Accept S272. The middleware no longer bypasses group-owned model restrictions for a single-group API key, while S88/S91 routing and pinned-account no-fallback regressions remain green.

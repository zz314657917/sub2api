### PASS: upstream-v0171-upstream-path-guard-triage

## Findings

- Upstream `017f6bbd5` is not an ancestor of this integration branch and its raw patch does not apply because the local gateway topology has since diverged.
- Its safety invariants are already present through local commit `31918bca5` and remain active after later gateway changes: all three `/responses/*subpath` route families reject non-forwardable suffixes before dispatch, Gemini request paths and AI Studio URL construction validate model segments, AI Studio GET suffixes are validated, and video task identifiers reject traversal/control-byte inputs before `PathEscape`.
- Local validation is at least as strict for whitespace because it rejects rather than trims a client-controlled path segment.

## Executed Checks

- `git fetch upstream --tags --prune`: `upstream/main` remained `00b859617` (`v0.1.171`).
- `git diff 017f6bbd5^ 017f6bbd5 | git apply --check --verbose`: did not apply, as expected for the divergent local topology; no patch was applied.
- `go test ./internal/service -run 'Test(SanitizedUpstreamPathSuffix|OpenAIResponsesRequestPathSuffix|AppendOpenAIResponsesRequestPathSuffix|ValidateEscapedUpstreamPathSegment|BuildGeminiAIStudioModelActionURL)' -count=1`: passed.
- `go test ./internal/server/routes -run 'TestGatewayRoutes(ResponsesSubpathRejectsNonConformingSubpaths|OpenAIResponsesCompactPathIsRegistered)' -count=1`: passed.
- `go test ./internal/handler -run '^TestGeminiV1BetaInvalidModelPathSegments$' -count=1`: passed.
- `go test ./internal/service -run 'Test.*(Video|Task).*' -count=1`: passed.
- `git diff --check`, `git status --short`, and `git ls-files -u`: clean.

## Unverified Risks

- No live upstream request, account credential, shared database, container, deployment, primary-worktree change, push, or merge to `main` was performed.

## Recommendation

Do not create S193 or duplicate-port `017f6bbd5`. Treat it as behaviorally covered and continue only when a later upstream candidate has a new bounded seam.

### PASS: upstream-v0168-codex-manifest-s132-integration

## Findings

- No blocking issue found in the scoped security and route review. The
  manifest handler requires existing API-key authentication, an assigned group,
  and the OpenAI platform before account selection can begin.
- The mainline Responses subpath guard was retained. The manifest selector is
  additive: ordinary `/v1/models` still calls the existing model-list handler,
  while only OpenAI-group probes with `client_version` reach the new path.
- Source tests use local in-process upstream HTTP servers. OAuth payloads pass
  through verbatim and API-key normalization is confined to the targeted
  compatibility field.

## Executed Checks

- Reviewed every changed path against the approved contract and the
  `review-and-verification` evidence-first checklist.
- `gofmt -w` on every changed Go file.
- `go test ./internal/service -run 'CodexModels' -count=1`
- `go test ./internal/server/routes -run 'CodexModels' -count=1`
- `go test ./... -run '^$'`
- `go build ./...`
- `git diff --cached --check`, `git diff --check`, `git ls-files -u`, and
  changed-path allowlist audit.

## Unverified Risks

- No real Codex client, API key, OpenAI/ChatGPT upstream, account selection,
  cache traffic, or external HTTP request was used.
- The handler's API-key/group rejection paths were code-reviewed and compile
  checked but not exercised through a full authenticated HTTP integration.
- No container, deployment, rollback, production configuration, or production
  traffic was touched.

## Recommendation

Safe to fast-forward locally as a source-level adapter. Before enabling for
users, run an isolated authenticated API-key smoke with a local fake upstream
to exercise the handler/middleware path, normal `/v1/models`, and the
`client_version` manifest route together.

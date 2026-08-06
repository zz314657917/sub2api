### PASS: upstream-v0171-codex-identity-probe-regression-s201

## Findings

- The S187 shared egress normalization correctly rewrites `codex-tui` identities to `codex_cli_rs`; two pre-existing OAuth probe regressions still asserted the superseded identity and untrimmed User-Agent.
- The S192 exact-term anchor behavior correctly initializes renewed quota windows at `StartsAt`; one pre-existing expired-semantic renewal regression still asserted the superseded local-midnight anchor.
- This task changes only those obsolete expectations. Production behavior, configuration, routes, dependencies, schemas, migrations, containers, deployment, the original E: worktree and the primary worktree remain unchanged.

## Executed Checks

- `gofmt -w internal/service/account_test_service_openai_compact_test.go internal/service/openai_gateway_service_test.go internal/service/subscription_assign_idempotency_test.go`: passed.
- `go test ./internal/service -run '^(TestAccountTestService_TestAccountConnection_OpenAIImageOAuthEnforcesFinalCodexIdentity|TestOpenAIGatewayService_RecordLunaIdentityPairsOfficialCodexHeaders|TestAssignSubscriptionRenewsExpiredSemanticMatch)$' -count=1`: passed.
- `go test ./internal/service -count=1`: passed.
- `go test ./cmd/server -run '^TestNonExistent$' -count=0`: passed.
- `git diff --check`, conflict-marker scan, unmerged-index check and staged allowlist audit: passed.

## Unverified Risks

- No real OpenAI/Codex upstream, subscription API, database, payment, container, deployment or production traffic was used. The conclusion is source-level regression coverage only.
- This recovery commit is in the temporary isolated branch because the original E: worktree is read-only in the current session. It must be cherry-picked onto that branch before any later integration decision.

## Recommendation

Commit the scoped recovery patch to `codex/upstream-v0171-integration-s201`; do not modify the primary worktree, push or deploy.

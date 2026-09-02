### PASS: upstream-v0185-anthropic-fallback-s287

## Implemented

- Added named Anthropic fallback beta constants without adding them to default
  OAuth/API-key mimic headers.
- Extended the existing pre-signature Anthropic body sanitizer so
  `fallbacks` requires `server-side-fallback-2026-07-01`, while
  `fallback_credit_token` accepts the server-side fallback beta or either
  supported fallback-credit beta.
- Cleared both Anthropic-only fallback fields on the Bedrock request path,
  including Claude Code request-body sanitization.
- Added focused regression coverage for direct sanitize behavior, API-key and
  OAuth-mimic request construction, Bedrock preparation, and context-management
  interaction.

## Scope

Only the approved S287 allowlist was changed. No provider, database, migration,
Ent, handler, frontend, container, deployment, or protected dirty path was
modified.

## Checks

- Focused fallback/context-management tests x10: PASS.
- Complete `go test ./internal/service`: PASS.
- `go test ./cmd/server -run '^$' -count=1`: PASS.
- `go build ./...`: PASS.
- Frontend typecheck and production build: PASS (existing workspace evidence).
- `gofmt -d`, `git diff --check`, and unmerged-path checks: PASS.
- Protected dirty aggregate hash remains
  `0e467987fd7aec5fc451983bdb8f8216f97ba69c`.

## Unverified

Real Anthropic/Bedrock provider behavior, database, container, deployment, and
browser smoke remain outside this Sprint contract.

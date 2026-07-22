### PASS: upstream-openai-inactive-workspace-plan-s104

## Findings

- No blocking issue was found.
- The runtime behavior matches upstream `d0b8760eb`: a token-derived plan type,
  including `k12`, is preserved, while `accounts/check` remains a fallback
  only when the current plan is empty.
- Expired and explicitly deactivated workspace candidates no longer win
  fallback selection; active and malformed-expiry candidates retain the
  existing best-effort behavior.
- No import identity, PAT/Agent Identity, persistence, scheduler, gateway,
  frontend, migration, billing, deployment, or container path changed.

## Executed Checks

- Test discovery listed all four required account-info/subscription tests.
- Focused service tests passed once, then passed for ten consecutive runs.
- Broader OpenAI OAuth/account-info regression tests passed, including OAuth
  URL, exchange-state, redirect, and no-refresh-token behavior.
- The three-file business diff was reviewed against upstream `d0b8760eb`; the
  local privacy-service behavior matches the upstream target and the OAuth
  hunk differs only in a shorter equivalent comment plus the explicit K12
  regression value.
- `gofmt -d`, `git diff --check`, conflict-marker scan, exact allowlist audit,
  and unmerged-index check passed.

## Unverified Risks

- No real K12 credential was used, so live `accounts/check`, token refresh,
  and upstream request behavior remain unverified.
- No deployment or running-container refresh/smoke was performed.
- Codex PAT (`at-*`) and the separate K12 `gpt-5.6-sol` 503 issue remain out of
  scope.

## Recommendation

- `PASS / source-only`: the scoped source, regression, contract, and evidence
  are ready for publication. Keep PAT support and live K12 credential smoke as
  separate work.

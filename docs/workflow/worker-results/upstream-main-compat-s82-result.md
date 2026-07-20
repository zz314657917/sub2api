### PASS: upstream-main-compat-s82-generator

## Changed Files

- `README.md`
- `deploy/config.example.yaml`
- `frontend/src/i18n/__tests__/wsModeLocaleDesc.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`

## Implemented Behavior

- README now documents the YAML and environment forms of the global v2 mode
  router prerequisite for account-level OpenAI Responses WS modes.
- README explicitly distinguishes local oversized-first-message
  `http_bridge_enabled` fallback from account-level mode selection.
- The example config explains that disabling the v2 router ignores account WS
  modes and retains legacy routing; no configuration value changed.
- English and Chinese account help text names the global prerequisite and the
  locally supported `ctx_pool` / `passthrough` selections.
- A focused locale test locks both descriptions and rejects unsupported
  account-level `http_bridge` wording.

## Commands Run

- Focused Vitest after adapting the upstream test to the local locale export
  shape: PASS, 1 file / 1 test.
- `npm.cmd run typecheck`: PASS.
- `npm.cmd run build`: PASS, 1080 modules transformed.
- README/config prerequisite assertions: PASS.
- Comment-stripped example-config comparison against baseline: PASS; no config
  value changed.
- Business diff review, `git diff --check`, unmerged-index scan, real
  conflict-marker scan, and protected primary-checkout hashes: PASS.

## Risks / Deferred Checks

- No browser smoke was run because this change affects static help copy only;
  locale import/typing and production compilation are covered.
- Existing Vite dynamic-import and large-chunk warnings remain non-blocking and
  are unrelated to S82.
- S82 does not add the upstream account-level `http_bridge` runtime feature.
  The local oversized-message bridge remains separate and unchanged.

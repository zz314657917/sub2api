### DONE: upstream-codex-image-tool-strip-policy-s68a-ui

## Summary

- Added a dedicated warm-white/terracotta `allow` / `strip` control beside the existing Codex image bridge card.
- Exposed the control for OpenAI OAuth, setup-token, and API Key accounts without broadening setup-token access to unrelated OpenAI settings.
- Loaded only the exact persisted `strip` value as `strip`; absent or unknown values resolve to `allow`.
- Cloned the latest `updatePayload.extra` or account `extra` before adding/removing `codex_image_generation_explicit_tool_policy`, preserving unknown keys.
- Added focused coverage for all three account types, both save paths, load behavior, unknown-key preservation, and setup-token control isolation.

## Verification

- `npm.cmd run test:run -- src/components/account/__tests__/EditAccountModal.spec.ts`: PASS (`25/25` tests).
- `npm.cmd run typecheck`: PASS.
- `git diff --check`: PASS.

The independent worktree reused the main repository's existing `frontend/node_modules` through a temporary junction for verification; the junction was removed afterward and did not enter the diff.

## Path Audit

Changed paths are limited to the contract allowlist:

- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/worker-results/upstream-codex-image-tool-strip-policy-s68a-ui-result.md`

No API, type, DTO, backend, migration, deployment, production configuration, or generic `extra` editor path was changed.

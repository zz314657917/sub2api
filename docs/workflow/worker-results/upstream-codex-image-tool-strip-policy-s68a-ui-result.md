### DONE: upstream-codex-image-tool-strip-policy-s68a-ui

## Summary

- Added a dedicated warm-white/terracotta `allow` / `strip` control beside the existing Codex image bridge card.
- Exposed the control for OpenAI OAuth, setup-token, and API Key accounts without broadening setup-token access to unrelated OpenAI settings.
- Matched backend precedence and normalization by reading a string top-level policy first, falling back to nested `extra.openai` only when needed, and treating trimmed case-insensitive `strip`, `remove`, and `drop` aliases as strip.
- Cloned the latest `updatePayload.extra` or account `extra` and its nested `openai` map before removing legacy nested policy state and normalizing the top-level policy, preserving neighboring unknown keys.
- Added focused coverage for all three account types, both save paths, top-level/nested precedence, unknown-key preservation, and setup-token control isolation.

## Verification

- `npm.cmd run test:run -- src/components/account/__tests__/EditAccountModal.spec.ts`: PASS (`30/30` tests).
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

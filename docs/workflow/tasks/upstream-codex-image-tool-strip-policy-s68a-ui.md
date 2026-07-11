# Task Contract: upstream-codex-image-tool-strip-policy-s68a-ui

## Task ID

`upstream-codex-image-tool-strip-policy-s68a-ui`

## Status

`approved`

## Role

You are the Generator worker for the account-maintenance UI half of the Codex explicit image tool strip policy prerequisite.

## Goal

Add a dedicated, maintainable `allow/strip` control to the existing OpenAI account edit modal, preserving all unknown `extra` keys and supporting OAuth, setup-token, and API Key accounts.

## Success Criteria

- A dedicated two-option control appears next to the existing Codex image-generation bridge controls for OpenAI OAuth, setup-token, and API Key accounts.
- Existing `strip` values load as `strip`; absent or unknown values load as `allow`.
- Saving `strip` writes only `codex_image_generation_explicit_tool_policy: "strip"`.
- Saving `allow` removes the policy key instead of persisting a redundant default.
- All unrelated and unknown account `extra` keys survive both save paths.
- Setup-token load/save uses a narrow policy block and does not enable unrelated OAuth/API Key WS, compact, passthrough, or quota controls.
- Styling follows the current warm-white/terracotta account modal rather than copying stale upstream slate styling.
- Focused component tests cover all three account types, load/save, and unknown-key preservation.

## Context

- Upstream UI reference: `f385cdceb`.
- Local i18n is split into `frontend/src/i18n/locales/{en,zh}/admin/accounts.ts`.
- The generic account update API replaces `extra` as a map, so this control must clone the current/update payload and change only the policy key.

## Allowed Paths

- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/worker-results/upstream-codex-image-tool-strip-policy-s68a-ui-result.md`

## Denied Paths

- Create/Bulk account modals, account API/types/stores, import/export, backend, migrations, deployment, and production configuration.
- Any raw JSON editor or generic `extra` editor.
- `knowledge/**` and global memories.

## Constraints

- Reuse existing modal components, icons, spacing, and color tokens.
- Do not broaden the existing OAuth/API Key settings block to setup-token; add narrow policy load/save logic for the three supported account types.
- Do not add API or type fields; use the existing `extra` map.
- Do not expose namespace-specific wording yet; S68b owns namespace expansion.
- Preserve all unrelated `extra` keys by cloning the latest `updatePayload.extra` or account `extra` before editing the policy key.

## Acceptance Commands

```powershell
Push-Location frontend
npm.cmd run test:run -- src/components/account/__tests__/EditAccountModal.spec.ts
npm.cmd run typecheck
Pop-Location
git diff --check
```

## Output

- Write `docs/workflow/worker-results/upstream-codex-image-tool-strip-policy-s68a-ui-result.md` with a DONE/BLOCKED/FAILED first line.
- Commit all contract changes and return the commit hash.

## Stop Rules

- Stop if a new API/type/DTO or generic raw JSON editor is required.
- Stop if setup-token support would require enabling unrelated OpenAI control logic.
- Stop if save behavior cannot prove preservation of unknown `extra` keys.
- Stop if implementation needs a path outside Allowed Paths.
- Do not repair unrelated frontend-suite drift.

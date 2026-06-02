### DONE: upstream-main-openai-endpoint-capability-ui-s2d

## Task ID
upstream-main-openai-endpoint-capability-ui-s2d

## Status
done

## Summary
- Ported the safe frontend subset of upstream `37044b83e fix(openai): clarify endpoint capability UI` onto `codex/upstream-main-openai-endpoint-capability-ui-s2d`.
- Added OpenAI API Key endpoint capability controls to create/edit account modals, backed by the existing `credentials.openai_capabilities` key.
- Clarified that Responses mode only applies to the text forwarding endpoint, and disabled that selector when text generation capability is disabled.
- Added modular English and Chinese i18n keys under `admin.accounts.openai`.
- Extended `EditAccountModal` tests for embeddings-only capability serialization, default capability omission, disabled Responses mode UI, and preservation of `openai_responses_supported`.

## Changed Files
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`
- `frontend/src/components/account/__tests__/EditAccountModal.spec.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `docs/workflow/main-log.md`
- `docs/workflow/tasks/upstream-main-openai-endpoint-capability-ui-s2d.md`
- `docs/workflow/worker-results/upstream-main-openai-endpoint-capability-ui-s2d-result.md`
- `docs/workflow/qa-reports/upstream-main-openai-endpoint-capability-ui-s2d-qa.md`

## Commands Run
```text
git status --short --branch -> clean before implementation on codex/upstream-main-openai-endpoint-capability-ui-s2d
git diff --check -> pass
corepack.cmd pnpm --dir frontend exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts -> pass, 1 file / 16 tests
corepack.cmd pnpm --dir frontend run typecheck -> pass
corepack.cmd pnpm --dir frontend run lint:check -> pass
```

## Scope Notes
- Upstream commit also edits monolithic `frontend/src/i18n/locales/en.ts` and `zh.ts`; local branch uses modular i18n, so the patch was manually ported.
- Backend capability routing and `openai_capabilities` support already exist locally; this Sprint did not change backend behavior.
- `CreateAccountModal` writes default capabilities by omission and writes a narrowed list only when OpenAI API Key capabilities are not both defaults.
- `EditAccountModal` hydrates capability state from `credentials.openai_capabilities` and preserves existing `extra.openai_responses_supported` probe metadata.

## Risks
- No browser screenshot or live runtime smoke was run for the modal UI.
- No dedicated `CreateAccountModal` Vitest was added; create-path behavior was covered by code review, shared helper behavior, `typecheck`, and `lint:check`.
- No real upstream credential smoke was run.

## Knowledge Candidates
- None.

## Contract Compliance
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

## Blocked Reason
- None.

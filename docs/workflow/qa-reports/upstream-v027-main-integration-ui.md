### PASS: upstream-v027-main-integration-ui

# UI Integration QA

## Scope

This report validates the fourteen applicable UI behaviors from the approved
`upstream-v027-ui` contract after their exact patch integration onto main base
`796003b8d`. The unavailable `0838e0e610` quota-editor item remains not
applicable: the approved baseline has no `UserPlatformQuotaModal.vue` or
equivalent administrator platform-quota editor, so no nonexistent spec was
included.

The worktree has the five UI/backend integration batches plus Anthropic
changes staged. This QA examined the complete staged UI patch and did not
modify code, Git state, dependencies, or the read-only `frontend/node_modules`
junction.

## Main Select Compatibility

The current main `Select.vue` keeps its observable selection sequence:
`update:modelValue(value)` is emitted before `change(value, option)`. Therefore
the status `v-model` is updated before `UserOrdersView` handles
`@change="handlePageChange(1)"`.

The integration test mounts the actual `Select`, `Pagination`, and
`UserOrdersView`. It first emits page 3 and observes an order request with
`page: 3`; it then changes status to `PENDING` and observes the final request
with `page: 1, status: 'PENDING'`. This verifies that the main Select behavior
does not regress the current Pagination/Orders adaptation. The Pagination
numeric-input spec also passed against the integrated source.

## Executed Checks

From `frontend`:

| Command | Result |
| --- | --- |
| Approved adjusted 14-file `pnpm.cmd exec vitest run ...` set | PASS, 14 files / 24 tests |
| `pnpm.cmd exec vue-tsc -b` | PASS, exit 0 |
| `pnpm.cmd exec vite build --outDir ../outputs/upstream-v027-dist` | PASS, exit 0; 1,922 modules transformed and 252 files emitted |

The 24 tests include the strengthened clipboard fallback exception paths,
real TOTP error-extractor coverage for `Error` and legacy Axios detail,
refund equal/over-balance boundaries, and the non-first-page order filter
regression.

From repository root:

| Command | Result |
| --- | --- |
| `git diff --cached --check` | PASS |
| `git diff HEAD --check` | PASS |
| staged and working-tree `git diff --name-only --diff-filter=U` | PASS, both empty |

## Browser Evidence

No browser process was started for this integration-only check, as directed.
The prior task-owned browser acceptance remains applicable to the unchanged UI
behaviors: real components covered dialog IDs, numeric pagination jump, amount
restoration, model Tab behavior, and TOTP retry digits. The only potentially
relevant main integration boundary, Select-to-Orders/Pagination event flow, is
covered by the actual mounted integration test above.

## Limits

The Vite build used the supplied read-only junction to the prior task's private
dependency sandbox. No install, deletion, lockfile change, browser action,
commit, push, deployment, real authentication, payment, or backend service
call occurred. Existing pnpm override, Browserslist age, and Vite chunk-size
messages were warnings only; all requested commands exited successfully.

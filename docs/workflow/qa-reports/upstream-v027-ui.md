### PASS: upstream-v027-ui

# Independent UI QA

## Scope and verdict

The approved UI contract is satisfied for the fourteen applicable upstream
behaviors. `0838e0e610` remains correctly **not applicable**: immutable base
`d6e6c34718e6eee6388391b346f40e1492e81d57` has neither
`UserPlatformQuotaModal.vue` nor an equivalent administrator platform-quota
editor. The amended contract correctly removes its nonexistent spec while
keeping all fourteen applicable specs mandatory.

Production diff review found the intended localized adaptations only:

| Behavior | QA conclusion |
| --- | --- |
| Subscription clear / payment config de-duplication | `clear()` clears loading; concurrent callers share the in-flight config promise. |
| Dialog IDs / pagination / model Tab | Module-level dialog counter produces distinct title IDs; numeric jump coercion works; empty Tab leaves focus while nonempty Tab commits the tag. |
| Refund / amount boundaries | Warning compares requested refund amount; balance equal to the requested amount is allowed and `50.01 > 50` warns. Invalid amount restores the current accepted DOM/state value. |
| TOTP | Digit inputs are state-bound and clear visibly after retry failure. Both a standard `Error` and legacy Axios `response.data.detail` pass through the real `extractApiErrorMessage` helper. |
| Clipboard / announcements / proxies / orders / registration | Fallback exceptions return `false`, show failure, preserve `copied=false`, and remove the textarea; partial bulk reads remain retained; overlapping proxy testing is deduplicated; status filtering resets from page 3 to page 1; cached disabled promo configuration hides the field. |

The first test review found inadequate regression coverage for clipboard
exceptions, status filtering from a non-first page, real TOTP error extraction,
and refund thresholds. The developer added coverage for all four before this
independent rerun. In particular, the old `useClipboard` spec did not test an
`execCommand` throw; the final suite tests both no-Clipboard and rejected
Clipboard-API fallback paths, including error toast and textarea cleanup.

## Independent checks

From `frontend`:

| Command | Result |
| --- | --- |
| Contract's adjusted 14-file `pnpm.cmd exec vitest run ...` set | PASS, 14 files / 24 tests |
| `pnpm.cmd exec vue-tsc -b` | PASS, exit 0 |

The exact Vitest set contains the fourteen contract specs for model tags,
refund dialog, base dialog, pagination, proxies, amount input, TOTP disable and
setup, clipboard, announcements, payment config, subscriptions, registration,
and user orders. The quota-editor spec is deliberately absent under the
approved not-applicable amendment.

From repository root:

| Command | Result |
| --- | --- |
| `git diff --check` | PASS |
| `git diff --name-only --diff-filter=U` | PASS, empty |

## Independent build evidence

`outputs/upstream-v027-ui/build-report.md` records an isolated private frontend
sandbox. It copied source/configuration without following the source
`node_modules` junction, used a private pnpm store, and verified final source
and sandbox SHA-256 inventories matched. The final frozen install, `vue-tsc -b`,
and Vite production build all exited 0; the build emitted 252 files after 1,922
modules transformed. Existing pnpm override, Browserslist age, and chunk-size
messages were warnings only.

## Browser evidence and cleanup

`outputs/upstream-v027-ui/runtime-report.md` records the task-owned mock-API
browser acceptance using real imported components. It verifies two distinct
dialog title IDs, numeric pagination jump, invalid amount restoration after
the pre-existing numeric normalization path, empty/nonempty model Tab behavior,
and visible TOTP digit clearing after a rejected enable attempt. Desktop and
390px screenshots are retained as `runtime-desktop.png` and
`runtime-mobile-390.png`.

The first Chrome attempt exited before attaching; the retry used the explicit
task profile with Edge and completed. The browser report records successful
`playwright-cli -s=v027-ui close`, exact termination of Vite PID `57324` whose
command line matched the task runtime config, and final
`ViteStillRunning=False` plus `RemainingTaskProfileProcesses=0`. Temporary
smoke HTML/TS, runtime config, dependency stub, and helper were removed.

## Limits

This is a local UI acceptance result with mocked browser API routes. It does
not exercise real authentication, payment, backend, deployment, or remote
services. No commit, push, dependency/lockfile change, or production data write
occurred.

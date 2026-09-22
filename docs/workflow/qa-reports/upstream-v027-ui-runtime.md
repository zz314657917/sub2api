### PASS: upstream-v027-ui

# Independent browser runtime acceptance

## Scope and isolation

- Harness imported the real `BaseDialog`, `Pagination`, `AmountInput`, `ModelTagInput`, and `TotpSetupModal` components from `frontend/src`.
- `frontend/v027-ui-smoke.html` and `frontend/v027-ui-smoke.ts` were temporary task-owned files and are removed after this report.
- TOTP requests were intercepted by a Playwright route only for `**/api/v1/user/totp/**`. The observed calls were to local Vite at `127.0.0.1:4328`: verification method (twice), setup, and enable. No real authentication or payment request was sent.
- The repository's read-only dependency tree is missing `@airwallex/airtracker`, which prevents Vite from traversing the API barrel when importing the real TOTP component. The temporary `outputs/upstream-v027-ui/vite.runtime.config.mjs` mapped that module to `outputs/upstream-v027-ui/airtracker-stub.ts`. This was limited to the browser smoke server, did not mock TOTP or any accepted component behavior, and does **not** establish that a production build passes.

## Browser startup

- Attempt 1: `playwright-cli -s=v027-ui open ... --persistent --profile E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/profile --headed` selected Chrome 103. Chrome launched with the explicit task profile then exited with code 9 before the debugging pipe connected; the CLI daemon exited with code 1.
- Retry: `playwright-cli -s=v027-ui open ... --browser msedge --persistent --profile E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/profile --idle-timeout 0` succeeded with Edge 139.0.3405.125. The CLI session reported the same explicit task-owned profile.

## Runtime scenarios

| Contract scenario | Browser evidence | Verdict |
| --- | --- | --- |
| Simultaneous dialog title IDs | Two real dialogs exposed `aria-labelledby` values `modal-title-1` and `modal-title-2`, resolving to `First runtime dialog` and `Second runtime dialog`. | PASS |
| Pagination numeric jump | Filled numeric page `5`, clicked `跳转`, and the mounted parent output changed from `1` to `5`. | PASS |
| Invalid amount correction | With real parent `v-model`, `12.50` first normalized through the pre-existing `watch(modelValue)` to accepted rendered state `12.5`; entering invalid `12.500` restored the input to `12.5` without a new accepted value. This confirms the requested restoration of accepted DOM/state. | PASS |
| Model empty/non-empty Tab | An empty model input allowed focus to leave the input. After entering `gpt-test`, Tab committed the tag and the mounted parent output became `gpt-test`. | PASS |
| TOTP retry digit synchronization | Mocked real-component flow: password verification -> mocked setup -> verify step -> six digits -> mocked rejected `/enable`. After failure the six visible digit inputs evaluated to `['', '', '', '', '', '']`. | PASS |

## Visual evidence

- Desktop 1440x1000: `outputs/upstream-v027-ui/runtime-desktop.png`
- Mobile 390x844: `outputs/upstream-v027-ui/runtime-mobile-390.png`
- Visual inspection: desktop preserves the common pagination jump controls; at 390px the real responsive pagination switches to previous/page/next without text overlap, and amount/model components remain within the viewport.

## Process cleanup plan

- `playwright-cli -s=v027-ui close`: exit 0 (`Browser 'v027-ui' closed`).
- Vite PID `57324` was read from the task-owned PID file and its command line matched `vite.runtime.config.mjs`; it was stopped directly. A final exact command-line check found `ViteStillRunning=False` and `RemainingTaskProfileProcesses=0` for the task profile / `cliDaemon.js v027-ui`.
- The smoke HTML/TS, the temporary Vite config, the Airtracker stub, and the run-code helper were removed. Screenshots, logs, PID records, and this report remain under the task-owned output directory.

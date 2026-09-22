### PASS: upstream-v027-ui private frontend build

## Scope

This report covers only the task-private dependency installation, TypeScript
project build, and Vite production build requested for the independent UI QA
environment. It does not replace the contract's focused Vitest or browser
acceptance evidence.

- Worktree: `E:/codex-worktrees/sub2api/upstream-v027-port`
- Sandbox: `E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/build-sandbox`
- Private pnpm store: `E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/pnpm-store`
- Production output: `E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/build-sandbox/dist`

## Isolation

The sandbox was created by copying only `frontend/src`, `frontend/public`, and
the required root frontend configuration files (`package.json`, `pnpm-lock.yaml`,
`.npmrc`, `index.html`, PostCSS/Tailwind/Vite/TypeScript configuration). The
source `frontend/node_modules` junction was not copied or followed. `outputs`
was not copied. The sandbox has its own installed `node_modules`; no install,
manifest, lockfile, or junction change was made under `frontend`.

At final verification:

- Source `src` and sandbox `src` file SHA-256 inventories matched exactly.
- Source and sandbox `package.json` SHA-256 matched:
  `735918D8B1F6D7C31C5169AE3E8863F0168677D1C23807002713C69B57512729`.
- Source and sandbox `pnpm-lock.yaml` SHA-256 matched:
  `8B545157E34CC0DDC1866A43B7147326B91549879EE6C3360F094DB300CE135E`.

The source changed twice while the worker was adding tests. The changed sandbox
files were refreshed before the final build, including `useClipboard.spec.ts`,
`AdminRefundDialog.balance.spec.ts`, `Totp.errors.spec.ts`, and
`UserOrdersView.filters.spec.ts`.

## Commands And Results

| Command | Working directory | Exit code | Result |
| --- | --- | ---: | --- |
| `pnpm.cmd install --frozen-lockfile --ignore-scripts --store-dir E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/pnpm-store` | sandbox | 0 | Installed 984 locked packages in the private sandbox. The Airwallex package resolved without a stub or externalization. |
| `pnpm.cmd exec vue-tsc -b` | sandbox | 0 | Final refreshed source typecheck passed. |
| `pnpm.cmd exec vite build --outDir E:/codex-worktrees/sub2api/upstream-v027-port/outputs/upstream-v027-ui/build-sandbox/dist` | sandbox | 0 | Final refreshed production build passed; 1,922 modules transformed and 252 files were emitted. |

Logs are retained in the sandbox as `pnpm-install.log`, `vue-tsc-final.log`, and
`vite-build-final.log`.

## Warnings

- pnpm 11 warns that the `package.json` `pnpm.overrides` field is no longer
  read. Frozen install still passed without modifying configuration.
- Vite reported existing dynamic/static import chunking warnings and chunks over
  500 kB. These are warnings only; the command exited 0.
- Browserslist data is reported as nine months old. No dependency refresh was
  performed because the task requires frozen lockfile validation.

## Limits

No focused Vitest command or browser session was run by this build-only task.
No commit, push, deployment, or repository business/contract/Git modification
was made.

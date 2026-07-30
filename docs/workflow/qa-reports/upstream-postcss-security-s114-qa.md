### PASS: upstream-postcss-security-s114

## Findings

- `frontend/pnpm-lock.yaml` resolves PostCSS to `8.5.23`; no PostCSS advisory
  appears in the fresh production audit output.
- The audit still reports two pre-existing high `xlsx` advisories. The existing
  `.github/audit-exceptions.yml` entries expired on `2026-07-06`, so the audit
  command and exception checker remain non-zero for that baseline.
- No frontend source, backend, database, deployment, container, knowledge, or
  unrelated worktree paths changed in S114.

## Executed Checks

- `corepack.cmd pnpm audit --prod --audit-level=high --json`: exit `1`; high
  findings were only `xlsx` (`GHSA-4r6h-8v6p-xvw6`, `GHSA-5pgg-2g8v-p4x9`).
- `python tools/check_pnpm_audit_exceptions.py --audit frontend/audit-s114-stdout.json --exceptions .github/audit-exceptions.yml`: exit `1` because both existing `xlsx` exceptions are expired.
- `corepack.cmd pnpm exec vue-tsc --noEmit`: PASS.
- `corepack.cmd pnpm run build`: PASS, 1091 modules.
- Lockfile inspection: PostCSS `8.5.23`; no `postcss@8.5.6` remains.
- `git diff --check`: PASS.
- Exact path audit: only the S114 manifest, lockfile, contract, status, main-log,
  and QA report are changed.

## Unverified Risks

- The expired `xlsx` exception policy is outside S114 and remains a separate
  security-maintenance task.
- No production deployment or authenticated browser smoke was run.

## Recommendation

`PASS / source-only`; integrate the PostCSS patch, but keep the expired `xlsx`
exception policy visible as a separate follow-up rather than silently extending
its deadline.

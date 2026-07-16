### DONE: upstream-main-compat-s77-frontend

## Task ID

upstream-main-compat-s77-frontend

## Status

done

## Summary

Removed the mobile-only `overflow-visible` override from TablePageLayout so
the existing `overflow-x-auto` table wrapper remains active on tablets and narrow
layouts. Added a source-level regression test for the responsive rule.

## Changed Files

- frontend/src/components/layout/TablePageLayout.vue
- frontend/src/components/layout/__tests__/TablePageLayout.spec.ts

## Commands Run

````text`
corepack.cmd pnpm --dir frontend exec vitest run src/components/layout/__tests__/TablePageLayout.spec.ts -> PASS (1 test)
corepack.cmd pnpm --dir frontend run typecheck -> PASS
corepack.cmd pnpm --dir frontend run build -> PASS
`````

## Risks

- Build retained existing Browserslist, chunk-size, dynamic-import, and Node DEP0190 warnings.
- No browser session smoke was run; this slice is guarded by the component source test and production build.

## Contract Compliance

- allowed_paths_only: yes
- denied_paths_touched: no
- success_criteria_met: yes
- stop_rules_triggered: no

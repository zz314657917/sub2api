# Task Contract: upstream-frontend-hardcoded-i18n-s103

- Task ID: `upstream-frontend-hardcoded-i18n-s103`
- Role: Planner / Generator / Evaluator
- Goal: Port the applicable portion of upstream `3401a971a` so the remaining
  user-visible frontend strings in the selected layout and custom-page paths
  use the existing `vue-i18n` message catalog.
- Success Criteria:
  - `AppHeader` uses localized `common.toggleMenu` and `common.userMenu`
    aria-labels in both supported locales.
  - `CustomPageView` renders localized `common.pageNotFound` on a non-OK page
    response without changing the fetch or HTML rendering flow.
  - The upstream payment custom-method placeholder is explicitly recorded as
    `skipped`: the local `PaymentProviderDialog` has no custom-method block or
    `customMethodDisplayName` field, so there is no safe local consumer.
  - English and Chinese common locale files contain the three applicable new
    message keys with appropriate translations and no duplicate object keys.
  - Existing component behavior, API calls, payment payloads, accessibility
    semantics, and responsive layout remain unchanged apart from the message
    lookup.
- Allowed Paths:
  - `frontend/src/components/layout/AppHeader.vue`
  - `frontend/src/i18n/locales/en/common.ts`
  - `frontend/src/i18n/locales/zh/common.ts`
  - `frontend/src/views/user/CustomPageView.vue`
  - `docs/workflow/tasks/upstream-frontend-hardcoded-i18n-s103.md`
  - `docs/workflow/qa-reports/upstream-frontend-hardcoded-i18n-s103-qa.md`
  - `docs/workflow/status.md`
  - `docs/workflow/main-log.md`
  - `knowledge/tasks/current-task.md`
- Denied Paths: backend, API contracts, payment execution, billing,
  persistence, migrations, generated files, unrelated frontend components,
  `frontend/src/i18n` loaders, deployment, containers, `VERSION`, and
  `knowledge/**` other than the current-task snapshot.
- Constraints:
  - Adapt the upstream patch manually to the current local context; do not
    merge or cherry-pick the upstream commit.
  - Reuse the existing `useI18n()` instances and locale object style.
  - Keep the exact existing English/Chinese message shape and do not replace
    unrelated hardcoded text in these files.
  - Do not add dependencies, schema changes, or runtime fallback behavior.
- Acceptance Commands:
  - `corepack.cmd pnpm --dir frontend exec vitest run src/components/layout/__tests__/AppHeader.spec.ts src/components/payment/__tests__/PaymentProviderDialog.spec.ts`
  - `corepack.cmd pnpm --dir frontend run typecheck`
  - `corepack.cmd pnpm --dir frontend run build`
  - Static checks confirm the three applicable targeted literals are no longer
    present in executable/template text, the missing payment prerequisite is
    documented, and all four business/locale files are in the exact allowlist.
  - `git diff --check`, conflict-marker scan, and unmerged-index check.
- Output: Scoped localization diff, focused payment-dialog regression,
  production type/build evidence, QA report, and final `PASS`, `FAIL`, or
  `BLOCKED` evidence.
- Stop Rules: Stop if the change requires modifying API/backend behavior,
  adding a locale loader or dependency, changing payment semantics, touching
  denied paths, or translating unrelated strings.

## Contract Review

`PASS / partial-port`: Upstream `3401a971a` is a presentation-only localization
patch. The local layout and custom-page components already expose `t`, and the
three common message locations fit the existing locale object structure. The
payment placeholder is intentionally skipped because its local UI prerequisite
is absent; adding it would silently expand this Sprint into a different
feature.

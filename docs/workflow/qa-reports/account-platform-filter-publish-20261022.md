# Account platform filter correction

Date: 2026-10-22.

## Finding and change

Account creation, editing and badges already support DeepSeek, Kimi and Zhipu.
The account-list dropdown and persisted-filter allowlist still contained only
the five older platforms. Add the three missing platforms to both lists.

This filters the stored account platform. An OpenAI relay account serving a
DeepSeek model remains an OpenAI account; model names do not rewrite account
classification.

The change touches only AccountTableFilters.vue, AccountsView.vue and the
existing AccountsView.filters.v0217.spec.ts test suite. Six new cases exercise
real filter events, request parameters, persistence, main/shared scope changes,
and remount restoration for each missing platform.

## Executed checks

Verification ran against an isolated exact publication snapshot, excluding
unrelated dirty source files:

- Eleven relevant frontend suites: 120 tests passed, including all 23 account
  filter cases.
- vue-tsc typecheck: passed.
- ESLint on the three changed frontend paths: passed.
- Canonical production frontend build: passed.
- Server build with the resulting frontend embedded: passed.
- Selected diff whitespace check: passed.
- Independent static review: no blocking issue; dropdown and saved-filter
  enumerations match, and existing assertions are preserved.

Evidence directory:
`E:/codex-runtime/sub2api-publish-20261022`.
The 12,158 captured original working-tree files were checked after verification:
only the three expected frontend files changed in this task.

## Limits

Component tests use mocked account APIs. No browser, deployed page, real account
data or running service acceptance is claimed. Deployment is a separate action.

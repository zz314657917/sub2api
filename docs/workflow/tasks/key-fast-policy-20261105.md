---
type: task-contract
status: approved
review_verdict: PASS
task_id: key-fast-policy-20261105
worker_model: gpt-6.1-sol
qa_worker_model: gpt-6.1-sol
base_commit: 0f6b739881ccf9e4ad423f464aeedcc9abd7f48a
spec_ref: docs/workflow/spec.md
last_verified: 2026-11-05
---

# Task Contract

## Task ID
key-fast-policy-20261105

## Role
Controller is Planner, integration owner and Final Evaluator. Follow the repository
Agent Matrix for independent Contract Reviewer, backend Developer, frontend Developer
and independent QA. Approval is effective only after an independent PASS records the
SHA-256 of these final contract bytes.

## Goal
Users can choose an OpenAI Fast policy for their own API key: passthrough (default),
force priority, or filter client-selected Fast. Keep administrator group enforcement
and global pass/filter/block authoritative, and preserve existing billing.

## Context
- Main: F:/mcplugins/sub2api.
- Candidate: E:/codex-worktrees/sub2api/key-fast-policy-20261105.
- Evidence: E:/codex-runtime/sub2api/key-fast-policy-20261105.
- Main's 4928 tracked files, 16 dirty paths and six untracked backend Go prerequisite
  files are frozen in before.json. Candidate includes their exact bytes. The original
  baseline build without those untracked files failed and remains as evidence;
  the corrected complete baseline is separate. Never revert or stage those inputs.
- This is implementation only. No new commit, push, container/service update, live
  provider, production database, credentials, or execution of production migrations.

## Success Criteria
1. Add persisted string openai_fast_policy with the exact values passthrough,
   force, filter. Default and old/empty records behave as passthrough. Additive
   SQL migration uses the next unused number (255, recheck before writing);
   only isolated test databases may execute it. No dependency changes.
2. Carry policy through own-key create/update/get/list DTOs, repository
   hydration/writes, minimal update fields, auth snapshots, trusted request context
   and cache invalidation. Create omissions default; update omission preserves.
   Reject unknown/non-string values with the established 400 validation behavior.
   Preserve owner checks, quota/rate-limit usage and managed Cafe key restrictions.
   Increment auth snapshot version so old cached keys reload with current policy.
3. Apply key policy only for actual OpenAI upstream accounts. Force adds or
   replaces service_tier with priority, even if missing/default/flex/ultrafast/
   unknown. Filter removes only normalized fast/priority, leaving other tiers
   unchanged. Passthrough preserves current request behavior. Read the preference
   only from server-authenticated key context; body/header spoofing cannot set it.
4. Priority: global policy is final; administrator group force overrides a key's
   filter; otherwise key policy controls the client tier. All existing group
   force/default behavior is unchanged. Global filter/block still applies to the
   resulting tier. Non-OpenAI accounts are unaffected by the new key setting.
5. Cover native Responses, OAuth passthrough, Chat Completions, Messages, normal WS
   and passthrough WS response.create including follow-up frames with omitted tier.
   Other WS events are byte-unchanged. Outbound tier, routing hint and recorded
   billing metadata reflect the final policy output. Preserve replay/error/cancel
   and billing ownership; do not change price cards, multipliers or historical bills.
   Explicitly cover raw Chat and Responses-to-raw-Chat streaming/nonstreaming:
   derive returned metadata from the final policy output, including replacement
   of an originally non-nil flex/priority value and removal by key/global filter.
6. Add a labelled three-option selector to the existing user key create/edit form.
   Chinese labels: 透传（默认）, 强制启用, 过滤. Explain Fast means priority,
   only OpenAI requests are affected, group/global policies take precedence,
   priority uses the existing priority prices, and WS sessions must reconnect
   after changes. Persist and load correctly; form reset returns to passthrough.
   Include matching English locale text and scoped mounted/serialization tests.
7. Backend tests demonstrate create/update validation, omitted updates, ownership
   and managed-key restrictions, repository and cache round trips and invalidation,
   trusted context, key x group x global precedence, HTTP ingress and both WS
   transports/follow-ups. Run relevant existing group/key/global/billing regressions.
   Frontend tests exercise actual create/save/edit/reset and API serialization,
   not only source-text matching. Run typecheck, scoped lint and production build.
8. Independent QA validates a frozen candidate and produces a first-line PASS/
   FAIL/BLOCKED report. Test failures from the baseline must be reproduced and
   separated, never weakened away. Controller verifies exact delivered hashes
   and preserves every non-task main file, HEAD and logical index.

## Allowed Paths
Backend Developer owns:
- backend/ent/schema/api_key.go
- backend/ent/apikey.go
- backend/ent/apikey/apikey.go
- backend/ent/apikey/where.go
- backend/ent/apikey_create.go
- backend/ent/apikey_update.go
- backend/ent/mutation.go
- backend/ent/migrate/schema.go
- backend/ent/runtime/runtime.go
- backend/migrations/255_api_key_openai_fast_policy.sql
- backend/internal/pkg/ctxkey/ctxkey.go
- backend/internal/handler/api_key_handler.go
- backend/internal/handler/dto/types.go
- backend/internal/handler/dto/mappers.go
- backend/internal/repository/api_key_repo.go
- backend/internal/service/api_key.go
- backend/internal/service/api_key_fast_policy.go
- backend/internal/service/api_key_service.go
- backend/internal/service/api_key_auth_cache.go
- backend/internal/service/api_key_auth_cache_impl.go
- backend/internal/service/openai_gateway_service.go
- backend/internal/service/openai_gateway_chat_completions_raw.go
- backend/internal/service/openai_gateway_responses_chat_fallback.go
- backend/internal/service/openai_policy_group_fast_20261103_test.go
- backend/internal/server/middleware/api_key_auth.go
- New *_key_fast_20261105_test.go files under backend/internal/service,
  backend/internal/repository, backend/internal/handler, backend/internal/handler/dto,
  backend/internal/server/middleware.
- docs/workflow/worker-results/key-fast-policy-20261105-backend.md

Frontend Developer owns:
- frontend/src/types/index.ts
- frontend/src/api/keys.ts
- frontend/src/views/user/KeysView.vue
- frontend/src/i18n/locales/zh/keys.ts
- frontend/src/i18n/locales/en/keys.ts
- New keyFastPolicy.spec.ts tests under frontend/src/api/__tests__ and
  frontend/src/views/user/__tests__.
- docs/workflow/worker-results/key-fast-policy-20261105-frontend.md

Controller owns task-specific workflow/QA evidence and isolated runtime fixtures.
Contract Reviewer owns docs/workflow/contract-reviews/key-fast-policy-20261105-review.md.
QA owns docs/workflow/qa-reports/key-fast-policy-20261105.md and external evidence.
Adding tests under a different descriptive name or an additional product owner
requires controller review before changing the approved contract.

## Denied Paths
- All product paths not explicitly allowed above.
- Existing dirty backend Go module/dependency, billing, Gemini and upstream fixes.
- Existing knowledge/**, shared docs/workflow/agent-matrix.md and main-log.md.
- Global memories, skills, templates, local credentials, production configuration.
- Price cards, historical accounting, provider integrations and live databases.

## Constraints
- Keep the two Developer write sets disjoint. Work only in the candidate.
- Do not broadly format, regenerate unrelated ent code, install dependencies,
  relax baseline tests, or change existing billing/security/managed-key ownership.
- No secret output, live API probes or broad process termination.
- The two additional raw Chat/fallback owners may change only to derive tier
  metadata from the final policy-transformed body. No billing or protocol rewrite.
- The existing group Fast test may change only its authentication snapshot version
  assertion from 20 to 21, matching the required cache-version bump. Preserve every
  group force, duplicate and permission assertion; do not weaken any regression.
- Use existing Go 1.26.10 toolchain/cache and candidate code; builds go outside main.
- Browser QA uses isolated profile/session with recorded exact PID ownership, then
  closes and verifies cleanup. Use loopback demo CRUD fixtures, not live accounts.

## Acceptance Commands
From candidate backend, Go 1.26.10:
- go test ./internal/service ./internal/handler ./internal/handler/dto
  ./internal/repository ./internal/server/middleware -run KeyFast20261105 -count=1
- Repeat the new relevant tests with -tags unit; use actual test names and report
  executed counts so no-tests-to-run is never counted as a pass.
- Relevant existing APIKey/group Fast/global policy/WS/billing targeted suites;
  preserve and reproduce any known baseline failure.
- go build -o E:/codex-runtime/sub2api/key-fast-policy-20261105/server.exe ./cmd/server
- gofmt check on changed Go files and scoped git diff --check.
From candidate frontend using existing installed dependencies:
- vitest run src/views/user/__tests__ src/api/__tests__/keyFastPolicy.spec.ts
  src/views/admin/__tests__/GroupsView.openaiFast.spec.ts
  src/views/admin/__tests__/groupsOpenAIFast.spec.ts
- vue-tsc --noEmit
- eslint on changed frontend sources/tests
- vite build --outDir E:/codex-runtime/sub2api/key-fast-policy-20261105/dist
- Actual built loopback page smoke for create/edit policy persistence and
  desktop/mobile light/dark accessibility/overflow plus browser cleanup.

## Output
Developer reports start ### DONE: key-fast-policy-20261105 or FAILED/BLOCKED and list
changed files, exact commands and counts, evidence, risks and knowledge_candidates.
QA report starts ### PASS: key-fast-policy-20261105 or FAIL/BLOCKED and binds the
frozen source/input manifest and evidence paths. No self-approval by Developers.

## Stop Rules
Stop and report exact findings if out-of-scope paths or owner conflicts are needed,
baseline inputs drift, a matrix model is unavailable, test credentials are missing
for required live operations, or an existing protection cannot be preserved.
Never execute real migration/deployment to satisfy a missing test.

# Task Contract: Upstream Google One Model Catalog S248

## Task ID

`upstream-google-one-model-catalog-s248`

## Role

- Planner / Final Evaluator: Codex Controller
- Implementation owner after repeated non-terminal worker returns: Codex Controller
- Independent QA Worker: `gpt-5.6-terra`

## Goal

Behaviorally adapt upstream source
`f98a056f75e93c81e3ab1cc8623db9e17b3dc432` from final merge
`844b1187855256eb45ef2f9bec8c4657ec450d63` so legacy Gemini Google One OAuth
accounts advertise and default-map only the conservative supported catalog:

- `gemini-2.0-flash`
- `gemini-2.5-flash`
- `gemini-2.5-pro`

Frozen product baseline: local `main@118ff2596`. Controller workflow commits
above that baseline must not alter product behavior.

## Success Criteria

1. A Gemini OAuth account with `oauth_type=google_one` is identified explicitly
   without changing Code Assist, API-key Gemini, Antigravity, or other account
   classification.
2. `geminicli.GoogleOneModels` contains exactly the three conservative models;
   image and 3.x models are absent. Mapping callers receive a defensive copy.
3. Admin available-models returns the conservative catalog for Google One OAuth
   while other Gemini OAuth accounts retain `DefaultModels`.
4. Google One with no explicit `model_mapping` gets the conservative identity
   mapping, including when the parsed raw mapping is empty.
5. A non-empty explicit Google One mapping remains authoritative and is not
   merged with or replaced by defaults.
6. Unsupported default models are not eligible through
   `IsModelSupported`; explicit mappings continue to work.
7. Business/evidence commits obey exact allowlists; all acceptance and
   protected-primary gates pass; independent Terra QA remains mandatory.

## Allowed Paths

Developer business commit:

- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/handler/admin/account_handler_available_models_test.go`
- `backend/internal/pkg/geminicli/models.go`
- `backend/internal/pkg/geminicli/models_test.go`
- `backend/internal/service/account.go`
- `backend/internal/service/account_google_one_s248_test.go`

Developer evidence commit only:

- `docs/workflow/worker-results/upstream-google-one-model-catalog-s248-result.md`

Independent QA evidence commit only:

- `docs/workflow/qa-reports/upstream-google-one-model-catalog-s248-qa.md`

## Denied Paths

- `backend/internal/service/account_wildcard_test.go`; the upstream owner is
  locally `//go:build unit`, and the repository's unrelated unit-tag suite has
  existing compile errors. Use the allowed self-contained default-tag test.
- All other backend, frontend, schema, migration, dependency, generated,
  Docker, deployment, knowledge, and workflow files except the active report.
- All twenty-two tracked and five untracked user-owned primary-worktree paths.
- Provider traffic, shared/production data, containers, browser automation,
  push, force operations, and history rewrites.

## Constraints

- Keep local account mapping cache/signature behavior; adapt only the Google One
  default branches and do not port unrelated upstream account refactors.
- Explicit non-empty mapping wins. Do not auto-merge conservative defaults into
  an explicit mapping.
- Return a fresh map from the package helper; do not expose mutable shared map
  state.
- Do not rename existing default models, change Antigravity/Gemini API-key/Code
  Assist behavior, add models, or alter routing, billing, scheduling, auth, or
  account persistence.
- The three named focused tests are S248 deliverables; baseline absence is
  expected. Add them before discovery.
- Do not install/update dependencies, call external services, or stage/format
  unrelated work.

## Acceptance Commands

From `backend/` in the isolated worktree:

```powershell
go test ./internal/pkg/geminicli -list '^TestGoogleOneModels_ExcludeUnsupportedNewAndImageModels$'
go test ./internal/pkg/geminicli -run '^TestGoogleOneModels_ExcludeUnsupportedNewAndImageModels$' -count=10
go test ./internal/handler/admin -list '^TestAccountHandlerGetAvailableModels_GeminiGoogleOneUsesConservativeCatalog$'
go test ./internal/handler/admin -run '^TestAccountHandlerGetAvailableModels_GeminiGoogleOneUsesConservativeCatalog$' -count=10
go test ./internal/service -list '^TestAccountGetModelMapping_GoogleOne(UsesConservativeDefaults|PreservesExplicitMapping)$'
go test ./internal/service -run '^TestAccountGetModelMapping_GoogleOne(UsesConservativeDefaults|PreservesExplicitMapping)$' -count=10
go test ./internal/pkg/geminicli -count=1
go test ./internal/handler/admin -count=1
go test ./internal/service -count=1
go test ./cmd/server -run '^$' -count=1
gofmt -l internal/handler/admin/account_handler.go internal/handler/admin/account_handler_available_models_test.go internal/pkg/geminicli/models.go internal/pkg/geminicli/models_test.go internal/service/account.go internal/service/account_google_one_s248_test.go
```

From the worktree root:

```powershell
git diff --check
git diff --cached --name-only
git ls-files -u
git merge-base --is-ancestor f98a056f75e93c81e3ab1cc8623db9e17b3dc432 upstream/main
git merge-base --is-ancestor 844b1187855256eb45ef2f9bec8c4657ec450d63 upstream/main
git log --oneline 844b1187855256eb45ef2f9bec8c4657ec450d63..upstream/main -- backend/internal/handler/admin/account_handler.go backend/internal/handler/admin/account_handler_available_models_test.go backend/internal/pkg/geminicli/models.go backend/internal/pkg/geminicli/models_test.go backend/internal/service/account.go backend/internal/service/account_wildcard_test.go
rg -n '^(<<<<<<< .+|=======$|>>>>>>> .+)$' backend/internal/handler/admin/account_handler.go backend/internal/handler/admin/account_handler_available_models_test.go backend/internal/pkg/geminicli/models.go backend/internal/pkg/geminicli/models_test.go backend/internal/service/account.go backend/internal/service/account_google_one_s248_test.go
```

Controller additionally verifies final-merge six-owner scope, the one-for-one
local service-test substitution, no later upstream touches, exact commit
allowlists, account mapping cache preservation, empty index/conflict state, and
the primary protected snapshot.

The twenty-two tracked user paths are:

- `backend/internal/handler/admin/cafe_room_handler.go`
- `backend/internal/handler/admin/cafe_room_handler_test.go`
- `backend/internal/repository/cafe_room_repo.go`
- `backend/internal/server/routes/admin.go`
- `backend/internal/service/cafe_public.go`
- `backend/internal/service/cafe_public_test.go`
- `backend/internal/service/cafe_room_service.go`
- `backend/internal/service/cafe_room_service_test.go`
- `frontend/src/api/admin/cafeRooms.ts`
- `frontend/src/features/pixelCafe/PixelCafePage.vue`
- `frontend/src/features/pixelCafe/__tests__/PixelCafePage.spec.ts`
- `frontend/src/features/pixelCafe/components/CafeScene.vue`
- `frontend/src/features/pixelCafe/components/SceneFallback.vue`
- `frontend/src/features/pixelCafe/components/__tests__/CafeScene.spec.ts`
- `frontend/src/features/pixelCafe/renderer/assetManifest.ts`
- `frontend/src/features/pixelCafe/renderer/createCafeRenderer.ts`
- `frontend/src/features/pixelCafe/renderer/sceneLayout.ts`
- `frontend/src/i18n/locales/en/admin/pixelCafe.ts`
- `frontend/src/i18n/locales/zh/admin/pixelCafe.ts`
- `frontend/src/types/pixelCafe.ts`
- `frontend/src/views/admin/pixelCafe/AdminCafeRoomsView.vue`
- `frontend/src/views/admin/pixelCafe/__tests__/AdminCafeRoomsView.spec.ts`

Their stable combined patch ID must remain
`941b1edf9df9e465a6100007edfc4a6715e38b5e`. The five untracked SHA-256 values
must remain:

- `e6cd621c9f2df7b5d4a5521e8904c95731996533761e01add8ba544b014e0952`
  `backend/internal/repository/cafe_room_account_option_test.go`
- `1e3830c11e13b586f09c254c1a468878a84f932a8615be58fb479cfd607d66ff`
  `frontend/src/views/admin/pixelCafe/components/CafeRoomAccountPicker.vue`
- `49ec0eaadeb4d49f0eb01853629769be601e8896c5eb3ee2d5ae98db83717c32`
  `frontend/src/views/admin/pixelCafe/components/__tests__/CafeRoomAccountPicker.spec.ts`
- `f21e77c5d3cc82727a516bc4b2cb901e53c2d7505a448d5dd551b74ddfb3ece0`
  `outputs/20260725-static-residential-socks5/静态住宅 IP (1)-sub2-socks5.json`
- `438fdda26586fa3a5857b927d7dbbfac4868bb55a6a1e8bfdac540296a497f4c`
  `outputs/20260731-static-residential-sub2/静态住宅 IP (3)-sub2api.json`

Primary staged/unmerged indexes must remain empty.

## Output

- Controller makes one business commit containing only the six local owners and
  one evidence commit containing only
  `docs/workflow/worker-results/upstream-google-one-model-catalog-s248-result.md`.
- Controller result first line: exactly
  `### DONE: upstream-google-one-model-catalog-s248`,
  `### BLOCKED: upstream-google-one-model-catalog-s248`, or
  `### FAILED: upstream-google-one-model-catalog-s248`.
- QA modifies only
  `docs/workflow/qa-reports/upstream-google-one-model-catalog-s248-qa.md`; first
  line is exactly `### PASS: upstream-google-one-model-catalog-s248`,
  `### FAIL: upstream-google-one-model-catalog-s248`, or
  `### BLOCKED: upstream-google-one-model-catalog-s248`.
- Reports include changed files, commands, key output, risks, compliance, and
  `knowledge_candidates`.

## Stop Rules

- Stop if `gpt-5.6-terra` is unavailable; do not replace the model.
- Stop if implementation requires the unit-tag owner, paths outside the
  allowlist, account cache/routing/persistence redesign, new models,
  dependencies, frontend/schema work, or external/shared state.
- Stop if, after adding the three required focused owners, fewer than four total
  tests are discovered, an outside-scope baseline fails, or protected-primary
  state changes unexpectedly.

## Budget

- worker_mode: stopped after four incomplete non-terminal returns; Controller takeover
- qa_worker_mode: native `gpt-5.6-terra`
- worker_model: `gpt-5.6-terra`
- qa_worker_model: `gpt-5.6-terra`
- developer_max_budget_usd: closed with zero worker commits
- qa_max_budget_usd: `0.10`
- worktree_root: `E:/codex-worktrees`

## Status

`contract-approved`

## Worker Output

Same requirements as `Output`.

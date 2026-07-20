### PASS: upstream-main-compat-s87

## Findings

- No P1/P2 backend behavior finding remains in the implemented S87 slices.
- API-key partial updates are covered through the real handler HTTP path and a
  direct service matrix, including omitted/null preservation, independent list
  clearing/replacement, invalid-input rejection, and no repository write on
  validation failure.
- Quota classification is Responses-only with exact root/subpath matching;
  `/v1/responsesx`, Chat Completions, Images, Messages, Usage, and Gemini paths
  do not receive the OpenAI quota envelope.
- `a05b87321` is correctly deferred. The local repository lacks the required
  `SubscriptionPlan.currency` schema/API chain, and the upstream migration
  number conflicts with the local migration sequence.

## Executed Checks

- Handler S87 test: PASS.
- Service S87 test: PASS.
- Middleware S87 quota envelope/path matrix: PASS.
- S85 `TestHandleFailoverError_` regression selection: PASS.
- S87 default-tag test discovery: PASS, all four required tests found.
- Frontend focused Vitest: PASS, 1 file / 2 tests (direct pnpm-store entrypoint).
- `gofmt`, `git diff --check`, exact allowlist, conflict-marker/unmerged-index,
  and S85 path-freeze gates: PASS.
- `npm.cmd run test:run -- src/components/channels/__tests__/AvailableChannelsTable.spec.ts`:
  PASS, 1 file / 2 tests.
- `npm.cmd run typecheck`: PASS.
- `npm.cmd run build`: PASS, 1088 modules transformed and production assets
  emitted to the existing backend web dist output.
- The frontend dependency tree was restored in the isolated worktree with
  pnpm `10.33.4` using `--frozen-lockfile`; no package manifest or lockfile
  changed. The previous broken `@airwallex/components-sdk` junction was a
  dependency-tree condition, not an S87 source failure.
- Final `git diff --check`, exact allowlist, unmerged-index/conflict scan,
  S87 test discovery, and S85 path-freeze gates: PASS.

## Unverified Risks

- No live OpenAI, Anthropic, Gemini, or authenticated browser request was run.
- The quota test exercises the middleware helper and path matrix directly;
  full API-key middleware integration remains covered only by the existing
  broader suite outside this focused default-tag contract.

## Recommendation

`PASS` — S87 is ready for scoped commit/review. Keep deployment, container
updates, and push as separate explicitly authorized actions. Live upstream
requests and authenticated browser smoke remain unverified.

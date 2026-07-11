### PASS: upstream-codex-image-tool-strip-policy-s68a

# QA Report

## Scope

- Integration head: `b863b2279`; audit baseline: `5a9ae78d8`.
- Contracts reviewed: `upstream-codex-image-tool-strip-policy-s68a-backend` and `upstream-codex-image-tool-strip-policy-s68a-ui`.
- Both worker results begin with the exact required `### DONE:` verdict.
- Backend commands used `GOTMPDIR=E:/codex-worktrees/sub2api/upstream-codex-image-tool-strip-policy-s68a-qa/.tmp/go-build`.

## Findings

- No implementation bug or behavioral regression was found in the integrated S68a policy prerequisite.
- Backend policy resolution matches the contract: non-OpenAI/nil/unset/unknown values default to `allow`; trimmed case-insensitive `strip`, `remove`, and `drop` normalize to `strip`; a top-level string wins over nested `extra.openai`, while a missing/non-string top-level value permits nested fallback.
- The policy is Codex-only. Managed HTTP applies it only after Codex client detection, and managed WS checks `isCodexCLI`; the ordinary-client HTTP regression test confirms a `strip` account does not strip an ordinary OpenAI request.
- Managed HTTP and WS remove the flat `image_generation` tool and matching tool choice while preserving ordinary functions. HTTP strip mode also disables the image-generation bridge, so the removed tool and bridge instructions are not re-injected. Existing default-allow and Spark injection/stripping behavior remain green.
- The flat helper is correctly bounded: `openai_codex_transform.go:646-687` reads only top-level `reqBody["tools"]` and removes only entries whose own type is `image_generation`. The focused helper test retains an `image_gen` namespace declaration. `additional_tools` and namespace traversal are not stripped in S68a.
- HTTP passthrough remains explicitly outside this policy step: `openai_gateway_service.go:2572-2576` returns through the passthrough path before policy resolution at `:2605-2609`. No apicompat, passthrough adapter, `additional_tools`, DTO, handler, repository, migration, billing, or deployment path changed. These deferred paths remain S68b scope.
- The account modal matches backend precedence and alias normalization. Saving canonicalizes strip to a top-level `"strip"`, saving allow removes the redundant top-level value, and both paths remove the nested legacy policy while cloning and preserving neighboring top-level/nested `extra` keys.
- Setup-token support stays narrow: only the dedicated policy card uses the three-account predicate. Passthrough, bridge, WS, Responses mode, image input, quota, and compact controls remain gated to OAuth/API Key or narrower account types; the component isolation test confirms the relevant controls are absent.
- Allowed/Denied audit found 17 changed paths: 13 are business/test/i18n paths explicitly allowed by the two contracts, two are worker results, and two are workflow metadata (`docs/workflow/status.md`, `docs/workflow/main-log.md`). There are zero Denied Path changes.

## Executed Checks

- Backend policy/HTTP/raw/WS contract command - PASS (`internal/service` package `6.176s`).
- Backend Spark-preservation and existing WS image-tool injection command - PASS (`6.389s`).
- Backend S67 GPT-5.6 effort/candidate/WS passthrough regression command - PASS (`6.367s`).
- Backend compile-only command `go test ./internal/service -run "^$" -count=1` - PASS (`6.242s`, no tests to run).
- `npm.cmd run test:run -- src/components/account/__tests__/EditAccountModal.spec.ts` - PASS: one file, `30/30` tests (`4.60s` total, `468ms` test time).
- `npm.cmd run typecheck` - PASS.
- The frontend worktree had no local dependencies, so a temporary `frontend/node_modules` junction targeted `F:/mcplugins/sub2api/frontend/node_modules`; it was removed in cleanup and `Test-Path frontend/node_modules` returned `False` afterward.
- Manual policy/UI boundary audit covering precedence, nested cleanup, neighbor preservation, setup-token isolation, Codex-only behavior, bridge suppression, ordinary-client preservation, and deferred S68b paths - PASS.
- `git diff --check 5a9ae78d8..HEAD` - PASS.
- Conflict-boundary scan across every changed path - PASS (`NO_CONFLICT_MARKERS`).
- Allowed/Denied union audit - PASS: 13 contract-allowed business/test/i18n paths, two worker results, two workflow metadata paths, zero Denied Paths.
- Worker result first-line audit - PASS for both backend and UI results.

## Unverified Risks

- No real Codex CLI request was sent to a live OpenAI HTTP or WebSocket upstream; managed transport behavior is covered by recorder and in-process WS tests.
- Namespace declarations, Responses Lite `additional_tools`, HTTP passthrough, and broader raw-payload expansion are intentionally unverified and unimplemented in S68a. They remain required S68b work.
- No full backend service or full frontend test suite was required; validation follows the approved focused contracts plus typecheck and compile checks.

## Recommendation

- PASS. S68a safely establishes the account policy and maintenance UI prerequisite. Proceed to a separately contracted S68b for namespace/`additional_tools`/passthrough expansion; do not treat this PASS as evidence that those deferred paths already strip image tools.

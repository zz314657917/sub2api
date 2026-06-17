### DONE: upstream-main-v0137-safe-patches-s15

## Summary

Ported selected low-risk upstream `v0.1.137` safety and compatibility patches without merging `upstream/main`.

## Commit Mapping

- `bbd970249` frontend `form-data` override: ported.
- `fa8f1749f` / `727ac3f68` token refresh non-retryable errors: ported.
- `c1c28ac7b` zstd upstream response decompression: ported.
- `ab9987b2e` non-JSON 2xx failover: ported.
- `6c7203d83` SSE `event:error` raw body preservation: ported.
- `edfd5e373` tool strict default false: ported.
- `c906bf000` / `a4ce73391` / `4f5f2788e` / `262fe1230` Chinese model fallback pricing and image-input token billing: ported.
- `142d8c361` DeepSeek `reasoning_effort=max -> xhigh`: ported.
- `6baf00d78` / `efbf6d209` protocol-aware thinking block filtering: equivalent local port.
- `56c6325d1` MiniMax `thinking.type=enabled -> adaptive`: equivalent local port.
- `a05d9e87c` thinking-enabled fallback reasoning effort: equivalent local port.
- `34b1e56e2` / `5c5283979`: covered by local tests/comments where relevant; doc-only upstream wording was not copied verbatim.

## Constraints Check

- Did not merge or rebase `upstream/main`.
- Did not modify `backend/ent/**`, `backend/migrations/**`, or `backend/cmd/server/VERSION`.
- Did not modify Studio Bridge, payment package UI, Canvas, tickets, or public page product files.

## Notes

- `frontend/pnpm-lock.yaml` was updated manually from upstream-equivalent diff because `corepack.cmd pnpm install --lockfile-only --ignore-scripts` exited 0 but did not update the lockfile in this local environment.
- The old zero-cost usage-log test used `deepseek-v4-flash`; that model now has fallback pricing from this Sprint, so the fixture was updated to `unknown-no-pricing-model`.

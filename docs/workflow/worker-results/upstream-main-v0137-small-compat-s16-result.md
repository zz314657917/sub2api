### DONE: upstream-main-v0137-small-compat-s16

## Summary

Ported a second small batch of upstream `v0.1.137` compatibility fixes after S15, without merging `upstream/main`.

## Commit Mapping

- `a67b10f46` / `44f579100` Responses sticky hash input anchor: equivalent local port. Local `ParsedRequest` is structured rather than raw-range based, so the port adds `ParsedRequest.Input` for `protocol=="responses"` and uses it only when `messages` produced no hash anchor.
- `b256f9114` streaming `max_tokens=1` Haiku probe intercept: ported. The `!stream` restriction was removed and tests cover streaming probes.
- `b88f8e4c0` / `2ce878892` OpenAI `/responses` probe tool capability: ported. Probe payload now requires `probe_ping` tool call, uses mapped upstream model when available, and marks 2xx responses without `function_call` as unsupported.
- `56c62c59c` / `9e9e154f5` API Key ACL denial message includes client IP: equivalent local port. This fork keeps the safer existing default of using trusted client IP rather than trusting forwarded headers by default.

## Constraints Check

- Did not merge or rebase `upstream/main`.
- Did not modify `backend/ent/**`, `backend/migrations/**`, or `backend/cmd/server/VERSION`.
- Did not modify Studio Bridge, payment package UI, Canvas, tickets, or public page product files.
- Kept S15 patches intact.

## Notes

- `backend/internal/server/middleware/api_key_auth_test.go` had a pre-existing compile blocker when testing the middleware package directly: duplicate `fakeSettingRepo` and a stub missing `DeleteWithAudit`. This was fixed as test infrastructure cleanup required by the new ACL tests.
- OpenAI image failover, Anthropic cooldown preservation, account-list parameter batching, OpenAI quota UI, cyber policy, channel monitor jitter, and Claude OAuth system prompt blocks were left for separate Sprints.

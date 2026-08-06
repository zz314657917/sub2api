### BLOCKED: upstream-v0171-tool-output-media-s200

## Findings

- Upstream `2bf9c6d56` fixes a real bridge gap: media embedded in a Responses `*_call_output.output` payload is rewritten to marker text and appended as `image_url` content in a user message immediately after the matching Chat tool reply.
- The upstream patch assumes a newer staged pipeline (`buildChatMessagesFromItems` followed by `normalizeChatMessagesWithToolOutputMedia`) that associates media with call IDs and preserves parallel tool-call/reply ordering.
- The local bridge is an older single-pass `responsesInputToChatMessages` implementation. It serializes every tool output directly into a Chat `tool` message and has no equivalent normalization hook. The only related local conversion is the reverse Anthropic-to-Responses bridge, which confirms the protocol boundary but cannot be reused as a direct implementation.

## Executed Checks

- Compared the full `2bf9c6d56` bridge diff (186 source lines plus 318 tests) with the local `ResponsesToChatCompletionsRequest`, `responsesInputToChatMessages`, `appendAssistantToolCall`, and `anthropic_to_responses` bridge structures.
- Verified that the local Chat bridge already handles direct user-image parts and namespace tool calls, but has no tool-output media association or tool-reply normalization phase.

## Unverified Risks

- No tool-output image fixture, Chat-only upstream, live model, deployment, or production traffic was exercised.
- A direct partial port could place a user image message between parallel tool calls and their replies, violating the Chat Completions tool-call adjacency requirement.

## Recommendation

Do not cherry-pick this patch. Create a separate cross-protocol design/contract that first introduces a tested local tool-call/reply normalization boundary, then ports the media extraction behavior with fixtures for plain text, JSON-escaped JSON, nested images, data URLs, duplicate call IDs, parallel calls, and intervening messages.

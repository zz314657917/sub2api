export const DEFAULT_CODEX_MODEL = 'gpt-6.1-sol'

export function buildCodexConfig(baseUrl: string, websockets = false): string {
  return `model_provider = "OpenAI"
model = "${DEFAULT_CODEX_MODEL}"
review_model = "${DEFAULT_CODEX_MODEL}"
model_reasoning_effort = "xhigh"

[model_providers.OpenAI]
name = "3Z API"
base_url = ${JSON.stringify(baseUrl)}
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1${websockets ? `
supports_websockets = true

[features]
responses_websockets_v2 = true` : ''}`
}

-- Refresh only known default Codex examples; preserve custom providers and article text.
UPDATE tutorial_pages
SET content_md = replace(content_md, $old$model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"$old$, $new$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$new$),
    updated_at = NOW()
WHERE slug = 'codex' AND strpos(content_md, $old$model = "gpt-5.5"
model_provider = "luoye"

[model_providers.luoye]
name = "luoye"
base_url = "https://ai.3zapi.top/v1"
env_key = "OPENAI_API_KEY"
wire_api = "responses"$old$) > 0;

UPDATE tutorial_pages
SET content_md = replace(content_md, $old$- model_provider = "luoye" 表示 Codex 使用下面这个自定义服务商。$old$, $new$- model_provider = "OpenAI" 与 [model_providers.OpenAI] 对应，服务商显示名称为 3Z API。$new$), updated_at = NOW()
WHERE slug = 'codex'
  AND strpos(content_md, $config$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$config$) > 0
  AND strpos(content_md, $old$- model_provider = "luoye" 表示 Codex 使用下面这个自定义服务商。$old$) > 0;

UPDATE tutorial_pages
SET content_md = replace(content_md, $old$- base_url 必须写成 `https://ai.3zapi.top/v1`，结尾带 `/v1`。$old$, $new$- Codex 的 base_url 使用根地址 `https://ai.3zapi.com`，无需追加 `/v1`。$new$), updated_at = NOW()
WHERE slug = 'codex'
  AND strpos(content_md, $config$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$config$) > 0
  AND strpos(content_md, $old$- base_url 必须写成 `https://ai.3zapi.top/v1`，结尾带 `/v1`。$old$) > 0;

UPDATE tutorial_pages
SET content_md = replace(content_md, $old$- env_key = "OPENAI_API_KEY" 要和 auth.json 里的字段名保持一致；CLI 环境变量方案也使用这个名字。$old$, $new$- requires_openai_auth = true 配合下方 auth.json 中的 OPENAI_API_KEY 使用。$new$), updated_at = NOW()
WHERE slug = 'codex'
  AND strpos(content_md, $config$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$config$) > 0
  AND strpos(content_md, $old$- env_key = "OPENAI_API_KEY" 要和 auth.json 里的字段名保持一致；CLI 环境变量方案也使用这个名字。$old$) > 0;

UPDATE tutorial_pages
SET content_md = replace(content_md, $old$1. Base URL 是否为 `https://ai.3zapi.top/v1`。$old$, $new$1. Base URL 是否为 `https://ai.3zapi.com`。$new$), updated_at = NOW()
WHERE slug = 'codex'
  AND strpos(content_md, $config$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$config$) > 0
  AND strpos(content_md, $old$1. Base URL 是否为 `https://ai.3zapi.top/v1`。$old$) > 0;

UPDATE tutorial_pages
SET content_md = replace(content_md, $old$- Base URL 报错：OpenAI 兼容地址必须写 `https://ai.3zapi.top/v1`，不要漏掉 `/v1`。$old$, $new$- Base URL 报错：本页 Codex 配置使用 `https://ai.3zapi.com` 根地址。$new$), updated_at = NOW()
WHERE slug = 'codex'
  AND strpos(content_md, $config$model_provider = "OpenAI"
model = "gpt-6-astra"
review_model = "gpt-5.5"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "3Z API"
base_url = "https://ai.3zapi.com"
wire_api = "responses"
requires_openai_auth = true
request_max_retries = 0
stream_max_retries = 1$config$) > 0
  AND strpos(content_md, $old$- Base URL 报错：OpenAI 兼容地址必须写 `https://ai.3zapi.top/v1`，不要漏掉 `/v1`。$old$) > 0;

-- Refresh the known default OpenAI address block in the getting-started article.
UPDATE tutorial_pages
SET content_md = replace(content_md,
    $old$OpenAI 兼容工具通常使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.top/v1
[[/command]]$old$,
    $new$Codex 使用根地址 https://ai.3zapi.com；其他要求 /v1 的 OpenAI 兼容工具使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.com/v1
[[/command]]$new$),
    updated_at = NOW()
WHERE slug = 'getting-started'
  AND strpos(content_md, $old$OpenAI 兼容工具通常使用：

[[command title="OpenAI 兼容 Base URL"]]
https://ai.3zapi.top/v1
[[/command]]$old$) > 0;

-- Only the saved default Codex address changes; other platforms remain intact.
UPDATE settings
SET value = jsonb_set(value::jsonb, '{platforms}', (
    SELECT jsonb_agg(
        CASE WHEN platform->>'id' = 'codex'
                  AND platform->>'base_url' IN ('https://ai.3zapi.top', 'https://ai.3zapi.top/v1')
             THEN jsonb_set(platform, '{base_url}', '"https://ai.3zapi.com"'::jsonb)
             ELSE platform END ORDER BY ordinal
    )
    FROM jsonb_array_elements(value::jsonb->'platforms') WITH ORDINALITY AS items(platform, ordinal)
))::text, updated_at = NOW()
WHERE key = 'quickstart_tutorial_config'
  AND jsonb_typeof(value::jsonb->'platforms') = 'array'
  AND EXISTS (
      SELECT 1 FROM jsonb_array_elements(value::jsonb->'platforms') AS platform
      WHERE platform->>'id' = 'codex'
        AND platform->>'base_url' IN ('https://ai.3zapi.top', 'https://ai.3zapi.top/v1')
  );

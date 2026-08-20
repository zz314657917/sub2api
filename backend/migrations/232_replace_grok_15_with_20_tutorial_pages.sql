DELETE FROM tutorial_pages
WHERE slug IN ('grok-imagine-1-5', 'grok-imagine-1-5-edit');

INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES
('grok-imagine-2-0-ext', 'Grok Imagine 2.0 Ext 图像生成', '使用 Grok Imagine 2.0 Ext 异步文生图。', '图像模型', 2249, 'published', $md$
# Grok Imagine 2.0 Ext 图像生成

模型 ID：`grok-imagine-2.0-ext`。请求 `POST https://ai.3zapi.cc/v1/images/generations`。

## 参数

- `model`：必填，固定为 `grok-imagine-2.0-ext`。
- `prompt`：必填，非空字符串。
- `n`：可选，`1` 到 `12`，默认 `1`。
- `size`：可选，支持 `1:1`、`2:3`、`3:2`、`3:4`、`4:3`、`9:16`、`16:9`，以及对应的兼容像素尺寸。
- `resolution`：可选，传 `quality`；它表示质量模式，不是 1K/2K/4K 像素档。
- `response_format`：可选，只能是 `url`；不支持 `b64_json` 或 `base64`。
- `Idempotency-Key`：强烈建议使用 UUID；网络超时重试时必须复用相同的 Key 和请求体。

此模型只支持文生图，不支持 `image_urls`、`stream`、`quality` 或 `style`。

## cURL

```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 7d8f6c7e-3f4f-4f92-a4d5-000000000001" \
  -d '{
    "model": "grok-imagine-2.0-ext",
    "prompt": "一只红苹果放在白色陶瓷盘上，干净的棚拍产品图",
    "n": 1,
    "size": "1:1",
    "resolution": "quality",
    "response_format": "url"
  }'
```

## 返回和轮询

提交成功返回 HTTP `202`，任务 ID 在 `data.id`。使用 `GET /v1/tasks/{task_id}` 轮询，直到 `completed` 或 `failed`。完成后遍历 `data.result.images[].url[]`，并尽快下载图片链接。

## 错误处理

- `400 invalid_request_error`：提示词为空或 JSON 不合法。
- `400 invalid_n`：`n` 不在 1–12。
- `400 invalid_size`：`size` 不在白名单。
- `400 invalid_response_format`：传入了非 `url` 格式。
- `400 invalid_quality`：误传 `quality` 字段，应改用 `resolution: "quality"`。
- `400 invalid_image_input`：传入了参考图；该模型不支持图生图。
- `401 authentication_error`：API Key 无效。
- `402 payment_required`：余额不足。
- `409 idempotency_error`：幂等 Key 冲突，按原 Key 和原请求体重试。
- `400 invalid_idempotency_key`：幂等 Key 不是合法的可见 ASCII 字符串；改用 UUID。
- `429 rate_limit_error`：遵循 `Retry-After` 退避。

错误通常位于 `error.type`、`error.code` 和 `error.message`；网络超时后不要立即换新的幂等 Key。
$md$),
('grok-imagine-image-2-0', 'Grok Imagine Image 2.0 图像生成与编辑', '使用 Grok Imagine Image 2.0 进行文生图、单图编辑和多图参考。', '图像模型', 2250, 'published', $md$
# Grok Imagine Image 2.0 图像生成与编辑

模型 ID：`grok-imagine-image-2.0`。请求 `POST https://ai.3zapi.cc/v1/images/generations`。

## 参数

- `model`：必填，固定为 `grok-imagine-image-2.0`。
- `prompt`：必填，最多 8000 个字符；多图参考时可用 `<IMAGE_0>`、`<IMAGE_1>`、`<IMAGE_2>` 指定图片作用。
- `n`：可选，`1` 到 `10`，默认 `1`。
- `aspect_ratio`：可选，支持 `auto`、`1:1`、`3:4`、`4:3`、`9:16`、`16:9`、`2:3`、`3:2`、`9:19.5`、`19.5:9`、`9:20`、`20:9`、`1:2`、`2:1`。
- `resolution`：可选，`1k` 或 `2k`。
- `quality`：仅文生图使用，`low` 或 `medium`；带 `image_urls` 时不要传。
- `image_urls`：可选，1–3 个公网图片 URL；不传表示文生图，传入表示编辑或多图参考。

## cURL：文生图

```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 7d8f6c7e-3f4f-4f92-a4d5-000000000002" \
  -d '{
    "model": "grok-imagine-image-2.0",
    "prompt": "雨夜中霓虹灯照亮的电影感城市街道",
    "n": 1,
    "aspect_ratio": "16:9",
    "resolution": "2k",
    "quality": "medium"
  }'
```

## cURL：参考图编辑

```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 7d8f6c7e-3f4f-4f92-a4d5-000000000003" \
  -d '{
    "model": "grok-imagine-image-2.0",
    "prompt": "保留产品主体，把背景改成蓝色渐变摄影棚",
    "n": 1,
    "aspect_ratio": "auto",
    "resolution": "1k",
    "image_urls": ["https://example.com/reference.png"]
  }'
```

## 返回和轮询

提交成功返回 HTTP `202`，保存 `data.id` 或 `data.poll_url`。轮询任务直到 `completed` 或 `failed`，成功时从 `data.result.images[].url[]` 读取全部图片。结果链接通常有有效期，应及时下载。

## 错误处理

- `400 invalid_request_error`：提示词、`aspect_ratio`、`resolution` 或 JSON 不合法。
- `400 invalid_n`：`n` 不在 1–10。
- `400 invalid_image_input`：参考图不是公网 URL、超过 3 张或格式无效。
- `400 invalid_quality`：带参考图时仍传了 `quality`。
- `401 authentication_error`：API Key 无效。
- `402 payment_required`：余额不足。
- `409 idempotency_error`：幂等 Key 冲突，复用原 Key 和原请求体。
- `400 invalid_idempotency_key`：幂等 Key 格式不合法；改用 UUID。
- `429 rate_limit_error`：降低并发并按照 `Retry-After` 退避。

程序应保留 `request_id`、任务 ID 和错误字段，不能记录 API Key；不要把提交成功的 HTTP 202 当成图片已完成。
$md$);

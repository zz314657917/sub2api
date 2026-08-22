INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES
('grok-imagine-1-5', 'Grok Imagine 1.5 图像生成', '使用 Grok Imagine 1.5 生成图像。', '图像模型', 2249, 'published', $md$
# Grok Imagine 1.5 图像生成

模型 ID：`grok-imagine-1.5-apimart`。请求 `POST https://ai.3zapi.cc/v1/images/generations`，并设置 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: 必填，固定为 `grok-imagine-1.5-apimart`。
- `prompt`: 必填，非空字符串。建议按“主体 + 场景 + 构图 + 光线 + 风格”描述。
- `size`: 可选，画幅比例或像素尺寸，例如 `1:1`、`16:9`、`9:16`、`1024x1024`。
- `n`: 可选，正整数，表示生成数量；建议先传 `1`，需要多张时遍历返回的 `data` 数组。

本地网关会将请求提交到异步图像任务，并自动轮询完成。`quality`、`background`、`resolution`、`mask_url` 等字段不是此模型的本地生成参数，传入前应删除。

## 使用场景示例

- 产品海报：在 `prompt` 中写清产品外形、品牌色、留白位置和文字区域。
- 社交媒体配图：使用 `16:9` 或 `9:16`，并说明主体在画面中的位置。
- 概念草图：先使用 `1:1` 和 `n: 1` 验证构图，再提高提示词细节。

## cURL

```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "grok-imagine-1.5-apimart",
    "prompt": "一张极简科技产品海报，银色耳机置于黑色桌面，冷色边缘光，右侧留白用于标题",
    "size": "16:9",
    "n": 1
  }'
```

## 返回结果

提交成功后，网关会等待任务完成并返回 OpenAI 兼容结果，通常从 `data[].url` 读取图片。程序应遍历 `data`，不要只取第一项。若上游任务失败，网关会返回错误而不是成功的图片 URL。

## 错误处理

- `400 invalid_request_error`: `model`、`prompt`、`size` 或 `n` 不合法；删除不支持的字段后重试。
- `401 authentication_error`: API Key 无效或缺失。
- `402 payment_required`: 余额不足。
- `403 permission_error`: 当前分组没有图像权限。
- `413 invalid_request_error`: 请求体或输入内容过大。
- `429 rate_limit_error`: 降低并发并使用指数退避。
- `502 upstream_error` / `503 api_error`: 上游暂时不可用，稍后重试。

错误通常为 `{"error":{"type":"...","message":"..."}}`。记录 `type`、`message` 和请求 ID（如有），不要记录 API Key。
$md$, NOW()),
('grok-imagine-1-5-edit', 'Grok Imagine 1.5 图像编辑', '使用 Grok Imagine 1.5 根据参考图进行编辑。', '图像模型', 2250, 'published', $md$
# Grok Imagine 1.5 图像编辑

模型 ID：`grok-imagine-1.5-edit-apimart`。请求 `POST https://ai.3zapi.cc/v1/images/edits`，认证头为 `Authorization: Bearer YOUR_API_KEY`。

## 参数

- `model`: 必填，固定为 `grok-imagine-1.5-edit-apimart`。
- `prompt`: 必填，说明要保留的内容和要修改的内容，例如“保留人物和服装，只把背景改成海边”。
- `image`: 必填，参考图文件；也可以使用 JSON 的 `image_urls` 传入一个可访问的图片 URL 或 `data:image/...;base64,...` 数据 URI。
- `image_urls`: 最多 1 个参考图。超过 1 个会在上传前返回参数错误。
- `size`: 可选，画幅比例或像素尺寸，例如 `1:1`、`4:3`、`1024x1024`。
- `n`: 可选，正整数；建议传 `1`。

编辑模型只支持单张参考图。`quality`、`background`、`resolution`、`mask_url` 等字段不是此模型的本地编辑参数，不能依赖这些字段改变输出。

## 使用场景示例

- 改背景：明确“保留主体，只替换背景”。
- 商品变体：明确保留商品结构、标志和材质，只修改颜色或环境。
- 风格转换：明确保留构图和主体，指定新的艺术风格。

## cURL

```bash
curl https://ai.3zapi.cc/v1/images/edits \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "model=grok-imagine-1.5-edit-apimart" \
  -F "prompt=保留耳机主体和材质，只把背景改成蓝色渐变摄影棚" \
  -F "size=1:1" \
  -F "n=1" \
  -F "image=@reference.png"
```

## 返回结果

网关会先上传参考图，再提交异步任务并自动轮询。成功时从返回的 `data[].url` 读取图片，并尽快下载结果链接。不要把“任务已提交”当成图片已经生成。

## 错误处理

- `400 invalid_request_error`: 缺少 `image`/`image_urls`、参考图超过 1 张、`prompt` 为空或参数类型错误。
- `401 authentication_error`: API Key 无效或缺失。
- `402 payment_required`: 余额不足。
- `403 permission_error`: 当前分组没有图像编辑权限。
- `413 invalid_request_error`: 参考图或请求体过大。
- `429 rate_limit_error`: 降低并发并退避重试。
- `502 upstream_error` / `503 api_error`: 上游暂时不可用，稍后重试。

错误通常为 `{"error":{"type":"...","message":"..."}}`。先检查参考图是否可读、格式是否有效，再决定是否重试；不要重复提交同一编辑任务造成重复扣费。
$md$, NOW());

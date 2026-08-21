INSERT INTO tutorial_pages (slug, title, description, category, sort_order, status, content_md, published_at)
VALUES
('seedream-5-0-pro', 'Seedream 5.0 Pro 图像生成', '使用 Seedream 5.0 Pro 进行文生图、图生图和多参考图融合。', '图像模型', 2251, 'published', $md$
# Seedream 5.0 Pro 图像生成

模型 ID：\`seedream-5-0-pro\`。请求\`POST https://ai.3zapi.cc/v1/images/generations\`。

## 参数

- \`model\`：必填，固定为 \`seedream-5-0-pro\`（兼容别名 \`seedream-5.0-pro\`）。
- \`prompt\`：必填，建议控制在 600 个英文单词以内。
- \`size\`：可选，支持 \`1:1\`、\`4:3\`、\`3:4\`、\`16:9\`、\`9:16\`、\`3:2\`、\`2:3\`、\`2:1\`、\`1:2\`、\`21:9\`，以及精确像素尺寸。
- \`resolution\`：可选，\`1K\`、\`1.5K\`、\`2K\`；精确像素 \`size\` 会优先于该字段。
- \`image_urls\`：可选，最多 10 张公网 URL 或完整 Data URI。
- \`n\`：固定为 \`1\`；这是单图模型。
- \`output_format\`：可选，\`png\` 或 \`jpeg\`。
- \`watermark\`、\`nsfw_check\`：可选布尔值。

Pro 不支持 \`n > 1\`、\`stream\`、\`tools\` 或超过 10 张参考图。

## cURL

\`\`\`bash
curl https://ai.3zapi.cc/v1/images/generations \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "seedream-5-0-pro",
    "prompt": "赛博朋克风格的城市夜景，霓虹灯反射在湿润街道上",
    "size": "16:9",
    "resolution": "2K",
    "n": 1
  }'
\`\`\`

## 返回和错误处理

提交成功后网关会轮询任务并返回图片 URL。常见错误：\`400 invalid_request_error\`（参数、比例、分辨率、\`n\` 或参考图数量错误）、\`401 authentication_error\`、\`402 payment_required\`、\`403 permission_error\`、\`429 rate_limit_error\`、\`502 upstream_error\` 和 \`503 api_error\`。记录 \`error.type\` 与 \`error.message\`，不要记录密钥。
$md$, NOW()),
('seedream-5-0-lite', 'Seedream 5.0 Lite 图像生成', '使用 Seedream 5.0 Lite 进行文生图、图生图和组图生成。', '图像模型', 2252, 'published', $md$
# Seedream 5.0 Lite 图像生成

模型 ID：\`seedream-5-0-lite\`。请求\`POST https://ai.3zapi.cc/v1/images/generations\`。

## 参数

- \`model\`：必填，固定为 \`seedream-5-0-lite\`（兼容别名 \`seedream-5.0-lite\`）。
- \`prompt\`：必填，非空字符串。
- \`size\`：可选，支持常用比例和 \`auto\`。
- \`resolution\`：可选，\`2K\`、\`3K\`、\`4K\`；Lite 不支持 \`1K\`。
- \`n\`：可选，\`1\` 到 \`15\`；\`image_urls\` 数量加 \`n\` 不得超过 15。
- \`image_urls\`：可选，公网 URL 或完整 Data URI。
- \`output_format\`：可选，\`jpeg\` 或 \`png\`。
- \`watermark\`：可选，布尔值。
- \`sequential_image_generation\`：可选，\`disabled\` 或 \`auto\`；\`n > 1\` 时网关默认使用 \`auto\`。

## cURL

\`\`\`bash
curl https://ai.3zapi.cc/v1/images/generations \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "seedream-5-0-lite",
    "prompt": "一只在花园里玩耍的金毛犬，阳光明媚，高清摄影风格",
    "size": "16:9",
    "resolution": "2K",
    "n": 1,
    "output_format": "png"
  }'
\`\`\`

## 返回和错误处理

提交后网关会轮询任务直到完成或失败，成功时从 \`data[].url\` 读取图片。\`400 invalid_request_error\` 常见于分辨率、比例、格式、\`n\` 或参考图加输出图超过 15；\`401 authentication_error\`、\`402 payment_required\`、\`403 permission_error\`、\`429 rate_limit_error\`、\`502 upstream_error\` 和 \`503 api_error\` 按状态处理。
$md$, NOW());

WITH parameter_blocks (slug, parameters_md) AS (
    VALUES
        ('gpt-image-2', $params$
- `model`: 固定为 `gpt-image-2`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`。
- `n`: 生成数量。
$params$),
        ('gpt-image-2-official', $params$
- `model`: 固定为 `gpt-image-2-official`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`。
- `n`: 生成数量。
- `quality`: 可选，图像质量。
- `background`: 可选，背景模式。
- `moderation`: 可选，内容审核模式。
- `output_format`: 可选，输出格式。
- `output_compression`: 可选，输出压缩质量。
- `mask_url`: 可选，可访问的蒙版图像 URL。
$params$),
        ('gemini-3-pro-image-preview', $params$
- `model`: 固定为 `gemini-3-pro-image-preview`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`。
- `image_urls`: 可选参考图，最多 14 张。
- `n`: 兼容流程中固定为 `1`。
$params$),
        ('gemini-3-pro-image-preview-official', $params$
- `model`: 固定为 `gemini-3-pro-image-preview-official`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`。
- `image_urls`: 可选参考图，最多 14 张。
- `n`: 兼容流程中固定为 `1`。
$params$),
        ('gemini-3-1-flash-image-preview', $params$
- `model`: 固定为 `gemini-3.1-flash-image-preview`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`，不要使用 `0.5k`。
- `image_urls`: 可选参考图，最多 14 张。
- `n`: 兼容流程中固定为 `1`。
- `google_search`: 可选，是否启用 Google 搜索。
- `google_image_search`: 可选，是否启用 Google 图片搜索。
$params$),
        ('gemini-3-1-flash-image-preview-official', $params$
- `model`: 固定为 `gemini-3.1-flash-image-preview-official`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位，可用 `1k`、`2k`、`4k`，不要使用 `0.5k`。
- `image_urls`: 可选参考图，最多 14 张。
- `n`: 兼容流程中固定为 `1`。
- `google_search`: 可选，是否启用 Google 搜索。
- `google_image_search`: 可选，是否启用 Google 图片搜索。
$params$),
        ('midjourney', $params$
- `model`: 固定为 `midjourney`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `version`: Midjourney 模型版本。
- `speed`: 生成速度模式。
- `quality`: 生成质量。
- `stylize`: 风格化强度。
- `chaos`: 结果差异程度。
- `weird`: 非常规风格强度。
- `stop`: 提前停止百分比。
- `niji`: 是否使用 Niji 模式。
- `raw`: 是否使用 Raw 模式。
- `tile`: 是否生成可平铺图像。
- `image_urls`: 可选参考图 URL 列表。
$params$),
        ('doubao-seedance-4-0', $params$
- `model`: 固定为 `doubao-seedance-4-0`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 分辨率档位。
- `n`: 生成数量。
- `image_urls`: 可选参考图；参考图数量加 `n` 不得超过 15。
$params$),
        ('doubao-seedance-4-5', $params$
- `model`: 固定为 `doubao-seedance-4-5`。
- `prompt`: 图像描述。
- `size`: 像素尺寸或比例，例如 `1024x1024`。
- `resolution`: 推荐仅使用 `2k` 或 `4k`。
- `n`: 生成数量。
- `image_urls`: 可选参考图；参考图数量加 `n` 不得超过 15。
$params$)
)
UPDATE tutorial_pages AS page
SET content_md = regexp_replace(
        page.content_md,
        E'## 参数\\r?\\n\\r?\\n.*?\\r?\\n## cURL',
        '## 参数' || E'\n\n' || btrim(parameter_blocks.parameters_md, E'\r\n') || E'\n\n## cURL',
        's'
    ),
    updated_at = NOW()
FROM parameter_blocks
WHERE page.slug = parameter_blocks.slug
  AND page.content_md ~ E'## 参数\\r?\\n\\r?\\n.*?\\r?\\n## cURL';

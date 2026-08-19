WITH curl_examples (slug, curl_md) AS (
    VALUES
        ('gpt-image-2', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "一只红色风筝",
    "size": "1024x1024",
    "resolution": "1k",
    "n": 1
  }'
```
$curl$),
        ('gpt-image-2-official', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2-official",
    "prompt": "一张干净的产品图",
    "size": "1024x1024",
    "resolution": "1k",
    "n": 1,
    "quality": "high",
    "background": "auto",
    "output_format": "png"
  }'
```
$curl$),
        ('gemini-3-pro-image-preview', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3-pro-image-preview",
    "prompt": "水彩山谷",
    "size": "1024x1024",
    "resolution": "2k",
    "n": 1
  }'
```
$curl$),
        ('gemini-3-pro-image-preview-official', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3-pro-image-preview-official",
    "prompt": "未来城市夜景",
    "size": "1024x1024",
    "resolution": "4k",
    "n": 1
  }'
```
$curl$),
        ('gemini-3-1-flash-image-preview', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.1-flash-image-preview",
    "prompt": "复古海报",
    "size": "1024x1024",
    "resolution": "1k",
    "n": 1,
    "google_search": true
  }'
```
$curl$),
        ('gemini-3-1-flash-image-preview-official', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.1-flash-image-preview-official",
    "prompt": "玻璃质感图标",
    "size": "1024x1024",
    "resolution": "2k",
    "n": 1,
    "google_image_search": true
  }'
```
$curl$),
        ('midjourney', $curl$
```bash
curl https://ai.3zapi.cc/v1/midjourney/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "midjourney",
    "prompt": "雨后的霓虹街道",
    "size": "1024x1024",
    "version": "6",
    "quality": "1",
    "stylize": 100,
    "chaos": 10,
    "raw": true
  }'
```
$curl$),
        ('doubao-seedance-4-0', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "doubao-seedance-4-0",
    "prompt": "晨雾森林",
    "size": "1024x1024",
    "resolution": "2k",
    "n": 1
  }'
```
$curl$),
        ('doubao-seedance-4-5', $curl$
```bash
curl https://ai.3zapi.cc/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "doubao-seedance-4-5",
    "prompt": "极简建筑渲染",
    "size": "1024x1024",
    "resolution": "4k",
    "n": 1
  }'
```
$curl$)
)
UPDATE tutorial_pages AS page
SET content_md = regexp_replace(
        replace(
            page.content_md,
            $literal$\n\n## cURL$literal$,
            E'\n\n## cURL'
        ),
        E'## cURL\\r?\\n\\r?\\n```bash\\r?\\n.*?\\r?\\n```',
        '## cURL' || E'\n\n' || btrim(curl_examples.curl_md, E'\r\n'),
        's'
    ),
    updated_at = NOW()
FROM curl_examples
WHERE page.slug = curl_examples.slug
  AND page.content_md ~ E'## cURL\\r?\\n\\r?\\n```bash\\r?\\n.*?\\r?\\n```';

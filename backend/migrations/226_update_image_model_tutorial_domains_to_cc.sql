UPDATE tutorial_pages
SET content_md = replace(
        content_md,
        'https://ai.3zapi.com',
        'https://ai.3zapi.cc'
    ),
    updated_at = NOW()
WHERE slug IN (
    'gpt-image-2',
    'gpt-image-2-official',
    'gemini-3-pro-image-preview',
    'gemini-3-pro-image-preview-official',
    'gemini-3-1-flash-image-preview',
    'gemini-3-1-flash-image-preview-official',
    'midjourney',
    'doubao-seedance-4-0',
    'doubao-seedance-4-5'
)
  AND content_md LIKE '%https://ai.3zapi.com%';

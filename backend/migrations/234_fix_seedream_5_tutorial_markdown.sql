UPDATE tutorial_pages
SET content_md = replace(content_md, E'\\`', '`'),
    updated_at = NOW()
WHERE slug IN ('seedream-5-0-pro', 'seedream-5-0-lite')
  AND strpos(content_md, E'\\`') > 0;

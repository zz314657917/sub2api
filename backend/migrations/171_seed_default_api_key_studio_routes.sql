-- Seed untouched default API keys with Studio Bridge multi-group routes.
-- User-customized keys are left unchanged: any existing group_id or non-empty
-- multi_group_routes means the owner already chose their routing.

WITH studio_settings AS (
    SELECT value::jsonb AS cfg
    FROM settings
    WHERE key = 'studio_bridge_luoye_ai'
      AND COALESCE(value, '') <> ''
      AND BTRIM(value) LIKE '{%'
      AND jsonb_typeof(value::jsonb) = 'object'
    LIMIT 1
),
configured_routes AS (
    SELECT jsonb_build_object(
        'group_id', chat_group.id,
        'priority', 1,
        'weight', 1,
        'cooldown_seconds', 30,
        'enabled', true,
        'text_only', true
    ) AS route
    FROM studio_settings
    JOIN groups chat_group
      ON chat_group.id = CASE
          WHEN BTRIM(COALESCE(studio_settings.cfg ->> 'default_chat_group', '')) ~ '^[0-9]+$'
          THEN BTRIM(studio_settings.cfg ->> 'default_chat_group')::bigint
          ELSE NULL
      END
     AND chat_group.deleted_at IS NULL

    UNION ALL

    SELECT jsonb_build_object(
        'group_id', image_group.id,
        'priority', 1,
        'weight', 1,
        'cooldown_seconds', 30,
        'enabled', true,
        'image_only', true
    ) AS route
    FROM studio_settings
    JOIN groups image_group
      ON image_group.id = CASE
          WHEN BTRIM(COALESCE(studio_settings.cfg ->> 'default_image_group', '')) ~ '^[0-9]+$'
          THEN BTRIM(studio_settings.cfg ->> 'default_image_group')::bigint
          ELSE NULL
      END
     AND image_group.deleted_at IS NULL

    UNION ALL

    SELECT jsonb_build_object(
        'group_id', video_group.id,
        'priority', 1,
        'weight', 1,
        'cooldown_seconds', 30,
        'enabled', true,
        'model_patterns', jsonb_build_array('doubao-seedance-*', '*-video-*')
    ) AS route
    FROM studio_settings
    JOIN groups video_group
      ON video_group.id = CASE
          WHEN BTRIM(COALESCE(studio_settings.cfg ->> 'default_video_group', '')) ~ '^[0-9]+$'
          THEN BTRIM(studio_settings.cfg ->> 'default_video_group')::bigint
          ELSE NULL
      END
     AND video_group.deleted_at IS NULL
),
route_payload AS (
    SELECT COALESCE(jsonb_agg(route), '[]'::jsonb) AS routes
    FROM configured_routes
)
UPDATE api_keys
SET multi_group_routes = route_payload.routes,
    updated_at = NOW()
FROM route_payload
WHERE api_keys.deleted_at IS NULL
  AND api_keys.name = '默认 API Key（勿删）'
  AND api_keys.group_id IS NULL
  AND COALESCE(jsonb_array_length(api_keys.multi_group_routes), 0) = 0
  AND jsonb_array_length(route_payload.routes) > 0;

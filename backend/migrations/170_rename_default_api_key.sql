-- Rename the automatically generated default API key without changing routing.
-- Legacy Studio Bridge keys from the earlier implementation are kept only when
-- they are the user's first visible key; otherwise they are soft-deleted so they
-- no longer appear as an extra API key.

UPDATE api_keys
SET name = '默认 API Key（勿删）',
    updated_at = NOW()
WHERE deleted_at IS NULL
  AND name = 'Default API Key';

WITH visible_keys AS (
    SELECT id,
           user_id,
           MIN(id) OVER (PARTITION BY user_id) AS first_key_id
    FROM api_keys
    WHERE deleted_at IS NULL
),
legacy_bridge_keys AS (
    SELECT id, user_id, first_key_id
    FROM visible_keys
    WHERE id IN (
        SELECT id
        FROM api_keys
        WHERE deleted_at IS NULL
          AND (name = 'studio-bridge' OR key LIKE 'sk-studio-bridge-%')
    )
)
UPDATE api_keys
SET name = '默认 API Key（勿删）',
    updated_at = NOW()
FROM legacy_bridge_keys
WHERE api_keys.id = legacy_bridge_keys.id
  AND legacy_bridge_keys.id = legacy_bridge_keys.first_key_id;

WITH visible_keys AS (
    SELECT id,
           user_id,
           MIN(id) OVER (PARTITION BY user_id) AS first_key_id
    FROM api_keys
    WHERE deleted_at IS NULL
),
legacy_bridge_keys AS (
    SELECT id, user_id, first_key_id
    FROM visible_keys
    WHERE id IN (
        SELECT id
        FROM api_keys
        WHERE deleted_at IS NULL
          AND (name = 'studio-bridge' OR key LIKE 'sk-studio-bridge-%')
    )
)
UPDATE api_keys
SET key = '__deleted__studio_bridge__' || api_keys.id || '__' || extract(epoch from now())::bigint,
    deleted_at = NOW(),
    updated_at = NOW()
FROM legacy_bridge_keys
WHERE api_keys.id = legacy_bridge_keys.id
  AND legacy_bridge_keys.id <> legacy_bridge_keys.first_key_id;

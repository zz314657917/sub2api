WITH ranked AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY
                CASE WHEN status = 'running' THEN 0 ELSE 1 END,
                created_at ASC,
                id ASC
        ) AS active_rank
    FROM image_creator_tasks
    WHERE status IN ('pending', 'running')
)
UPDATE image_creator_tasks AS task
SET status = 'failed',
    error_message = COALESCE(NULLIF(task.error_message, ''), 'another image generation task is already active; please submit again later'),
    completed_at = COALESCE(task.completed_at, NOW()),
    updated_at = NOW()
FROM ranked
WHERE task.id = ranked.id
  AND ranked.active_rank > 1;

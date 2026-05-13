CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_image_creator_tasks_user_active_unique
    ON image_creator_tasks(user_id)
    WHERE status IN ('pending', 'running');

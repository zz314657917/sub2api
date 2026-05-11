ALTER TABLE image_creator_tasks
    DROP CONSTRAINT IF EXISTS chk_image_creator_tasks_count;

ALTER TABLE image_creator_tasks
    ADD CONSTRAINT chk_image_creator_tasks_count
        CHECK (image_count BETWEEN 1 AND 8);

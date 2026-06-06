-- 给分组增加运行时用途标记，用于 API Key 智能多分组路由自动分流。
-- inference: 普通推理/聊天/responses
-- image: 图片生成
-- video: 视频生成
-- embedding: Embedding

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS routing_scope VARCHAR(20) NOT NULL DEFAULT 'inference';

UPDATE groups
SET routing_scope = 'image'
WHERE allow_image_generation = TRUE
  AND (routing_scope IS NULL OR routing_scope = 'inference');

CREATE INDEX IF NOT EXISTS idx_groups_routing_scope ON groups(routing_scope);

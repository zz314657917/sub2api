-- /admin/groups daily usage rollups. Historical data is published by the aggregation job.
CREATE TABLE IF NOT EXISTS usage_group_daily_rollups (
    bucket_date DATE NOT NULL,
    group_id BIGINT NOT NULL,
    actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bucket_date, group_id)
);

CREATE TABLE IF NOT EXISTS usage_group_rollup_state (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    closed_before DATE NOT NULL DEFAULT DATE '1970-01-01',
    retained_from TIMESTAMPTZ NOT NULL DEFAULT TIMESTAMPTZ '1970-01-01 00:00:00+00',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO usage_group_rollup_state (id, closed_before, retained_from)
VALUES (1, DATE '1970-01-01', TIMESTAMPTZ '1970-01-01 00:00:00+00')
ON CONFLICT (id) DO NOTHING;

CREATE OR REPLACE FUNCTION invalidate_group_usage_rollup_state()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE affected_date DATE; published_before DATE;
BEGIN
    IF TG_OP = 'DELETE' THEN
        affected_date := (OLD.created_at AT TIME ZONE 'Asia/Shanghai')::date;
    ELSIF OLD.group_id IS NULL THEN
        affected_date := (NEW.created_at AT TIME ZONE 'Asia/Shanghai')::date;
    ELSIF NEW.group_id IS NULL THEN
        affected_date := (OLD.created_at AT TIME ZONE 'Asia/Shanghai')::date;
    ELSE
        affected_date := LEAST((OLD.created_at AT TIME ZONE 'Asia/Shanghai')::date, (NEW.created_at AT TIME ZONE 'Asia/Shanghai')::date);
    END IF;
    SELECT closed_before INTO published_before FROM usage_group_rollup_state WHERE id = 1 FOR UPDATE;
    IF published_before > affected_date THEN
        UPDATE usage_group_rollup_state SET closed_before = LEAST(closed_before, affected_date), updated_at = NOW() WHERE id = 1;
    END IF;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION invalidate_group_usage_rollup_state_after_insert()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE affected_date DATE; published_before DATE;
BEGIN
    SELECT MIN((created_at AT TIME ZONE 'Asia/Shanghai')::date) INTO affected_date FROM inserted_usage_logs WHERE group_id IS NOT NULL;
    IF affected_date IS NULL THEN RETURN NULL; END IF;
    SELECT closed_before INTO published_before FROM usage_group_rollup_state WHERE id = 1 FOR KEY SHARE;
    IF published_before > affected_date THEN
        UPDATE usage_group_rollup_state SET closed_before = LEAST(closed_before, affected_date), updated_at = NOW() WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS usage_logs_group_rollup_invalidate_insert ON usage_logs;
CREATE TRIGGER usage_logs_group_rollup_invalidate_insert AFTER INSERT ON usage_logs REFERENCING NEW TABLE AS inserted_usage_logs FOR EACH STATEMENT EXECUTE FUNCTION invalidate_group_usage_rollup_state_after_insert();
DROP TRIGGER IF EXISTS usage_logs_group_rollup_invalidate_delete ON usage_logs;
CREATE TRIGGER usage_logs_group_rollup_invalidate_delete AFTER DELETE ON usage_logs FOR EACH ROW WHEN (OLD.group_id IS NOT NULL) EXECUTE FUNCTION invalidate_group_usage_rollup_state();
DROP TRIGGER IF EXISTS usage_logs_group_rollup_invalidate_update ON usage_logs;
CREATE TRIGGER usage_logs_group_rollup_invalidate_update AFTER UPDATE OF created_at, group_id, actual_cost ON usage_logs FOR EACH ROW WHEN ((OLD.created_at IS DISTINCT FROM NEW.created_at OR OLD.group_id IS DISTINCT FROM NEW.group_id OR OLD.actual_cost IS DISTINCT FROM NEW.actual_cost) AND (OLD.group_id IS NOT NULL OR NEW.group_id IS NOT NULL)) EXECUTE FUNCTION invalidate_group_usage_rollup_state();

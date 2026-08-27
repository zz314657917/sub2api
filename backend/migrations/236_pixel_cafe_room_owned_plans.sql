-- Pixel Cafe S266: every active Cafe room owns one editable plan. Historical
-- rounds, seats and orders keep their original plan references.

DO $$
DECLARE
    duplicate_room RECORD;
    copied_plan_id BIGINT;
BEGIN
    FOR duplicate_room IN
        SELECT ranked.room_id, ranked.plan_id
        FROM (
            SELECT id AS room_id,
                   plan_id,
                   ROW_NUMBER() OVER (PARTITION BY plan_id ORDER BY id) AS owner_rank
            FROM cafe_rooms
            WHERE deleted_at IS NULL
        ) AS ranked
        WHERE ranked.owner_rank > 1
        ORDER BY ranked.plan_id, ranked.room_id
    LOOP
        INSERT INTO group_buy_plans (
            title, description, product_key, total_shares, subscription_tier,
            max_buyers, seat_count, price_per_share, price_per_seat, price_label,
            quota_per_share_label, quota_label, max_shares_per_user,
            fulfillment_timeout_minutes, target_group_id, tier_group_ids,
            tier_rules, validity_days, timeout_minutes, launch_mode,
            fulfillment_mode, room_key_quota_usd, room_key_rate_limit_5h,
            room_key_rate_limit_1d, room_key_rate_limit_7d,
            auto_create_room_key, refund_mode, agreement_text, status,
            sort_order, last_round_created_at, deleted_at, created_at, updated_at
        )
        SELECT
            room.name, source.description, source.product_key,
            source.total_shares, source.subscription_tier, source.max_buyers,
            source.seat_count, source.price_per_share, source.price_per_seat,
            source.price_label, source.quota_per_share_label,
            source.quota_label, source.max_shares_per_user,
            source.fulfillment_timeout_minutes, source.target_group_id,
            source.tier_group_ids, source.tier_rules, source.validity_days,
            source.timeout_minutes, source.launch_mode,
            source.fulfillment_mode, source.room_key_quota_usd,
            source.room_key_rate_limit_5h, source.room_key_rate_limit_1d,
            source.room_key_rate_limit_7d, source.auto_create_room_key,
            source.refund_mode, source.agreement_text, source.status,
            source.sort_order, source.last_round_created_at, NULL, NOW(), NOW()
        FROM group_buy_plans AS source
        JOIN cafe_rooms AS room ON room.id = duplicate_room.room_id
        WHERE source.id = duplicate_room.plan_id
        RETURNING id INTO copied_plan_id;

        UPDATE cafe_rooms
        SET plan_id = copied_plan_id,
            updated_at = NOW()
        WHERE id = duplicate_room.room_id
          AND plan_id = duplicate_room.plan_id
          AND deleted_at IS NULL;
    END LOOP;
END $$;

DROP INDEX IF EXISTS idx_cafe_rooms_plan;
CREATE UNIQUE INDEX IF NOT EXISTS idx_cafe_rooms_plan_active_unique
    ON cafe_rooms(plan_id)
    WHERE deleted_at IS NULL;

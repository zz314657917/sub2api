package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GroupBuyPlan holds the schema definition for group-buying plan templates.
type GroupBuyPlan struct {
	ent.Schema
}

func (GroupBuyPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_plans"},
	}
}

func (GroupBuyPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (GroupBuyPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			MaxLen(120).
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("product_key").
			MaxLen(64).
			Default("token_pinpinpin"),
		field.Int("total_shares").
			Default(10).
			Positive(),
		// subscription_tier is deliberately constrained by the Cafe service and
		// migration. Generic group-buy plans keep their existing product_key
		// semantics and are not interpreted as Cafe subscriptions.
		field.String("subscription_tier").
			MaxLen(16).
			Default("plus"),
		field.Int("max_buyers").
			Default(4).
			Positive(),
		field.Int("seat_count").
			Comment("Legacy alias of total_shares for v1 API compatibility").
			Default(10).
			Positive(),
		field.Float("price_per_share").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("price_per_seat").
			Comment("Legacy alias of price_per_share for v1 API compatibility").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("price_label").
			MaxLen(120).
			Default(""),
		field.String("quota_per_share_label").
			MaxLen(255).
			Default(""),
		field.String("quota_label").
			Comment("Legacy alias of quota_per_share_label for v1 API compatibility").
			MaxLen(255).
			Default(""),
		field.Int("max_shares_per_user").
			Default(10),
		field.Int("fulfillment_timeout_minutes").
			Default(1440).
			Positive(),
		field.Int64("target_group_id"),
		field.JSON("tier_group_ids", map[string]int64{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.JSON("tier_rules", []domain.GroupBuyTierRule{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Int("validity_days").
			Default(30),
		field.Int("timeout_minutes").
			Default(1440),
		field.String("launch_mode").
			MaxLen(16).
			Default("auto"),
		field.String("fulfillment_mode").
			MaxLen(32).
			Default("aggregate_tier").
			Comment("Fulfillment mode: aggregate_tier or room_subscription."),
		field.Float("room_key_quota_usd").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("room_key_rate_limit_5h").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("room_key_rate_limit_1d").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Float("room_key_rate_limit_7d").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Bool("auto_create_room_key").
			Default(true),
		field.String("refund_mode").
			MaxLen(32).
			Default("balance_credit"),
		field.String("agreement_text").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.Int("sort_order").
			Default(0),
		field.Time("last_round_created_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("deleted_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (GroupBuyPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("target_group", Group.Type).
			Ref("group_buy_plans").
			Field("target_group_id").
			Unique().
			Required(),
		edge.To("rounds", GroupBuyRound.Type),
		edge.To("seats", GroupBuySeat.Type),
		edge.To("events", GroupBuyEvent.Type),
		edge.To("cafe_rooms", CafeRoom.Type),
	}
}

func (GroupBuyPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_key"),
		index.Fields("status"),
		index.Fields("sort_order"),
		index.Fields("target_group_id"),
		index.Fields("fulfillment_mode"),
		index.Fields("deleted_at"),
	}
}

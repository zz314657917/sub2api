package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GroupBuySeat holds the schema definition for a user's seat in a round.
type GroupBuySeat struct {
	ent.Schema
}

func (GroupBuySeat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_seats"},
	}
}

func (GroupBuySeat) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("round_id"),
		field.Int64("plan_id"),
		field.Int64("user_id"),
		field.Int64("order_id").
			Optional().
			Nillable(),
		// A Cafe membership aggregates several seat/order batches. It stays
		// nullable so every pre-S252 generic and Cafe row remains readable.
		field.Int64("membership_id").
			Optional().
			Nillable(),
		field.String("status").
			MaxLen(24).
			Default("locked"),
		field.Int("share_count").
			Default(1),
		field.Int("seat_no").
			Optional().
			Nillable(),
		field.JSON("policy_snapshot", domain.GroupBuyPolicySnapshot{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Int64("subscription_id").
			Optional().
			Nillable(),
		field.Int64("bound_api_key_id").
			Optional().
			Nillable(),
		field.Time("locked_until").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("activated_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("bound_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("refund_processed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("refund_note").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (GroupBuySeat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("round", GroupBuyRound.Type).
			Ref("seats").
			Field("round_id").
			Unique().
			Required(),
		edge.From("plan", GroupBuyPlan.Type).
			Ref("seats").
			Field("plan_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("group_buy_seats").
			Field("user_id").
			Unique().
			Required(),
		edge.From("order", PaymentOrder.Type).
			Ref("group_buy_seat").
			Field("order_id").
			Unique(),
		edge.From("membership", CafeRoundMembership.Type).
			Ref("payment_batches").
			Field("membership_id").
			Unique(),
		edge.From("subscription", UserSubscription.Type).
			Ref("group_buy_seats").
			Field("subscription_id").
			Unique(),
		edge.From("bound_api_key", APIKey.Type).
			Ref("group_buy_seats").
			Field("bound_api_key_id").
			Unique(),
		edge.To("refunds", GroupBuyRefund.Type),
		edge.To("events", GroupBuyEvent.Type),
		edge.To("account_bindings", APIKeyAccountBinding.Type),
	}
}

func (GroupBuySeat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("round_id"),
		index.Fields("plan_id"),
		index.Fields("user_id"),
		index.Fields("order_id").
			Unique().
			Annotations(entsql.IndexWhere("order_id IS NOT NULL")),
		index.Fields("status"),
		index.Fields("membership_id"),
		index.Fields("locked_until"),
		index.Fields("expires_at"),
		index.Fields("round_id", "user_id"),
		index.Fields("round_id", "seat_no").
			Unique().
			Annotations(entsql.IndexWhere("seat_no IS NOT NULL AND status IN ('locked', 'paid', 'active', 'refund_pending', 'refund_processing')")),
	}
}

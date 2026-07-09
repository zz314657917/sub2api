package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// GroupBuyRefund records idempotent refund handling for a group-buy share batch.
type GroupBuyRefund struct {
	ent.Schema
}

func (GroupBuyRefund) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_refunds"},
	}
}

func (GroupBuyRefund) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("seat_id"),
		field.Int64("order_id").
			Optional().
			Nillable(),
		field.Int64("user_id"),
		field.String("mode").
			MaxLen(32),
		field.String("status").
			MaxLen(24).
			Default("processing"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("idempotency_key").
			MaxLen(120),
		field.String("note").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("processed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
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

func (GroupBuyRefund) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("seat", GroupBuySeat.Type).
			Ref("refunds").
			Field("seat_id").
			Unique().
			Required(),
		edge.From("order", PaymentOrder.Type).
			Ref("group_buy_refunds").
			Field("order_id").
			Unique(),
		edge.From("user", User.Type).
			Ref("group_buy_refunds").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (GroupBuyRefund) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("seat_id").
			Unique(),
		index.Fields("idempotency_key").
			Unique(),
		index.Fields("order_id"),
		index.Fields("status"),
		index.Fields("user_id"),
	}
}

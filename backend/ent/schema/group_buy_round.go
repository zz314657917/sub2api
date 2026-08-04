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

// GroupBuyRound holds the schema definition for a concrete group-buying round.
type GroupBuyRound struct {
	ent.Schema
}

func (GroupBuyRound) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_rounds"},
	}
}

func (GroupBuyRound) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("plan_id"),
		field.Int64("cafe_room_id").
			Optional().
			Nillable(),
		field.Int64("assigned_account_id").
			Optional().
			Nillable(),
		field.String("room_code_snapshot").
			Optional().
			Nillable(),
		field.String("room_name_snapshot").
			Optional().
			Nillable(),
		field.String("status").
			MaxLen(20).
			Default("open"),
		field.Int("total_shares").
			Positive(),
		field.Int("paid_shares").
			Default(0),
		field.Int("reserved_shares").
			Default(0),
		field.Int("total_seats").
			Comment("Legacy alias of total_shares for v1 API compatibility").
			Positive(),
		field.Int("paid_seats").
			Comment("Legacy alias of paid_shares for v1 API compatibility").
			Default(0),
		field.Int("reserved_seats").
			Comment("Legacy alias of reserved_shares for v1 API compatibility").
			Default(0),
		field.Time("deadline_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("started_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("closed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("activated_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("entitlement_expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("activation_token").
			Optional().
			Nillable(),
		field.String("close_reason").
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

func (GroupBuyRound) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("plan", GroupBuyPlan.Type).
			Ref("rounds").
			Field("plan_id").
			Unique().
			Required(),
		edge.To("seats", GroupBuySeat.Type),
		edge.To("events", GroupBuyEvent.Type),
		edge.From("cafe_room", CafeRoom.Type).
			Ref("rounds").
			Field("cafe_room_id").
			Unique(),
		edge.From("assigned_account", Account.Type).
			Ref("cafe_rounds").
			Field("assigned_account_id").
			Unique(),
		edge.To("account_bindings", APIKeyAccountBinding.Type),
	}
}

func (GroupBuyRound) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("plan_id"),
		index.Fields("cafe_room_id"),
		index.Fields("assigned_account_id"),
		index.Fields("activation_token").
			Unique().
			Annotations(entsql.IndexWhere("activation_token IS NOT NULL")),
		index.Fields("status"),
		index.Fields("deadline_at"),
		index.Fields("created_at"),
	}
}

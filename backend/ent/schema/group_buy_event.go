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

// GroupBuyEvent records user-visible activity and admin audit events.
type GroupBuyEvent struct {
	ent.Schema
}

func (GroupBuyEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_events"},
	}
}

func (GroupBuyEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("plan_id").
			Optional().
			Nillable(),
		field.Int64("round_id").
			Optional().
			Nillable(),
		field.Int64("seat_id").
			Optional().
			Nillable(),
		field.Int64("user_id").
			Optional().
			Nillable(),
		field.String("event_type").
			MaxLen(48),
		field.String("message").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.JSON("metadata", map[string]any{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (GroupBuyEvent) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("plan", GroupBuyPlan.Type).
			Ref("events").
			Field("plan_id").
			Unique(),
		edge.From("round", GroupBuyRound.Type).
			Ref("events").
			Field("round_id").
			Unique(),
		edge.From("seat", GroupBuySeat.Type).
			Ref("events").
			Field("seat_id").
			Unique(),
		edge.From("user", User.Type).
			Ref("group_buy_events").
			Field("user_id").
			Unique(),
	}
}

func (GroupBuyEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("plan_id"),
		index.Fields("round_id"),
		index.Fields("seat_id"),
		index.Fields("user_id"),
		index.Fields("event_type"),
		index.Fields("created_at"),
	}
}

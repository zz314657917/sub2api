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

// CafeRoundMembership is the durable per-user allocation for a share based
// Cafe round. Payment batches remain GroupBuySeat rows and point here.
type CafeRoundMembership struct{ ent.Schema }

func (CafeRoundMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "cafe_round_memberships"}}
}

func (CafeRoundMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("round_id"),
		field.Int64("user_id"),
		field.String("status").MaxLen(24).Default("locked"),
		field.Int("paid_shares").Default(0),
		field.Int("reserved_shares").Default(0),
		field.Int64("bound_api_key_id").Optional().Nillable(),
		field.Time("activated_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").Immutable().Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (CafeRoundMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("round", GroupBuyRound.Type).
			Ref("cafe_memberships").
			Field("round_id").
			Unique().
			Required(),
		edge.To("payment_batches", GroupBuySeat.Type),
		edge.To("account_bindings", APIKeyAccountBinding.Type),
	}
}

func (CafeRoundMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("round_id", "user_id").Unique(),
		index.Fields("round_id", "status"),
		index.Fields("bound_api_key_id").Unique().Annotations(entsql.IndexWhere("bound_api_key_id IS NOT NULL")),
		index.Fields("expires_at"),
	}
}

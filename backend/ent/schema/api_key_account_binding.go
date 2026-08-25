package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// APIKeyAccountBinding pins a managed API key to one account for one room seat.
type APIKeyAccountBinding struct {
	ent.Schema
}

func (APIKeyAccountBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_key_account_bindings"},
	}
}

func (APIKeyAccountBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.TimeMixin{}}
}

func (APIKeyAccountBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("api_key_id"),
		field.Int64("user_id"),
		field.Int64("group_id"),
		field.Int64("account_id"),
		field.Int64("cafe_room_id"),
		field.Int64("round_id"),
		field.Int64("seat_id").
			Optional().
			Nillable(),
		field.Int64("membership_id").
			Optional().
			Nillable(),
		field.String("status").
			MaxLen(20).
			Default("active"),
		field.Bool("strict_mode").
			Default(true),
		field.Time("starts_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int64("replaced_by_binding_id").
			Optional().
			Nillable(),
		field.Time("migrated_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (APIKeyAccountBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api_key", APIKey.Type).
			Ref("account_bindings").
			Field("api_key_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("account_bindings").
			Field("user_id").
			Unique().
			Required(),
		edge.From("group", Group.Type).
			Ref("account_bindings").
			Field("group_id").
			Unique().
			Required(),
		edge.From("account", Account.Type).
			Ref("account_bindings").
			Field("account_id").
			Unique().
			Required(),
		edge.From("cafe_room", CafeRoom.Type).
			Ref("account_bindings").
			Field("cafe_room_id").
			Unique().
			Required(),
		edge.From("round", GroupBuyRound.Type).
			Ref("account_bindings").
			Field("round_id").
			Unique().
			Required(),
		edge.From("seat", GroupBuySeat.Type).
			Ref("account_bindings").
			Field("seat_id").
			Unique(),
		edge.From("membership", CafeRoundMembership.Type).
			Ref("account_bindings").
			Field("membership_id").
			Unique(),
	}
}

func (APIKeyAccountBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api_key_id").
			Unique().
			StorageKey("idx_api_key_account_bindings_active_key").
			Annotations(entsql.IndexWhere("status = 'active'")),
		index.Fields("seat_id").
			Unique().
			StorageKey("idx_api_key_account_bindings_active_seat").
			Annotations(entsql.IndexWhere("status = 'active' AND seat_id IS NOT NULL")),
		index.Fields("membership_id").
			Unique().
			StorageKey("idx_api_key_account_bindings_active_membership").
			Annotations(entsql.IndexWhere("status = 'active' AND membership_id IS NOT NULL")),
		index.Fields("user_id", "group_id", "status"),
		index.Fields("account_id", "status"),
		index.Fields("cafe_room_id", "status"),
		index.Fields("round_id"),
		index.Fields("expires_at"),
		index.Fields("replaced_by_binding_id"),
	}
}

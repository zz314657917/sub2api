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

// GroupBuyEntitlement stores the per-user TokenPinPinPin aggregate entitlement.
type GroupBuyEntitlement struct {
	ent.Schema
}

func (GroupBuyEntitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "group_buy_entitlements"},
	}
}

func (GroupBuyEntitlement) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("product_key").
			MaxLen(64).
			Default("token_pinpinpin"),
		field.String("status").
			MaxLen(20).
			Default("inactive"),
		field.Int("active_share_count").
			Default(0),
		field.Int64("target_group_id").
			Optional().
			Nillable(),
		field.Int64("subscription_id").
			Optional().
			Nillable(),
		field.Int64("bound_api_key_id").
			Optional().
			Nillable(),
		field.Time("last_activated_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("refreshed_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("deactivated_at").
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

func (GroupBuyEntitlement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("group_buy_entitlements").
			Field("user_id").
			Unique().
			Required(),
		edge.From("target_group", Group.Type).
			Ref("group_buy_entitlements").
			Field("target_group_id").
			Unique(),
		edge.From("subscription", UserSubscription.Type).
			Ref("group_buy_entitlements").
			Field("subscription_id").
			Unique(),
		edge.From("bound_api_key", APIKey.Type).
			Ref("group_buy_entitlements").
			Field("bound_api_key_id").
			Unique(),
	}
}

func (GroupBuyEntitlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "product_key").Unique(),
		index.Fields("status"),
		index.Fields("target_group_id"),
		index.Fields("subscription_id"),
		index.Fields("bound_api_key_id"),
		index.Fields("expires_at"),
	}
}

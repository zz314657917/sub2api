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

// CafeRoom is the long-lived operational and visual shell for a Pixel Cafe room.
type CafeRoom struct {
	ent.Schema
}

func (CafeRoom) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cafe_rooms"},
	}
}

func (CafeRoom) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (CafeRoom) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			MaxLen(64).
			NotEmpty(),
		field.String("name").
			MaxLen(120).
			NotEmpty(),
		field.Int64("plan_id"),
		field.Int64("account_id").
			Optional().
			Nillable(),
		field.String("zone_key").
			MaxLen(32).
			Default("featured"),
		field.String("theme_key").
			MaxLen(64).
			Default("warm_wood"),
		field.String("scene_slot_key").
			MaxLen(120).
			Default(""),
		field.String("status").
			MaxLen(20).
			Default("draft"),
		field.Bool("featured").
			Default(false),
		field.Int("sort_order").
			Default(0),
		field.JSON("metadata", map[string]any{}).
			Default(func() map[string]any { return map[string]any{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
	}
}

func (CafeRoom) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("plan", GroupBuyPlan.Type).
			Ref("cafe_rooms").
			Field("plan_id").
			Unique().
			Required(),
		edge.From("account", Account.Type).
			Ref("cafe_rooms").
			Field("account_id").
			Unique(),
		edge.To("rounds", GroupBuyRound.Type),
		edge.To("account_bindings", APIKeyAccountBinding.Type),
	}
}

func (CafeRoom) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").
			Unique().
			StorageKey("idx_cafe_rooms_code_active").
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("zone_key", "status", "sort_order"),
		index.Fields("plan_id"),
		index.Fields("account_id"),
		index.Fields("status"),
		index.Fields("featured", "sort_order"),
		index.Fields("deleted_at"),
	}
}

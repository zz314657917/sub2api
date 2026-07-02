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

// InvoiceRequest holds invoice applications submitted by users.
type InvoiceRequest struct {
	ent.Schema
}

func (InvoiceRequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invoice_requests"},
	}
}

func (InvoiceRequest) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.String("currency").
			MaxLen(10).
			Default("CNY"),
		field.String("invoice_type").
			MaxLen(30),
		field.String("title").
			MaxLen(255),
		field.String("tax_number").
			MaxLen(64),
		field.String("remark").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("status").
			MaxLen(30).
			Default("pending"),
		field.String("admin_note").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("invoice_no").
			Optional().
			Nillable().
			MaxLen(128),
		field.String("file_name").
			Optional().
			Nillable().
			MaxLen(255),
		field.String("file_path").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int64("file_size").
			Optional().
			Nillable(),
		field.String("file_content_type").
			Optional().
			Nillable().
			MaxLen(128),
		field.Int64("reviewed_by").
			Optional().
			Nillable(),
		field.Time("reviewed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("issued_at").
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

func (InvoiceRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("invoice_requests").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (InvoiceRequest) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("currency"),
	}
}

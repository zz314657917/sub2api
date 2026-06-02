package schema

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SupportTicket struct {
	ent.Schema
}

func (SupportTicket) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "support_tickets"},
	}
}

func (SupportTicket) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("title").
			MaxLen(200).
			NotEmpty(),
		field.String("status").
			MaxLen(20).
			Default("open").
			Validate(validateSupportTicketStatus),
		field.String("ticket_type").
			MaxLen(20).
			Default("support").
			Validate(validateSupportTicketType),
		field.String("system_key").
			MaxLen(80).
			Optional().
			Nillable(),
		field.String("last_message_preview").
			MaxLen(240).
			Default(""),
		field.Time("last_message_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Int("user_unread_count").
			NonNegative().
			Default(0),
		field.Int("admin_unread_count").
			NonNegative().
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("closed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("support_tickets").
			Field("user_id").
			Required().
			Unique(),
		edge.To("messages", SupportTicketMessage.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (SupportTicket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "last_message_at", "id"),
		index.Fields("status", "last_message_at", "id"),
		index.Fields("ticket_type", "user_id", "last_message_at", "id"),
	}
}

func validateSupportTicketStatus(value string) error {
	switch value {
	case "open", "pending_admin", "pending_user", "closed":
		return nil
	default:
		return fmt.Errorf("must be one of open, pending_admin, pending_user, closed")
	}
}

func validateSupportTicketType(value string) error {
	switch value {
	case "support", "system":
		return nil
	default:
		return fmt.Errorf("must be one of support, system")
	}
}

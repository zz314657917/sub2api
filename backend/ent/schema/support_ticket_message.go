package schema

import (
	"encoding/json"
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

type SupportTicketMessage struct {
	ent.Schema
}

func (SupportTicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "support_ticket_messages"},
	}
}

func (SupportTicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id"),
		field.String("sender_type").
			MaxLen(20).
			Validate(validateSupportTicketSenderType),
		field.Int64("sender_user_id").
			Optional().
			Nillable(),
		field.String("content").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty(),
		field.String("event_type").
			MaxLen(80).
			Optional().
			Nillable(),
		field.String("event_key").
			MaxLen(160).
			Optional().
			Nillable(),
		field.JSON("metadata", json.RawMessage{}).
			Default(json.RawMessage("{}")),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (SupportTicketMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ticket", SupportTicket.Type).
			Ref("messages").
			Field("ticket_id").
			Required().
			Unique(),
	}
}

func (SupportTicketMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ticket_id", "created_at", "id"),
		index.Fields("ticket_id", "event_key"),
		index.Fields("sender_user_id"),
	}
}

func validateSupportTicketSenderType(value string) error {
	switch value {
	case "user", "admin", "system":
		return nil
	default:
		return fmt.Errorf("must be one of user, admin, system")
	}
}

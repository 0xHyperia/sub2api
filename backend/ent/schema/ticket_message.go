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

type TicketMessage struct{ ent.Schema }

func (TicketMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ticket_messages"}}
}

func (TicketMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("ticket_id"),
		field.Int64("sender_user_id").Optional().Nillable(),
		field.String("sender_type").MaxLen(20).Immutable(),
		field.String("event_type").MaxLen(32).Default("").Immutable(),
		field.String("content").SchemaType(map[string]string{dialect.Postgres: "text"}).Default("").Immutable(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (TicketMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ticket", Ticket.Type).Ref("messages").Field("ticket_id").Unique().Required(),
		edge.From("sender", User.Type).Ref("ticket_messages").Field("sender_user_id").Unique(),
		edge.To("attachments", TicketAttachment.Type),
	}
}

func (TicketMessage) Indexes() []ent.Index {
	return []ent.Index{index.Fields("ticket_id", "created_at")}
}

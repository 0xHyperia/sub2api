package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TicketAttachment struct{ ent.Schema }

func (TicketAttachment) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ticket_attachments"}}
}

func (TicketAttachment) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("message_id"),
		field.String("object_key").MaxLen(1024).Unique().Immutable(),
		field.String("original_name").MaxLen(255).Immutable(),
		field.String("content_type").MaxLen(100).Immutable(),
		field.Int64("size_bytes").Immutable(),
		field.String("sha256").MaxLen(64).Immutable(),
		field.Time("created_at").Immutable().Default(time.Now),
	}
}

func (TicketAttachment) Edges() []ent.Edge {
	return []ent.Edge{edge.From("message", TicketMessage.Type).Ref("attachments").Field("message_id").Unique().Required()}
}

func (TicketAttachment) Indexes() []ent.Index {
	return []ent.Index{index.Fields("message_id")}
}

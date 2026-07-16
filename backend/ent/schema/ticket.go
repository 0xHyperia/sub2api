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

type Ticket struct{ ent.Schema }

func (Ticket) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "tickets"}}
}

func (Ticket) Fields() []ent.Field {
	return []ent.Field{
		field.String("number").MaxLen(32).Unique().Immutable(),
		field.Int64("user_id"),
		field.Int64("category_id"),
		field.String("subject").MaxLen(200),
		field.String("status").MaxLen(20).Default("open"),
		field.Int("user_unread_count").Default(0),
		field.Int("admin_unread_count").Default(0),
		field.String("last_actor_type").MaxLen(20).Default("user"),
		field.Time("last_message_at").Default(time.Now),
		field.Time("closed_at").Optional().Nillable(),
		field.Int64("closed_by_user_id").Optional().Nillable(),
		field.String("closed_by_role").Optional().Nillable().MaxLen(20),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Ticket) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("tickets").Field("user_id").Unique().Required(),
		edge.From("category", TicketCategory.Type).Ref("tickets").Field("category_id").Unique().Required(),
		edge.To("messages", TicketMessage.Type),
	}
}

func (Ticket) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "last_message_at"),
		index.Fields("status", "last_message_at"),
		index.Fields("category_id", "last_message_at"),
		index.Fields("admin_unread_count", "last_message_at"),
	}
}

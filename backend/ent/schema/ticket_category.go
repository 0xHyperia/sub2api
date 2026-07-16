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

type TicketCategory struct{ ent.Schema }

func (TicketCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ticket_categories"}}
}

func (TicketCategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").MaxLen(64).Unique().Immutable(),
		field.String("name_zh").MaxLen(100),
		field.String("name_en").MaxLen(100).Default(""),
		field.Bool("active").Default(true),
		field.Int("sort_order").Default(0),
		field.Time("created_at").Immutable().Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (TicketCategory) Edges() []ent.Edge {
	return []ent.Edge{edge.To("tickets", Ticket.Type)}
}

func (TicketCategory) Indexes() []ent.Index {
	return []ent.Index{index.Fields("active", "sort_order")}
}

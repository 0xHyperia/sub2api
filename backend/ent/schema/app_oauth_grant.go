package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AppOAuthGrant represents one user's consent for one OAuth client.
type AppOAuthGrant struct {
	ent.Schema
}

func (AppOAuthGrant) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "app_oauth_grants"}}
}

func (AppOAuthGrant) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.TimeMixin{}}
}

func (AppOAuthGrant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("grant_id").MaxLen(64).NotEmpty().Unique(),
		field.String("client_id").MaxLen(64).NotEmpty(),
		field.JSON("scopes", []string{}).Default(func() []string { return []string{} }).SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Enum("status").Values("active", "revoked").Default("active"),
		field.Int("grant_version").Positive().Default(1),
		field.Time("first_authorized_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("last_authorized_at").Default(time.Now).SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("revoked_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AppOAuthGrant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "client_id").Unique(),
		index.Fields("user_id", "status"),
		index.Fields("client_id"),
	}
}

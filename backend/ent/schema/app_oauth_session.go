package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AppOAuthSession represents one installation's refresh-token family.
type AppOAuthSession struct {
	ent.Schema
}

func (AppOAuthSession) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "app_oauth_sessions"}}
}

func (AppOAuthSession) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.TimeMixin{}}
}

func (AppOAuthSession) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("app_grant_id"),
		field.String("session_id").MaxLen(64).NotEmpty().Unique(),
		field.String("installation_id_hash").MaxLen(64).NotEmpty(),
		field.String("token_family_id").MaxLen(64).NotEmpty().Unique(),
		field.String("device_name").MaxLen(200).Default(""),
		field.String("platform").MaxLen(40).Default(""),
		field.JSON("scopes", []string{}).Default(func() []string { return []string{} }).SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.Enum("status").Values("active", "revoked").Default("active"),
		field.Time("last_used_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("revoked_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AppOAuthSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("app_grant_id", "installation_id_hash").Unique(),
		index.Fields("app_grant_id", "status"),
	}
}

package schema

import (
	"fmt"

	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

var appAuthorizationStatuses = map[string]struct{}{
	"active":  {},
	"revoked": {},
}

func validateAppAuthorizationStatus(value string) error {
	if _, ok := appAuthorizationStatuses[value]; ok {
		return nil
	}
	return fmt.Errorf("invalid app authorization status %q", value)
}

// AppAuthorization represents one user-approved ZeroAgent OAuth grant.
type AppAuthorization struct {
	ent.Schema
}

func (AppAuthorization) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "app_authorizations"}}
}

func (AppAuthorization) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.TimeMixin{}}
}

func (AppAuthorization) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.String("grant_id").MaxLen(64).NotEmpty().Unique(),
		field.String("client_id").MaxLen(64).NotEmpty(),
		field.String("device_name").MaxLen(200).Default(""),
		field.String("platform").MaxLen(40).Default(""),
		field.JSON("scopes", []string{}).
			Default(func() []string { return []string{} }).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}),
		field.String("token_family_id").MaxLen(64).NotEmpty().Unique(),
		field.String("status").MaxLen(20).Default("active").Validate(validateAppAuthorizationStatus),
		field.Time("last_used_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("revoked_at").Optional().Nillable().SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AppAuthorization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "status"),
		index.Fields("client_id"),
		index.Fields("token_family_id"),
	}
}

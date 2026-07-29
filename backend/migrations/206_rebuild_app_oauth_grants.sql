-- OAuth v2 intentionally invalidates the unused legacy authorization model.
-- Application consent is unique per user/client; device token families live in
-- separate installation sessions.
DROP TABLE IF EXISTS app_authorizations;

CREATE TABLE app_oauth_grants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grant_id VARCHAR(64) NOT NULL UNIQUE,
    client_id VARCHAR(64) NOT NULL,
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    grant_version INTEGER NOT NULL DEFAULT 1,
    first_authorized_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_authorized_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT app_oauth_grants_user_client_key UNIQUE (user_id, client_id),
    CONSTRAINT app_oauth_grants_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT app_oauth_grants_version_check CHECK (grant_version > 0),
    CONSTRAINT app_oauth_grants_scopes_array_check CHECK (jsonb_typeof(scopes) = 'array')
);

CREATE INDEX app_oauth_grants_user_status_idx
    ON app_oauth_grants (user_id, status);

CREATE INDEX app_oauth_grants_client_id_idx
    ON app_oauth_grants (client_id);

CREATE TABLE app_oauth_sessions (
    id BIGSERIAL PRIMARY KEY,
    app_grant_id BIGINT NOT NULL REFERENCES app_oauth_grants(id) ON DELETE CASCADE,
    session_id VARCHAR(64) NOT NULL UNIQUE,
    installation_id_hash VARCHAR(64) NOT NULL,
    token_family_id VARCHAR(64) NOT NULL UNIQUE,
    device_name VARCHAR(200) NOT NULL DEFAULT '',
    platform VARCHAR(40) NOT NULL DEFAULT '',
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT app_oauth_sessions_grant_installation_key UNIQUE (app_grant_id, installation_id_hash),
    CONSTRAINT app_oauth_sessions_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT app_oauth_sessions_scopes_array_check CHECK (jsonb_typeof(scopes) = 'array')
);

CREATE INDEX app_oauth_sessions_grant_status_idx
    ON app_oauth_sessions (app_grant_id, status);

CREATE INDEX app_oauth_sessions_last_used_idx
    ON app_oauth_sessions (last_used_at)
    WHERE status = 'active';

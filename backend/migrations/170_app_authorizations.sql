CREATE TABLE IF NOT EXISTS app_authorizations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grant_id VARCHAR(64) NOT NULL,
    client_id VARCHAR(64) NOT NULL,
    device_name VARCHAR(200) NOT NULL DEFAULT '',
    platform VARCHAR(40) NOT NULL DEFAULT '',
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    token_family_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT app_authorizations_status_check CHECK (status IN ('active', 'revoked')),
    CONSTRAINT app_authorizations_grant_id_key UNIQUE (grant_id),
    CONSTRAINT app_authorizations_token_family_id_key UNIQUE (token_family_id)
);

CREATE INDEX IF NOT EXISTS app_authorizations_user_status_idx
    ON app_authorizations (user_id, status);

CREATE INDEX IF NOT EXISTS app_authorizations_client_id_idx
    ON app_authorizations (client_id);

CREATE INDEX IF NOT EXISTS app_authorizations_token_family_id_idx
    ON app_authorizations (token_family_id);

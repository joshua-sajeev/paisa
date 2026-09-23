-- +goose Up

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    icon_key TEXT NOT NULL DEFAULT 'bank',
    balance BIGINT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_accounts_active_name
ON accounts (name)
WHERE is_archived = FALSE;

CREATE UNIQUE INDEX uq_accounts_primary
ON accounts (is_primary)
WHERE is_primary = TRUE;

-- +goose Down

DROP INDEX uq_accounts_active_name;
DROP TABLE accounts;

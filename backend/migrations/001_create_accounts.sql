-- +goose Up

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_accounts_active_name
ON accounts (name)
WHERE is_archived = FALSE;

-- +goose Down

DROP INDEX uq_accounts_active_name;
DROP TABLE accounts;

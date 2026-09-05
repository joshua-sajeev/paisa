-- +goose Up

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    from_account_id UUID,
    to_account_id UUID,
    jar_id UUID,
    amount BIGINT NOT NULL,
    occurred_at DATE NOT NULL,
    is_master_income BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_transactions_from_account
        FOREIGN KEY (from_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_transactions_to_account
        FOREIGN KEY (to_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_transactions_jar
        FOREIGN KEY (jar_id)
        REFERENCES jars(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_transactions_type
        CHECK (type IN ('income', 'expense', 'transfer')),

    CONSTRAINT chk_transactions_amount
        CHECK (amount > 0),

    CONSTRAINT chk_transactions_income_accounts
        CHECK (
            type <> 'income'
            OR (
                from_account_id IS NULL
                AND to_account_id IS NOT NULL
            )
        ),

    CONSTRAINT chk_transactions_expense_accounts
        CHECK (
            type <> 'expense'
            OR (
                from_account_id IS NOT NULL
                AND to_account_id IS NULL
            )
        ),

    CONSTRAINT chk_transactions_transfer_accounts
        CHECK (
            type <> 'transfer'
            OR (
                from_account_id IS NOT NULL
                AND to_account_id IS NOT NULL
                AND from_account_id <> to_account_id
            )
        )
);

CREATE TABLE templates (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    from_account_id UUID,
    to_account_id UUID,
    jar_id UUID,
    amount BIGINT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_templates_from_account
        FOREIGN KEY (from_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_templates_to_account
        FOREIGN KEY (to_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_templates_jar
        FOREIGN KEY (jar_id)
        REFERENCES jars(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_templates_type
        CHECK (type IN ('income', 'expense', 'transfer')),

    CONSTRAINT chk_templates_amount
        CHECK (amount IS NULL OR amount > 0)
);

CREATE UNIQUE INDEX uq_templates_active_name
ON templates (name)
WHERE is_archived = FALSE;

CREATE TABLE jar_allocations (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    jar_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_jar_allocations_transaction
        FOREIGN KEY (transaction_id)
        REFERENCES transactions(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_jar_allocations_jar
        FOREIGN KEY (jar_id)
        REFERENCES jars(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_jar_allocations_amount
        CHECK (amount > 0),

    CONSTRAINT uq_jar_allocations_transaction_jar
        UNIQUE (transaction_id, jar_id)
);



CREATE INDEX idx_transactions_from_account
    ON transactions (from_account_id);

CREATE INDEX idx_transactions_to_account
    ON transactions (to_account_id);

CREATE INDEX idx_transactions_jar
    ON transactions (jar_id);

CREATE INDEX idx_transactions_occurred_at
    ON transactions (occurred_at);

CREATE INDEX idx_jar_allocations_jar
    ON jar_allocations (jar_id);

-- +goose Down


DROP INDEX idx_jar_allocations_jar;

DROP INDEX idx_transactions_occurred_at;
DROP INDEX idx_transactions_jar;
DROP INDEX idx_transactions_to_account;
DROP INDEX idx_transactions_from_account;

DROP TABLE jar_allocations;
DROP INDEX uq_templates_active_name;
DROP TABLE templates;
DROP TABLE transactions;

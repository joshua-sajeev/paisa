-- +goose Up

CREATE TABLE goals (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    target BIGINT NOT NULL,
    deadline DATE NOT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_goals_target
        CHECK (target > 0)
);

CREATE UNIQUE INDEX uq_goals_active_name
ON goals (name)
WHERE is_archived = FALSE;

CREATE TABLE contributions (
    id UUID PRIMARY KEY,
    goal_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    occurred_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_contributions_goal
        FOREIGN KEY (goal_id)
        REFERENCES goals(id),

    CONSTRAINT chk_contributions_amount
        CHECK (amount > 0)
);

-- +goose Down

DROP TABLE contributions;
DROP INDEX uq_goals_active_name;
DROP TABLE goals;

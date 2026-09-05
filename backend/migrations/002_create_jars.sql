-- +goose Up

CREATE TABLE jars (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    allocation_type TEXT NOT NULL,
    allocation_value BIGINT NOT NULL DEFAULT 0,

    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_jars_allocation_type
        CHECK (allocation_type IN ('percentage', 'remainder')),

    CONSTRAINT chk_jars_allocation
        CHECK (
            (
                allocation_type = 'percentage'
                AND allocation_value BETWEEN 1 AND 100
            )
            OR
            (
                allocation_type = 'remainder'
                AND allocation_value = 0
            )
        )
);

CREATE UNIQUE INDEX uq_jars_active_name
ON jars (name)
WHERE is_archived = FALSE;

CREATE UNIQUE INDEX uq_jars_single_remainder
ON jars (allocation_type)
WHERE allocation_type = 'remainder'
  AND is_archived = FALSE;

-- +goose Down

DROP INDEX uq_jars_single_remainder;
DROP INDEX uq_jars_active_name;
DROP TABLE jars;

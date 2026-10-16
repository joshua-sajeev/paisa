-- +goose Up

ALTER TABLE jars ADD COLUMN rollover BOOLEAN NOT NULL DEFAULT TRUE;
UPDATE jars SET rollover = FALSE WHERE LOWER(name) = 'leisure';

-- +goose Down

ALTER TABLE jars DROP COLUMN rollover;

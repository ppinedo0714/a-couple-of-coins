-- +goose Up
CREATE TABLE import_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','done','failed')),
    source_type   VARCHAR(20) NOT NULL DEFAULT 'csv',
    file_name     VARCHAR(255) NOT NULL,
    rows_total    INTEGER,
    rows_imported INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ
);

-- +goose Down
DROP TABLE import_jobs;

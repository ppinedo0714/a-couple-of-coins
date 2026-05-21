-- +goose Up
CREATE TABLE account_balance_snapshots (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    balance    NUMERIC(15,2) NOT NULL,
    date       DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, date)
);

-- +goose Down
DROP TABLE account_balance_snapshots;

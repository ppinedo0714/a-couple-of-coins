-- +goose Up
CREATE TABLE transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id    UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    category_id   UUID REFERENCES categories(id) ON DELETE SET NULL,
    amount        NUMERIC(15,2) NOT NULL,
    description   VARCHAR(500) NOT NULL,
    merchant_name VARCHAR(255),
    date          DATE NOT NULL,
    source        VARCHAR(20) NOT NULL CHECK (source IN ('manual','csv','bank')),
    classified    BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX transactions_user_id_idx ON transactions(user_id);
CREATE INDEX transactions_account_id_idx ON transactions(account_id);
CREATE INDEX transactions_date_idx ON transactions(date);

-- +goose Down
DROP TABLE transactions;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    salary_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY UNIQUE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expiry TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    limit_cents BIGINT NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    balance_cents BIGINT NOT NULL DEFAULT 0,
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    budget_id INT REFERENCES budgets(id) ON DELETE CASCADE,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    amount_in_cents BIGINT NOT NULL,
    type TEXT NOT NULL,
    is_pending BOOLEAN NOT NULL DEFAULT FALSE,
    is_debt BOOLEAN DEFAULT FALSE,
    tags JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS depots (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS trades (
    id SERIAL PRIMARY KEY,
    depot_id INT NOT NULL REFERENCES depots(id) ON DELETE CASCADE,
    wallet_transaction_id INT REFERENCES transactions(id) ON DELETE SET NULL,
    wkn TEXT NOT NULL,
    type TEXT NOT NULL,
    quantity DOUBLE PRECISION NOT NULL,
    total_in_cents BIGINT NOT NULL,
    fees_in_cents BIGINT NOT NULL DEFAULT 0,
    taxes_in_cents BIGINT NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transaction_templates (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day INT NOT NULL,
    budget_id INT REFERENCES budgets(id) ON DELETE SET NULL,
    wallet_id INT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    amount_in_cents BIGINT NOT NULL,
    type TEXT NOT NULL,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_budget_id ON transactions(budget_id);
CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_depots_user_id ON depots(user_id);
CREATE INDEX IF NOT EXISTS idx_trades_depot_id ON trades(depot_id);
CREATE INDEX IF NOT EXISTS idx_transaction_templates_user_id ON transaction_templates(user_id);

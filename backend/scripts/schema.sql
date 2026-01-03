CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    session_token VARCHAR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry TIMESTAMP NOT NULL,
    UNIQUE(session_token),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    name TEXT NOT NULL,
    limit_cents BIGINT NOT NULL,
    UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    name TEXT NOT NULL,
    UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    budget_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    description TEXT NOT NULL,
    amount_in_cents BIGINT NOT NULL,
    type TEXT NOT NULL,
    is_pending BOOLEAN NOT NULL,
    is_debt BOOLEAN,
    tags JSONB,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (budget_id) REFERENCES budgets(id),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_budget_id ON transactions(budget_id);
CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
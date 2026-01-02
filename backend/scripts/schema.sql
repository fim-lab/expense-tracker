CREATE TABLE users (
    id INT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    session_token VARCHAR(512) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry TIMESTAMP NOT NULL,
    UNIQUE(session_token),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE budgets (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    name TEXT NOT NULL,
    limit_cents BIGINT NOT NULL,
    UNIQUE(user_id, name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL,
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    budget_id UUID NOT NULL,
    description TEXT NOT NULL,
    amount_in_cents BIGINT NOT NULL,
    wallet TEXT NOT NULL,
    type TEXT NOT NULL,
    is_pending BOOLEAN NOT NULL,
    is_debt BOOLEAN,
    tags JSONB,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (budget_id) REFERENCES budgets(id)
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_transactions_budget_id ON transactions(budget_id);
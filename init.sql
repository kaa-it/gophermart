CREATE TABLE IF NOT EXISTS users
(
    id        SERIAL PRIMARY KEY,
    login     TEXT NOT NULL UNIQUE,
    password  TEXT NOT NULL,
    current   DECIMAL DEFAULT 0.0,
    withdrawn DECIMAL DEFAULT 0.0
);

CREATE TABLE IF NOT EXISTS sessions
(
    id            SERIAL PRIMARY KEY,
    user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token TEXT      NOT NULL,
    expired       TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    number TEXT PRIMARY KEY,
    user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT,
    accrual DECIMAL,
    uploaded_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS withdrawals
(
    number TEXT PRIMARY KEY,
    user_id       INT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    sum DECIMAL NOT NULL,
    processed_at TIMESTAMP NOT NULL
);
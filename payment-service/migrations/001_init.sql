--ACCOUNTS 
CREATE TABLE IF NOT EXISTS accounts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL UNIQUE, 
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);


--INBOX 
CREATE TABLE IF NOT EXISTS inbox_messages (
    message_id TEXT PRIMARY KEY, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


--PAYMENT TRANSACTIONS 
CREATE TABLE IF NOT EXISTS payment_transactions (
    id         BIGSERIAL PRIMARY KEY,
    order_id   TEXT NOT NULL UNIQUE, 
    user_id    BIGINT NOT NULL,
    amount     BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


--OUTBOX 
CREATE TABLE IF NOT EXISTS outbox_messages (
    id         UUID PRIMARY KEY,
    type       TEXT NOT NULL,        
    payload    JSONB NOT NULL,
    sent       BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_unsent ON outbox_messages (sent, created_at);
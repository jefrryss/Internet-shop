CREATE TABLE orders (
    id TEXT PRIMARY KEY,          
    user_id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,   
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    topic VARCHAR(100) NOT NULL,
    key VARCHAR(100),
    payload JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_outbox_processed ON outbox(processed) WHERE processed = FALSE;
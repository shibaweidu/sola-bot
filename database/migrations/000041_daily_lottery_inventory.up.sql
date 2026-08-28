CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS daily_lottery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id BIGINT NOT NULL,
    code TEXT NOT NULL,
    amount INTEGER NOT NULL,
    batch_name TEXT NOT NULL DEFAULT '',
    redeem_url TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'available',
    assigned_user BIGINT NOT NULL DEFAULT 0,
    assigned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT daily_lottery_codes_amount_positive CHECK (amount > 0),
    CONSTRAINT daily_lottery_codes_status_valid CHECK (status IN ('available', 'assigned', 'used')),
    CONSTRAINT daily_lottery_codes_https_url CHECK (redeem_url = '' OR redeem_url ~ '^https://')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_lottery_codes_chat_code
    ON daily_lottery_codes (chat_id, code);
CREATE INDEX IF NOT EXISTS idx_daily_lottery_codes_lookup
    ON daily_lottery_codes (chat_id, amount, status, created_at);

CREATE TABLE IF NOT EXISTS welcome_buttons (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    label TEXT NOT NULL,
    action_type TEXT NOT NULL,
    action_value TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT welcome_buttons_action_valid CHECK (action_type IN ('daily_lottery', 'link')),
    CONSTRAINT welcome_buttons_https_url CHECK (action_type = 'daily_lottery' OR action_value ~ '^https://')
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_welcome_buttons_chat_label
    ON welcome_buttons (chat_id, label);
CREATE INDEX IF NOT EXISTS idx_welcome_buttons_chat_order
    ON welcome_buttons (chat_id, sort_order);

CREATE TABLE IF NOT EXISTS point_center_configs (
    chat_id BIGINT PRIMARY KEY,
    invite_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    inviter_reward INTEGER NOT NULL DEFAULT 100,
    invitee_reward INTEGER NOT NULL DEFAULT 20,
    sign_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sign_reward INTEGER NOT NULL DEFAULT 1,
    exchange_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    exchange_minimum INTEGER NOT NULL DEFAULT 10,
    exchange_rate INTEGER NOT NULL DEFAULT 1,
    exchange_url TEXT NOT NULL DEFAULT '',
    exchange_instructions TEXT NOT NULL DEFAULT '',
    purchase_url TEXT NOT NULL DEFAULT '',
    purchase_text TEXT NOT NULL DEFAULT '',
    shop_url TEXT NOT NULL DEFAULT '',
    shop_text TEXT NOT NULL DEFAULT '',
    invite_text TEXT NOT NULL DEFAULT '',
    points_text TEXT NOT NULL DEFAULT '',
    sign_text TEXT NOT NULL DEFAULT '',
    exchange_text TEXT NOT NULL DEFAULT '',
    rank_text TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS referral_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    bot_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    inviter_id BIGINT NOT NULL,
    use_count INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_referral_codes_chat_inviter ON referral_codes(chat_id, inviter_id);

CREATE TABLE IF NOT EXISTS referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_id UUID NOT NULL REFERENCES referral_codes(id),
    bot_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    inviter_id BIGINT NOT NULL,
    invitee_id BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'started',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMPTZ,
    rewarded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_referral_bot_invitee UNIQUE (bot_id, invitee_id)
);
CREATE INDEX IF NOT EXISTS idx_referrals_chat_inviter ON referrals(chat_id, inviter_id);

CREATE TABLE IF NOT EXISTS daily_sign_records (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    sign_date DATE NOT NULL,
    reward INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_daily_sign_user_chat_day UNIQUE (chat_id, user_id, sign_date)
);

CREATE TABLE IF NOT EXISTS exchange_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    batch_name TEXT NOT NULL DEFAULT '',
    amount INTEGER NOT NULL DEFAULT 1,
    redeem_url TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'available',
    assigned_user BIGINT NOT NULL DEFAULT 0,
    assigned_chat BIGINT NOT NULL DEFAULT 0,
    assigned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_exchange_codes_available ON exchange_codes(status, amount, created_at);

CREATE TABLE IF NOT EXISTS exchange_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    points_spent INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    code_id UUID NOT NULL REFERENCES exchange_codes(id),
    status TEXT NOT NULL DEFAULT 'assigned',
    redeem_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_exchange_orders_chat_user ON exchange_orders(chat_id, user_id, created_at);

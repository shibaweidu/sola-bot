CREATE TABLE IF NOT EXISTS daily_lottery_configs (
    chat_id BIGINT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    daily_attempts INTEGER NOT NULL DEFAULT 3,
    cost_points INTEGER NOT NULL DEFAULT 0,
    guarantee_on_last BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT daily_lottery_attempts_fixed CHECK (daily_attempts = 3),
    CONSTRAINT daily_lottery_cost_nonnegative CHECK (cost_points >= 0)
);

CREATE TABLE IF NOT EXISTS daily_lottery_prizes (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    amount INTEGER NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT daily_lottery_prize_amount_positive CHECK (amount > 0),
    CONSTRAINT daily_lottery_prize_weight_positive CHECK (weight > 0),
    CONSTRAINT daily_lottery_prize_chat_amount_unique UNIQUE (chat_id, amount)
);
CREATE INDEX IF NOT EXISTS idx_daily_lottery_prizes_chat ON daily_lottery_prizes(chat_id, enabled);

CREATE TABLE IF NOT EXISTS daily_lottery_attempts (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    draw_date VARCHAR(10) NOT NULL,
    attempt_no INTEGER NOT NULL,
    result VARCHAR(16) NOT NULL,
    amount INTEGER NOT NULL DEFAULT 0,
    code_id UUID,
    cost_points INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT daily_lottery_attempt_no_valid CHECK (attempt_no BETWEEN 1 AND 3),
    CONSTRAINT daily_lottery_result_valid CHECK (result IN ('won', 'lost')),
    CONSTRAINT daily_lottery_attempt_unique UNIQUE (chat_id, user_id, draw_date, attempt_no)
);
CREATE INDEX IF NOT EXISTS idx_daily_lottery_attempts_user_day ON daily_lottery_attempts(chat_id, user_id, draw_date);
CREATE INDEX IF NOT EXISTS idx_daily_lottery_attempts_code ON daily_lottery_attempts(code_id);

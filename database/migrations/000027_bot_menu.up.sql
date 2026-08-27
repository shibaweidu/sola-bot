CREATE TABLE IF NOT EXISTS bot_menu_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_bot_id BIGINT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('member', 'admin')),
    label TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '',
    action_type TEXT NOT NULL CHECK (action_type IN ('builtin', 'link')),
    action_key TEXT NOT NULL DEFAULT '',
    action_value TEXT NOT NULL DEFAULT '',
    row_index INTEGER NOT NULL DEFAULT 0 CHECK (row_index >= 0),
    column_index INTEGER NOT NULL DEFAULT 0 CHECK (column_index >= 0 AND column_index < 4),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT idx_bot_menu_bot_role_label UNIQUE (telegram_bot_id, role, label)
);

CREATE INDEX IF NOT EXISTS idx_bot_menu_bot_role_order
    ON bot_menu_items (telegram_bot_id, role, row_index, column_index);

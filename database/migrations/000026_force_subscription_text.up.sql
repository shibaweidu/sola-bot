ALTER TABLE chat_admin_configs
    ADD COLUMN IF NOT EXISTS force_subscribe_message TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS force_subscribe_kick_message TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS force_subscribe_channel_labels TEXT NOT NULL DEFAULT '';

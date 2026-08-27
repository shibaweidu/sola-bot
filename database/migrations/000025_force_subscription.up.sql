ALTER TABLE chat_admin_configs
    ADD COLUMN IF NOT EXISTS force_subscribe_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS force_subscribe_channels TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS force_subscribe_action TEXT NOT NULL DEFAULT 'mute';

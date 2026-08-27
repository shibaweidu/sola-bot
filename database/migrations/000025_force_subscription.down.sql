ALTER TABLE chat_admin_configs
    DROP COLUMN IF EXISTS force_subscribe_action,
    DROP COLUMN IF EXISTS force_subscribe_channels,
    DROP COLUMN IF EXISTS force_subscribe_enabled;

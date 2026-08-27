ALTER TABLE chat_admin_configs
    DROP COLUMN IF EXISTS force_subscribe_channel_labels,
    DROP COLUMN IF EXISTS force_subscribe_kick_message,
    DROP COLUMN IF EXISTS force_subscribe_message;

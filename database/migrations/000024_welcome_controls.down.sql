ALTER TABLE chat_admin_configs
    DROP COLUMN IF EXISTS welcome_delete_seconds,
    DROP COLUMN IF EXISTS welcome_enabled;

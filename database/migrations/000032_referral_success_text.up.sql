ALTER TABLE point_center_configs
    ADD COLUMN IF NOT EXISTS invite_success_text TEXT NOT NULL DEFAULT '';

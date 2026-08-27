ALTER TABLE point_center_configs
    DROP COLUMN IF EXISTS invite_join_url;

ALTER TABLE point_center_configs
    DROP COLUMN IF EXISTS invite_join_text;

ALTER TABLE point_center_configs
    ALTER COLUMN invitee_reward SET DEFAULT 0;

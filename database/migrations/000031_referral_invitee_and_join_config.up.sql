ALTER TABLE point_center_configs
    ALTER COLUMN invitee_reward SET DEFAULT 20;

ALTER TABLE point_center_configs
    ADD COLUMN IF NOT EXISTS invite_join_url TEXT NOT NULL DEFAULT '';

ALTER TABLE point_center_configs
    ADD COLUMN IF NOT EXISTS invite_join_text TEXT NOT NULL DEFAULT '';

-- Existing installations were migrated to inviter-only rewards by 000030.
-- Restore a useful default for those rows while allowing the admin to edit it.
UPDATE point_center_configs
SET invitee_reward = inviter_reward,
    updated_at = NOW()
WHERE invitee_reward = 0;

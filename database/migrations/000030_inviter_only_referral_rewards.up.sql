ALTER TABLE point_center_configs
    ALTER COLUMN invitee_reward SET DEFAULT 0;

UPDATE point_center_configs
SET invitee_reward = 0,
    updated_at = NOW()
WHERE invitee_reward <> 0;

DELETE FROM bot_menu_items
WHERE action_type = 'builtin'
  AND action_key = 'hide_keyboard';

ALTER TABLE daily_lottery_attempts
    DROP CONSTRAINT IF EXISTS daily_lottery_attempt_no_valid;

DELETE FROM daily_lottery_attempts
WHERE attempt_no > 3;

ALTER TABLE daily_lottery_attempts
    ADD CONSTRAINT daily_lottery_attempt_no_valid CHECK (attempt_no BETWEEN 1 AND 3);

ALTER TABLE daily_lottery_configs
    DROP COLUMN IF EXISTS paid_enabled;

-- Telegram clients provide the keyboard collapse/expand control in the input field.
-- Keep legacy menu rows for compatibility, but do not show them by default.
UPDATE bot_menu_items
SET enabled = FALSE,
    updated_at = NOW()
WHERE action_type = 'builtin'
  AND action_key = 'hide_keyboard';

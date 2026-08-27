ALTER TABLE bot_menu_items
    DROP CONSTRAINT IF EXISTS bot_menu_items_link_mode_check;

ALTER TABLE bot_menu_items
    DROP COLUMN IF EXISTS link_buttons_json,
    DROP COLUMN IF EXISTS link_mode;

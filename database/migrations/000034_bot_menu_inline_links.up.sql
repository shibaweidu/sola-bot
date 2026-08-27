ALTER TABLE bot_menu_items
    ADD COLUMN IF NOT EXISTS link_mode VARCHAR(16) NOT NULL DEFAULT 'text',
    ADD COLUMN IF NOT EXISTS link_buttons_json JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE bot_menu_items
    DROP CONSTRAINT IF EXISTS bot_menu_items_link_mode_check;

ALTER TABLE bot_menu_items
    ADD CONSTRAINT bot_menu_items_link_mode_check
    CHECK (link_mode IN ('text', 'buttons'));
